package maatq

import (
	log "github.com/Sirupsen/logrus"
	"time"
)

// 根据时间调度的消息
type PriorityMessage struct {
	Message
	T int64 `json:"t"` // 下一次执行的时间
}

func (pm *PriorityMessage) IsDue() bool {
	return time.Now().Unix() >= pm.T
}

func (pm *PriorityMessage) ToLogFields() log.Fields {
	v := (&pm.Message).ToLogFields()
	v["t"] = pm.T
	return v
}
