package maatq

import (
	"testing"
	"time"
)

func TestCheckItem(t *testing.T) {
	i := NewCheckItem("Test", time.Second, "testing item")
	i.SetDeadFunc(func(c *checkItem) error {
		t.Log("I'm dead.")
		return nil
	})
	i.SetAliveFunc(func(c *checkItem) error {
		t.Log("I'm alive.")
		return nil
	})
	time.Sleep(time.Second)

	// Check isTimeout()
	if !i.isTimeout() {
		t.Error("isTimeout error")
	}

	i.dead()

	i.Alive()
	i.Alive()
	i.Alive()
	i.dead()
	i.Alive()
}

func TestHealthChecker(t *testing.T) {
	c := NewHealthChecker(time.Second)
	i := NewCheckItem("Test", time.Second, "testing item")
	i.SetDeadFunc(func(c *checkItem) error {
		t.Log("I'm dead.")
		return nil
	})
	i.SetAliveFunc(func(c *checkItem) error {
		t.Log("I'm alive.")
		return nil
	})
	c.AddItem(i)
}
