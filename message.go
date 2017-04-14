package mataq

import (
	log "github.com/Sirupsen/logrus"
)

type Message struct {
	Id        *string
	Event     *string
	Timestamp *int64
	Try       *int
	Data      *interface{}
}

func (m *Message) GetId() string {
	if m.Id == nil {
		return ""
	}
	return *(m.Id)
}

func (m *Message) GetEvent() string {
	if m.Event == nil {
		return ""
	}
	return *(m.Event)
}

func (m *Message) GetTimestamp() int64 {
	if m.Timestamp == nil {
		return 0
	}
	return *(m.Timestamp)
}

func (m *Message) GetTry() int {
	if m.Try == nil {
		return 0
	}
	return *(m.Try)
}

func (m *Message) GetData() interface{} {
	if m.Data == nil {
		return nil
	}

	return *(m.Data)
}

func (m *Message) ToLogFields() log.Fields {
	return log.Fields{
		"eventId":   m.GetId(),
		"event":     m.GetEvent(),
		"timestamp": m.GetTimestamp(),
		"try":       m.GetTry(),
		"data":      m.GetData(),
	}
}
