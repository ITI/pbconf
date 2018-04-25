//line gen/parse.yy:2

// Do Not Edit:  go tool yacc -o parse.go gen/parse.yy
package reports

import __yyfmt__ "fmt"

//line gen/parse.yy:3
import (
	"bufio"
	"errors"
	"io"
	//"fmt"
)

type op struct {
	op        string
	fieldlist string
	source    string
	limit     string
	clauses   cmpd_clause
}
type relational_cond struct {
	op1      string
	operator string
	op2      string
}
type cmpd_clause struct {
	clause       []relational_cond
	logical_cond []string
}

var ast []op

func init() {
}

func Parse(s io.Reader) ([]op, error) {
	ast = make([]op, 0)
	retVal := yyParse(NewLexer(bufio.NewReader(s)))
	if retVal == 1 {
		return ast, errors.New("Error parsing the input. Cannot proceed with report")
	}
	return ast, nil
}

//line gen/parse.yy:42
type yySymType struct {
	yys int
	lit string
	cmp cmpd_clause
}

const TOKSELECT = 57346
const TOKFROM = 57347
const TOKDB = 57348
const TOKLOG = 57349
const TOKCME = 57350
const WHERE = 57351
const AND = 57352
const OR = 57353
const CONTAINS = 57354
const TOKLIMIT = 57355
const COMMA = 57356
const DOT = 57357
const ASTERISK = 57358
const COMPARISON = 57359
const EQUAL = 57360
const TOKWORD = 57361
const TOKNUMBR = 57362
const TOKTIME = 57363
const TOKMSG = 57364
const TOKDIFF = 57365
const TOKCOMMITID = 57366

var yyToknames = [...]string{
	"$end",
	"error",
	"$unk",
	"TOKSELECT",
	"TOKFROM",
	"TOKDB",
	"TOKLOG",
	"TOKCME",
	"WHERE",
	"AND",
	"OR",
	"CONTAINS",
	"TOKLIMIT",
	"COMMA",
	"DOT",
	"ASTERISK",
	"COMPARISON",
	"EQUAL",
	"TOKWORD",
	"TOKNUMBR",
	"TOKTIME",
	"TOKMSG",
	"TOKDIFF",
	"TOKCOMMITID",
}
var yyStatenames = [...]string{}

const yyEofCode = 1
const yyErrCode = 2
const yyInitialStackSize = 16

//line gen/parse.yy:247

//line yacctab:1
var yyExca = [...]int{
	-1, 1,
	1, -1,
	-2, 0,
}

const yyNprod = 33
const yyPrivate = 57344

var yyTokenNames []string
var yyStates []string

const yyLast = 75

