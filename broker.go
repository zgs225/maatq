package maatq

import (
	"encoding/json"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/go-redis/redis"
	"github.com/google/uuid"
)

// 用于代理启动Workers和Scheduler，并且提供对外HTTP API
type Broker struct {
	scheduler *Scheduler
	group     *WorkerGroup
	config    *BrokerOptions
	redis     *redis.Client
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
		redis: redis.NewClient(&redis.Options{
			Addr:     config.Addr,
			Password: config.Password,
			DB:       0,
		}),
	}
	if config.Scheduler {
		broker.scheduler = NewDefaultScheduler(config.Addr, config.Password)
		h, err := broker.scheduler.loads()
		log.Debug("Loading dumps: ", h)
		if err == nil && h.Len() > 0 {
			log.Debug("Dumps loaded")
			broker.scheduler.heap = h
		}
		if err != nil {
			log.Error("Dumps load error: ", err)
		}
	}
	go broker.handleSignals()
	return broker, nil
}

func (b *Broker) ServeLoop(addr string) {
	ch := make(chan error)
	go b.group.ServeLoop()
	if b.config.Scheduler {
		go b.scheduler.ServeLoop()
	}
	go b.ServeHttp(addr, ch)
	log.Error(<-ch)
}

func (b *Broker) ServeHttp(addr string, ch chan error) {
	log.Info("Http serve: ", addr)
	handler := b.newHttpServer()
	ch <- http.ListenAndServe(addr, handler)
}

func (b *Broker) newHttpServer() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
	mux.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
	mux.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
	mux.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	mux.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))

	mux.HandleFunc("/v1/messages/dispatch", func(w http.ResponseWriter, r *http.Request) {
		var m Message
		err := json.NewDecoder(r.Body).Decode(&m)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Server", "mataq/1.0")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			resp := response{
				Ok:   false,
				Err:  err.Error(),
				Code: 100,
			}
			json.NewEncoder(w).Encode(&resp)
			return
		}

		id := uuid.New()
		m.Id = id.String()
		m.Timestamp = time.Now().Unix()
		m.Try = 0

		if err := b.Enqueue(DefaultQueue, &m); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			resp := response{
				Ok:   false,
				Err:  err.Error(),
				Code: 101,
			}
			json.NewEncoder(w).Encode(&resp)
		} else {
			w.WriteHeader(http.StatusOK)
			resp := response{
				Ok:      true,
				EventId: m.Id,
			}
			json.NewEncoder(w).Encode(&resp)
		}
	})

	mux.HandleFunc("/v1/messages/delay", func(w http.ResponseWriter, r *http.Request) {
		var (
			m   Message
			req delayRequest
		)
		err := json.NewDecoder(r.Body).Decode(&req)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Server", "mataq/1.0")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			resp := response{
				Ok:   false,
				Err:  err.Error(),
				Code: 100,
			}
			json.NewEncoder(w).Encode(&resp)
			return
		}

		m.Id = uuid.New().String()
		m.Event = req.Event
		m.Data = req.Data
		m.Timestamp = time.Now().Unix()
		m.Try = 0
		d, err := time.ParseDuration(req.Delay)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			resp := response{
				Ok:   false,
				Err:  err.Error(),
				Code: 103,
			}
			json.NewEncoder(w).Encode(&resp)
			return
		}
		b.scheduler.Delay(&m, d)
		w.WriteHeader(http.StatusOK)
		resp := response{
			Ok:      true,
			EventId: m.Id,
		}
		json.NewEncoder(w).Encode(&resp)
	})

	mux.HandleFunc("/v1/messages/period", func(w http.ResponseWriter, r *http.Request) {
		var (
			m   Message
			req periodRequest
		)
		err := json.NewDecoder(r.Body).Decode(&req)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Server", "mataq/1.0")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			resp := response{
				Ok:   false,
				Err:  err.Error(),
				Code: 100,
			}
			json.NewEncoder(w).Encode(&resp)
			return
		}

		m.Id = uuid.New().String()
		m.Event = req.Event
		m.Data = req.Data
		m.Timestamp = time.Now().Unix()
		m.Try = 0
		p, err := NewPeriod(req.Period)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			resp := response{
				Ok:   false,
				Err:  err.Error(),
				Code: 104,
			}
			json.NewEncoder(w).Encode(&resp)
			return
		}
		b.scheduler.Period(&m, p)
		w.WriteHeader(http.StatusOK)
		resp := response{
			Ok:      true,
			EventId: m.Id,
		}
		json.NewEncoder(w).Encode(&resp)
	})

	mux.HandleFunc("/v1/messages/crontab", func(w http.ResponseWriter, r *http.Request) {
		var (
			m   Message
			req crontabRequest
		)
		err := json.NewDecoder(r.Body).Decode(&req)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Server", "mataq/1.0")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			resp := response{
				Ok:   false,
				Err:  err.Error(),
				Code: 100,
			}
			json.NewEncoder(w).Encode(&resp)
			return
		}

		m.Id = uuid.New().String()
		m.Event = req.Event
		m.Data = req.Data
		m.Timestamp = time.Now().Unix()
		m.Try = 0
		cron, err := NewCrontab(req.Crontab)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			resp := response{
				Ok:   false,
				Err:  err.Error(),
				Code: 105,
			}
			json.NewEncoder(w).Encode(&resp)
			return
		}
		b.scheduler.Crontab(&m, cron)
		w.WriteHeader(http.StatusOK)
		resp := response{
			Ok:      true,
			EventId: m.Id,
		}
		json.NewEncoder(w).Encode(&resp)
	})

	return mux
}

func (b *Broker) Enqueue(queue string, m *Message) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	cmd := b.redis.RPush(queue, data)
	return cmd.Err()
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

func (b *Broker) Dumps() error {
	return b.scheduler.dumps()
}

func (b *Broker) handleSignals() {
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
	<-ch
	log.Warn("Prepare to safe exit...")
	if b.scheduler != nil {
		log.Warn("Dumping heap...")
		if err := b.scheduler.dumps(); err != nil {
			log.Error("调度器保存错误: ", err)
		}
	}
	b.group.cleanup()
	os.Exit(0)
}
