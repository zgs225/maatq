package maatq

import (
	"container/heap"
	"testing"
)

func TestHeap(t *testing.T) {
	h := newHeap()

	for i := 10; i > 0; i-- {
		heap.Push(h, &PriorityMessage{
			T: int64(i),
		})
	}

	l := h.Len()
	if l != 10 {
		t.Errorf("Heap length error: expected[%d] got[%d]", 10, l)
	}

	for j := 1; j <= 10; j++ {
		v := heap.Pop(h).(*PriorityMessage)
		if v.T != int64(j) {
			t.Error("Heap pop error")
		}
	}
}
