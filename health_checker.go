package maatq

import (
	"errors"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
)

var (
	errHealthCheckItemExists    = errors.New("Health check item already exists.")
	errHealthCheckItemNotExists = errors.New("Health check item not exists.")
	healthCheckMinInternval     = time.Second
)

type checkItem struct {
	mu        sync.Mutex
	name      string
	timestamp int64
	alive     bool
	aliveFunc func(*checkItem) error
	comment   string
	deadline  time.Duration
	deadFunc  func(*checkItem) error
	logger    *log.Entry
}

func NewCheckItem(n string, d time.Duration, c string) *checkItem {
	v := &checkItem{
		name:      n,
		timestamp: time.Now().Unix(),
		alive:     true,
		deadline:  d,
		logger:    log.WithField("健康检查项", n),
		comment:   c,
	}
	return v
}

// healthChecker 健康度检查器
type healthChecker struct {
	items    map[string]*checkItem
	interval time.Duration
}

func NewHealthChecker(i time.Duration) *healthChecker {
	if i < healthCheckMinInternval {
		i = healthCheckMinInternval
	}
	return &healthChecker{
		items:    make(map[string]*checkItem),
		interval: i,
	}
}

// AddItem 生成新的健康度检查项
// 参数列表:
//     - name: 检查项名称
//     - d: 超时死亡时间
func (c *healthChecker) AddItem(i *checkItem) error {
	_, ok := c.items[i.name]
	if ok {
		return errHealthCheckItemExists
	}
	c.items[i.name] = i
	return nil
}

func (c *healthChecker) ServeLoop() error {
	for {
		for _, i := range c.items {
			if i.isTimeout() && i.isAlive() {
				i.dead()
			}
		}
		time.Sleep(c.interval)
	}
}

func (c *checkItem) isTimeout() bool {
	return time.Unix(c.timestamp, 0).Add(c.deadline).Before(time.Now())
}

func (c *checkItem) isAlive() bool {
	return c.alive
}

func (c *checkItem) dead() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.alive = false
	c.logger.Error("已死亡")
	if c.deadFunc != nil {
		if err := c.deadFunc(c); err != nil {
			c.logger.WithError(err).Error("死亡回调错误")
		}
	}
}

func (c *checkItem) Alive() {
	c.mu.Lock()
	defer c.mu.Unlock()
	dead := !c.alive
	c.alive = true
	c.timestamp = time.Now().Unix()
	if dead && c.aliveFunc != nil {
		if err := c.aliveFunc(c); err != nil {
			c.logger.WithError(err).Error("复活回调错误")
		}
	}
}

func (c *checkItem) SetDeadFunc(f func(*checkItem) error) {
	c.deadFunc = f
}

func (c *checkItem) SetAliveFunc(f func(*checkItem) error) {
	c.aliveFunc = f
}
