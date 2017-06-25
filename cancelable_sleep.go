package maatq

import (
	"sync/atomic"
	"time"
)

type cancelSleep struct {
	quit chan int
	flag int32
}

func (cs *cancelSleep) Sleep(d time.Duration) {
	if atomic.LoadInt32(&cs.flag) == 0 {
		atomic.StoreInt32(&cs.flag, 1)
		go func() { time.AfterFunc(d, cs.Cancel) }()
		<-cs.quit
	}
}

func (cs *cancelSleep) Cancel() {
	if atomic.LoadInt32(&cs.flag) == 1 {
		cs.quit <- 1
		atomic.StoreInt32(&cs.flag, 0)
	}
}

func newCancelSleep() *cancelSleep {
	return &cancelSleep{
		quit: make(chan int, 1),
	}
}
