package maatq

import (
	"container/heap"
	"sync"
)

// minHeap to hold peroidic messages, minHeap implements
// container/heap.Interface
type minHeap struct {
	mu    sync.Mutex
	Items *[]*PriorityMessage
}

func (h minHeap) Len() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(*h.Items)
}

func (h minHeap) Less(i, j int) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return (*h.Items)[i].T < (*h.Items)[j].T
}

func (h minHeap) Swap(i, j int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	(*h.Items)[i], (*h.Items)[j] = (*h.Items)[j], (*h.Items)[i]
}

func (h minHeap) Push(x interface{}) {
	h.mu.Lock()
	defer h.mu.Unlock()
	*h.Items = append(*h.Items, x.(*PriorityMessage))
}

func (h minHeap) Pop() interface{} {
	h.mu.Lock()
	defer h.mu.Unlock()
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
