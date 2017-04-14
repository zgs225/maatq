package mataq

import (
	"errors"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/go-redis/redis"
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
}

var (
	ParallelError = errors.New("parallel should gte 0")
)

type GroupOptions struct {
	Parallel int
	Addr     string
	Password string
	Try      int
}

type WorkerGroup struct {
	C       chan int
	Workers []*Worker
	options *GroupOptions
}

func (g *WorkerGroup) ServeLoop() {
	for _, worker := range g.Workers {
		go worker.Work()
	}
	g.wait()
}

func (g *WorkerGroup) AddEventHandler(name string, handler EventHandler) {
	log.Warningf("Event[%s] handled by Func[%s]", name, GetFunctionName(handler))
	for _, worker := range g.Workers {
		worker.AddEventHandler(name, handler)
	}
}

func (g *WorkerGroup) wait() {
	for i := 0; i < g.options.Parallel; i++ {
		<-g.C
	}
}

func (g *WorkerGroup) initWorkers() {
	for i := 0; i < g.options.Parallel; i++ {
		// 初始化Worker
		c := &Worker{Try: g.options.Try, C: g.C, Id: i}
		g.Workers[i] = c

		c.Client = redis.NewClient(&redis.Options{
			Addr:     g.options.Addr,
			Password: g.options.Password,
			DB:       0,
		})
		c.EventHandlers = make(map[string]EventHandler)
		c.initLog()
		c.checkConn()
	}
}

func (g *WorkerGroup) handleSignals() {
	sigC := make(chan os.Signal, 3)
	signal.Notify(sigC, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		for sig := range sigC {
			if sig == os.Interrupt || sig == syscall.SIGTERM || sig == syscall.SIGQUIT {
				log.Warn("Prepare to safe exit...")
				var delay = false

				for _, worker := range g.Workers {
					worker.Mu.Lock()
					defer worker.Mu.Unlock()

					if worker.CurrentMsg != nil {
						delay = true
					}
				}

				if delay {
					log.Warn("Mataqd will exit after 5 seconds")
					timer := time.NewTimer(time.Second * 5)
					<-timer.C
				}

				for _, worker := range g.Workers {
					worker.PushBackCurrentMsg()
					worker.Client.Close()
				}

				os.Exit(0)
			}
		}
	}()
}

// 获取监听队列的 Group
func NewWorkerGroup(opt *GroupOptions) (*WorkerGroup, error) {
	if opt.Parallel < 0 {
		return nil, ParallelError
	}

	if opt.Parallel == 0 {
		opt.Parallel = runtime.NumCPU()
	}

	ptr := &WorkerGroup{
		C:       make(chan int, opt.Parallel),
		Workers: make([]*Worker, opt.Parallel),
		options: opt,
	}

	ptr.initWorkers()
	ptr.handleSignals()

	return ptr, nil
}
