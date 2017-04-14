package mataq

import (
	"encoding/json"
	"errors"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/go-redis/redis"
)

var (
	EventAlreadyExists = errors.New("event already exists")
	DefaultQueue       = "mataq:default"
	DefaultFailedQueue = "mataq:default:failed"
)

type EventHandler func(arg interface{}) (interface{}, error)

func (h EventHandler) Call(arg interface{}) (interface{}, error) {
	return h(arg)
}

// 消费者用于从队列中取消息，并且分配任务
// 每个消费者包含一个 redis client 和 一个事件和函数的对应列表
type Consumer struct {
	Id            int
	Client        *redis.Client
	EventHandlers map[string]EventHandler
	Try           int
	C             chan int
	Logger        *log.Entry
	Mu            sync.Mutex
	CurrentMsg    *string
}

func (c *Consumer) AddEventHandler(event string, handler EventHandler) error {
	if _, ok := c.EventHandlers[event]; ok {
		return EventAlreadyExists
	}
	c.EventHandlers[event] = handler
	return nil
}

func (c *Consumer) RemoveEventHandler(event string) {
	delete(c.EventHandlers, event)
}

func (c *Consumer) checkConn() {
	cmd := c.Client.Ping()
	if err := cmd.Err(); err != nil {
		c.Logger.Panic(err)
	}
}

func (c *Consumer) initLog() {
	c.Logger = log.WithFields(log.Fields{
		"workerId": c.Id,
		"queue":    DefaultQueue,
	})
}

func (c *Consumer) Work() {
	c.Logger.WithField("try", c.Try).Info("worker started")

	for {
		result, err := c.Client.BLPop(0, DefaultQueue).Result()
		if err != nil {
			c.Logger.Error(err)
			continue
		}
		c.CurrentMsg = &(result[1])
		c.Mu.Lock()

		c.Logger.WithFields(log.Fields{
			"msg": result[1],
		}).Debug("message recieved")
		c.handle(result[1])
		c.CurrentMsg = nil
		c.Mu.Unlock()
	}

	c.C <- 1
}

func (c *Consumer) PushBackCurrentMsg() {
	if c.CurrentMsg != nil {
		c.Client.LPush(DefaultQueue, *(c.CurrentMsg))
	}
}

func (c *Consumer) handle(msg string) {
	var (
		message Message
		handler EventHandler
		event   string
	)

	err := json.Unmarshal([]byte(msg), &message)

	if err != nil {
		c.Logger.Errorf("json unmarshal error: %v", err)
		return
	}

	if !c.checkMessage(&message) {
		return
	}

	event = message.Event
	handler = c.EventHandlers[event]
	result, err := handler.Call(message.Data)

	if err != nil {
		c.Logger.WithFields(message.ToLogFields()).Error(err)
		if message.Try < c.Try {
			c.requeue(&message)
		} else {
			c.enqueueFailed(&message)
		}
		c.notify(&message, false, err.Error(), nil)
	} else {
		c.Logger.WithFields(message.ToLogFields()).Info("success")
		c.notify(&message, true, "", result)
	}
}

func (c *Consumer) enqueueFailed(message *Message) {
	bytes, _ := json.Marshal(*message)
	c.Client.RPush(DefaultFailedQueue, string(bytes[:]))
}

func (c *Consumer) notify(message *Message, success bool, errMsg string, data interface{}) {
	var (
		jResp map[string]interface{}
		resp  string
	)

	jResp = map[string]interface{}{
		"success":   success,
		"error":     errMsg,
		"data":      data,
		"timestamp": time.Now().Unix(),
	}

	bytes, _ := json.Marshal(jResp)
	resp = string(bytes[:])
	c.Logger.WithField("eventId", message.Id).Debug(resp)

	c.Client.Set(message.Id, resp, 0)
}

func (c *Consumer) requeue(message *Message) {
	message.Try += 1
	message.Timestamp = time.Now().Unix()
	bytes, _ := json.Marshal(*message)
	c.Client.RPush(DefaultQueue, string(bytes[:]))
}

func (c *Consumer) checkMessage(message *Message) bool {
	var (
		checked = true
		err     string
	)

	if message.Event == "" {
		checked = false
		err = "field event required"
		c.Logger.WithFields(message.ToLogFields()).Error(err)
		return checked
	} else {
		event := message.Event
		_, ok := c.EventHandlers[event]

		if !ok {
			checked = false
			err = "event handler for event not found"
			c.Logger.WithFields(message.ToLogFields()).Error(err)
			return checked
		}
	}

	if len(message.Id) == 0 {
		checked = false
		err = "event id required"
		c.Logger.WithFields(message.ToLogFields()).Error(err)
		return checked
	}

	return checked
}
