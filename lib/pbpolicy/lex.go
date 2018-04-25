// CAUTION: Generated file - DO NOT EDIT.

package policy // golex -o lex.go gen/lex.l

import (
	"bufio"
	"log"
)

type yylexer struct {
	src     *bufio.Reader
	buf     []byte
	empty   bool
	current byte
}

func NewLexer(src *bufio.Reader) (y *yylexer) {
	y = &yylexer{src: src}
	if b, err := src.ReadByte(); err == nil {
		y.current = b
	}
	return
}

func (y *yylexer) getc() byte {
	if y.current != 0 {
		y.buf = append(y.buf, y.current)
	}
	y.current = 0
	if b, err := y.src.ReadByte(); err == nil {
		y.current = b
	}
	return y.current
}

func (y yylexer) Error(e string) {
	log.Println(e)
}

func (y *yylexer) Lex(lval *yySymType) int {
	c := y.current
	if y.empty {
		c, y.empty = y.getc(), false
	}

yystate0:

	y.buf = y.buf[:0]

	goto yystart1

	goto yystate0 // silence unused label error
	goto yystate1 // silence unused label error
yystate1:
	c = y.getc()
yystart1:
	switch {
	default:
		goto yyrule7
	case c == '#':
		goto yystate4
	case c == '-' || c == '.' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'q' || c >= 's' && c <= 'z':
		goto yystate5
	case c == '\t' || c == '\n' || c == '\r' || c == ' ':
		goto yystate3
	case c == '\x00':
		goto yystate2
	case c == 'r':
		goto yystate6
	case c == '{':
		goto yystate14
	case c == '}':
		goto yystate15
	}

yystate2:
	c = y.getc()
	goto yyrule2

yystate3:
	c = y.getc()
	switch {
	default:
		goto yyrule1
	case c == '\t' || c == '\n' || c == '\r' || c == ' ':
		goto yystate3
	}

yystate4:
	c = y.getc()
	switch {
	default:
		goto yyrule5
	case c >= '\x01' && c <= '\t' || c >= '\v' && c <= 'ÿ':
		goto yystate4
	}

yystate5:
	c = y.getc()
	switch {
	default:
		goto yyrule7
	case c == '-' || c == '.' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate5
	}

yystate6:
	c = y.getc()
	switch {
	default:
		goto yyrule7
	case c == '-' || c == '.' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'd' || c >= 'f' && c <= 'z':
		goto yystate5
	case c == 'e':
		goto yystate7
	}

yystate7:
	c = y.getc()
	switch {
	default:
		goto yyrule7
	case c == '-' || c == '.' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'p' || c >= 'r' && c <= 'z':
		goto yystate5
	case c == 'q':
		goto yystate8
	}

yystate8:
	c = y.getc()
	switch {
	default:
		goto yyrule7
	case c == '-' || c == '.' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 't' || c >= 'v' && c <= 'z':
		goto yystate5
	case c == 'u':
		goto yystate9
	}

yystate9:
	c = y.getc()
	switch {
	default:
		goto yyrule7
	case c == '-' || c == '.' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'h' || c >= 'j' && c <= 'z':
		goto yystate5
	case c == 'i':
		goto yystate10
	}

yystate10:
	c = y.getc()
	switch {
	default:
		goto yyrule7
	case c == '-' || c == '.' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'q' || c >= 's' && c <= 'z':
		goto yystate5
	case c == 'r':
		goto yystate11
	}

yystate11:
	c = y.getc()
	switch {
	default:
		goto yyrule7
	case c == '-' || c == '.' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'd' || c >= 'f' && c <= 'z':
		goto yystate5
	case c == 'e':
		goto yystate12
	}

yystate12:
	c = y.getc()
	switch {
	default:
		goto yyrule7
	case c == '-' || c == '.' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'r' || c >= 't' && c <= 'z':
		goto yystate5
	case c == 's':
		goto yystate13
	}

yystate13:
	c = y.getc()
	switch {
	default:
		goto yyrule6
	case c == '-' || c == '.' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate5
	}

yystate14:
	c = y.getc()
	goto yyrule3

yystate15:
	c = y.getc()
	goto yyrule4

yyrule1: // [ \t\r\n]+

	goto yystate0
yyrule2: // \0
	{
		return 0
	}
yyrule3: // {
	{
		return TOKLBRACE
	}
yyrule4: // }
	{
		return TOKRBRACE
	}
yyrule5: // #.*
	{
		lval.lit = string(y.buf)
		return TOKCOMMENT
		goto yystate0
	}
yyrule6: // requires
	{
		lval.lit = string(y.buf)
		return TOKREQUIRE
		goto yystate0
	}
yyrule7: // [\-.a-zA-Z0-9]*
	{
		lval.lit = string(y.buf)
		return TOKWORD
		goto yystate0
	}
	panic("unreachable")

	goto yyabort // silence unused label error

yyabort: // no lexem recognized
	y.empty = true
	return int(c)
}
