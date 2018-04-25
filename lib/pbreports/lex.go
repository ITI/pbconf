// CAUTION: Generated file - DO NOT EDIT.

package reports // golex -o lex.go gen/lex.l

import (
	"bufio"
	"fmt"
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
	fmt.Println("Got error in lexer parser. Do not proceed")
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
		goto yyrule27
	case c == '*':
		goto yystate4
	case c == ',':
		goto yystate5
	case c == '.':
		goto yystate6
	case c == ':' || c == 'B' || c == 'E' || c >= 'G' && c <= 'K' || c == 'N' || c >= 'P' && c <= 'R' || c >= 'T' && c <= 'V' || c >= 'X' && c <= 'Z' || c == 'b' || c == 'e' || c >= 'g' && c <= 'k' || c == 'n' || c >= 'p' && c <= 'r' || c >= 't' && c <= 'v' || c >= 'x' && c <= 'z':
		goto yystate23
	case c == '<':
		goto yystate26
	case c == '=':
		goto yystate29
	case c == '>':
		goto yystate30
	case c == 'A':
		goto yystate32
	case c == 'C':
		goto yystate36
	case c == 'D':
		goto yystate65
	case c == 'F':
		goto yystate72
	case c == 'L':
		goto yystate78
	case c == 'M':
		goto yystate89
	case c == 'O':
		goto yystate101
	case c == 'S':
		goto yystate103
	case c == 'W':
		goto yystate113
	case c == '\t' || c == '\r' || c == ' ':
		goto yystate3
	case c == '\x00':
		goto yystate2
	case c == 'a':
		goto yystate121
	case c == 'c':
		goto yystate122
	case c == 'd':
		goto yystate124
	case c == 'f':
		goto yystate125
	case c == 'l':
		goto yystate126
	case c == 'm':
		goto yystate127
	case c == 'o':
		goto yystate128
	case c == 's':
		goto yystate129
	case c == 'w':
		goto yystate130
	case c == '|':
		goto yystate131
	case c >= '0' && c <= '9':
		goto yystate7
	}

yystate2:
	c = y.getc()
	goto yyrule2

yystate3:
	c = y.getc()
	switch {
	default:
		goto yyrule1
	case c == '\t' || c == '\r' || c == ' ':
		goto yystate3
	}

yystate4:
	c = y.getc()
	goto yyrule18

yystate5:
	c = y.getc()
	goto yyrule16

yystate6:
	c = y.getc()
	goto yyrule17

yystate7:
	c = y.getc()
	switch {
	default:
		goto yyrule27
	case c == '-':
		goto yystate8
	case c == ':' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'g' || c >= 'i' && c <= 'l' || c >= 'o' && c <= 'r' || c == 't' || c >= 'v' && c <= 'z':
		goto yystate23
	case c == 'h' || c == 'm' || c == 'n' || c == 's' || c == 'u':
		goto yystate24
	case c == '|':
		goto yystate25
	case c >= '0' && c <= '9':
		goto yystate7
	}

yystate8:
	c = y.getc()
	switch {
	default:
		goto yyabort
	case c >= '0' && c <= '9':
		goto yystate9
	}

yystate9:
	c = y.getc()
	switch {
	default:
		goto yyabort
	case c == '-':
		goto yystate10
	case c >= '0' && c <= '9':
		goto yystate9
	}

yystate10:
	c = y.getc()
	switch {
	default:
		goto yyabort
	case c >= '0' && c <= '9':
		goto yystate11
	}

yystate11:
	c = y.getc()
	switch {
	default:
		goto yyabort
	case c == 'T':
		goto yystate12
	case c >= '0' && c <= '9':
		goto yystate11
	}

yystate12:
	c = y.getc()
	switch {
	default:
		goto yyabort
	case c >= '0' && c <= '9':
		goto yystate13
	}

