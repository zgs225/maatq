package maatq

import (
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
