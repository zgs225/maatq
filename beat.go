package maatq

import (
	"bytes"
	"container/heap"
	"encoding/gob"
	"encoding/json"
	"errors"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/go-redis/redis"
)

// The periodic task Scheduler

func init() {
	gob.Register(&minHeap{})
	gob.Register(&PriorityMessage{})
	gob.Register(&Message{})
	gob.Register(&Crontab{})
	gob.Register(&Period{})
}

const (
	DEFAULT_MAX_INTERVAL time.Duration = 5 * time.Minute
	MAATQ_DUMPS_KEY                    = "maatq:heap:dumps"
)

type Scheduler struct {
	mu           sync.Mutex
	interval     time.Duration
	heap         *minHeap
	logger       *log.Entry
	lastSyncTime time.Time
	isRunning    bool
	r            *redis.Client
	csleep       *cancelSleep
	health       *checkItem
}

func (s *Scheduler) toJSON() string {
	return s.heap.String()
}

func (s *Scheduler) SetInterval(v time.Duration) {
	s.interval = v
}

func (s *Scheduler) ServeLoop() {
	s.logger.Info("Starting...")
	s.logger.Debugf("Ticking with max interval %s", s.interval.String())

	for s.isRunning {
		d, err := s.tick()

		if err != nil {
			s.logger.Errorf("Error in a tick: %v", err)
			s.dumps()
		}

		if int64(d) > 0 {
			s.logger.Debugf("Waking up in %s", d.String())
			s.csleep.Sleep(d)
		}

		s.health.Alive()
	}
}

// Delay a message in give duration
func (s *Scheduler) Delay(m *Message, d time.Duration) {
	t := time.Now().Add(d)
	pm := &PriorityMessage{*m, t.Unix(), nil}
	s.mu.Lock()
	heap.Push(s.heap, pm)
	s.mu.Unlock()
	s.csleep.Cancel()
}

// 添加周期执行的任务
func (s *Scheduler) Period(m *Message, p *Period) {
	s.logger.WithFields(m.ToLogFields()).WithField("period", time.Second*time.Duration(p.Cycle)).Info("Periodic message recieved")
	t := p.Next()
	pm := &PriorityMessage{*m, t.Unix(), p}
	s.mu.Lock()
	heap.Push(s.heap, pm)
	s.mu.Unlock()
	s.csleep.Cancel()
}

// 添加Crontab任务
func (s *Scheduler) Crontab(m *Message, cron *Crontab) {
	s.logger.WithFields(m.ToLogFields()).WithField("crontab", cron.Text).Info("Crontab message recieved")
	t := cron.Next()
	pm := &PriorityMessage{*m, t.Unix(), cron}
	s.mu.Lock()
	heap.Push(s.heap, pm)
	s.mu.Unlock()
	s.csleep.Cancel()
}

// 取消一个任务
func (s *Scheduler) Cancel(id string) bool {
	s.csleep.Cancel()
	items := *s.heap.Items
	for i := 0; i < s.heap.Len(); i++ {
		item := items[i]
		if item.Id == id {
			s.mu.Lock()
			m1 := heap.Remove(s.heap, i)
			s.mu.Unlock()
			m2 := m1.(*PriorityMessage)
			log.WithFields(m2.ToLogFields()).Warn("Canceld")
			return true
		}
	}
	return false
}

// Run a tick, one iteration of the scheduler, executes one due task per call.
// Returns preferred delay duration for next call
func (s *Scheduler) tick() (time.Duration, error) {
	if s.heap.Len() <= 0 {
		return s.interval, nil
	}

	s.mu.Lock()
	m, ok := heap.Pop(s.heap).(*PriorityMessage)
	s.mu.Unlock()
	if !ok {
		return time.Duration(0), errors.New("Message asserts as *PriorityMessage error.")
	}
	s.logger.WithFields(m.ToLogFields()).Debug("priority message recieved.")

	if m.IsDue() {
		b, err := json.Marshal(m.Message)
		if err != nil {
			s.logger.Error(err)
			return time.Duration(0), err
		}
		s.logger.WithField("msg", string(b)).Debugf("Priority message push to queue %s", m.GetWorkQueue())
		s.r.RPush(m.GetWorkQueue(), string(b))
		if m.IsPeriodic() {
			m.T = m.P.Next().Unix()
			s.mu.Lock()
			heap.Push(s.heap, m)
			s.mu.Unlock()
		}
		return time.Duration(0), nil
	} else {
		d := time.Unix(m.T, 0).Sub(time.Now())
		if d > s.interval {
			d = s.interval
		}
		s.mu.Lock()
		heap.Push(s.heap, m)
		s.mu.Unlock()
		return d, nil
	}
}

// Mark scheduler as not running
func (s *Scheduler) shutdown() {
	s.isRunning = false
}

func (s *Scheduler) dumps() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(s.heap); err != nil {
		return err
	}
	return s.r.Set(MAATQ_DUMPS_KEY, buf.Bytes(), 0).Err()
}

func (s *Scheduler) loads() (*minHeap, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var heap minHeap
	b, err := s.r.Get(MAATQ_DUMPS_KEY).Bytes()
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(b)
	if err := gob.NewDecoder(buf).Decode(&heap); err != nil {
		return nil, err
	}
	return &heap, nil
}

func NewDefaultScheduler(addr, password string) *Scheduler {
	h := newHeap()
	l := log.WithFields(log.Fields{
		"workerId": "scheduler",
	})
	return &Scheduler{
		interval:  DEFAULT_MAX_INTERVAL,
		heap:      h,
		logger:    l,
		isRunning: true,
		r: redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password,
			DB:       0,
		}),
		csleep: newCancelSleep(),
		health: NewCheckItem("Schedular", DEFAULT_MAX_INTERVAL+time.Second, "Task schedular"),
	}
}
