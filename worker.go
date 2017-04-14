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
type Worker struct {
	Id            int
	Client        *redis.Client
	EventHandlers map[string]EventHandler
	Try           int
	C             chan int
	Logger        *log.Entry
	Mu            sync.Mutex
	CurrentMsg    *string
}

func (w *Worker) AddEventHandler(event string, handler EventHandler) error {
	if _, ok := w.EventHandlers[event]; ok {
		return EventAlreadyExists
	}
	w.EventHandlers[event] = handler
	return nil
}

func (w *Worker) RemoveEventHandler(event string) {
	delete(w.EventHandlers, event)
}

func (w *Worker) checkConn() {
	cmd := w.Client.Ping()
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
	w.Logger.WithField("try", w.Try).Info("worker started")

	for {
		result, err := w.Client.BLPop(0, DefaultQueue).Result()
		if err != nil {
			w.Logger.Error(err)
			continue
		}
		w.CurrentMsg = &(result[1])
		w.Mu.Lock()

		w.Logger.WithFields(log.Fields{
			"msg": result[1],
		}).Debug("message recieved")
		w.handle(result[1])
		w.CurrentMsg = nil
		w.Mu.Unlock()
	}

	w.C <- 1
}

func (w *Worker) PushBackCurrentMsg() {
	if w.CurrentMsg != nil {
		w.Client.LPush(DefaultQueue, *(w.CurrentMsg))
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
	handler = w.EventHandlers[event]
	result, err := handler.Call(message.Data)

	if err != nil {
		w.Logger.WithFields(message.ToLogFields()).Error(err)
		if message.Try < w.Try {
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
	w.Client.RPush(DefaultFailedQueue, string(bytes[:]))
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

	w.Client.Set(message.Id, resp, 0)
}

func (w *Worker) requeue(message *Message) {
	message.Try += 1
	message.Timestamp = time.Now().Unix()
	bytes, _ := json.Marshal(*message)
	w.Client.RPush(DefaultQueue, string(bytes[:]))
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
		_, ok := w.EventHandlers[event]

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
