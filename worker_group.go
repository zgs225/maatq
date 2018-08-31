package maatq

import (
	"errors"
	"runtime"
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
	ErrParallel = errors.New("parallel should gte 0")
)

type GroupOptions struct {
	Parallel int
	Addr     string
	Password string
	Try      int
	Queues   []string
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
		err := worker.AddEventHandler(name, handler)
		if err != nil {
			log.Fatal(err)
		}
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
		c := &Worker{try: g.options.Try, c: g.C, Id: i}
		g.Workers[i] = c

		for _, q := range g.options.Queues {
			c.queues = append(c.queues, queueName(q))
		}

		c.client = redis.NewClient(&redis.Options{
			Addr:     g.options.Addr,
			Password: g.options.Password,
			DB:       0,
		})
		c.eventHandlers = make(map[string]EventHandler)
		c.initLog()
		c.checkConn()
	}
}

func (g *WorkerGroup) cleanup() {
	var delay = false

	for _, worker := range g.Workers {
		worker.mu.Lock()
		defer worker.mu.Unlock()

		if worker.cm != nil {
			delay = true
		}
	}

	if delay {
		log.Warn("Maatqd will exit after 5 seconds")
		timer := time.NewTimer(time.Second * 5)
		<-timer.C
	}

	for _, worker := range g.Workers {
		worker.pushBackCurrentMsg()
		worker.client.Close()
	}
}

// 获取监听队列的 Group
func NewWorkerGroup(opt *GroupOptions) (*WorkerGroup, error) {
	if opt.Parallel < 0 {
		return nil, ErrParallel
	}

	if opt.Parallel == 0 {
		opt.Parallel = runtime.NumCPU()
	}

	if len(opt.Queues) == 0 {
		return nil, errors.New("No queues for listening")
	}

	ptr := &WorkerGroup{
		C:       make(chan int, opt.Parallel),
		Workers: make([]*Worker, opt.Parallel),
		options: opt,
	}

	ptr.initWorkers()

	return ptr, nil
}
