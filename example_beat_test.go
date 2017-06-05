package maatq_test

import (
	"fmt"

	"github.com/zgs225/maatq"
)

// This example demostrate how to use a scheduler
func Example_ServeLoop() {
	s := maatq.NewDefaultScheduler()
	s.ServeLoop()
	fmt.Println("hello")
	// Output: hello
}
