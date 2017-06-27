package maatq

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestCronLexer(t *testing.T) {
	lexer := newCronLexer("* 30 */2 * *")
	if token, _ := lexer.NextToken(); token.t != cronTokenTypes_Asterisk {
		t.Error("期望", cronTokenTypes_Asterisk.TokenName(), "结果", token.t.TokenName())
	}
	if token, _ := lexer.NextToken(); token.t != cronTokenTypes_Number {
		t.Error("期望", cronTokenTypes_Number.TokenName(), "结果", token.t.TokenName())
	}
	if token, _ := lexer.NextToken(); token.t != cronTokenTypes_Asterisk {
		t.Error("期望", cronTokenTypes_Asterisk.TokenName(), "结果", token.t.TokenName())
	}
	if token, _ := lexer.NextToken(); token.t != cronTokenTypes_Slash {
		t.Error("期望", cronTokenTypes_Slash.TokenName(), "结果", token.t.TokenName())
	}
	if token, _ := lexer.NextToken(); token.t != cronTokenTypes_Number {
		t.Error("期望", cronTokenTypes_Number.TokenName(), "结果", token.t.TokenName())
	}
	if token, _ := lexer.NextToken(); token.t != cronTokenTypes_Asterisk {
		t.Error("期望", cronTokenTypes_Asterisk.TokenName(), "结果", token.t.TokenName())
	}
	if token, _ := lexer.NextToken(); token.t != cronTokenTypes_Asterisk {
		t.Error("期望", cronTokenTypes_Asterisk.TokenName(), "结果", token.t.TokenName())
	}
	if token, _ := lexer.NextToken(); token.t != cronTokenTypes_EOF {
		t.Error("期望", cronTokenTypes_EOF.TokenName(), "结果", token.t.TokenName())
	}
}

func TestCronParser(t *testing.T) {
	lexer := newCronLexer("* 30 */2 * *")
	parser, _ := newCronParser(lexer, 5)
	if token := parser.L(0); token.t != cronTokenTypes_Asterisk {
		t.Error("期望", cronTokenTypes_Asterisk.TokenName(), "结果", token.t.TokenName())
	}
	if token := parser.L(1); token.t != cronTokenTypes_Number {
		t.Error("期望", cronTokenTypes_Number.TokenName(), "结果", token.t.TokenName())
	}
	if token := parser.L(2); token.t != cronTokenTypes_Asterisk {
		t.Error("期望", cronTokenTypes_Asterisk.TokenName(), "结果", token.t.TokenName())
	}
	if token := parser.L(3); token.t != cronTokenTypes_Slash {
		t.Error("期望", cronTokenTypes_Slash.TokenName(), "结果", token.t.TokenName())
	}
	if token := parser.L(4); token.t != cronTokenTypes_Number {
		t.Error("期望", cronTokenTypes_Number.TokenName(), "结果", token.t.TokenName())
	}
}

func TestCronParserConsume(t *testing.T) {
	lexer := newCronLexer("* 30 */2 * *")
	parser, _ := newCronParser(lexer, 5)
	if token, _ := parser.Consume(); token.t != cronTokenTypes_Asterisk {
		t.Error("期望", cronTokenTypes_Asterisk.TokenName(), "结果", token.t.TokenName())
	}
	if token, _ := parser.Consume(); token.t != cronTokenTypes_Number {
		t.Error("期望", cronTokenTypes_Number.TokenName(), "结果", token.t.TokenName())
	}
	if token, _ := parser.Consume(); token.t != cronTokenTypes_Asterisk {
		t.Error("期望", cronTokenTypes_Asterisk.TokenName(), "结果", token.t.TokenName())
	}
	if token, _ := parser.Consume(); token.t != cronTokenTypes_Slash {
		t.Error("期望", cronTokenTypes_Slash.TokenName(), "结果", token.t.TokenName())
	}
	if token, _ := parser.Consume(); token.t != cronTokenTypes_Number {
		t.Error("期望", cronTokenTypes_Number.TokenName(), "结果", token.t.TokenName())
	}
	if token, _ := parser.Consume(); token.t != cronTokenTypes_Asterisk {
		t.Error("期望", cronTokenTypes_Asterisk.TokenName(), "结果", token.t.TokenName())
	}
	if token, _ := parser.Consume(); token.t != cronTokenTypes_Asterisk {
		t.Error("期望", cronTokenTypes_Asterisk.TokenName(), "结果", token.t.TokenName())
	}
	if token, _ := parser.Consume(); token.t != cronTokenTypes_EOF {
		t.Error("期望", cronTokenTypes_EOF.TokenName(), "结果", token.t.TokenName())
	}
}

