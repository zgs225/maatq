package maatq_test

import (
	"fmt"

	"github.com/zgs225/maatq"
)

// This example demostrate how to use a scheduler
func Example_ServeLoop() {
	s := maatq.NewDefaultScheduler("127.0.0.1:6379", "")
	s.ServeLoop()
	fmt.Println("hello")
	// Output: hello
}
