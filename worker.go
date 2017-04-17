package maatq

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
	DefaultQueue       = "maatq:default"
	DefaultFailedQueue = "maatq:default:failed"
)

type EventHandler func(arg interface{}) (interface{}, error)

func (h EventHandler) Call(arg interface{}) (interface{}, error) {
	return h(arg)
}

// 消费者用于从队列中取消息，并且分配任务
// 每个消费者包含一个 redis client 和 一个事件和函数的对应列表
type Worker struct {
	Id     int
	Logger *log.Entry

	client        *redis.Client
	eventHandlers map[string]EventHandler
	try           int
	c             chan int
	mu            sync.Mutex
	currentMsg    *string
}

func (w *Worker) AddEventHandler(event string, handler EventHandler) error {
	if _, ok := w.eventHandlers[event]; ok {
		return EventAlreadyExists
	}
	w.eventHandlers[event] = handler
	return nil
}

func (w *Worker) RemoveEventHandler(event string) {
	delete(w.eventHandlers, event)
}

func (w *Worker) checkConn() {
	cmd := w.client.Ping()
	if err := cmd.Err(); err != nil {
		w.Logger.Panic(err)
	}
}

func (w *Worker) initLog() {
	w.Logger = log.WithFields(log.Fields{
		"workerId": w.Id,
		"queue":    DefaultQueue,
	})
}

func (w *Worker) Work() {
	w.Logger.WithField("try", w.try).Info("worker started")

	for {
		result, err := w.client.BLPop(0, DefaultQueue).Result()
		if err != nil {
			w.Logger.Error(err)
			continue
		}
		w.currentMsg = &(result[1])
		w.mu.Lock()

		w.Logger.WithFields(log.Fields{
			"msg": result[1],
		}).Debug("message recieved")
		w.handle(result[1])
		w.currentMsg = nil
		w.mu.Unlock()
	}

	// 这个代码永远不会运行到
	// w.c <- 1
}

func (w *Worker) pushBackCurrentMsg() {
	if w.currentMsg != nil {
		w.client.LPush(DefaultQueue, *(w.currentMsg))
	}
}

func (w *Worker) handle(msg string) {
	var (
		message Message
		handler EventHandler
		event   string
	)

	err := json.Unmarshal([]byte(msg), &message)

	if err != nil {
		w.Logger.Errorf("json unmarshal error: %v", err)
		return
	}

	if !w.checkMessage(&message) {
		return
	}

	event = message.Event
	handler = w.eventHandlers[event]
	result, err := handler.Call(message.Data)

	if err != nil {
		w.Logger.WithFields(message.ToLogFields()).Error(err)
		if message.Try < w.try {
			w.requeue(&message)
		} else {
			w.enqueueFailed(&message)
		}
		w.notify(&message, false, err.Error(), nil)
	} else {
		w.Logger.WithFields(message.ToLogFields()).Info("success")
		w.notify(&message, true, "", result)
	}
}

func (w *Worker) enqueueFailed(message *Message) {
	bytes, _ := json.Marshal(*message)
	w.client.RPush(DefaultFailedQueue, string(bytes[:]))
}

func (w *Worker) notify(message *Message, success bool, errMsg string, data interface{}) {
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
	w.Logger.WithField("eventId", message.Id).Debug(resp)

	w.client.Set(message.Id, resp, 0)
}

func (w *Worker) requeue(message *Message) {
	message.Try += 1
	message.Timestamp = time.Now().Unix()
	bytes, _ := json.Marshal(*message)
	w.client.RPush(DefaultQueue, string(bytes[:]))
}

func (w *Worker) checkMessage(message *Message) bool {
	var (
		checked = true
		err     string
	)

	if message.Event == "" {
		checked = false
		err = "field event required"
		w.Logger.WithFields(message.ToLogFields()).Error(err)
		return checked
	} else {
		event := message.Event
		_, ok := w.eventHandlers[event]

		if !ok {
			checked = false
			err = "event handler for event not found"
			w.Logger.WithFields(message.ToLogFields()).Error(err)
			return checked
		}
	}

	if len(message.Id) == 0 {
		checked = false
		err = "event id required"
		w.Logger.WithFields(message.ToLogFields()).Error(err)
		return checked
	}

	return checked
}