yystate13:
	c = y.getc()
	switch {
	default:
		goto yyabort
	case c == ':':
		goto yystate14
	case c >= '0' && c <= '9':
		goto yystate13
	}

yystate14:
	c = y.getc()
	switch {
	default:
		goto yyabort
	case c >= '0' && c <= '9':
		goto yystate15
	}

yystate15:
	c = y.getc()
	switch {
	default:
		goto yyabort
	case c == ':':
		goto yystate16
	case c >= '0' && c <= '9':
		goto yystate15
	}

yystate16:
	c = y.getc()
	switch {
	default:
		goto yyabort
	case c >= '0' && c <= '9':
		goto yystate17
	}

yystate17:
	c = y.getc()
	switch {
	default:
		goto yyabort
	case c == '-':
		goto yystate18
	case c == '.':
		goto yystate22
	case c >= '0' && c <= '9':
		goto yystate17
	}

yystate18:
	c = y.getc()
	switch {
	default:
		goto yyabort
	case c >= '0' && c <= '9':
		goto yystate19
	}

yystate19:
	c = y.getc()
	switch {
	default:
		goto yyabort
	case c == ':':
		goto yystate20
	case c >= '0' && c <= '9':
		goto yystate19
	}

yystate20:
	c = y.getc()
	switch {
	default:
		goto yyabort
	case c >= '0' && c <= '9':
		goto yystate21
	}

yystate21:
	c = y.getc()
	switch {
	default:
		goto yyrule25
	case c >= '0' && c <= '9':
		goto yystate21
	}

yystate22:
	c = y.getc()
	switch {
	default:
		goto yyabort
	case c == '-':
		goto yystate18
	case c >= '0' && c <= '9':
		goto yystate22
	}

yystate23:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate23
	}

yystate24:
	c = y.getc()
	switch {
	default:
		goto yyrule26
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate23
	}

yystate25:
	c = y.getc()
	goto yyrule26

yystate26:
	c = y.getc()
	switch {
	default:
		goto yyrule22
	case c == '=':
		goto yystate27
	case c == '>':
		goto yystate28
	}

yystate27:
	c = y.getc()
	goto yyrule19

yystate28:
	c = y.getc()
	goto yyrule21

yystate29:
	c = y.getc()
	goto yyrule24

yystate30:
	c = y.getc()
	switch {
	default:
		goto yyrule23
	case c == '=':
		goto yystate31
	}

yystate31:
	c = y.getc()
	goto yyrule20

yystate32:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'N':
		goto yystate33
	case c == 'n':
		goto yystate35
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'M' || c >= 'O' && c <= 'Z' || c >= 'a' && c <= 'm' || c >= 'o' && c <= 'z':
		goto yystate23
	}

yystate33:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'D':
		goto yystate34
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'C' || c >= 'E' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate23
	}

yystate34:
	c = y.getc()
	switch {
	default:
		goto yyrule10
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate23
	}

yystate35:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'd':
		goto yystate34
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'c' || c >= 'e' && c <= 'z':
		goto yystate23
	}

yystate36:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'M':
		goto yystate37
	case c == 'O':
		goto yystate39
	case c == 'o':
		goto yystate52
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'L' || c == 'N' || c >= 'P' && c <= 'Z' || c >= 'a' && c <= 'n' || c >= 'p' && c <= 'z':
		goto yystate23
	}

yystate37:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'E':
		goto yystate38
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'D' || c >= 'F' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate23
	}

yystate38:
	c = y.getc()
	switch {
	default:
		goto yyrule6
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate23
	}

yystate39:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'M':
		goto yystate40
	case c == 'N':
		goto yystate46
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'L' || c >= 'O' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate23
	}

yystate40:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'M':
		goto yystate41
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'L' || c >= 'N' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate23
	}

yystate41:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'I':
		goto yystate42
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'H' || c >= 'J' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate23
	}

yystate42:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'T':
		goto yystate43
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'S' || c >= 'U' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate23
	}

