package maatq

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

// This example demostrate how to use a scheduler
// func Example_ServeLoop() {
// 	s := maatq.NewDefaultScheduler("127.0.0.1:6379", "")
// 	s.ServeLoop()
// 	fmt.Println("hello")
// 	// Output: hello
// }

func TestDumpsAndLoads(t *testing.T) {
	s := NewDefaultScheduler("127.0.0.1:6379", "")
	for i := 0; i < 10; i++ {
		s.Delay(&Message{
			Id:    fmt.Sprintf("ID(%d)", i),
			Event: "hello",
			Try:   3,
			Data:  "yuez",
		}, time.Duration(i+1)*time.Second)
	}
	if err := s.dumps(); err != nil {
		t.Error("Dumps error: ", err)
	}
	h, err := s.loads()
	if err != nil {
		t.Error("Loads error: ", err)
	}
	t.Log(h.Items)
	if !reflect.DeepEqual(h.Items, s.heap.Items) {
		t.Error("Loads heap does not equals original heap: ", h.Items, s.heap.Items)
	}
}