var yyAct = [...]int{

	56, 46, 57, 35, 8, 60, 41, 10, 66, 69,
	11, 7, 55, 59, 58, 37, 19, 64, 36, 20,
	51, 34, 70, 65, 65, 48, 47, 49, 61, 55,
	53, 52, 37, 28, 26, 38, 39, 25, 50, 43,
	44, 24, 42, 23, 45, 22, 21, 30, 54, 14,
	33, 29, 43, 44, 62, 63, 16, 17, 15, 5,
	32, 31, 18, 67, 68, 27, 13, 12, 9, 40,
	6, 4, 3, 2, 1,
}
var yyPact = [...]int{

	55, -1000, -1000, -1000, -1000, -12, 62, 61, -1000, 35,
	-1000, -1000, 50, 54, -3, 31, 30, 28, 26, -1000,
	-1000, 18, 15, 58, 14, 38, 52, 51, 41, 1,
	-4, 13, 13, -18, -1000, 29, 32, 8, 42, 42,
	17, 20, 0, 12, 11, 10, -7, -1000, -1000, -19,
	9, -1000, 8, 8, 4, -1000, -1000, 5, -1000, -1000,
	-10, -1000, -7, -7, -11, -1000, 3, -1000, -1000, -1000,
	-1000,
}
var yyPgo = [...]int{

	0, 74, 73, 72, 71, 70, 3, 2, 69, 1,
	0, 68,
}
var yyR1 = [...]int{

	0, 1, 1, 1, 3, 3, 3, 3, 3, 3,
	3, 4, 4, 2, 2, 8, 8, 6, 6, 6,
	9, 9, 10, 10, 10, 7, 7, 5, 5, 11,
	11, 11, 11,
}
var yyR2 = [...]int{

	0, 1, 1, 1, 6, 8, 8, 10, 10, 12,
	8, 6, 8, 6, 8, 3, 5, 3, 5, 5,
	1, 1, 1, 1, 1, 1, 2, 1, 1, 1,
	1, 3, 3,
}
var yyChk = [...]int{

	-1000, -1, -2, -3, -4, 4, -5, 23, 16, -11,
	19, 22, 5, 5, 14, 8, 6, 7, 8, 19,
	22, 15, 15, 15, 15, 19, 19, 7, 19, 13,
	9, 9, 9, 9, 20, -6, 22, 19, -6, -6,
	-8, 24, 13, 10, 11, 12, -9, 18, 17, 10,
	18, 20, 19, 19, -7, 19, -10, -7, 21, 20,
	24, 19, -9, -9, 13, 19, 18, -10, -10, 20,
	19,
}
var yyDef = [...]int{

	0, -2, 1, 2, 3, 0, 0, 0, 27, 28,
	29, 30, 0, 0, 0, 0, 0, 0, 0, 31,
	32, 0, 0, 0, 0, 4, 11, 13, 0, 0,
	0, 0, 0, 0, 5, 6, 0, 0, 12, 14,
	10, 0, 0, 0, 0, 0, 0, 20, 21, 0,
	0, 7, 0, 0, 8, 25, 17, 22, 23, 24,
	0, 15, 0, 0, 0, 26, 0, 18, 19, 9,
	16,
}
var yyTok1 = [...]int{

	1,
}
var yyTok2 = [...]int{

	2, 3, 4, 5, 6, 7, 8, 9, 10, 11,
	12, 13, 14, 15, 16, 17, 18, 19, 20, 21,
	22, 23, 24,
}
var yyTok3 = [...]int{
	0,
}

var yyErrorMessages = [...]struct {
	state int
	token int
	msg   string
}{}

//line yaccpar:1

/*	parser for yacc output	*/

var (
	yyDebug        = 0
	yyErrorVerbose = false
)

type yyLexer interface {
	Lex(lval *yySymType) int
	Error(s string)
}

type yyParser interface {
	Parse(yyLexer) int
	Lookahead() int
}

type yyParserImpl struct {
	lval  yySymType
	stack [yyInitialStackSize]yySymType
	char  int
}

func (p *yyParserImpl) Lookahead() int {
	return p.char
}

func yyNewParser() yyParser {
	return &yyParserImpl{}
}

const yyFlag = -1000

func yyTokname(c int) string {
	if c >= 1 && c-1 < len(yyToknames) {
		if yyToknames[c-1] != "" {
			return yyToknames[c-1]
		}
	}
	return __yyfmt__.Sprintf("tok-%v", c)
}

func yyStatname(s int) string {
	if s >= 0 && s < len(yyStatenames) {
		if yyStatenames[s] != "" {
			return yyStatenames[s]
		}
	}
	return __yyfmt__.Sprintf("state-%v", s)
}

func yyErrorMessage(state, lookAhead int) string {
	const TOKSTART = 4

	if !yyErrorVerbose {
		return "syntax error"
	}

	for _, e := range yyErrorMessages {
		if e.state == state && e.token == lookAhead {
			return "syntax error: " + e.msg
		}
	}

	res := "syntax error: unexpected " + yyTokname(lookAhead)

	// To match Bison, suggest at most four expected tokens.
	expected := make([]int, 0, 4)

	// Look for shiftable tokens.
	base := yyPact[state]
	for tok := TOKSTART; tok-1 < len(yyToknames); tok++ {
		if n := base + tok; n >= 0 && n < yyLast && yyChk[yyAct[n]] == tok {
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}
	}

	if yyDef[state] == -2 {
		i := 0
		for yyExca[i] != -1 || yyExca[i+1] != state {
			i += 2
		}

		// Look for tokens that we accept or reduce.
		for i += 2; yyExca[i] >= 0; i += 2 {
			tok := yyExca[i]
			if tok < TOKSTART || yyExca[i+1] == 0 {
				continue
			}
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}

		// If the default action is to accept or reduce, give up.
		if yyExca[i+1] != 0 {
			return res
		}
	}

	for i, tok := range expected {
		if i == 0 {
			res += ", expecting "
		} else {
			res += " or "
		}
		res += yyTokname(tok)
	}
	return res
}