yystate43:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'I':
		goto yystate44
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'H' || c >= 'J' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate23
	}

yystate44:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'D':
		goto yystate45
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'C' || c >= 'E' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate23
	}

yystate45:
	c = y.getc()
	switch {
	default:
		goto yyrule15
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate23
	}

yystate46:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'T':
		goto yystate47
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'S' || c >= 'U' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate23
	}

yystate47:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'A':
		goto yystate48
	case c >= '0' && c <= ':' || c >= 'B' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate23
	}

yystate48:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'I':
		goto yystate49
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'H' || c >= 'J' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate23
	}

yystate49:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'N':
		goto yystate50
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'M' || c >= 'O' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate23
	}

yystate50:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'S':
		goto yystate51
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'R' || c >= 'T' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate23
	}

yystate51:
	c = y.getc()
	switch {
	default:
		goto yyrule12
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate23
	}

yystate52:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'm':
		goto yystate53
	case c == 'n':
		goto yystate60
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'l' || c >= 'o' && c <= 'z':
		goto yystate23
	}

yystate53:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'm':
		goto yystate54
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'l' || c >= 'n' && c <= 'z':
		goto yystate23
	}

yystate54:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'i':
		goto yystate55
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'h' || c >= 'j' && c <= 'z':
		goto yystate23
	}

yystate55:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 't':
		goto yystate56
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 's' || c >= 'u' && c <= 'z':
		goto yystate23
	}

yystate56:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'I' || c == 'i':
		goto yystate57
	case c == '|':
		goto yystate58
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'H' || c >= 'J' && c <= 'Z' || c >= 'a' && c <= 'h' || c >= 'j' && c <= 'z':
		goto yystate23
	}

yystate57:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'd':
		goto yystate45
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'c' || c >= 'e' && c <= 'z':
		goto yystate23
	}

yystate58:
	c = y.getc()
	switch {
	default:
		goto yyabort
	case c == 'd':
		goto yystate59
	}

yystate59:
	c = y.getc()
	goto yyrule15

yystate60:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 't':
		goto yystate61
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 's' || c >= 'u' && c <= 'z':
		goto yystate23
	}

yystate61:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'a':
		goto yystate62
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'Z' || c >= 'b' && c <= 'z':
		goto yystate23
	}

yystate62:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'i':
		goto yystate63
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'h' || c >= 'j' && c <= 'z':
		goto yystate23
	}

yystate63:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'n':
		goto yystate64
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'm' || c >= 'o' && c <= 'z':
		goto yystate23
	}

yystate64:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 's':
		goto yystate51
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'r' || c >= 't' && c <= 'z':
		goto yystate23
	}

yystate65:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'B':
		goto yystate66
	case c == 'I':
		goto yystate67
	case c == 'i':
		goto yystate70
	case c >= '0' && c <= ':' || c == 'A' || c >= 'C' && c <= 'H' || c >= 'J' && c <= 'Z' || c >= 'a' && c <= 'h' || c >= 'j' && c <= 'z':
		goto yystate23
	}

yystate66:
	c = y.getc()
	switch {
	default:
		goto yyrule7
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate23
	}

yystate67:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'F':
		goto yystate68
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'E' || c >= 'G' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate23
	}

yystate68:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'F':
		goto yystate69
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'E' || c >= 'G' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate23
	}

yystate69:
	c = y.getc()
	switch {
	default:
		goto yyrule14
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate23
	}

yystate70:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'f':
		goto yystate71
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'e' || c >= 'g' && c <= 'z':
		goto yystate23
	}

yystate71:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'f':
		goto yystate69
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'e' || c >= 'g' && c <= 'z':
		goto yystate23
	}

yystate72:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'R':
		goto yystate73
	case c == 'r':
		goto yystate76
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'Q' || c >= 'S' && c <= 'Z' || c >= 'a' && c <= 'q' || c >= 's' && c <= 'z':
		goto yystate23
	}

