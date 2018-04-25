package ontology_test

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
	"bytes"
	"fmt"
	"testing"
	"strings"
	"github.com/iti/pbconf/lib/pbontology"
)

func begin(t *testing.T, name string) {
	fmt.Printf("##################### Begin Ontology::%s #####################\n", name)
}

func end(t *testing.T, name string) {
	fmt.Printf("###################### End Ontology::%s ######################\n", name)
}

func testingError(t *testing.T, test string, format string, a ...interface{}) {
	t.Errorf(test+":: "+format, a...)
}

func buildInnerConfiguration(Op string, Key string, Val string, Svc string) (string) {
	props := "{\"Op\":\"" + Op + "\",\"Key\":\"" + Key + "\",\"Val\":\"" + Val + "\",\"Svc\":\"" + Svc + "\"}"
	return props
}

func buildOuterConfiguration(ontology string, ontologizer string, individual string, cfgs []string) (string) {
	r := strings.Join(cfgs, ",")
	return "{\"ontology\":\"" + ontology + "\", \"ontologizer\":\"" + ontologizer + "\", \"individual\":\"" + individual + "\", \"properties\":[" + r + "]}"

}

func buildCombinedConfiguration(cfgs []string) (string) {
	r := strings.Join(cfgs, ",")
	return r
	//return "{\"ontology\":\"config\", \"properties\":[" + r + "]}"
}

func issueConfiguration(cfg string) (bool, string, error)  {
	var buffer bytes.Buffer
	buffer.WriteString(string(cfg))
	return ontology.ValidateAgainstOntology(buffer)
}

func testBasicPasswordPolicy(t *testing.T, test string) {
	//Make sure we reset the ontology server policy and configuration ontologies before unit tests
	resetResult := ontology.ResetOntologyServer()

	if len(resetResult) > 0 {
		testingError(t, test, "Error reseting ontology server for unit tests\n")
	}

	if !ontology.CheckServerAvailability() {
		return
	}

	//Create sel421a with 10 char pass
	ocs := []string{}
	cfgs := []string{}
	cfgs = append(cfgs, buildInnerConfiguration("variable", "type", "SEL421", ""))
	cfgs = append(cfgs, buildInnerConfiguration("password", "level2", "abcdefGHIK", ""))
	ocs = append(ocs, buildOuterConfiguration("config", "SEL421", "sel421a", cfgs))
	status, explanation, error := issueConfiguration(buildCombinedConfiguration(ocs))
	if status == false {
		if error == nil {
			testingError(t, test, "Error validating, with explanation : %s" + explanation)
		} else {
			testingError(t, test, "Error validating, with explanation : %s, and error %s" + explanation, error.Error())
		}
	}

	//Create sel421b with 8 char pass
	ocsb := []string{}
	cfgsb := []string{}
	cfgsb = append(cfgsb, buildInnerConfiguration("variable", "type", "SEL421", ""))
	cfgsb = append(cfgsb, buildInnerConfiguration("password", "level2", "abcdEFGH", ""))
	ocsb = append(ocsb, buildOuterConfiguration("config", "SEL421", "sel421b", cfgsb))
	statusb, explanationb, errorb := issueConfiguration(buildCombinedConfiguration(ocsb))
	if statusb == false {
		if errorb == nil {
			testingError(t, test, "Error validating, with explanation : %s" + explanationb)
		} else {
			testingError(t, test, "Error validating, with explanation : %s, and error %s" + explanationb, errorb.Error())
		}
	}

	//Set a password policy requiring minimum 12 characters (this should break both configurations)
	var buffer bytes.Buffer
	buffer.WriteString("{\"ontology\":\"policy\",\"data\":[[{\"Class\":\"SEL421\",\"Axioms\":[{\"s\":\"password.level2\",\"p\":\"min-length\",\"o\":\"12\"},{\"s\":\"password.level2\",\"p\":\"max-length\",\"o\":\"16\"}]}]]}")
	status, explanation, e := ontology.ValidateAgainstOntology(buffer)

	if status == false {
		testingError(t, test, "Could not send policy to ontology server successfully. Got explanation: %v\n, error:%s\n", explanation, e)
	}

	//Get the current state (empty result is considered valid)
	//We should get both items back as errored individuals
	currentState := ontology.GetOntologyInconsistencies()

	if len(currentState) == 0 {
		testingError(t, test, "Should have received an error here, but didn't A\n")
	}

	//Update the policy to only require passwords be > 6 length
	var abuffer bytes.Buffer
	abuffer.WriteString("{\"ontology\":\"policy\",\"data\":[[{\"Class\":\"SEL421\",\"Axioms\":[{\"s\":\"password.level2\",\"p\":\"min-length\",\"o\":\"9\"},{\"s\":\"password.level2\",\"p\":\"max-length\",\"o\":\"16\"}]}]]}")
	status, explanation, e = ontology.ValidateAgainstOntology(abuffer)

	if status == false {
		testingError(t, test, "Could not send policy to ontology server successfully. Got explanation: %v\n, error:%s\n", explanation, e)
	}

	//Get the current state (empty result is considered valid)
	//We should get 1 error item here now that the requirement has been increased
	currentState = ontology.GetOntologyInconsistencies()
	if len(currentState) == 0 {
		testingError(t, test, "Should have received an error here, but didn't B\n", explanation, e)
	}

	//Update the policy to only require passwords be > 4 length
	var bbuffer bytes.Buffer
	bbuffer.WriteString("{\"ontology\":\"policy\",\"data\":[[{\"Class\":\"SEL421\",\"Axioms\":[{\"s\":\"password.level2\",\"p\":\"min-length\",\"o\":\"6\"},{\"s\":\"password.level2\",\"p\":\"max-length\",\"o\":\"16\"}]}]]}")
	status, explanation, e = ontology.ValidateAgainstOntology(bbuffer)

	if status == false {
		testingError(t, test, "Could not send policy to ontology server successfully. Got explanation: %v\n, error:%s\n", explanation, e)
	}

	//Now get the final state, which should be valid (so empty response)
	currentState = ontology.GetOntologyInconsistencies()
	if len(currentState) > 0 {
		testingError(t, test, "Shouldn't receive any type of error after the issue was corrected \n", explanation, e)
	}
}

func TestOntologyVerificationHandler(t *testing.T) {
	test := "TestOntology"
	begin(t, test)
	defer end(t, test)

	testBasicPasswordPolicy(t, test);
}
