package maatq

import (
	"sync/atomic"
	"time"
)

type cancelSleep struct {
	quit chan bool
	flag int32
}

func (cs *cancelSleep) Sleep(d time.Duration) {
	atomic.StoreInt32(&cs.flag, 1)
	select {
	case <-cs.quit:
		atomic.StoreInt32(&cs.flag, 0)
		break
	case <-time.After(d):
		atomic.StoreInt32(&cs.flag, 0)
		break
	}
}

func (cs *cancelSleep) Cancel() {
	if cs.flag == 1 {
		cs.quit <- true
	}
}

func newCancelSleep() *cancelSleep {
	return &cancelSleep{
		quit: make(chan bool, 1),
	}
}