yystate73:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'O':
		goto yystate74
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'N' || c >= 'P' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate23
	}

yystate74:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'M':
		goto yystate75
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'L' || c >= 'N' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate23
	}

yystate75:
	c = y.getc()
	switch {
	default:
		goto yyrule4
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate23
	}

yystate76:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'o':
		goto yystate77
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'n' || c >= 'p' && c <= 'z':
		goto yystate23
	}

yystate77:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'm':
		goto yystate75
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'l' || c >= 'n' && c <= 'z':
		goto yystate23
	}

yystate78:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'I':
		goto yystate79
	case c == 'O':
		goto yystate83
	case c == 'i':
		goto yystate85
	case c == 'o':
		goto yystate88
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'H' || c >= 'J' && c <= 'N' || c >= 'P' && c <= 'Z' || c >= 'a' && c <= 'h' || c >= 'j' && c <= 'n' || c >= 'p' && c <= 'z':
		goto yystate23
	}

yystate79:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'M':
		goto yystate80
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'L' || c >= 'N' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate23
	}

yystate80:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'I':
		goto yystate81
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'H' || c >= 'J' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate23
	}

yystate81:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'T':
		goto yystate82
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'S' || c >= 'U' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate23
	}

yystate82:
	c = y.getc()
	switch {
	default:
		goto yyrule13
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate23
	}

yystate83:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'G':
		goto yystate84
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'F' || c >= 'H' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate23
	}

yystate84:
	c = y.getc()
	switch {
	default:
		goto yyrule5
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate23
	}

yystate85:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'm':
		goto yystate86
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'l' || c >= 'n' && c <= 'z':
		goto yystate23
	}

yystate86:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'i':
		goto yystate87
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'h' || c >= 'j' && c <= 'z':
		goto yystate23
	}

yystate87:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 't':
		goto yystate82
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 's' || c >= 'u' && c <= 'z':
		goto yystate23
	}

yystate88:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'g':
		goto yystate84
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'f' || c >= 'h' && c <= 'z':
		goto yystate23
	}

yystate89:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'E':
		goto yystate90
	case c == 'e':
		goto yystate96
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'D' || c >= 'F' && c <= 'Z' || c >= 'a' && c <= 'd' || c >= 'f' && c <= 'z':
		goto yystate23
	}

yystate90:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'S':
		goto yystate91
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'R' || c >= 'T' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate23
	}

yystate91:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'S':
		goto yystate92
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'R' || c >= 'T' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate23
	}

yystate92:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'A':
		goto yystate93
	case c >= '0' && c <= ':' || c >= 'B' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate23
	}

yystate93:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'G':
		goto yystate94
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'F' || c >= 'H' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate23
	}

yystate94:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'E':
		goto yystate95
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'D' || c >= 'F' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate23
	}

yystate95:
	c = y.getc()
	switch {
	default:
		goto yyrule8
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate23
	}

yystate96:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 's':
		goto yystate97
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'r' || c >= 't' && c <= 'z':
		goto yystate23
	}

yystate97:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 's':
		goto yystate98
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'r' || c >= 't' && c <= 'z':
		goto yystate23
	}

yystate98:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'a':
		goto yystate99
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'Z' || c >= 'b' && c <= 'z':
		goto yystate23
	}

yystate99:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'g':
		goto yystate100
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'f' || c >= 'h' && c <= 'z':
		goto yystate23
	}

yystate100:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'e':
		goto yystate95
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'd' || c >= 'f' && c <= 'z':
		goto yystate23
	}

yystate101:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'R' || c == 'r':
		goto yystate102
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'Q' || c >= 'S' && c <= 'Z' || c >= 'a' && c <= 'q' || c >= 's' && c <= 'z':
		goto yystate23
	}

