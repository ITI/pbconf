package policy_test

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

import ()

type DSLCase struct {
	I int    // Index
	C string // Input
	O string // Expected Result
}

var Cases = []DSLCase{
	{1, // Case 1
		`SEL421 {
	password.level2 min-length 4
	password.level2 max-length 16
	password.level2 complexity MIXEDCASE
	requires password.level2
}`,

		// Result
		`[{"Class":"SEL421","Axioms":[{"s":"password.level2","p":"min-length","o":"4"},{"s":"password.level2","p":"max-length","o":"16"},{"s":"password.level2","p":"complexity","o":"MIXEDCASE"},{"s":"SEL421","p":"requires","o":"password.level2"}]}]`,
	},

	{2, // Case 2
		`SEL421 { password.level2 min-length 4 password.level2 max-length 16 password.level2 complexity MIXEDCASE requires password.level2 }`,

		// Result
		`[{"Class":"SEL421","Axioms":[{"s":"password.level2","p":"min-length","o":"4"},{"s":"password.level2","p":"max-length","o":"16"},{"s":"password.level2","p":"complexity","o":"MIXEDCASE"},{"s":"SEL421","p":"requires","o":"password.level2"}]}]`,
	},

	{3, // Case 3
		`SEL421 {
	password.level2 min-length 4
}`,

		// Result
		`[{"Class":"SEL421","Axioms":[{"s":"password.level2","p":"min-length","o":"4"}]}]`,
	},

	{4, // Case 4
		`SEL421 { password.level2 min-length 4 }`,

		// Result
		`[{"Class":"SEL421","Axioms":[{"s":"password.level2","p":"min-length","o":"4"}]}]`,
	},

	{5, // Case 5
		``,

		// Result
		`[]`,
	},

	{6, // Case 6
		`SEL421 { one two three }`,

		// Result
		`[{"Class":"SEL421","Axioms":[{"s":"one","p":"two","o":"three"}]}]`,
	},

	{7, // Case 7
		`SEL421 { password.level2 min-length 4 }`,

		// Result
		`[{"Class":"SEL421","Axioms":[{"s":"password.level2","p":"min-length","o":"4"}]}]`,
	},

	{8, // Case 8
		`SEL421 {
	password.level2 min-length 4
	password.level2 max-length 16
	password.level2 complexity MIXEDCASE
	requires password.level2
}
foo {
	one two three
	four five six
	requires foobar
}`,
		// Result
		`[{"Class":"SEL421","Axioms":[{"s":"password.level2","p":"min-length","o":"4"},{"s":"password.level2","p":"max-length","o":"16"},{"s":"password.level2","p":"complexity","o":"MIXEDCASE"},{"s":"SEL421","p":"requires","o":"password.level2"}]},{"Class":"foo","Axioms":[{"s":"one","p":"two","o":"three"},{"s":"four","p":"five","o":"six"},{"s":"foo","p":"requires","o":"foobar"}]}]`,
	},
}
