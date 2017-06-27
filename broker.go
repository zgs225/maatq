package maatq

import (
	"time"
)

// 用于代理启动Workers和Scheduler，并且提供对外HTTP API
type Broker struct {
	scheduler *Scheduler
	group     *WorkerGroup
	config    *BrokerOptions
}

type BrokerOptions struct {
	Parallel  int
	Addr      string
	Password  string
	Try       int
	Queues    []string
	Scheduler bool
}

func NewBroker(config *BrokerOptions) (*Broker, error) {
	group, err := NewWorkerGroup(&GroupOptions{
		Parallel: config.Parallel,
		Addr:     config.Addr,
		Password: config.Password,
		Try:      config.Try,
		Queues:   config.Queues,
	})
	if err != nil {
		return nil, err
	}
	broker := &Broker{
		group:  group,
		config: config,
	}
	if config.Scheduler {
		broker.scheduler = NewDefaultScheduler(config.Addr, config.Password)
	}
	return broker, nil
}

func (b *Broker) ServeLoop() {
	ch := make(chan bool)
	go b.group.ServeLoop()
	if b.config.Scheduler {
		go b.scheduler.ServeLoop()
	}
	<-ch
}

func (b *Broker) AddEventHandler(event string, handler EventHandler) {
	b.group.AddEventHandler(event, handler)
}

func (b *Broker) Delay(m *Message, d time.Duration) {
	if b.config.Scheduler {
		b.scheduler.Delay(m, d)
	}
}

func (b *Broker) Period(m *Message, p *Period) {
	if b.config.Scheduler {
		b.scheduler.Period(m, p)
	}
}

func (b *Broker) Crontab(m *Message, cron *Crontab) {
	if b.config.Scheduler {
		b.scheduler.Crontab(m, cron)
	}
}