yystate102:
	c = y.getc()
	switch {
	default:
		goto yyrule11
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate23
	}

yystate103:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'E':
		goto yystate104
	case c == 'e':
		goto yystate109
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'D' || c >= 'F' && c <= 'Z' || c >= 'a' && c <= 'd' || c >= 'f' && c <= 'z':
		goto yystate23
	}

yystate104:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'L':
		goto yystate105
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'K' || c >= 'M' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate23
	}

yystate105:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'E':
		goto yystate106
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'D' || c >= 'F' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate23
	}

yystate106:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'C':
		goto yystate107
	case c >= '0' && c <= ':' || c == 'A' || c == 'B' || c >= 'D' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate23
	}

yystate107:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'T':
		goto yystate108
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'S' || c >= 'U' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate23
	}

yystate108:
	c = y.getc()
	switch {
	default:
		goto yyrule3
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate23
	}

yystate109:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'l':
		goto yystate110
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'k' || c >= 'm' && c <= 'z':
		goto yystate23
	}

yystate110:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'e':
		goto yystate111
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'd' || c >= 'f' && c <= 'z':
		goto yystate23
	}

yystate111:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'c':
		goto yystate112
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'Z' || c == 'a' || c == 'b' || c >= 'd' && c <= 'z':
		goto yystate23
	}

yystate112:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 't':
		goto yystate108
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 's' || c >= 'u' && c <= 'z':
		goto yystate23
	}

yystate113:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'H':
		goto yystate114
	case c == 'h':
		goto yystate118
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'G' || c >= 'I' && c <= 'Z' || c >= 'a' && c <= 'g' || c >= 'i' && c <= 'z':
		goto yystate23
	}

yystate114:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'E':
		goto yystate115
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'D' || c >= 'F' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate23
	}

yystate115:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'R':
		goto yystate116
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'Q' || c >= 'S' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate23
	}

yystate116:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'E':
		goto yystate117
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'D' || c >= 'F' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate23
	}

yystate117:
	c = y.getc()
	switch {
	default:
		goto yyrule9
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z':
		goto yystate23
	}

yystate118:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'e':
		goto yystate119
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'd' || c >= 'f' && c <= 'z':
		goto yystate23
	}

yystate119:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'r':
		goto yystate120
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'q' || c >= 's' && c <= 'z':
		goto yystate23
	}

yystate120:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'e':
		goto yystate117
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'd' || c >= 'f' && c <= 'z':
		goto yystate23
	}

yystate121:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'n':
		goto yystate35
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'm' || c >= 'o' && c <= 'z':
		goto yystate23
	}

yystate122:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'm':
		goto yystate123
	case c == 'o':
		goto yystate52
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'l' || c == 'n' || c >= 'p' && c <= 'z':
		goto yystate23
	}

yystate123:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'e':
		goto yystate38
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'd' || c >= 'f' && c <= 'z':
		goto yystate23
	}

yystate124:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'b':
		goto yystate66
	case c == 'i':
		goto yystate70
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'Z' || c == 'a' || c >= 'c' && c <= 'h' || c >= 'j' && c <= 'z':
		goto yystate23
	}

yystate125:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'r':
		goto yystate76
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'q' || c >= 's' && c <= 'z':
		goto yystate23
	}

yystate126:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'i':
		goto yystate85
	case c == 'o':
		goto yystate88
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'h' || c >= 'j' && c <= 'n' || c >= 'p' && c <= 'z':
		goto yystate23
	}

yystate127:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'e':
		goto yystate96
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'd' || c >= 'f' && c <= 'z':
		goto yystate23
	}

yystate128:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'r':
		goto yystate102
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'q' || c >= 's' && c <= 'z':
		goto yystate23
	}

yystate129:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'e':
		goto yystate109
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'd' || c >= 'f' && c <= 'z':
		goto yystate23
	}

