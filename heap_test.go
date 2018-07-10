package maatq

import (
	"container/heap"
	"testing"
)

func TestHeap(t *testing.T) {
	h := newHeap()

	for i := 500000; i > 0; i-- {
		heap.Push(h, &PriorityMessage{
			T: int64(i),
		})
	}

	l := h.Len()
	if l != 500000 {
		t.Errorf("Heap length error: expected[%d] got[%d]", 500000, l)
	}

	for j := 1; j <= 500000; j++ {
		v := heap.Pop(h).(*PriorityMessage)
		if v.T != int64(j) {
			t.Error("Heap pop error")
		}
	}
}
