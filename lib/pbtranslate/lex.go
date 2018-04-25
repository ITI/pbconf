// CAUTION: Generated file - DO NOT EDIT.

package pbtranslate // golex -o lex.go gen/lex.l

import (
	"bufio"
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
	log.Error(e)
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
		goto yyrule8
	case c == 'O':
		goto yystate5
	case c == 'P':
		goto yystate9
	case c == 'S':
		goto yystate17
	case c == '\t' || c == '\n' || c == '\r' || c == ' ':
		goto yystate3
	case c == '\x00':
		goto yystate2
	case c == 'o':
		goto yystate25
	case c == 'p':
		goto yystate27
	case c == 's':
		goto yystate34
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'N' || c == 'Q' || c == 'R' || c >= 'T' && c <= 'Z' || c >= 'a' && c <= 'n' || c == 'q' || c == 'r' || c >= 't' && c <= 'z':
		goto yystate4
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
		goto yyrule8
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate4
	}

yystate5:
	c = y.getc()
	switch {
	default:
		goto yyrule8
	case c == 'F':
		goto yystate6
	case c == 'N':
		goto yystate8
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'E' || c >= 'G' && c <= 'M' || c >= 'O' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate4
	}

yystate6:
	c = y.getc()
	switch {
	default:
		goto yyrule8
	case c == 'F':
		goto yystate7
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'E' || c >= 'G' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate4
	}

yystate7:
	c = y.getc()
	switch {
	default:
		goto yyrule7
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate4
	}

yystate8:
	c = y.getc()
	switch {
	default:
		goto yyrule6
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate4
	}

yystate9:
	c = y.getc()
	switch {
	default:
		goto yyrule8
	case c == 'A':
		goto yystate10
	case c >= '0' && c <= '9' || c >= 'B' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate4
	}

yystate10:
	c = y.getc()
	switch {
	default:
		goto yyrule8
	case c == 'S':
		goto yystate11
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'R' || c >= 'T' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate4
	}

yystate11:
	c = y.getc()
	switch {
	default:
		goto yyrule8
	case c == 'S':
		goto yystate12
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'R' || c >= 'T' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate4
	}

yystate12:
	c = y.getc()
	switch {
	default:
		goto yyrule8
	case c == 'W':
		goto yystate13
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'V' || c >= 'X' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate4
	}

yystate13:
	c = y.getc()
	switch {
	default:
		goto yyrule8
	case c == 'O':
		goto yystate14
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'N' || c >= 'P' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate4
	}

yystate14:
	c = y.getc()
	switch {
	default:
		goto yyrule8
	case c == 'R':
		goto yystate15
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Q' || c >= 'S' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate4
	}

yystate15:
	c = y.getc()
	switch {
	default:
		goto yyrule8
	case c == 'D':
		goto yystate16
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'C' || c >= 'E' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate4
	}

yystate16:
	c = y.getc()
	switch {
	default:
		goto yyrule4
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate4
	}

yystate17:
	c = y.getc()
	switch {
	default:
		goto yyrule8
	case c == 'E':
		goto yystate18
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'D' || c >= 'F' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate4
	}

yystate18:
	c = y.getc()
	switch {
	default:
		goto yyrule8
	case c == 'R':
		goto yystate19
	case c == 'T':
		goto yystate24
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Q' || c == 'S' || c >= 'U' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate4
	}

yystate19:
	c = y.getc()
	switch {
	default:
		goto yyrule8
	case c == 'V':
		goto yystate20
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'U' || c >= 'W' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate4
	}

yystate20:
	c = y.getc()
	switch {
	default:
		goto yyrule8
	case c == 'I':
		goto yystate21
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'H' || c >= 'J' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate4
	}

yystate21:
	c = y.getc()
	switch {
	default:
		goto yyrule8
	case c == 'C':
		goto yystate22
	case c >= '0' && c <= '9' || c == 'A' || c == 'B' || c >= 'D' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate4
	}

yystate22:
	c = y.getc()
	switch {
	default:
		goto yyrule8
	case c == 'E':
		goto yystate23
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'D' || c >= 'F' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate4
	}

