package maatq

import (
	"container/heap"
	"encoding/json"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/go-redis/redis"
)

// The periodic task scheduler

const (
	DEFAULT_MAX_INTERVAL time.Duration = 5 * time.Minute
	SYNC_INTERVAL                      = 5 * time.Minute
)

type scheduler struct {
	interval     time.Duration
	heap         *minHeap
	logger       *log.Entry
	lastSyncTime time.Time
	isRunning    bool
	r            *redis.Client
	csleep       *cancelSleep
}

func (s *scheduler) SetInterval(v time.Duration) {
	s.interval = v
}

func (s *scheduler) ServeLoop() {
	s.logger.Info("Starting...")
	s.logger.Debugf("Ticking with max interval %s", s.interval.String())

	for s.isRunning {
		d, err := s.tick()

		if err != nil {
			s.logger.Errorf("Error in a tick: %v", err)
			s.shutdown()
			s.sync()
		}

		if int64(d) > 0 {
			s.logger.Debugf("Waking up in %s", d.String())
			s.csleep.Sleep(d)
		}

		if s.shouldSync() {
			s.sync()
		}
	}
}

// Delay a message in give duration
func (s *scheduler) Delay(m *Message, d time.Duration) {
	t := time.Now().Add(d)
	pm := &PriorityMessage{*m, t.Unix()}
	heap.Push(s.heap, pm)
	s.csleep.Cancel()
}

// Run a tick, one iteration of the scheduler, executes one due task per call.
// Returns preferred delay duration for next call
func (s *scheduler) tick() (time.Duration, error) {
	if s.heap.Len() <= 0 {
		return s.interval, nil
	}

	m := heap.Pop(s.heap).(*PriorityMessage)
	s.logger.WithFields(m.ToLogFields()).Debug("priority message recieved.")

	if m.IsDue() {
		b, err := json.Marshal(m.Message)
		if err != nil {
			s.logger.Error(err)
			return time.Duration(0), err
		}
		s.logger.WithField("msg", string(b)).Debugf("Priority message push to queue %s", DefaultQueue)
		s.r.RPush(DefaultQueue, string(b))
		return time.Duration(0), nil
	} else {
		d := time.Unix(m.T, 0).Sub(time.Now())
		if d > s.interval {
			d = s.interval
		}
		heap.Push(s.heap, m)
		return d, nil
	}
}

func (s *scheduler) shouldSync() bool {
	if s.lastSyncTime.IsZero() {
		return true
	}
	return time.Now().Sub(s.lastSyncTime) >= SYNC_INTERVAL
}

// Sync task from redis
func (s *scheduler) sync() {
}

// Mark scheduler as not running
func (s *scheduler) shutdown() {
	s.isRunning = false
}

func NewDefaultScheduler(addr, password string) *scheduler {
	h := newHeap()
	l := log.WithFields(log.Fields{
		"workerId": "scheduler",
	})
	return &scheduler{
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
	}
}
