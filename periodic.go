package maatq

import (
	"time"
)

type Periodicor interface {
	Next() time.Time
}

type Period struct {
	Begin time.Time // 周期的开始时间
	Cycle int64     // 秒为单位
}

func (p *Period) Next() time.Time {
	return p.nextPeriodFrom(time.Now())
}

func (p *Period) nextPeriodFrom(from time.Time) time.Time {
	i := from.Unix()
	j := p.Begin.Unix()
	m := (i - j) / p.Cycle
	if m < 0 {
		m = 0
	}
	m++
	s := j + m*p.Cycle
	return time.Unix(s, 0)
}
