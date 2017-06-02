package maatq

import (
	"container/heap"
	"sync"
)

// minHeap to hold peroidic messages, minHeap implements
// container/heap.Interface
type minHeap struct {
	mu    sync.Mutex
	items *[]*PriorityMessage
}

func (h minHeap) Len() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(*h.items)
}

func (h minHeap) Less(i, j int) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return (*h.items)[i].T < (*h.items)[j].T
}

func (h minHeap) Swap(i, j int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	(*h.items)[i], (*h.items)[j] = (*h.items)[j], (*h.items)[i]
}

func (h minHeap) Push(x interface{}) {
	h.mu.Lock()
	defer h.mu.Unlock()
	*h.items = append(*h.items, x.(*PriorityMessage))
}

func (h minHeap) Pop() interface{} {
	h.mu.Lock()
	defer h.mu.Unlock()
	n := len(*h.items) - 1
	v := (*h.items)[n]
	*h.items = (*h.items)[0:n]
	return v
}

// newHeap get a pointer of minHeap
func newHeap() *minHeap {
	s := make([]*PriorityMessage, 0)
	h := &minHeap{
		items: &s,
	}
	heap.Init(h)

	return h
}
