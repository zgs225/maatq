package maatq

import (
	"testing"
	"time"
)

func TestPeriod(t *testing.T) {
	b := time.Date(2017, time.June, 26, 12, 3, 5, 0, time.Local)
	p := &Period{
		Begin: b,
		Cycle: 10,
	}
	n := p.nextPeriodFrom(b)
	if n.Sub(b) != time.Second*10 {
		t.Error("下一次应该是在10秒后: ", n.Sub(b))
	}

	m := p.Next()
	t.Log("从现在起下一次执行时间是: ", m)
}
