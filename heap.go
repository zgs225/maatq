package maatq

import (
	"bytes"
	"container/heap"
	"encoding/json"
)

// minHeap to hold peroidic messages, minHeap implements
// container/heap.Interface
type minHeap struct {
	Items *[]*PriorityMessage
}

func (h minHeap) String() string {
	b := bytes.Buffer{}
	if err := json.NewEncoder(&b).Encode(h.Items); err != nil {
		return err.Error()
	}
	return b.String()
}

func (h minHeap) Len() int {
	if h.Items == nil {
		return 0
	}
	return len(*h.Items)
}

func (h minHeap) Less(i, j int) bool {
	return (*h.Items)[i].T < (*h.Items)[j].T
}

func (h minHeap) Swap(i, j int) {
	(*h.Items)[i], (*h.Items)[j] = (*h.Items)[j], (*h.Items)[i]
}

func (h minHeap) Push(x interface{}) {
	*h.Items = append(*h.Items, x.(*PriorityMessage))
}

func (h minHeap) Pop() interface{} {
	n := len(*h.Items) - 1
	v := (*h.Items)[n]
	*h.Items = (*h.Items)[0:n]
	return v
}

// newHeap get a pointer of minHeap
func newHeap() *minHeap {
	s := make([]*PriorityMessage, 0)
	h := &minHeap{
		Items: &s,
	}
	heap.Init(h)

	return h
}
