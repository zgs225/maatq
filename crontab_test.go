package maatq

import (
	"reflect"
	"testing"
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
		if !reflect.DeepEqual(cron.minutes, minutes) {
			t.Error(s, "分钟解析错误: 期望", minutes, "结果", cron.minutes)
		}
		hours := makeRangeOfInt8(int8(0), int8(23), 1)
		if !reflect.DeepEqual(cron.hours, hours) {
			t.Error(s, "小时解析错误: 期望", hours, "结果", cron.hours)
		}
	}

	{
		s := "*/3 */2 * * *"
		cron, err := NewCrontab(s)
		if err != nil {
			t.Error(s, err)
		}
		minutes := makeRangeOfInt8(int8(0), int8(59), 3)
		if !reflect.DeepEqual(cron.minutes, minutes) {
			t.Error(s, "分钟解析错误: 期望", minutes, "结果", cron.minutes)
		}
		hours := makeRangeOfInt8(int8(0), int8(23), 2)
		if !reflect.DeepEqual(cron.hours, hours) {
			t.Error(s, "小时解析错误: 期望", hours, "结果", cron.hours)
		}
	}

	{
		s := "13 21 * * *"
		cron, err := NewCrontab(s)
		if err != nil {
			t.Error(s, err)
		}
		minutes := []int8{int8(13)}
		if !reflect.DeepEqual(cron.minutes, minutes) {
			t.Error(s, "分钟解析错误: 期望", minutes, "结果", cron.minutes)
		}
		hours := []int8{int8(21)}
		if !reflect.DeepEqual(cron.hours, hours) {
			t.Error(s, "小时解析错误: 期望", hours, "结果", cron.hours)
		}
	}

	{
		s := "0,5,15,20 0,12,23 * * *"
		cron, err := NewCrontab(s)
		if err != nil {
			t.Error(s, err)
		}
		minutes := []int8{int8(0), int8(5), int8(15), int8(20)}
		if !reflect.DeepEqual(cron.minutes, minutes) {
			t.Error(s, "分钟解析错误: 期望", minutes, "结果", cron.minutes)
		}
		hours := []int8{int8(0), int8(12), int8(23)}
		if !reflect.DeepEqual(cron.hours, hours) {
			t.Error(s, "小时解析错误: 期望", hours, "结果", cron.hours)
		}
	}

	{
		s := "0-20 9-17 * * *"
		cron, err := NewCrontab(s)
		if err != nil {
			t.Error(s, err)
		}
		minutes := makeRangeOfInt8(int8(0), int8(20), 1)
		if !reflect.DeepEqual(cron.minutes, minutes) {
			t.Error(s, "分钟解析错误: 期望", minutes, "结果", cron.minutes)
		}
		hours := makeRangeOfInt8(int8(9), int8(17), 1)
		if !reflect.DeepEqual(cron.hours, hours) {
			t.Error(s, "小时解析错误: 期望", hours, "结果", cron.hours)
		}
	}

	{
		s := "0-20/3 1-12/2 * * *"
		cron, err := NewCrontab(s)
		if err != nil {
			t.Error(s, err)
		}
		minutes := makeRangeOfInt8(int8(0), int8(20), 3)
		if !reflect.DeepEqual(cron.minutes, minutes) {
			t.Error(s, "分钟解析错误: 期望", minutes, "结果", cron.minutes)
		}
		hours := makeRangeOfInt8(int8(1), int8(12), 2)
		if !reflect.DeepEqual(cron.hours, hours) {
			t.Error(s, "小时解析错误: 期望", hours, "结果", cron.hours)
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
