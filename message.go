package maatq

import (
	"encoding/json"
	"time"

	log "github.com/Sirupsen/logrus"
)

type Message struct {
	Id        string      `json:"id"`
	Event     string      `json:"event"`
	Timestamp int64       `json:"timestamp"`
	Try       int         `json:"try"`
	Data      interface{} `json:"data,omitempty"`
}

func (m *Message) ToLogFields() log.Fields {
	return log.Fields{
		"eventId":   m.Id,
		"event":     m.Event,
		"timestamp": m.Timestamp,
		"try":       m.Try,
		"data":      m.Data,
	}
}

// 处理中的消息的结构
type handlingMessage struct {
	Queue     string
	Msg       *Message
	Error     error
	Result    interface{}
	StartTime time.Time
	EndTime   time.Time
}

func newHandlingMessage(queue, msg string) (*handlingMessage, error) {
	var (
		m  Message
		rv *handlingMessage
	)
	if err := json.Unmarshal([]byte(msg), &m); err != nil {
		return nil, err
	}

	rv = &handlingMessage{
		Queue:     queue,
		Msg:       &m,
		StartTime: time.Now(),
	}

	return rv, nil
}

// 获取开始和结束的毫秒
func (hm *handlingMessage) milliSeconds() float64 {
	d := hm.EndTime.Sub(hm.StartTime)
	ms := d / time.Microsecond
	nsec := d % time.Microsecond
	return float64(ms) + float64(nsec)/(1e6)
}