func TestCrontab(t *testing.T) {
	{
		s := "* * * * *"
		cron, err := NewCrontab(s)
		if err != nil {
			t.Error(s, err)
		}
		minutes := makeRangeOfInt8(int8(0), int8(59), 1)
		if !reflect.DeepEqual(cron.Minutes, minutes) {
			t.Error(s, "分钟解析错误: 期望", minutes, "结果", cron.Minutes)
		}
		hours := makeRangeOfInt8(int8(0), int8(23), 1)
		if !reflect.DeepEqual(cron.Hours, hours) {
			t.Error(s, "小时解析错误: 期望", hours, "结果", cron.Hours)
		}
		daysOfMonth := makeRangeOfInt8(int8(0), int8(31), 1)
		if !reflect.DeepEqual(cron.DaysOfMonth, daysOfMonth) {
			t.Error(s, "天解析错误: 期望", daysOfMonth, "结果", cron.DaysOfMonth)
		}
		months := makeRangeOfInt8(int8(0), int8(12), 1)
		if !reflect.DeepEqual(cron.Months, months) {
			t.Error(s, "月解析错误: 期望", months, "结果", cron.Months)
		}
		daysOfWeek := []int8{}
		if !reflect.DeepEqual(cron.DaysOfWeek, daysOfWeek) {
			t.Error(s, "周解析错误: 期望", daysOfWeek, "结果", cron.DaysOfWeek)
		}
	}

	{
		s := "*/3 */2 */4 */3 */5"
		cron, err := NewCrontab(s)
		if err != nil {
			t.Error(s, err)
		}
		minutes := makeRangeOfInt8(int8(0), int8(59), 3)
		if !reflect.DeepEqual(cron.Minutes, minutes) {
			t.Error(s, "分钟解析错误: 期望", minutes, "结果", cron.Minutes)
		}
		hours := makeRangeOfInt8(int8(0), int8(23), 2)
		if !reflect.DeepEqual(cron.Hours, hours) {
			t.Error(s, "小时解析错误: 期望", hours, "结果", cron.Hours)
		}
		daysOfMonth := makeRangeOfInt8(int8(0), int8(31), 4)
		if !reflect.DeepEqual(cron.DaysOfMonth, daysOfMonth) {
			t.Error(s, "天解析错误: 期望", daysOfMonth, "结果", cron.DaysOfMonth)
		}
		months := makeRangeOfInt8(int8(0), int8(12), 3)
		if !reflect.DeepEqual(cron.Months, months) {
			t.Error(s, "月解析错误: 期望", months, "结果", cron.Months)
		}
		daysOfWeek := makeRangeOfInt8(int8(0), int8(7), 5)
		if !reflect.DeepEqual(cron.DaysOfWeek, daysOfWeek) {
			t.Error(s, "周解析错误: 期望", daysOfWeek, "结果", cron.DaysOfWeek)
		}
	}

	{
		s := "13 21 18 10 6"
		cron, err := NewCrontab(s)
		if err != nil {
			t.Error(s, err)
		}
		minutes := []int8{int8(13)}
		if !reflect.DeepEqual(cron.Minutes, minutes) {
			t.Error(s, "分钟解析错误: 期望", minutes, "结果", cron.Minutes)
		}
		hours := []int8{int8(21)}
		if !reflect.DeepEqual(cron.Hours, hours) {
			t.Error(s, "小时解析错误: 期望", hours, "结果", cron.Hours)
		}
		daysOfMonth := []int8{int8(18)}
		if !reflect.DeepEqual(cron.DaysOfMonth, daysOfMonth) {
			t.Error(s, "天解析错误: 期望", daysOfMonth, "结果", cron.DaysOfMonth)
		}
		months := []int8{int8(10)}
		if !reflect.DeepEqual(cron.Months, months) {
			t.Error(s, "月解析错误: 期望", months, "结果", cron.Months)
		}
		daysOfWeek := []int8{int8(6)}
		if !reflect.DeepEqual(cron.DaysOfWeek, daysOfWeek) {
			t.Error(s, "周解析错误: 期望", daysOfWeek, "结果", cron.DaysOfWeek)
		}
	}

	{
		s := "0,5,15,20 0,12,23 1,3,5 1,4,7,10 1,2,3,4,5"
		cron, err := NewCrontab(s)
		if err != nil {
			t.Error(s, err)
		}
		minutes := []int8{int8(0), int8(5), int8(15), int8(20)}
		if !reflect.DeepEqual(cron.Minutes, minutes) {
			t.Error(s, "分钟解析错误: 期望", minutes, "结果", cron.Minutes)
		}
		hours := []int8{int8(0), int8(12), int8(23)}
		if !reflect.DeepEqual(cron.Hours, hours) {
			t.Error(s, "小时解析错误: 期望", hours, "结果", cron.Hours)
		}
		daysOfMonth := []int8{int8(1), int8(3), int8(5)}
		if !reflect.DeepEqual(cron.DaysOfMonth, daysOfMonth) {
			t.Error(s, "天解析错误: 期望", daysOfMonth, "结果", cron.DaysOfMonth)
		}
		months := []int8{int8(1), int8(4), int8(7), int8(10)}
		if !reflect.DeepEqual(cron.Months, months) {
			t.Error(s, "月解析错误: 期望", months, "结果", cron.Months)
		}
		daysOfWeek := []int8{int8(1), int8(2), int8(3), int8(4), int8(5)}
		if !reflect.DeepEqual(cron.DaysOfWeek, daysOfWeek) {
			t.Error(s, "周解析错误: 期望", daysOfWeek, "结果", cron.DaysOfWeek)
		}
	}

	{
		s := "0-20 9-17 1-13 1-6 1-5"
		cron, err := NewCrontab(s)
		if err != nil {
			t.Error(s, err)
		}
		minutes := makeRangeOfInt8(int8(0), int8(20), 1)
		if !reflect.DeepEqual(cron.Minutes, minutes) {
			t.Error(s, "分钟解析错误: 期望", minutes, "结果", cron.Minutes)
		}
		hours := makeRangeOfInt8(int8(9), int8(17), 1)
		if !reflect.DeepEqual(cron.Hours, hours) {
			t.Error(s, "小时解析错误: 期望", hours, "结果", cron.Hours)
		}
		daysOfMonth := makeRangeOfInt8(int8(1), int8(13), 1)
		if !reflect.DeepEqual(cron.DaysOfMonth, daysOfMonth) {
			t.Error(s, "天解析错误: 期望", daysOfMonth, "结果", cron.DaysOfMonth)
		}
		months := makeRangeOfInt8(int8(1), int8(6), 1)
		if !reflect.DeepEqual(cron.Months, months) {
			t.Error(s, "月解析错误: 期望", months, "结果", cron.Months)
		}
		daysOfWeek := makeRangeOfInt8(int8(1), int8(5), 1)
		if !reflect.DeepEqual(cron.DaysOfWeek, daysOfWeek) {
			t.Error(s, "周解析错误: 期望", daysOfWeek, "结果", cron.DaysOfWeek)
		}
	}

	{
		s := "0-20/3 1-12/2 2-30/5 1-6/2 1-5/2"
		cron, err := NewCrontab(s)
		if err != nil {
			t.Error(s, err)
		}
		minutes := makeRangeOfInt8(int8(0), int8(20), 3)
		if !reflect.DeepEqual(cron.Minutes, minutes) {
			t.Error(s, "分钟解析错误: 期望", minutes, "结果", cron.Minutes)
		}
		hours := makeRangeOfInt8(int8(1), int8(12), 2)
		if !reflect.DeepEqual(cron.Hours, hours) {
			t.Error(s, "小时解析错误: 期望", hours, "结果", cron.Hours)
		}
		daysOfMonth := makeRangeOfInt8(int8(2), int8(30), 5)
		if !reflect.DeepEqual(cron.DaysOfMonth, daysOfMonth) {
			t.Error(s, "天解析错误: 期望", daysOfMonth, "结果", cron.DaysOfMonth)
		}
		months := makeRangeOfInt8(int8(1), int8(6), 2)
		if !reflect.DeepEqual(cron.Months, months) {
			t.Error(s, "月解析错误: 期望", months, "结果", cron.Months)
		}
		daysOfWeek := makeRangeOfInt8(int8(1), int8(5), 2)
		if !reflect.DeepEqual(cron.DaysOfWeek, daysOfWeek) {
			t.Error(s, "周解析错误: 期望", daysOfWeek, "结果", cron.DaysOfWeek)
		}
	}

	{
		for i, m := range []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"} {
			for j, w := range []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"} {
				s := fmt.Sprintf("* 1-12/2 1,2,3 %s %s", m, w)
				cron, err := NewCrontab(s)
				if err != nil {
					t.Error(s, err)
				}
				minutes := makeRangeOfInt8(int8(0), int8(59), 1)
				if !reflect.DeepEqual(cron.Minutes, minutes) {
					t.Error(s, "分钟解析错误: 期望", minutes, "结果", cron.Minutes)
				}
				hours := makeRangeOfInt8(int8(1), int8(12), 2)
				if !reflect.DeepEqual(cron.Hours, hours) {
					t.Error(s, "小时解析错误: 期望", hours, "结果", cron.Hours)
				}
				daysOfMonth := []int8{int8(1), int8(2), int8(3)}
				if !reflect.DeepEqual(cron.DaysOfMonth, daysOfMonth) {
					t.Error(s, "天解析错误: 期望", daysOfMonth, "结果", cron.DaysOfMonth)
				}
				months := []int8{int8(i + 1)}
				if !reflect.DeepEqual(cron.Months, months) {
					t.Error(s, "月解析错误: 期望", months, "结果", cron.Months)
				}
				daysOfWeek := []int8{int8(j)}
				if !reflect.DeepEqual(cron.DaysOfWeek, daysOfWeek) {
					t.Error(s, "周解析错误: 期望", daysOfWeek, "结果", cron.DaysOfWeek)
				}
			}
		}
	}

	{
		s := "0-/20/3 * * * *"
		_, err := NewCrontab(s)
		if err == nil {
			t.Error(s, "期望解析失败，结果解析成功")
		}
	}
}

