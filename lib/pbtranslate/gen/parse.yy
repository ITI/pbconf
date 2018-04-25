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
// Do Not Edit:  go tool yacc -o parse.go gen/parse.yy
package pbtranslate
import  (
    "io"
    "bufio"
    "errors"
)

type op struct {
    Op string
    Key string
    Val string
    Svc string
}

var ast []op

func Parse(s io.Reader) ([]op, error) {
    ast = make([]op, 0)
    retVal := yyParse(NewLexer(bufio.NewReader(s)))
    if retVal == 1 {
        return ast, errors.New("Got error parsing the input, need to abort")
    }
    return ast, nil
}
%}

%union {
    lit string
}

%token TOKSET
%token TOKPASSWORD
%token TOKSERVICE
%token TOKSTATE
%token TOKWORD
%%
commands: /* empty */
        | commands command
        ;

command:
       set_service
       |
       set_password
       |
       set_var
       |
       svc_config
       ;

set_service:
        TOKSET TOKSERVICE TOKWORD TOKSTATE {
                ast = append(ast, op{
                    Op: "service",
                    Key: $3.lit,
                    Val: $4.lit,
                })
           }
           ;

set_password:
        TOKSET TOKPASSWORD TOKWORD TOKWORD {
                ast = append(ast, op{
                    Op: "password",
                    Key: $3.lit,
                    Val: $4.lit,
                })
            }
            ;

set_var:
        TOKSET TOKWORD TOKWORD {
                ast = append(ast, op{
                    Op: "variable",
                    Key: $2.lit,
                    Val: $3.lit,
                })
       }
       ;

svc_config:
        TOKSERVICE TOKWORD TOKWORD arg_list {
                ast = append(ast, op{
                    Op: "service_option",
                    Key: $3.lit,
                    Val: $4.lit,
                    Svc: $2.lit,
                 })
        }
        ;

arg_list: /* empty */
        |
        TOKWORD arg_list
        ;
%%
