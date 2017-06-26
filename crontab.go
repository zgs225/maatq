package maatq

import (
	"bytes"
	"fmt"
	"io"
	"unicode"
)

type Crontab struct {
	minute     []int  // 0 - 59
	hour       []int  // 0 - 23
	dayOfMonth []int  // 0 - 31
	month      []int  // 0 - 12
	dayOfWeek  []int  // 0 - 7 (0或者7是周日, 或者使用名字)
	text       string // Cron字符串
}

type cronTokenType int32

const (
	cronTokenTypes_Number cronTokenType = iota
	cronTokenTypes_Word
	cronTokenTypes_Asterisk
	cronTokenTypes_Hyphen
	cronTokenTypes_Slash
	cronTokenTypes_Comma
	cronTokenTypes_EOF
)

func (t cronTokenType) TokenName() string {
	switch t {
	case cronTokenTypes_Number:
		return "<NUM>"
	case cronTokenTypes_Word:
		return "<WORD>"
	case cronTokenTypes_Asterisk:
		return "<ASTERISK>"
	case cronTokenTypes_Hyphen:
		return "<HYPHEN>"
	case cronTokenTypes_Slash:
		return "<SLASH>"
	case cronTokenTypes_Comma:
		return "<COMMA>"
	case cronTokenTypes_EOF:
		return "<EOF>"
	default:
		return "<UNKNOWN>"
	}
}

type cronToken struct {
	t cronTokenType
	v []byte
}

type cronLexer struct {
	b   byte
	r   io.ByteScanner
	err error
}

func (lexer *cronLexer) readByte() {
	lexer.b, lexer.err = lexer.r.ReadByte()
}

func (lexer *cronLexer) NextToken() (*cronToken, error) {
	lexer.readByte()
	if lexer.err != nil {
		goto return_errors
	}
	if unicode.IsSpace(rune(lexer.b)) {
		for {
			lexer.readByte()
			if lexer.err != nil {
				goto return_errors
			}
			if !unicode.IsSpace(rune(lexer.b)) {
				break
			}
		}
	}

	if unicode.IsDigit(rune(lexer.b)) {
		b := new(bytes.Buffer)
		lexer.err = b.WriteByte(lexer.b)
		if lexer.err != nil {
			goto return_errors
		}
		for {
			lexer.readByte()
			if !unicode.IsDigit(rune(lexer.b)) {
				lexer.r.UnreadByte()
				return &cronToken{
					t: cronTokenTypes_Number,
					v: b.Bytes(),
				}, nil
			} else {
				lexer.err = b.WriteByte(lexer.b)
				if lexer.err != nil {
					goto return_errors
				}
			}
		}
	}

	if lexer.b == '*' {
		return &cronToken{
			t: cronTokenTypes_Asterisk,
			v: []byte{lexer.b},
		}, nil
	}

	if lexer.b == '-' {
		return &cronToken{
			t: cronTokenTypes_Hyphen,
			v: []byte{lexer.b},
		}, nil
	}

	if lexer.b == '/' {
		return &cronToken{
			t: cronTokenTypes_Slash,
			v: []byte{lexer.b},
		}, nil
	}

	if lexer.b == ',' {
		return &cronToken{
			t: cronTokenTypes_Comma,
			v: []byte{lexer.b},
		}, nil
	}

	if unicode.IsLetter(rune(lexer.b)) {
		b := new(bytes.Buffer)
		lexer.err = b.WriteByte(lexer.b)
		if lexer.err != nil {
			goto return_errors
		}
		for {
			lexer.readByte()
			if !unicode.IsLetter(rune(lexer.b)) {
				lexer.r.UnreadByte()
				return &cronToken{
					t: cronTokenTypes_Word,
					v: b.Bytes(),
				}, nil
			} else {
				lexer.err = b.WriteByte(lexer.b)
				if lexer.err != nil {
					goto return_errors
				}
			}
		}
	}

	return nil, fmt.Errorf("语法错误: 不支持字符%c", lexer.b)

return_errors:
	if lexer.err == io.EOF {
		return &cronToken{
			t: cronTokenTypes_EOF,
			v: []byte("<EOF>"),
		}, nil
	}
	return nil, lexer.err
}

func newCronLexer(s string) *cronLexer {
	return &cronLexer{
		r: bytes.NewBufferString(s),
	}
}

// 将Crontab的字符串解析成*Crontab实例
func NewCrontab(cron string) (*Crontab, error) {
	var v Crontab

	return &v, nil
}
