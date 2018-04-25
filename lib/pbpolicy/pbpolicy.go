package policy

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
	"encoding/json"
	"io"

	change "github.com/iti/pbconf/lib/pbchange"
	logging "github.com/iti/pbconf/lib/pblogger"
	ontology "github.com/iti/pbconf/lib/pbontology"
)

func ParsePolicy(b io.Reader) error {
	_, err := Parse(b)
	return err
}

/**
 * This aggregates all policy, including the new / edited policy and sends it out to the Ontology server for validation
 * This results in ontology, ontologizer, and a data key which is an array of arrays containing target + array of axioms
 */
func ValidateAgainstOntology(log logging.Logger, policy *change.ChangeData) (bool, string, error) {
	//aggregate all the policies Content
	changeEng, err := change.GetCMEngine(nil)
	if err != nil {
		return false, "", err
	}

	polList, err := changeEng.ListObjects(change.POLICY)
	if err != nil {
		return false, "", err
	}
	var buf bytes.Buffer
	var ct int = 0
	buf.WriteString("[")
	for _, policyName := range polList {
		pol, err := changeEng.GetObject(change.POLICY, policyName)
		if err != nil {
			return false, "", err
		}
		if policy != nil && policyName == policy.Content.Object {
			continue //skip the saved content if its a policy edit
		} else {
			var tBuf bytes.Buffer
			if ct != 0 {
				buf.WriteString(",")
			}
			tBuf.Write(pol.Content.Files["Rules"])
			ast, err := Parse(bytes.NewReader(tBuf.Bytes()))
			if err != nil {
				log.Debug(err.Error())
				return false, "Error parsing rules into AST", err
			}
			j, err := json.Marshal(ast)
			if err != nil {
				log.Debug(err.Error())
				return false, "Error parsing AST into JSON", err
			}

			buf.WriteString(string(j))
			ct++
		}
	}

	//always write the prospective changed/new policy
	if policy != nil {
		var tBuf bytes.Buffer
		tBuf.Write(policy.Content.Files["Rules"])
		ast, err := Parse(bytes.NewReader(tBuf.Bytes()))
		if err != nil {
			log.Debug(err.Error())
			return false, "Error parsing rules into AST", err
		}
		j, err := json.Marshal(ast)
		if err != nil {
			log.Debug(err.Error())
			return false, "Error parsing AST into JSON", err
		}

		if ct != 0 {
			buf.WriteString(",")
		}
		buf.WriteString(string(j))
	}

	buf.WriteString("]")

	var buffer bytes.Buffer
	buffer.WriteString("{\"ontology\":\"policy\",")
	buffer.WriteString("\"data\":" + buf.String() + "}")

	log.Debug("=== Buffer state being sent to ontology ===")
	log.Debug(buffer.String())

	status, explanation, err := ontology.ValidateAgainstOntology(buffer)
	if err != nil {
		log.Debug("Error validating : " + err.Error())
	}
	return status, explanation, err
}

func LogOntologyInconsistencies(log logging.Logger) {
	inconsistencies := ontology.GetOntologyInconsistencies()
	if len(inconsistencies) > 0 {
		//We have inconsistencies, have to parse
		log.Log("DEBUG", "%s", "Current inconsistencies")
		log.Log("DEBUG", "%s", string(inconsistencies))
	}
}
