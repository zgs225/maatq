package maatq

import (
	"time"
)

type cancelSleep struct {
	quit chan int
}

func (cs *cancelSleep) Sleep(d time.Duration) {
	go func() { time.AfterFunc(d, cs.Cancel) }()
	<-cs.quit
}

func (cs *cancelSleep) Cancel() {
	cs.quit <- 1
}

func newCancelSleep() *cancelSleep {
	return &cancelSleep{
		quit: make(chan int, 1),
	}
}
