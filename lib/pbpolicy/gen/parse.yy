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
package policy
import  (
    "io"
    "bufio"
    "errors"
    "fmt"
)

type axiom struct {
    Subject string `json:"s"`
    Predicate string `json:"p"`
    Object string `json:"o"`
}

type pol struct {
    Class string
    Axioms []axiom
}

func NewPolicy(name string) *pol {
    r := &pol{Class: name}
    r.Axioms = make([]axiom, 0)
    return r
}

var ast []pol

func Parse(s io.Reader) ([]pol, error) {
    ast = make([]pol, 0)
    l := NewLexer(bufio.NewReader(s))
    retVal := yyParse(l)
    if retVal == 1 {
        return ast, errors.New("Got error parsing the input, need to abort")
    }
    return ast, nil
}

var CUR_CLASS *pol

%}

%union {
    lit string
}


%token TOKCOMMENT
%token TOKLBRACE
%token TOKRBRACE
%token TOKREQUIRE
%token TOKWORD

%%


statements: /* empty */
          | statements class
          ;

class:
     class_start axioms TOKRBRACE
      {
        ast = append(ast, *CUR_CLASS)
        CUR_CLASS = nil
      }
      ;

class_start:
            TOKWORD TOKLBRACE
            {
              if CUR_CLASS != nil {
                fmt.Println("There's an error here")
              } else {
                CUR_CLASS = NewPolicy($1.lit)
              }
            }
            ;

axioms:
      axioms axiom
      | axiom
      ;

axiom:
     std_axiom
     | require_axiom
     ;

std_axiom:
      TOKWORD TOKWORD TOKWORD
      {
        if CUR_CLASS == nil {
            fmt.Println("Error Here")
        } else {
            CUR_CLASS.Axioms = append(CUR_CLASS.Axioms, axiom{
                Subject: $1.lit,
                Predicate: $2.lit,
                Object: $3.lit,
            })
        }
      }
      ;

require_axiom:
      TOKREQUIRE TOKWORD
      {
        if CUR_CLASS == nil {
            fmt.Println("Error Here")
        } else {
            CUR_CLASS.Axioms = append(CUR_CLASS.Axioms, axiom{
                Subject: CUR_CLASS.Class,
                Predicate: "requires",
                Object: $2.lit,
            })
        }
      }
      ;

%%

