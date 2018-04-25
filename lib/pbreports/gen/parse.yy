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
package reports
import  (
    "io"
    "bufio"
    "errors"
)

type op struct {
    op string
    fieldlist string
    source string
    limit string
    clauses cmpd_clause
}
type relational_cond struct {
    op1 string
    operator string
    op2 string
}
type cmpd_clause struct {
    clause []relational_cond
    logical_cond []string
}
var ast []op

func init() {
}

func Parse(s io.Reader)  ([]op, error) {
    ast = make([]op, 0)
    retVal := yyParse(NewLexer(bufio.NewReader(s)))
    if retVal == 1 {
    return ast, errors.New("Error parsing the input. Cannot proceed with report")
    }
    return ast, nil
}
%}

%union {
    lit string
    cmp cmpd_clause
}

%token TOKSELECT
%token TOKFROM

%token TOKDB
%token TOKLOG
%token TOKCME

%token WHERE
%token AND
%token OR
%token CONTAINS
%token TOKLIMIT

%token COMMA
%token DOT
%token ASTERISK
%token COMPARISON
%token EQUAL

%token TOKWORD
%token TOKNUMBR
%token TOKTIME
%token TOKMSG
%token TOKDIFF
%token TOKCOMMITID
%%
command:
    select_log
    |
    select_cme
    |
    select_db
    ;

select_cme:
    TOKSELECT select_fields TOKFROM TOKCME DOT TOKWORD {
        ast = append(ast, op{
                    op: "selectcme",
                    fieldlist: $2.lit,
                    source: $6.lit,
        })
    }
    |
    TOKSELECT select_fields TOKFROM TOKCME DOT TOKWORD TOKLIMIT TOKNUMBR {
        ast = append(ast, op{
                    op: "selectcme",
                    fieldlist: $2.lit,
                    source: $6.lit,
                    limit: $8.lit,
        })
    }
    |
    TOKSELECT select_fields TOKFROM TOKCME DOT TOKWORD WHERE search_condition{
        ast = append(ast, op{
                    op: "selectcme",
                    fieldlist: $2.lit,
                    source: $6.lit,
                    clauses: $8.cmp,
        })
    }
    |
    TOKSELECT select_fields TOKFROM TOKCME DOT TOKWORD WHERE search_condition TOKLIMIT TOKNUMBR {
        ast = append(ast, op{
                    op: "selectcme",
                    fieldlist: $2.lit,
                    source: $6.lit,
                    clauses: $8.cmp,
                    limit: $10.lit,
        })
    }
    |
    TOKSELECT select_fields TOKFROM TOKCME DOT TOKWORD WHERE TOKMSG CONTAINS multi_word {
        clause := []relational_cond{relational_cond{op1:$8.lit, operator:$9.lit, op2:$10.lit}}
        cmp_cs := cmpd_clause {clause, []string{}}
        ast = append(ast, op{
                    op:"selectcme",
                    fieldlist: $2.lit,
                    source: $6.lit,
                    clauses: cmp_cs,
        })
    }
    |
    TOKSELECT select_fields TOKFROM TOKCME DOT TOKWORD WHERE TOKMSG CONTAINS multi_word TOKLIMIT TOKNUMBR{
        clause := []relational_cond{relational_cond{op1:$8.lit, operator:$9.lit, op2:$10.lit}}
        cmp_cs := cmpd_clause {clause, []string{}}
        ast = append(ast, op{
                    op:"selectcme",
                    fieldlist: $2.lit,
                    source: $6.lit,
                    clauses: cmp_cs,
                    limit: $12.lit,
        })
    }
    |
    TOKSELECT TOKDIFF TOKFROM TOKCME DOT TOKWORD WHERE diff_condition{
        ast = append(ast, op{
                   op: "selectcme",
                   fieldlist: $2.lit,
                   source: $6.lit,
                   clauses: $8.cmp,
        })
    }
    ;

select_db:
    TOKSELECT select_fields TOKFROM TOKDB DOT TOKWORD {
        ast = append(ast, op{
                    op: "selectdb",
                    fieldlist: $2.lit,
                    source: $6.lit,
        })
    }
    |
    TOKSELECT select_fields TOKFROM TOKDB DOT TOKWORD WHERE search_condition {
        ast = append(ast, op{
                   op: "selectdb",
                   fieldlist: $2.lit,
                   source: $6.lit,
                   clauses: $8.cmp,
        })
    }
    ;

select_log:
    TOKSELECT select_fields TOKFROM TOKLOG DOT TOKLOG {
        ast = append(ast, op{
                    op: "selectlog",
                    fieldlist: $2.lit,
        })
    }
    |
    TOKSELECT select_fields TOKFROM TOKLOG DOT TOKLOG WHERE search_condition {
        ast = append(ast, op{
                    op: "selectlog",
                    fieldlist: $2.lit,
                    clauses: $8.cmp,
        })
    }
    ;
diff_condition:
    TOKCOMMITID EQUAL TOKWORD {
        clause := []relational_cond{relational_cond{op1:$1.lit, operator:$2.lit, op2:$3.lit}}
        $$.cmp = cmpd_clause {clause, []string{}}
    }
    |
    diff_condition AND TOKCOMMITID EQUAL TOKWORD {
        clauses := append($1.cmp.clause, []relational_cond{relational_cond{op1:$3.lit, operator:$4.lit, op2:$5.lit}}...)
        conditionals := append($1.cmp.logical_cond, $2.lit)
        $$.cmp = cmpd_clause{clauses, conditionals}
    }
    ;
search_condition:
    TOKWORD comparison_op second_operand {
        clause := []relational_cond{relational_cond{op1:$1.lit, operator:$2.lit, op2:$3.lit}}
        $$.cmp = cmpd_clause {clause, []string{}}
    }
    |
    search_condition AND TOKWORD comparison_op second_operand {
        clauses := append($1.cmp.clause, []relational_cond{relational_cond{op1:$3.lit, operator:$4.lit, op2:$5.lit}}...)
        conditionals := append($1.cmp.logical_cond, $2.lit)
        $$.cmp = cmpd_clause{clauses, conditionals}
    }
    |
    search_condition OR TOKWORD comparison_op second_operand {
        clauses := append($1.cmp.clause, []relational_cond{relational_cond{op1:$3.lit, operator:$4.lit, op2:$5.lit}}...)
        conditionals := append($1.cmp.logical_cond, $2.lit)
        $$.cmp = cmpd_clause{clauses, conditionals}
    }
    ;
comparison_op:
    EQUAL {$$.lit = $1.lit}
    |
    COMPARISON {$$.lit = $1.lit}
    ;
second_operand:
    multi_word {$$.lit = $1.lit}
    |
    TOKTIME {$$.lit = $1.lit}
    |
    TOKNUMBR {$$.lit = $1.lit}
    ;
multi_word:
    TOKWORD {$$.lit = $1.lit}
    |
    multi_word TOKWORD {$$.lit = $1.lit + " " + $2.lit}
    ;
select_fields:
    ASTERISK {$$.lit = $1.lit;}
    |
    select_expr {$$.lit = $1.lit;}
    ;
select_expr:
    TOKWORD { $$.lit = $1.lit;}
    |
    TOKMSG { $$.lit = $1.lit;}
    |
    select_expr COMMA TOKWORD {$$.lit = $1.lit + $2.lit + $3.lit}
    |
    select_expr COMMA TOKMSG {$$.lit = $1.lit + $2.lit + $3.lit}
    ;
%%