func TestCrontabNext(t *testing.T) {
	{
		var (
			now, next time.Time
		)
		cron, _ := NewCrontab("* * * * *")
		now = time.Date(2017, 6, 1, 0, 0, 0, 0, time.Local)
		for i := 0; i < 100; i++ {
			next = cron.nextFrom(now)
			if next.Sub(now) != time.Minute {
				t.Error("下一次周期计算错误:", now, next)
				break
			} else {
				now = next
			}
		}
	}

	{
		var (
			now, next time.Time
		)
		cron, _ := NewCrontab("*/2 * * * *")
		now = time.Date(2017, 6, 1, 0, 0, 0, 0, time.Local)
		for i := 0; i < 100; i++ {
			next = cron.nextFrom(now)
			if next.Sub(now) != 2*time.Minute {
				t.Error("下一次周期计算错误:", now, next)
				break
			} else {
				now = next
			}
		}
	}

	{
		var (
			now, next time.Time
		)
		cron, _ := NewCrontab("0 0 3 * *")
		now = time.Date(2017, 6, 1, 0, 0, 0, 0, time.Local)
		for i := 0; i < 100; i++ {
			next = cron.nextFrom(now)
			if next.Day() != 3 || next.Minute() != 0 || next.Hour() != 0 {
				t.Error("[*/2 * 3 * *] 下一次周期计算错误:", now, next)
				break
			} else {
				now = next
			}
		}
	}

	{
		var (
			now, next time.Time
		)
		cron, _ := NewCrontab("0 0 0 * 1-5")
		now = time.Date(2017, 6, 1, 0, 0, 0, 0, time.Local)
		for i := 0; i < 100; i++ {
			next = cron.nextFrom(now)
			// t.Log(next, next.Weekday())
			now = next
		}
	}

	{
		var (
			now, next time.Time
		)
		cron, _ := NewCrontab("*/30 * 1 4-7 *")
		now = time.Date(2017, 6, 1, 0, 0, 0, 0, time.Local)
		for i := 0; i < 100; i++ {
			next = cron.nextFrom(now)
			t.Log(next, next.Weekday())
			now = next
		}
	}
}

func BenchmarkCrontab(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewCrontab("* * * * *")
	}
}

func BenchmarkCrontabNext(b *testing.B) {
	cron, _ := NewCrontab("*/30 * 1 4-7 *")
	for i := 0; i < b.N; i++ {
		cron.Next()
	}
}