func yylex1(lex yyLexer, lval *yySymType) (char, token int) {
	token = 0
	char = lex.Lex(lval)
	if char <= 0 {
		token = yyTok1[0]
		goto out
	}
	if char < len(yyTok1) {
		token = yyTok1[char]
		goto out
	}
	if char >= yyPrivate {
		if char < yyPrivate+len(yyTok2) {
			token = yyTok2[char-yyPrivate]
			goto out
		}
	}
	for i := 0; i < len(yyTok3); i += 2 {
		token = yyTok3[i+0]
		if token == char {
			token = yyTok3[i+1]
			goto out
		}
	}

out:
	if token == 0 {
		token = yyTok2[1] /* unknown char */
	}
	if yyDebug >= 3 {
		__yyfmt__.Printf("lex %s(%d)\n", yyTokname(token), uint(char))
	}
	return char, token
}

func yyParse(yylex yyLexer) int {
	return yyNewParser().Parse(yylex)
}

func (yyrcvr *yyParserImpl) Parse(yylex yyLexer) int {
	var yyn int
	var yyVAL yySymType
	var yyDollar []yySymType
	_ = yyDollar // silence set and not used
	yyS := yyrcvr.stack[:]

	Nerrs := 0   /* number of errors */
	Errflag := 0 /* error recovery flag */
	yystate := 0
	yyrcvr.char = -1
	yytoken := -1 // yyrcvr.char translated into internal numbering
	defer func() {
		// Make sure we report no lookahead when not parsing.
		yystate = -1
		yyrcvr.char = -1
		yytoken = -1
	}()
	yyp := -1
	goto yystack

ret0:
	return 0

ret1:
	return 1

yystack:
	/* put a state and value onto the stack */
	if yyDebug >= 4 {
		__yyfmt__.Printf("char %v in %v\n", yyTokname(yytoken), yyStatname(yystate))
	}

	yyp++
	if yyp >= len(yyS) {
		nyys := make([]yySymType, len(yyS)*2)
		copy(nyys, yyS)
		yyS = nyys
	}
	yyS[yyp] = yyVAL
	yyS[yyp].yys = yystate

yynewstate:
	yyn = yyPact[yystate]
	if yyn <= yyFlag {
		goto yydefault /* simple state */
	}
	if yyrcvr.char < 0 {
		yyrcvr.char, yytoken = yylex1(yylex, &yyrcvr.lval)
	}
	yyn += yytoken
	if yyn < 0 || yyn >= yyLast {
		goto yydefault
	}
	yyn = yyAct[yyn]
	if yyChk[yyn] == yytoken { /* valid shift */
		yyrcvr.char = -1
		yytoken = -1
		yyVAL = yyrcvr.lval
		yystate = yyn
		if Errflag > 0 {
			Errflag--
		}
		goto yystack
	}

yydefault:
	/* default state action */
	yyn = yyDef[yystate]
	if yyn == -2 {
		if yyrcvr.char < 0 {
			yyrcvr.char, yytoken = yylex1(yylex, &yyrcvr.lval)
		}

		/* look through exception table */
		xi := 0
		for {
			if yyExca[xi+0] == -1 && yyExca[xi+1] == yystate {
				break
			}
			xi += 2
		}
		for xi += 2; ; xi += 2 {
			yyn = yyExca[xi+0]
			if yyn < 0 || yyn == yytoken {
				break
			}
		}
		yyn = yyExca[xi+1]
		if yyn < 0 {
			goto ret0
		}
	}
	if yyn == 0 {
		/* error ... attempt to resume parsing */
		switch Errflag {
		case 0: /* brand new error */
			yylex.Error(yyErrorMessage(yystate, yytoken))
			Nerrs++
			if yyDebug >= 1 {
				__yyfmt__.Printf("%s", yyStatname(yystate))
				__yyfmt__.Printf(" saw %s\n", yyTokname(yytoken))
			}
			fallthrough

		case 1, 2: /* incompletely recovered error ... try again */
			Errflag = 3

			/* find a state where "error" is a legal shift action */
			for yyp >= 0 {
				yyn = yyPact[yyS[yyp].yys] + yyErrCode
				if yyn >= 0 && yyn < yyLast {
					yystate = yyAct[yyn] /* simulate a shift of "error" */
					if yyChk[yystate] == yyErrCode {
						goto yystack
					}
				}

				/* the current p has no shift on "error", pop stack */
				if yyDebug >= 2 {
					__yyfmt__.Printf("error recovery pops state %d\n", yyS[yyp].yys)
				}
				yyp--
			}
			/* there is no state on the stack with an error shift ... abort */
			goto ret1

		case 3: /* no shift yet; clobber input char */
			if yyDebug >= 2 {
				__yyfmt__.Printf("error recovery discards %s\n", yyTokname(yytoken))
			}
			if yytoken == yyEofCode {
				goto ret1
			}
			yyrcvr.char = -1
			yytoken = -1
			goto yynewstate /* try again in the same state */
		}
	}

	/* reduction by production yyn */
	if yyDebug >= 2 {
		__yyfmt__.Printf("reduce %v in:\n\t%v\n", yyn, yyStatname(yystate))
	}

	yynt := yyn
	yypt := yyp
	_ = yypt // guard against "declared and not used"

	yyp -= yyR2[yyn]
	// yyp is now the index of $0. Perform the default action. Iff the
	// reduced production is ε, $1 is possibly out of range.
	if yyp+1 >= len(yyS) {
		nyys := make([]yySymType, len(yyS)*2)
		copy(nyys, yyS)
		yyS = nyys
	}
	yyVAL = yyS[yyp+1]

	/* consult goto table to find next state */
	yyn = yyR1[yyn]
	yyg := yyPgo[yyn]
	yyj := yyg + yyS[yyp].yys + 1

	if yyj >= yyLast {
		yystate = yyAct[yyg]
	} else {
		yystate = yyAct[yyj]
		if yyChk[yystate] != -yyn {
			yystate = yyAct[yyg]
		}
	}
	// dummy call; replaced with literal code
	switch yynt {

	case 4:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line gen/parse.yy:82
		{
			ast = append(ast, op{
				op:        "selectcme",
				fieldlist: yyDollar[2].lit,
				source:    yyDollar[6].lit,
			})
		}
	case 5:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line gen/parse.yy:90
		{
			ast = append(ast, op{
				op:        "selectcme",
				fieldlist: yyDollar[2].lit,
				source:    yyDollar[6].lit,
				limit:     yyDollar[8].lit,
			})
		}
	case 6:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line gen/parse.yy:99
		{
			ast = append(ast, op{
				op:        "selectcme",
				fieldlist: yyDollar[2].lit,
				source:    yyDollar[6].lit,
				clauses:   yyDollar[8].cmp,
			})
		}
	case 7:
		yyDollar = yyS[yypt-10 : yypt+1]
		//line gen/parse.yy:108
		{
			ast = append(ast, op{
				op:        "selectcme",
				fieldlist: yyDollar[2].lit,
				source:    yyDollar[6].lit,
				clauses:   yyDollar[8].cmp,
				limit:     yyDollar[10].lit,
			})
		}
	case 8:
		yyDollar = yyS[yypt-10 : yypt+1]
		//line gen/parse.yy:118
		{
			clause := []relational_cond{relational_cond{op1: yyDollar[8].lit, operator: yyDollar[9].lit, op2: yyDollar[10].lit}}
			cmp_cs := cmpd_clause{clause, []string{}}
			ast = append(ast, op{
				op:        "selectcme",
				fieldlist: yyDollar[2].lit,
				source:    yyDollar[6].lit,
				clauses:   cmp_cs,
			})
		}
	case 9:
		yyDollar = yyS[yypt-12 : yypt+1]
		//line gen/parse.yy:129
		{
			clause := []relational_cond{relational_cond{op1: yyDollar[8].lit, operator: yyDollar[9].lit, op2: yyDollar[10].lit}}
			cmp_cs := cmpd_clause{clause, []string{}}
			ast = append(ast, op{
				op:        "selectcme",
				fieldlist: yyDollar[2].lit,
				source:    yyDollar[6].lit,
				clauses:   cmp_cs,
				limit:     yyDollar[12].lit,
			})
		}
	case 10:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line gen/parse.yy:141
		{
			ast = append(ast, op{
				op:        "selectcme",
				fieldlist: yyDollar[2].lit,
				source:    yyDollar[6].lit,
				clauses:   yyDollar[8].cmp,
			})
		}
	case 11:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line gen/parse.yy:152
		{
			ast = append(ast, op{
				op:        "selectdb",
				fieldlist: yyDollar[2].lit,
				source:    yyDollar[6].lit,
			})
		}
	case 12:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line gen/parse.yy:160
		{
			ast = append(ast, op{
				op:        "selectdb",
				fieldlist: yyDollar[2].lit,
				source:    yyDollar[6].lit,
				clauses:   yyDollar[8].cmp,
			})
		}
	case 13:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line gen/parse.yy:171
		{
			ast = append(ast, op{
				op:        "selectlog",
				fieldlist: yyDollar[2].lit,
			})
		}
	case 14:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line gen/parse.yy:178
		{
			ast = append(ast, op{
				op:        "selectlog",
				fieldlist: yyDollar[2].lit,
				clauses:   yyDollar[8].cmp,
			})
		}
	case 15:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line gen/parse.yy:187
		{
			clause := []relational_cond{relational_cond{op1: yyDollar[1].lit, operator: yyDollar[2].lit, op2: yyDollar[3].lit}}
			yyVAL.cmp = cmpd_clause{clause, []string{}}
		}
	case 16:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line gen/parse.yy:192
		{
			clauses := append(yyDollar[1].cmp.clause, []relational_cond{relational_cond{op1: yyDollar[3].lit, operator: yyDollar[4].lit, op2: yyDollar[5].lit}}...)
			conditionals := append(yyDollar[1].cmp.logical_cond, yyDollar[2].lit)
			yyVAL.cmp = cmpd_clause{clauses, conditionals}
		}
	case 17:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line gen/parse.yy:199
		{
			clause := []relational_cond{relational_cond{op1: yyDollar[1].lit, operator: yyDollar[2].lit, op2: yyDollar[3].lit}}
			yyVAL.cmp = cmpd_clause{clause, []string{}}
		}
	case 18:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line gen/parse.yy:204
		{
			clauses := append(yyDollar[1].cmp.clause, []relational_cond{relational_cond{op1: yyDollar[3].lit, operator: yyDollar[4].lit, op2: yyDollar[5].lit}}...)
			conditionals := append(yyDollar[1].cmp.logical_cond, yyDollar[2].lit)
			yyVAL.cmp = cmpd_clause{clauses, conditionals}
		}
	case 19:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line gen/parse.yy:210
		{
			clauses := append(yyDollar[1].cmp.clause, []relational_cond{relational_cond{op1: yyDollar[3].lit, operator: yyDollar[4].lit, op2: yyDollar[5].lit}}...)
			conditionals := append(yyDollar[1].cmp.logical_cond, yyDollar[2].lit)
			yyVAL.cmp = cmpd_clause{clauses, conditionals}
		}
	case 20:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line gen/parse.yy:217
		{
			yyVAL.lit = yyDollar[1].lit
		}
	case 21:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line gen/parse.yy:219
		{
			yyVAL.lit = yyDollar[1].lit
		}
	case 22:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line gen/parse.yy:222
		{
			yyVAL.lit = yyDollar[1].lit
		}
	case 23:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line gen/parse.yy:224
		{
			yyVAL.lit = yyDollar[1].lit
		}
	case 24:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line gen/parse.yy:226
		{
			yyVAL.lit = yyDollar[1].lit
		}
	case 25:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line gen/parse.yy:229
		{
			yyVAL.lit = yyDollar[1].lit
		}
	case 26:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line gen/parse.yy:231
		{
			yyVAL.lit = yyDollar[1].lit + " " + yyDollar[2].lit
		}
	case 27:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line gen/parse.yy:234
		{
			yyVAL.lit = yyDollar[1].lit
		}
	case 28:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line gen/parse.yy:236
		{
			yyVAL.lit = yyDollar[1].lit
		}
	case 29:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line gen/parse.yy:239
		{
			yyVAL.lit = yyDollar[1].lit
		}
	case 30:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line gen/parse.yy:241
		{
			yyVAL.lit = yyDollar[1].lit
		}
	case 31:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line gen/parse.yy:243
		{
			yyVAL.lit = yyDollar[1].lit + yyDollar[2].lit + yyDollar[3].lit
		}
	case 32:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line gen/parse.yy:245
		{
			yyVAL.lit = yyDollar[1].lit + yyDollar[2].lit + yyDollar[3].lit
		}
	}
	goto yystack /* stack new state and value */
}
