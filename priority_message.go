package maatq

// 根据时间调度的消息
type PriorityMessage struct {
	Message
	T int64 `json:"t"` // 下一次执行的时间
}