yystate23:
	c = y.getc()
	switch {
	default:
		goto yyrule5
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate4
	}

yystate24:
	c = y.getc()
	switch {
	default:
		goto yyrule3
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate4
	}

yystate25:
	c = y.getc()
	switch {
	default:
		goto yyrule8
	case c == 'f':
		goto yystate26
	case c == 'n':
		goto yystate8
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'e' || c >= 'g' && c <= 'm' || c >= 'o' && c <= 'z':
		goto yystate4
	}

yystate26:
	c = y.getc()
	switch {
	default:
		goto yyrule8
	case c == 'f':
		goto yystate7
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'e' || c >= 'g' && c <= 'z':
		goto yystate4
	}

yystate27:
	c = y.getc()
	switch {
	default:
		goto yyrule8
	case c == 'a':
		goto yystate28
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c >= 'b' && c <= 'z':
		goto yystate4
	}

yystate28:
	c = y.getc()
	switch {
	default:
		goto yyrule8
	case c == 's':
		goto yystate29
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'r' || c >= 't' && c <= 'z':
		goto yystate4
	}

yystate29:
	c = y.getc()
	switch {
	default:
		goto yyrule8
	case c == 's':
		goto yystate30
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'r' || c >= 't' && c <= 'z':
		goto yystate4
	}

yystate30:
	c = y.getc()
	switch {
	default:
		goto yyrule8
	case c == 'w':
		goto yystate31
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'v' || c >= 'x' && c <= 'z':
		goto yystate4
	}

yystate31:
	c = y.getc()
	switch {
	default:
		goto yyrule8
	case c == 'o':
		goto yystate32
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'n' || c >= 'p' && c <= 'z':
		goto yystate4
	}

yystate32:
	c = y.getc()
	switch {
	default:
		goto yyrule8
	case c == 'r':
		goto yystate33
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'q' || c >= 's' && c <= 'z':
		goto yystate4
	}

yystate33:
	c = y.getc()
	switch {
	default:
		goto yyrule8
	case c == 'd':
		goto yystate16
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'c' || c >= 'e' && c <= 'z':
		goto yystate4
	}

yystate34:
	c = y.getc()
	switch {
	default:
		goto yyrule8
	case c == 'e':
		goto yystate35
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'd' || c >= 'f' && c <= 'z':
		goto yystate4
	}

yystate35:
	c = y.getc()
	switch {
	default:
		goto yyrule8
	case c == 'r':
		goto yystate36
	case c == 't':
		goto yystate24
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'q' || c == 's' || c >= 'u' && c <= 'z':
		goto yystate4
	}

yystate36:
	c = y.getc()
	switch {
	default:
		goto yyrule8
	case c == 'v':
		goto yystate37
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'u' || c >= 'w' && c <= 'z':
		goto yystate4
	}

yystate37:
	c = y.getc()
	switch {
	default:
		goto yyrule8
	case c == 'i':
		goto yystate38
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'h' || c >= 'j' && c <= 'z':
		goto yystate4
	}

yystate38:
	c = y.getc()
	switch {
	default:
		goto yyrule8
	case c == 'c':
		goto yystate39
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == 'a' || c == 'b' || c >= 'd' && c <= 'z':
		goto yystate4
	}

yystate39:
	c = y.getc()
	switch {
	default:
		goto yyrule8
	case c == 'e':
		goto yystate23
	case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'd' || c >= 'f' && c <= 'z':
		goto yystate4
	}

yyrule1: // [ \t\r\n]+

	goto yystate0
yyrule2: // \0
	{
		return 0
	}
yyrule3: // SET|set
	{
		return TOKSET
	}
yyrule4: // PASSWORD|password
	{
		return TOKPASSWORD
	}
yyrule5: // SERVICE|service
	{
		return TOKSERVICE
	}
yyrule6: // ON|on
	{
		lval.lit = string(y.buf)
		return TOKSTATE
		goto yystate0
	}
yyrule7: // OFF|off
	{
		lval.lit = string(y.buf)
		return TOKSTATE
		goto yystate0
	}
yyrule8: // [a-zA-Z0-9]*
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