yystate130:
	c = y.getc()
	switch {
	default:
		goto yyrule28
	case c == 'h':
		goto yystate118
	case c >= '0' && c <= ':' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'g' || c >= 'i' && c <= 'z':
		goto yystate23
	}

yystate131:
	c = y.getc()
	switch {
	default:
		goto yyabort
	case c == 'e':
		goto yystate132
	case c == 'h':
		goto yystate142
	case c == 'i':
		goto yystate146
	case c == 'n':
		goto yystate152
	case c == 'o':
		goto yystate154
	case c == 'r':
		goto yystate166
	}

yystate132:
	c = y.getc()
	switch {
	default:
		goto yyabort
	case c == 'l':
		goto yystate133
	case c == 's':
		goto yystate137
	}

yystate133:
	c = y.getc()
	switch {
	default:
		goto yyabort
	case c == 'e':
		goto yystate134
	}

yystate134:
	c = y.getc()
	switch {
	default:
		goto yyabort
	case c == 'c':
		goto yystate135
	}

yystate135:
	c = y.getc()
	switch {
	default:
		goto yyabort
	case c == 't':
		goto yystate136
	}

yystate136:
	c = y.getc()
	goto yyrule3

yystate137:
	c = y.getc()
	switch {
	default:
		goto yyabort
	case c == 's':
		goto yystate138
	}

yystate138:
	c = y.getc()
	switch {
	default:
		goto yyabort
	case c == 'a':
		goto yystate139
	}

yystate139:
	c = y.getc()
	switch {
	default:
		goto yyabort
	case c == 'g':
		goto yystate140
	}

yystate140:
	c = y.getc()
	switch {
	default:
		goto yyabort
	case c == 'e':
		goto yystate141
	}

yystate141:
	c = y.getc()
	goto yyrule8

yystate142:
	c = y.getc()
	switch {
	default:
		goto yyabort
	case c == 'e':
		goto yystate143
	}

yystate143:
	c = y.getc()
	switch {
	default:
		goto yyabort
	case c == 'r':
		goto yystate144
	}

yystate144:
	c = y.getc()
	switch {
	default:
		goto yyabort
	case c == 'e':
		goto yystate145
	}

yystate145:
	c = y.getc()
	goto yyrule9

yystate146:
	c = y.getc()
	switch {
	default:
		goto yyabort
	case c == 'f':
		goto yystate147
	case c == 'm':
		goto yystate149
	}

yystate147:
	c = y.getc()
	switch {
	default:
		goto yyabort
	case c == 'f':
		goto yystate148
	}

yystate148:
	c = y.getc()
	goto yyrule14

yystate149:
	c = y.getc()
	switch {
	default:
		goto yyabort
	case c == 'i':
		goto yystate150
	}

yystate150:
	c = y.getc()
	switch {
	default:
		goto yyabort
	case c == 't':
		goto yystate151
	}

yystate151:
	c = y.getc()
	goto yyrule13

yystate152:
	c = y.getc()
	switch {
	default:
		goto yyabort
	case c == 'd':
		goto yystate153
	}

yystate153:
	c = y.getc()
	goto yyrule10

yystate154:
	c = y.getc()
	switch {
	default:
		goto yyabort
	case c == 'g':
		goto yystate155
	case c == 'm':
		goto yystate156
	case c == 'n':
		goto yystate160
	}

yystate155:
	c = y.getc()
	goto yyrule5

yystate156:
	c = y.getc()
	switch {
	default:
		goto yyabort
	case c == 'm':
		goto yystate157
	}

yystate157:
	c = y.getc()
	switch {
	default:
		goto yyabort
	case c == 'i':
		goto yystate158
	}

yystate158:
	c = y.getc()
	switch {
	default:
		goto yyabort
	case c == 't':
		goto yystate159
	}

yystate159:
	c = y.getc()
	switch {
	default:
		goto yyabort
	case c == 'I' || c == 'i' || c == '|':
		goto yystate58
	}

