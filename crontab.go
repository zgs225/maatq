package maatq

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"unicode"
)

type Crontab struct {
	minutes     []int8 // 0 - 59
	hours       []int8 // 0 - 23
	daysOfMonth []int8 // 0 - 31
	months      []int8 // 0 - 12
	daysOfWeek  []int8 // 0 - 7 (0或者7是周日, 或者使用名字)
	text        string // Cron字符串
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

func (t *cronToken) IntVal() (int, error) {
	return strconv.Atoi(string(t.v))
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

type cronParser struct {
	lexer     *cronLexer
	lookahead []*cronToken
	cap       int32
	head      int32
}

func (p *cronParser) L(i int32) *cronToken {
	idx := (i + p.head) % p.cap
	return p.lookahead[idx]
}

func (p *cronParser) Consume() (*cronToken, error) {
	t1 := p.L(0)
	p.head = (p.head + 1) % p.cap
	t2, err := p.lexer.NextToken()
	if err != nil {
		return nil, err
	}
	p.lookahead[(p.head+p.cap-1)%p.cap] = t2
	return t1, nil
}

func (p *cronParser) Parse(cron *Crontab) error {
	if err := p.parseMinutes(cron); err != nil {
		return err
	}
	return nil
}

func (p *cronParser) Match(t cronTokenType) error {
	token := p.L(0)
	if token.t != t {
		return fmt.Errorf("语法错误: 应该是%s, 实际: %s, 值: %s", t.TokenName(),
			token.t.TokenName(), string(token.v))
	}
	return nil
}

// */2 形式的结构
// 返回 step cron token, error
func (p *cronParser) stepedAsterisk() (*cronToken, error) {
	if err := p.Match(cronTokenTypes_Asterisk); err != nil {
		return nil, err
	}
	if _, err := p.Consume(); err != nil {
		return nil, err
	}
	if err := p.Match(cronTokenTypes_Slash); err != nil {
		return nil, err
	}
	if _, err := p.Consume(); err != nil {
		return nil, err
	}
	if err := p.Match(cronTokenTypes_Number); err != nil {
		return nil, err
	}
	num, err := p.Consume()
	if err != nil {
		return nil, err
	}
	return num, nil
}

// *
func (p *cronParser) asterisk() error {
	if err := p.Match(cronTokenTypes_Asterisk); err != nil {
		return err
	}
	if _, err := p.Consume(); err != nil {
		return err
	}
	return nil
}

// 10-30/2
// 返回 begin, end, step int value
func (p *cronParser) stepedRange() (int, int, int, error) {
	// Number
	if err := p.Match(cronTokenTypes_Number); err != nil {
		return 0, 0, 0, err
	}
	begin, err := p.Consume()
	if err != nil {
		return 0, 0, 0, err
	}
	v1, err := begin.IntVal()
	if err != nil {
		return 0, 0, 0, err
	}
	// Hyphen
	if err := p.Match(cronTokenTypes_Hyphen); err != nil {
		return 0, 0, 0, err
	}
	_, err = p.Consume()
	if err != nil {
		return 0, 0, 0, err
	}
	// Number
	if err := p.Match(cronTokenTypes_Number); err != nil {
		return 0, 0, 0, err
	}
	end, err := p.Consume()
	if err != nil {
		return 0, 0, 0, err
	}
	v2, err := end.IntVal()
	if err != nil {
		return 0, 0, 0, err
	}
	// Slash
	if err := p.Match(cronTokenTypes_Slash); err != nil {
		return 0, 0, 0, err
	}
	_, err = p.Consume()
	if err != nil {
		return 0, 0, 0, err
	}
	// Number
	if err := p.Match(cronTokenTypes_Number); err != nil {
		return 0, 0, 0, err
	}
	step, err := p.Consume()
	if err != nil {
		return 0, 0, 0, err
	}
	v3, err := step.IntVal()
	if err != nil {
		return 0, 0, 0, err
	}
	return v1, v2, v3, nil
}

// 10-30
// 返回 begin, end cron token
func (p *cronParser) cronRange() (int, int, error) {
	if err := p.Match(cronTokenTypes_Number); err != nil {
		return 0, 0, err
	}
	begin, err := p.Consume()
	if err != nil {
		return 0, 0, err
	}
	v1, err := begin.IntVal()
	if err != nil {
		return 0, 0, err
	}
	if err := p.Match(cronTokenTypes_Hyphen); err != nil {
		return 0, 0, err
	}
	if _, err := p.Consume(); err != nil {
		return 0, 0, err
	}
	if err := p.Match(cronTokenTypes_Number); err != nil {
		return 0, 0, err
	}
	end, err := p.Consume()
	if err != nil {
		return 0, 0, err
	}
	v2, err := end.IntVal()
	if err != nil {
		return 0, 0, err
	}
	return v1, v2, nil
}

// 3
func (p *cronParser) number() (int, error) {
	if err := p.Match(cronTokenTypes_Number); err != nil {
		return 0, err
	}
	token, err := p.Consume()
	if err != nil {
		return 0, err
	}
	return token.IntVal()
}

func (p *cronParser) parseMinutes(cron *Crontab) error {
	head := p.L(0)
	// 当分钟是 *
	if head.t == cronTokenTypes_Asterisk {
		if p.L(1).t == cronTokenTypes_Slash { // 当分钟是 */2 类似的模式
			token, err := p.stepedAsterisk()
			if err != nil {
				return err
			}
			step, err := token.IntVal()
			if err != nil {
				return err
			}
			cron.minutes = makeRangeOfInt8(int8(0), int8(59), step)
			return nil
		} else { // 当分钟是 *
			if err := p.asterisk(); err != nil {
				return err
			}
			cron.minutes = makeRangeOfInt8(int8(0), int8(59), 1)
			return nil
		}
	} else if head.t == cronTokenTypes_Number {
		if p.L(1).t == cronTokenTypes_Hyphen { // 当分钟是 0-59
			if p.L(3).t == cronTokenTypes_Slash { // 是 0-59/3
				v1, v2, v3, err := p.stepedRange()
				if err != nil {
					return err
				}
				if v1 < 0 || v1 > 59 {
					return fmt.Errorf("语法错误: 分钟取值范围是0-59, 实际: %d", v1)
				}
				if v2 < 0 || v2 > 59 {
					return fmt.Errorf("语法错误: 分钟取值范围是0-59, 实际: %d", v2)
				}
				if v1 > v2 {
					return fmt.Errorf("语法错误: 分钟取值范围错误, 实际: %d-%d", v1, v2)
				}
				cron.minutes = makeRangeOfInt8(int8(v1), int8(v2), v3)
			} else {
				v1, v2, err := p.cronRange()
				if err != nil {
					return err
				}
				if v1 < 0 || v1 > 59 {
					return fmt.Errorf("语法错误: 分钟取值范围是0-59, 实际: %d", v1)
				}
				if v2 < 0 || v2 > 59 {
					return fmt.Errorf("语法错误: 分钟取值范围是0-59, 实际: %d", v2)
				}
				if v1 > v2 {
					return fmt.Errorf("语法错误: 分钟取值范围错误, 实际: %d-%d", v1, v2)
				}
				cron.minutes = makeRangeOfInt8(int8(v1), int8(v2), 1)
			}
		} else if p.L(1).t == cronTokenTypes_Comma { // 当分钟是 0,13,20
			return p.list(cron)
		} else { // 单纯的数字
			v, err := p.number()
			if err != nil {
				return err
			}
			if v < 0 || v > 59 {
				return fmt.Errorf("语法错误: 分钟取值范围是0-59, 实际: %d", v)
			}
			cron.minutes = []int8{int8(v)}
		}
		return nil
	} else {
		return fmt.Errorf("语法错误: 应该是%s或者%s，实际: %s，值: %s",
			cronTokenTypes_Asterisk.TokenName(), cronTokenTypes_Number.TokenName(),
			head.t.TokenName(), string(head.v))
	}
}

func (p *cronParser) list(cron *Crontab) error {
	if err := p.Match(cronTokenTypes_Number); err != nil {
		return err
	}
	t1, err := p.Consume()
	if err != nil {
		return err
	}
	v1, err := t1.IntVal()
	if err != nil {
		return err
	}
	if v1 < 0 || v1 > 59 {
		return fmt.Errorf("语法错误: 分钟取值范围是0-59, 实际: %d", v1)
	}
	if err := p.Match(cronTokenTypes_Comma); err != nil {
		return err
	}
	p.Consume()
	cron.minutes = append(cron.minutes, int8(v1))

	if p.L(1).t == cronTokenTypes_Comma { // Remain list
		return p.list(cron)
	} else {
		if err := p.Match(cronTokenTypes_Number); err != nil {
			return err
		}
		t2, err := p.Consume()
		if err != nil {
			return err
		}
		v2, err := t2.IntVal()
		if err != nil {
			return err
		}
		if v2 < 0 || v2 > 59 {
			return fmt.Errorf("语法错误: 分钟取值范围是0-59, 实际: %d", v2)
		}
		cron.minutes = append(cron.minutes, int8(v2))
		return nil
	}
}

func newCronParser(lexer *cronLexer, cap int32) (*cronParser, error) {
	p := &cronParser{
		lexer:     lexer,
		lookahead: make([]*cronToken, cap, cap),
		cap:       cap,
		head:      0,
	}
	for i := 0; int32(i) < cap; i++ {
		t, err := lexer.NextToken()
		if err != nil {
			return nil, err
		}
		p.lookahead[i] = t
	}
	return p, nil
}

// 将Crontab的字符串解析成*Crontab实例
func NewCrontab(cron string) (*Crontab, error) {
	var v = Crontab{
		text: cron,
	}

	lexer := newCronLexer(cron)
	parser, err := newCronParser(lexer, 5)
	if err != nil {
		return nil, err
	}

	if err := parser.Parse(&v); err != nil {
		return nil, err
	}

	return &v, nil
}
