package maatq

import (
	"testing"
	"time"
)

func TestCancelSleep(t *testing.T) {
	cs := newCancelSleep()
	ch := make(chan bool)
	be := time.Now()
	go func() {
		cs.Sleep(time.Second)
		ch <- true
	}()
	time.Sleep(500 * time.Millisecond)
	cs.Cancel()
	<-ch
	if time.Since(be) >= time.Second {
		t.Error("Cancel sleep error")
	}
	t.Log("Sleep", time.Since(be))
}
