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
package reports // golex -o lex.go gen/lex.l

import (
    "bufio"
    "fmt"
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
    fmt.Println("Got error in lexer parser. Do not proceed")
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

[ \t\r]+
\0                                                                  return 0
SELECT|[s|S]elect                                                   return TOKSELECT
FROM|[f|F]rom                                                       return TOKFROM

LOG|[l|L]og                                                         return TOKLOG
CME|cme                                                             return TOKCME
DB|db                                                               return TOKDB
MESSAGE|[M|m]essage                                                 lval.lit = "Message";return TOKMSG

WHERE|[w|W]here                                                     return WHERE
AND|[a|A]nd                                                         lval.lit = string(y.buf); return AND
OR|[O|o]r                                                           lval.lit = string(y.buf); return OR
CONTAINS|[C|c]ontains                                               lval.lit = "CONTAINS"; return CONTAINS
LIMIT|[l|L]imit                                                     return TOKLIMIT
DIFF|[D|d]iff                                                       lval.lit = "DIFF";return TOKDIFF
COMMITID|[C|c]ommit[I|i]d                                           lval.lit = "COMMITID";return TOKCOMMITID

,                                                                   lval.lit = string(y.buf); return COMMA
\.                                                                  return DOT
\*                                                                  lval.lit = string(y.buf); return ASTERISK

"<="  |
">="  |
"<>"  |
"<"   |
">"                                                                 lval.lit = string(y.buf); return COMPARISON

"="                                                                 lval.lit = string(y.buf); return EQUAL

[0-9]+-[0-9]+-[0-9]+T[0-9]+:[0-9]+:[0-9]+\.?[0-9]*-[0-9]+:[0-9]+    lval.lit = string(y.buf); return TOKTIME
[0-9]+[ns|us|ms|s|m|h]                                              lval.lit = string(y.buf); return TOKTIME
[0-9]*                                                              lval.lit = string(y.buf); return TOKNUMBR
[a-zA-Z0-9:]+                                                        lval.lit = string(y.buf); return TOKWORD
%%
    y.empty=true
    return int(c)
}
