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
	ErrEventAlreadyExists = errors.New("event already exists")
	DefaultFailedQueue    = "maatq:default:failed"
	DefaultQueue          = "maatq:default"
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
	cm            *handlingMessage
	queues        []string
}

func (w *Worker) AddEventHandler(event string, handler EventHandler) error {
	if _, ok := w.eventHandlers[event]; ok {
		return ErrEventAlreadyExists
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
		"queues":   w.queues,
	})
}

func (w *Worker) Work() {
	w.Logger.WithField("try", w.try).Info("Worker started")

	for {
		result, err := w.client.BLPop(0, w.queues...).Result()
		if err != nil {
			w.Logger.Error(err)
			continue
		}

		w.Logger.WithFields(log.Fields{
			"msg": result[1],
		}).Debugf("[%s] message recieved", result[0])

		cm, err := newHandlingMessage(result[0], result[1])
		if err != nil {
			w.Logger.Error(err)
			continue
		}
		w.cm = cm

		w.processCurrentMsg()
	}

	// 这个代码永远不会运行到
	// w.c <- 1
}

// 处理当前消息
func (w *Worker) processCurrentMsg() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.handle()
	w.cm = nil
}

func (w *Worker) pushBackCurrentMsg() {
	if w.cm != nil {
		bytes, _ := json.Marshal(w.cm.Msg)
		w.client.LPush(w.cm.Queue, bytes)
	}
}

func (w *Worker) handle() {
	var (
		hm      *handlingMessage = w.cm
		message Message          = *(hm.Msg)
		handler EventHandler
		event   string
	)

	if !w.checkMessage(&message) {
		return
	}

	event = message.Event
	handler = w.eventHandlers[event]
	result, err := handler.Call(message.Data)

	if err != nil {
		w.cm.Error = err
		w.cm.EndTime = time.Now()
		w.Logger.WithFields(message.ToLogFields()).Errorf("[%.2fms] [%s]: %v", w.cm.milliSeconds(), "fail", err)
		if message.Try < w.try {
			w.requeue()
		} else {
			w.enqueueFailed()
		}
		w.notify(false, err.Error(), nil)
	} else {
		w.cm.EndTime = time.Now()
		w.Logger.WithFields(message.ToLogFields()).Infof("[%.2fms] [%s]", w.cm.milliSeconds(), "ok")
		w.notify(true, "", result)
	}
}

func (w *Worker) enqueueFailed() {
	bytes, _ := json.Marshal(w.cm.Msg)
	w.client.RPush(DefaultFailedQueue, string(bytes[:]))
}

func (w *Worker) notify(success bool, errMsg string, data interface{}) {
	var (
		message *Message = w.cm.Msg
		jResp   map[string]interface{}
		resp    string
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

func (w *Worker) requeue() {
	message := w.cm.Msg
	message.Try += 1
	message.Timestamp = time.Now().Unix()
	bytes, _ := json.Marshal(message)
	w.client.RPush(w.cm.Queue, string(bytes))
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