yystate160:
	c = y.getc()
	switch {
	default:
		goto yyabort
	case c == 't':
		goto yystate161
	}

yystate161:
	c = y.getc()
	switch {
	default:
		goto yyabort
	case c == 'a':
		goto yystate162
	}

yystate162:
	c = y.getc()
	switch {
	default:
		goto yyabort
	case c == 'i':
		goto yystate163
	}

yystate163:
	c = y.getc()
	switch {
	default:
		goto yyabort
	case c == 'n':
		goto yystate164
	}

yystate164:
	c = y.getc()
	switch {
	default:
		goto yyabort
	case c == 's':
		goto yystate165
	}

yystate165:
	c = y.getc()
	goto yyrule12

yystate166:
	c = y.getc()
	switch {
	default:
		goto yyrule11
	case c == 'o':
		goto yystate167
	}

yystate167:
	c = y.getc()
	switch {
	default:
		goto yyabort
	case c == 'm':
		goto yystate168
	}

yystate168:
	c = y.getc()
	goto yyrule4

yyrule1: // [ \t\r]+

	goto yystate0
yyrule2: // \0
	{
		return 0
	}
yyrule3: // SELECT|[s|S]elect
	{
		return TOKSELECT
	}
yyrule4: // FROM|[f|F]rom
	{
		return TOKFROM
	}
yyrule5: // LOG|[l|L]og
	{
		return TOKLOG
	}
yyrule6: // CME|cme
	{
		return TOKCME
	}
yyrule7: // DB|db
	{
		return TOKDB
	}
yyrule8: // MESSAGE|[M|m]essage
	{
		lval.lit = "Message"
		return TOKMSG
		goto yystate0
	}
yyrule9: // WHERE|[w|W]here
	{
		return WHERE
	}
yyrule10: // AND|[a|A]nd
	{
		lval.lit = string(y.buf)
		return AND
		goto yystate0
	}
yyrule11: // OR|[O|o]r
	{
		lval.lit = string(y.buf)
		return OR
		goto yystate0
	}
yyrule12: // CONTAINS|[C|c]ontains
	{
		lval.lit = "CONTAINS"
		return CONTAINS
		goto yystate0
	}
yyrule13: // LIMIT|[l|L]imit
	{
		return TOKLIMIT
	}
yyrule14: // DIFF|[D|d]iff
	{
		lval.lit = "DIFF"
		return TOKDIFF
		goto yystate0
	}
yyrule15: // COMMITID|[C|c]ommit[I|i]d
	{
		lval.lit = "COMMITID"
		return TOKCOMMITID
		goto yystate0
	}
yyrule16: // ,
	{
		lval.lit = string(y.buf)
		return COMMA
		goto yystate0
	}
yyrule17: // \.
	{
		return DOT
	}
yyrule18: // \*
	{
		lval.lit = string(y.buf)
		return ASTERISK
		goto yystate0
	}
yyrule19: // "<="
yyrule20: // ">="
yyrule21: // "<>"
yyrule22: // "<"
yyrule23: // ">"
	{
		lval.lit = string(y.buf)
		return COMPARISON
		goto yystate0
	}
yyrule24: // "="
	{
		lval.lit = string(y.buf)
		return EQUAL
		goto yystate0
	}
yyrule25: // [0-9]+-[0-9]+-[0-9]+T[0-9]+:[0-9]+:[0-9]+\.?[0-9]*-[0-9]+:[0-9]+
	{
		lval.lit = string(y.buf)
		return TOKTIME
		goto yystate0
	}
yyrule26: // [0-9]+[ns|us|ms|s|m|h]
	{
		lval.lit = string(y.buf)
		return TOKTIME
		goto yystate0
	}
yyrule27: // [0-9]*
	{
		lval.lit = string(y.buf)
		return TOKNUMBR
		goto yystate0
	}
yyrule28: // [a-zA-Z0-9:]+
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
