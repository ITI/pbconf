/***********************************************************************
   Copyright 2018 Information Trust Institute

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
***********************************************************************/

%{
package pbtranslate // golex -o lex.go gen/lex.l

import (
    "bufio"
)

type yylexer struct{
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
%}

%yyc c
%yyn c = y.getc()
%%
    y.buf = y.buf[:0]

[ \t\r\n]+
\0                  return 0

SET|set             return TOKSET
PASSWORD|password   return TOKPASSWORD
SERVICE|service     return TOKSERVICE
ON|on               lval.lit = string(y.buf); return TOKSTATE
OFF|off             lval.lit = string(y.buf); return TOKSTATE
[a-zA-Z0-9]*        lval.lit = string(y.buf); return TOKWORD
%%
    y.empty=true
    return int(c)
}
