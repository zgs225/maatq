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

var (
	ParallelError = errors.New("parallel should gte 0")
)

type ConsumerGroup struct {
	Parallel  int
	C         chan int
	Consumers []*Consumer

	try           int
	redisAddr     string
	redisPassword string
}

func (g *ConsumerGroup) ServeLoop() {
	for _, consumer := range g.Consumers {
		go consumer.Work()
	}
	g.wait()
}

func (g *ConsumerGroup) AddEventHandler(name string, handler EventHandler) {
	log.Warningf("Event[%s] handled by Func[%s]", name, GetFunctionName(handler))
	for _, consumer := range g.Consumers {
		consumer.AddEventHandler(name, handler)
	}
}

func (g *ConsumerGroup) wait() {
	for i := 0; i < g.Parallel; i++ {
		<-g.C
	}
}

func (g *ConsumerGroup) initConsumers() {
	for i := 0; i < g.Parallel; i++ {
		// 初始化Consumer
		c := &Consumer{Try: g.try, C: g.C, Id: i}
		g.Consumers[i] = c

		c.Client = redis.NewClient(&redis.Options{
			Addr:     g.redisAddr,
			Password: g.redisPassword,
			DB:       0,
		})
		c.EventHandlers = make(map[string]EventHandler)
		c.InitLog()
		c.CheckConn()
	}
}

func (g *ConsumerGroup) handleSignals() {
	sigC := make(chan os.Signal, 3)
	signal.Notify(sigC, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		for sig := range sigC {
			if sig == os.Interrupt || sig == syscall.SIGTERM || sig == syscall.SIGQUIT {
				log.Warn("Prepare to safe exit...")
				var delay = false

				for _, consumer := range g.Consumers {
					consumer.Mu.Lock()
					defer consumer.Mu.Unlock()

					if consumer.CurrentMsg != nil {
						delay = true
					}
				}

				if delay {
					log.Warn("Mataqd will exit after 5 seconds")
					timer := time.NewTimer(time.Second * 5)
					<-timer.C
				}

				for _, consumer := range g.Consumers {
					consumer.PushBackCurrentMsg()
				}

				os.Exit(0)
			}
		}
	}()
}

// 获取监听队列的 Group
func NewConsumerGroup(parallel int, addr, password string, try int) (*ConsumerGroup, error) {
	if parallel < 0 {
		return nil, ParallelError
	}

	if parallel == 0 {
		parallel = runtime.NumCPU()
	}

	ptr := &ConsumerGroup{
		Parallel:      parallel,
		C:             make(chan int, parallel),
		Consumers:     make([]*Consumer, parallel),
		try:           try,
		redisAddr:     addr,
		redisPassword: password,
	}

	ptr.initConsumers()
	ptr.handleSignals()

	return ptr, nil
}
