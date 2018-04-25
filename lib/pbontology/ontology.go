package ontology

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
	"bufio"
	"bytes"
	"encoding/json"
	logging "github.com/iti/pbconf/lib/pblogger"
	"net"
	"strings"
)

//Accepts a buffer and emits it to the ontology for processing
//Returns the consistency of the ontology post-operations
//Returns explanation if inconsistency is found
//Returns error if problems unrelated to ontology validation occur
func ValidateAgainstOntology(buffer bytes.Buffer) (bool, string, error) {
	explanation := ""

	//Connect to the server on port 9090
	conn, err := net.Dial("tcp", "localhost:9090")
	if err != nil {
		//Fallback to VALID, request failed
		//This should only occur if an ontology server isn't started
		explanation = "Error connecting to Ontology server, forcing true for now (ValidateAgainstOntology)."
		return true, explanation, nil
	}

	//At this point, we've made a connection, so try to read the connection string from server
	connectStatus, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil || len(string(connectStatus)) == 0 {
		explanation = "Invalid connection status = " + string(connectStatus)
		return false, explanation, err
	}

	//Write out buffer to server
	conn.Write([]byte(buffer.String() + "\n"))
	bufferStatus, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil || len(string(bufferStatus)) == 0 {
		//Fallback to invalid, request failed
		explanation = "Invalid result = " + string(bufferStatus)
		return false, explanation, err
	}

	//Close connection
	exitStr := "exit"
	conn.Write([]byte(exitStr + "\n"))
	exitStatus, err := bufio.NewReader(conn).ReadString('\n')
	if len(exitStatus) > 0 {
		explanation = "Exit status was " + string(exitStatus)
	}

	byt := []byte(bufferStatus)
	var dat map[string]interface{}
	if err := json.Unmarshal(byt, &dat); err != nil {
		explanation := "Failed to parse response"
		return false, explanation, err
	}

	status := dat["status"].(string)
	explanation = dat["explanation"].(string)

	if strings.Compare(status, "VALID") == 0 {
		return true, explanation, nil
	} else {
		return false, explanation, nil
	}
}

func CheckServerAvailability() bool {
	conn, err := net.Dial("tcp", "localhost:9090")
	if err != nil {
		return false
	}

	connectStatus, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil || len(string(connectStatus)) == 0 {
		return false
	}

	return true
}

func ResetOntologyServer() string {
	explanation := ""

	conn, err := net.Dial("tcp", "localhost:9090")
	if err != nil {
		explanation = "Error connecting to Ontology server, forcing true for now."
		return ""
	}

	connectStatus, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil || len(string(connectStatus)) == 0 {
		explanation = "Invalid connection status = " + string(connectStatus)
	}

	conn.Write([]byte("reset" + "\n"))
	bufferStatus, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil || len(string(bufferStatus)) == 0 {
		explanation = "Failed to reset ontology server"
	}

	exitStr := "exit"
	conn.Write([]byte(exitStr + "\n"))
	exitStatus, err := bufio.NewReader(conn).ReadString('\n')
	if len(exitStatus) > 0 {
		explanation = "Exit status was " + string(exitStatus)
	}

	if len(bufferStatus) == 0 {
		explanation = ""
	}

	return explanation
}

//This function checks the whole ontology and reports any device config inconsistencies in a parseable format.
func GetOntologyInconsistencies() string {
	explanation := ""
	l, _ := logging.GetLogger("Policy API")
	logging.SetLevel("DEBUG", "Policy API")

	conn, err := net.Dial("tcp", "localhost:9090")
	if err != nil {
		explanation = "Error connecting to Ontology server, forcing true for now."
		return ""
	}

	//At this point, we've made a connection, so try to read the connection string from server
	connectStatus, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil || len(string(connectStatus)) == 0 {
		explanation = "Invalid connection status = " + string(connectStatus)
	}

	conn.Write([]byte("validate" + "\n"))
	bufferStatus, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil || len(string(bufferStatus)) == 0 {
		//Fallback to invalid, request failed
		explanation = "Invalid result = " + string(bufferStatus)
	}

	exitStr := "exit"
	conn.Write([]byte(exitStr + "\n"))
	exitStatus, err := bufio.NewReader(conn).ReadString('\n')

	if len(exitStatus) > 0 {
		explanation = "Exit status was " + string(exitStatus)
	}

	explanation = string(bufferStatus)

	//Remove when it's just a new line char
	if len(explanation) <= 2 {
		explanation = ""
	}

	if len(explanation) > 0 {
		l.Log("DEBUG", "Explanation = %s, %i", explanation, len(explanation))
	}

	return explanation
}
