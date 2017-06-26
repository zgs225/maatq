package maatq

import (
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
