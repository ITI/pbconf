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


import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/iti/pbconf/lib/pbpolicy"
	"testing"
)

func TestDSL(t *testing.T) {
	for _, c := range Cases {
		e := dodsltests(c)
		if e != nil {
			t.Errorf(e.Error())
		}
	}
}

func dodsltests(input DSLCase) error {
	test := strings.NewReader(input.C)

	a, err := policy.Parse(test)
	if err != nil {
		return err
	}

	j, err := json.Marshal(a)

	if err != nil {
		return err
	}

	if string(j) != input.O {
		return fmt.Errorf("Case %d\nExpected: %s\n Got: %s", input.I, input.O, j)
	}

	return nil
}
