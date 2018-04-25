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

// This file contains the integration tests for testing all the /node routes.
// As a part of this testing, the database is NOT mocked, but instead a test
// database is created for every test. This necessarily tests all the database
// functionality as well.
package node_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"github.com/gorilla/mux"
	config "github.com/iti/pbconf/lib/pbconfig"
	"github.com/iti/pbconf/lib/pbdatabase"
	logging "github.com/iti/pbconf/lib/pblogger"

	"github.com/iti/pbconf/lib/pbnode"
)

var logLevel = "DEBUG"

func setupDB(t *testing.T, dbFile string) pbdatabase.AppDatabase {
	logging.InitLogger(logLevel, &config.Config{}, "")
	os.Remove(dbFile)
	dbHandle := pbdatabase.Open(dbFile, logLevel)
	if dbHandle.Ping() != nil {
		t.Error("Could not create test database file")
	}
	dbHandle.LoadSchema()
	return dbHandle
}

func setupApiHandler(dbHandle pbdatabase.AppDatabase) *mux.Router {
	apiHandler := node.NewAPIHandler(logLevel, dbHandle)
	muxRouter := mux.NewRouter()
	apiHandler.AddAPIEndpoints(muxRouter)
	return muxRouter
}

func createNewRequest(t *testing.T, verb, urlPath string, body io.Reader) *http.Request {
	req, err := http.NewRequest(verb, urlPath, body)
	if err != nil {
		t.Error(err.Error())
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req
}

func createNodeInDb(t *testing.T, dbHandle pbdatabase.AppDatabase, name string) pbdatabase.PbNode {
	node := pbdatabase.PbNode{Name: name}
	if err := node.Create(dbHandle); err != nil {
		t.Error(err.Error())
	}
	return node
}

func testingError(t *testing.T, test string, format string, a ...interface{}) {
	t.Errorf(test + ":: " + format, a...)
}

func begin(t *testing.T, name string) {
	fmt.Printf("##################### Begin Node::%s #####################\n", name)
}

func end(t *testing.T, name string) {
	fmt.Printf("###################### End Node::%s ######################\n", name)
}

/****** Test all the /node handlers *************/
func TestHeadBaseHandler(t *testing.T) {
	test := "TestHeadBaseHandler"
	begin(t, test)
	defer end(t, test)
	dbFile := "test_headnode.db"
	dbHandle := setupDB(t, dbFile)
	defer os.Remove(dbFile)
	defer dbHandle.Close()
	muxRouter := setupApiHandler(dbHandle)
	//muxRouter.HandleFunc("/node", apiHandler.handleBaseRoute)

	//test1 should return error, node table is not touched yet
	req := createNewRequest(t, "HEAD", "https://localhost:8080/node", nil)
	writer := httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "Did not get StatusOK from headBaseHandler")
	}
	if writer.Header().Get("Last-Modified") != "" {
		testingError(t, test, "Should not get a Last-Modified flag here yet.")
	}
	//test2 now touch the nodes table and redo the request
	//set up items in the database
	createNodeInDb(t, dbHandle, "FirstNode")
	req = createNewRequest(t, "HEAD", "https://localhost:8080/node", nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "Did not get StatusOK from getBaseHandler")
	}
	if writer.Header().Get("Last-Modified") == "" {
		testingError(t, test, "test 2: Did not get a Last-Modified flag.")
	}
	//test Versioning
	req = createNewRequest(t, "HEAD", "https://localhost:8080/v2/node", nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Should not get StatusOK, wrong version sent")
	}

	req = createNewRequest(t, "HEAD", "https://localhost:8080/v1/node", nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "Did not get StatusOK, version is now not 1?")
	}
	//this doesnt even reach our global.ProcessVersioning, it is not a route that is registered (only v[0-9]+ routes are registered)
	req = createNewRequest(t, "HEAD", "https://localhost:8080/v1.0/node", nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Should not get StatusOK, version in wrong format sent")
	}
}

func TestGetBaseHandler(t *testing.T) {
	test := "TestGetBaseHandler"
	begin(t, test)
	defer end(t, test)
	dbFile := "test_getnode.db"
	dbHandle := setupDB(t, dbFile)
	defer os.Remove(dbFile)
	defer dbHandle.Close()
	muxRouter := setupApiHandler(dbHandle)

	//set up items in the database
	createNodeInDb(t, dbHandle, "FirstNode")
	createNodeInDb(t, dbHandle, "SecondNode")

	req := createNewRequest(t, "GET", "https://localhost:8080/node", nil)
	writer := httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "Did not get StatusOK from getBaseHandler")
	}
	//test Versioning
	req = createNewRequest(t, "GET", "https://localhost:8080/v2/node", nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Should not get StatusOK, wrong version sent")
	}

	req = createNewRequest(t, "GET", "https://localhost:8080/v1/node", nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "Did not get StatusOK, version is now not 1?")
	}
	//this doesnt even reach our global.ProcessVersioning, it is not a route that is registered (only v[0-9]+ routes are registered)
	req = createNewRequest(t, "GET", "https://localhost:8080/v1.0/node", nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Should not get StatusOK, version in wrong format sent")
	}
}

func TestPatchBaseHandler(t *testing.T) {
	test := "TestPatchBaseHandler"
	begin(t, test)
	defer end(t, test)
	dbFile := "test_patchBaseNode.db"
	dbHandle := setupDB(t, dbFile)
	defer os.Remove(dbFile)
	defer dbHandle.Close()
	muxRouter := setupApiHandler(dbHandle)

	//set up items in the database
	testNode := pbdatabase.PbNode{Name: "Hello", ConfigItems: []pbdatabase.ConfigItem{
		{Key: "NodeAttr1", Value: "Attr1Value"},
		{Key: "NodeAttr2", Value: "2ndAttrValue"},
	},
	}
	if err := testNode.Create(dbHandle); err != nil {
		testingError(t, test, "setup Create: " + err.Error())
	}

	//test 1 update the node name, leaving config items empty, should NOT update the name, name is immutable
	newNode := pbdatabase.PbNode{Id: testNode.Id, Name: "Jello",
		ConfigItems: []pbdatabase.ConfigItem{}}
	jsonStr, err := json.Marshal(newNode)
	if err != nil {
		testingError(t, test, "test 1: Error marshaling the node")
	}
	req := createNewRequest(t, "PATCH", "https://localhost:8080/node", bytes.NewBuffer(jsonStr))
	writer := httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Should not have updated a node with name change.")
	}
	node3 := pbdatabase.PbNode{Id: testNode.Id}
	if err := node3.Get(dbHandle); err != nil {
		testingError(t, test, "db node3 Get error:" + err.Error())
	}
	if node3.Name == "Jello" {
		testingError(t, test, "Updated the node name through patch route. This should have failed.")
	}

	//test2 update existing config item
	newNode = pbdatabase.PbNode{Id: testNode.Id, Name: "Hello",
		ConfigItems: []pbdatabase.ConfigItem{{Key: "NodeAttr2", Value: "New 2ndAttrValue"}}}
	jsonStr, err = json.Marshal(newNode)
	if err != nil {
		testingError(t, test, "test 2: Error marshaling the node")
	}
	req = createNewRequest(t, "PATCH", "https://localhost:8080/node", bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "Test 2:Did not update a node.")
	}
	cfgItem := pbdatabase.PbNodeConfigItem{NodeId: testNode.Id, ConfigItem: pbdatabase.ConfigItem{Key: "NodeAttr2"}}
	if err := cfgItem.Get(dbHandle); err != nil {
		testingError(t, test, "Test 2:cfg get error:" + err.Error())
	}
	if cfgItem.Value != "New 2ndAttrValue" {
		testingError(t, test, "Not able to update config item")
	}
	//test3 add new config item
	newNode = pbdatabase.PbNode{Id: testNode.Id, Name: "Hello",
		ConfigItems: []pbdatabase.ConfigItem{{Key: "JelloAttr", Value: "Yellow"}}}
	jsonStr, err = json.Marshal(newNode)
	if err != nil {
		testingError(t, test, "test 3: Error marshaling the node")
	}
	req = createNewRequest(t, "PATCH", "https://localhost:8080/node", bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "Test 3: Did not update a node.")
	}
	cfgItem = pbdatabase.PbNodeConfigItem{NodeId: testNode.Id, ConfigItem: pbdatabase.ConfigItem{Key: "JelloAttr"}}
	exists, err := cfgItem.Exists(dbHandle)
	if !exists || err != nil {
		testingError(t, test, "Test 3: Could not add a config item to existing node")
	}
	//test 4: try to simulate decoder error
	jsonStr = []byte(``)
	req = createNewRequest(t, "PATCH", "https://localhost:8080/node", bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusBadRequest {
		testingError(t, test, "Test4: Should not have been able to PATCH a node with decoder error.")
	}
	//test Versioning
	newNode = pbdatabase.PbNode{Id: testNode.Id, Name: "Hello",
		ConfigItems: []pbdatabase.ConfigItem{{Key: "Random", Value: "YellowRandom"}}}
	jsonStr, err = json.Marshal(newNode)
	if err != nil {
		testingError(t, test, "test versioning: Error marshaling the node")
	}
	req = createNewRequest(t, "PATCH", "https://localhost:8080/v2/node", bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Test 5: Should not get StatusOK, wrong version sent")
	}
	//this doesnt even reach our global.ProcessVersioning, it is not a route that is registered (only v[0-9]+ routes are registered)
	req = createNewRequest(t, "PATCH", "https://localhost:8080/v1.0/node", bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Test6: Should not get StatusOK, version in wrong format sent")
	}
	req = createNewRequest(t, "PATCH", "https://localhost:8080/v1/node", bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "Test7: Did not get StatusOK, version is now not 1?")
	}
}

func TestPostBaseHandler(t *testing.T) {
	test := "TestPostBaseHandler"
	begin(t, test)
	defer end(t, test)

	dbFile := "test_postBaseNode.db"
	dbHandle := setupDB(t, dbFile)
	defer os.Remove(dbFile)
	defer dbHandle.Close()
	muxRouter := setupApiHandler(dbHandle)

	//test 1 create a node, should succeed when id is not specified
	testNode := pbdatabase.PbNode{Name: "NewNode", ConfigItems: []pbdatabase.ConfigItem{
		{Key: "NodeAttr1", Value: "Attr1Value"},
		{Key: "NodeAttr2", Value: "2ndAttrValue"},
	},
	}
	jsonStr, err := json.Marshal(testNode)
	if err != nil {
		testingError(t, test, "test 1: Error marshaling the node")
	}
	req := createNewRequest(t, "POST", "https://localhost:8080/node", bytes.NewBuffer(jsonStr))
	writer := httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusCreated {
		testingError(t, test, "Did not create a node when id was not specified")
	}

	//test 2  specify name and id of node, post should not suceed
	testNode2 := pbdatabase.PbNode{Name: "NewNode", Id: 22, ConfigItems: []pbdatabase.ConfigItem{
		{Key: "NodeAttr1", Value: "Attr1Value"},
		{Key: "NodeAttr2", Value: "2ndAttrValue"},
	},
	}
	jsonStr, err = json.Marshal(testNode2)
	if err != nil {
		testingError(t, test, "test 2: Error marshaling the node")
	}
	req = createNewRequest(t, "POST", "https://localhost:8080/node", bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusCreated {
		testingError(t, test, "Should not be able to create a node with specified id on POST request.")
	}

	//test 3 update the node, should not be able to update the node
	testNode3 := pbdatabase.PbNode{Name: "NewNode3", Id: testNode.Id, ConfigItems: []pbdatabase.ConfigItem{
		{Key: "NodeAttr1", Value: "Attr1Value"},
		{Key: "NodeAttr2", Value: "2ndAttrValue"},
	},
	}
	jsonStr, err = json.Marshal(testNode3)
	if err != nil {
		testingError(t, test, "test 3: Error marshaling the node")
	}
	req = createNewRequest(t, "POST", "https://localhost:8080/node", bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "test 3: Should not be able to update a node with specified id on POST request.")
	}

	//test 4 :try to trigger validation Error
	jsonStr2 := []byte(`{"Name":"", "ConfigItems":[{"Key":"xx", "Value":"yy"}]}`)
	req = createNewRequest(t, "POST", "https://localhost:8080/node", bytes.NewBuffer(jsonStr2))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "test 4: Should not be able to update a node with validation error.")
	}
	//test5:trigger decoder error
	jsonStr2 = []byte(``)
	req = createNewRequest(t, "POST", "https://localhost:8080/node", bytes.NewBuffer(jsonStr2))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "test 5: Should not be able to update a node with validation error.")
	}
	//test versioning of route
	testNode = pbdatabase.PbNode{Name: "NewNodeVV", ConfigItems: []pbdatabase.ConfigItem{
		{Key: "NodeVVAttr1", Value: "Attr1Value"},
		{Key: "NodeVVAttr2", Value: "2ndAttrValue"},
	},
	}
	jsonStr, err = json.Marshal(testNode)
	if err != nil {
		testingError(t, test, "test 6: Error marshaling the node")
	}
	req = createNewRequest(t, "POST", "https://localhost:8080/v2/node", bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Test 5: Should not get StatusOK, wrong version sent")
	}

	req = createNewRequest(t, "POST", "https://localhost:8080/v1.0/node", bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Test6: Should not get StatusOK, version in wrong format sent")
	}
	req = createNewRequest(t, "POST", "https://localhost:8080/v1/node", bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusCreated {
		testingError(t, test, "Test7: Did not get StatusOK, version is now not 1?")
	}
}

/****** Test all the /node/{nodeid} handlers *************/
func TestGetWIdHandler(t *testing.T) {
	test := "TestGetWIdHandler"
	begin(t, test)
	defer end(t, test)

	dbFile := "test_getWIDNode.db"
	dbHandle := setupDB(t, dbFile)
	defer os.Remove(dbFile)
	defer dbHandle.Close()
	muxRouter := setupApiHandler(dbHandle)

	//set up our node in the database
	testNode := pbdatabase.PbNode{Name: "NewNode", ConfigItems: []pbdatabase.ConfigItem{
		{Key: "NodeAttr1", Value: "Attr1Value"},
		{Key: "NodeAttr2", Value: "2ndAttrValue"},
	},
	}
	err := testNode.Create(dbHandle)
	if err != nil {
		testingError(t, test, "Error setting up node in the database")
	}
	//test1: things are ok?
	fmt.Println("Test 1 TestGetWIdHandler")
	req := createNewRequest(t, "GET", fmt.Sprintf("https://localhost:8080/node/%d", testNode.Id), nil)
	writer := httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)

	if writer.Code != http.StatusOK {
		testingError(t, test, "Test1: Could not GET a node with specified id.")
	}
	//test Versioning
	req = createNewRequest(t, "GET", fmt.Sprintf("https://localhost:8080/v2/node/%d", testNode.Id), nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Should not get StatusOK, wrong version sent")
	}

	req = createNewRequest(t, "GET", fmt.Sprintf("https://localhost:8080/v1/node/%d", testNode.Id), nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "Did not get StatusOK, version is now not 1?")
	}

	req = createNewRequest(t, "GET", fmt.Sprintf("https://localhost:8080/v1.0/node/%d", testNode.Id), nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Should not get StatusOK, version in wrong format sent")
	}
}

func TestPatchWIdHandler(t *testing.T) {
	test := "TestPatchWIdHandler"
	begin(t, test)
	defer end(t, test)

	dbFile := "test_patchWIDNode.db"
	dbHandle := setupDB(t, dbFile)
	defer os.Remove(dbFile)
	defer dbHandle.Close()
	muxRouter := setupApiHandler(dbHandle)

	//set up our node in the database
	testNode := pbdatabase.PbNode{Name: "NewNode", ConfigItems: []pbdatabase.ConfigItem{
		{Key: "NodeAttr1", Value: "Attr1Value"},
		{Key: "NodeAttr2", Value: "2ndAttrValue"},
	},
	}
	err := testNode.Create(dbHandle)
	if err != nil {
		testingError(t, test, "Error setting up node in the database")
	}
	testNodeId := fmt.Sprintf("%v", testNode.Id)
	//try to update non existent node
	var jsonStr = []byte(`{"Name":"NewNode"}`)
	req := createNewRequest(t, "PATCH", "https://localhost:8080/node/4", bytes.NewBuffer(jsonStr))
	writer := httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "should not be able to update node with specified id that does not exist. %v", writer.Code)
	}
	//test 1 update the node name, leaving config items empty, should NOT update the name
	newNode := pbdatabase.PbNode{Id: testNode.Id, Name: "Jello",
		ConfigItems: []pbdatabase.ConfigItem{}}
	jsonStr, err = json.Marshal(newNode)
	if err != nil {
		testingError(t, test, "test 1: Error marshaling the node")
	}
	req = createNewRequest(t, "PATCH", "https://localhost:8080/node/"+testNodeId, bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Test 1: Did not update a node.")
	}
	node3 := pbdatabase.PbNode{Id: testNode.Id}
	if err := node3.Get(dbHandle); err != nil {
		testingError(t, test, "Test 1: " + err.Error())
	}
	if node3.Name == "Jello" {
		testingError(t, test, "Test 1: Updated the node name through patch route, when it should have failed.")
	}

	//test2 update exisiting config item
	newNode = pbdatabase.PbNode{Id: testNode.Id, Name: "NewNode",
		ConfigItems: []pbdatabase.ConfigItem{{Key: "NodeAttr2", Value: "New 2ndAttrValue"}}}
	jsonStr, err = json.Marshal(newNode)
	if err != nil {
		testingError(t, test, "test 2: Error marshaling the node")
	}
	req = createNewRequest(t, "PATCH", "https://localhost:8080/node/"+testNodeId, bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "Test 2:Did not update a node.")
	}
	cfgItem := pbdatabase.PbNodeConfigItem{NodeId: testNode.Id, ConfigItem: pbdatabase.ConfigItem{Key: "NodeAttr2"}}
	if err := cfgItem.Get(dbHandle); err != nil {
		testingError(t, test, "Test2: " + err.Error())
	}
	if cfgItem.Value != "New 2ndAttrValue" {
		testingError(t, test, "test 2:Not able to update config item")
	}
	//test3 add new config item
	newNode = pbdatabase.PbNode{Id: testNode.Id, Name: "NewNode",
		ConfigItems: []pbdatabase.ConfigItem{{Key: "JelloAttr", Value: "Yellow"}}}
	jsonStr, err = json.Marshal(newNode)
	if err != nil {
		testingError(t, test, "test 3: Error marshaling the node")
	}
	req = createNewRequest(t, "PATCH", "https://localhost:8080/node/"+testNodeId, bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "Test 3: Did not update a node.")
	}
	cfgItem = pbdatabase.PbNodeConfigItem{NodeId: testNode.Id, ConfigItem: pbdatabase.ConfigItem{Key: "JelloAttr"}}
	exists, err := cfgItem.Exists(dbHandle)
	if !exists || err != nil {
		testingError(t, test, "Test 3: Could not add a config item to existing node")
	}
	//test 4: illegal id in route
	req = createNewRequest(t, "PATCH", "https://localhost:8080/node/ee", bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusBadRequest {
		testingError(t, test, "Test4: Should not have been able to PATCH a node with illegal id.")
	}
	//test 5: try to simulate decoder error
	jsonStr = []byte(``)
	req = createNewRequest(t, "PATCH", "https://localhost:8080/node/"+testNodeId, bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusBadRequest {
		testingError(t, test, "Test5: Should not have been able to PATCH a node with illegal id.")
	}
	//test Versioning, tests 6-8
	newNode = pbdatabase.PbNode{Id: testNode.Id, Name: "NewNode",
		ConfigItems: []pbdatabase.ConfigItem{{Key: "JelloVersionAttr", Value: "YellowVersion"}}}
	jsonStr, err = json.Marshal(newNode)
	if err != nil {
		testingError(t, test, "test 6-8: Error marshaling the node")
	}

	req = createNewRequest(t, "PATCH", fmt.Sprintf("https://localhost:8080/v2/node/%d", testNode.Id), bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Test6: Should not get StatusOK, wrong version sent")
	}

	req = createNewRequest(t, "PATCH", fmt.Sprintf("https://localhost:8080/v1/node/%d", testNode.Id), bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "Test7: Did not get StatusOK, version is now not 1?")
	}

	req = createNewRequest(t, "PATCH", fmt.Sprintf("https://localhost:8080/v1.0/node/%d", testNode.Id), bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Test8: Should not get StatusOK, version in wrong format sent")
	}
}

func TestDeleteWIdHandler(t *testing.T) {
	test := "TestDeleteWIdHandler"
	begin(t, test)
	defer end(t, test)

	dbFile := "test_deleteWIDNode.db"
	dbHandle := setupDB(t, dbFile)
	defer os.Remove(dbFile)
	defer dbHandle.Close()
	muxRouter := setupApiHandler(dbHandle)

	//set up our node in the database
	testNode := pbdatabase.PbNode{Name: "NewNode", ConfigItems: []pbdatabase.ConfigItem{
		{Key: "NodeAttr1", Value: "Attr1Value"},
		{Key: "NodeAttr2", Value: "2ndAttrValue"},
	},
	}
	err := testNode.Create(dbHandle)
	if err != nil {
		testingError(t, test, "Error setting up node in the database")
	}
	testNodeId := fmt.Sprintf("%v", testNode.Id)
	req := createNewRequest(t, "DELETE", "https://localhost:8080/node/5", nil)
	writer := httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "Test 1: was able to delete a node with specified id that does not exist.")
	}
	//delete existing node, should go through
	req = createNewRequest(t, "DELETE", "https://localhost:8080/node/"+testNodeId, nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "Test 2: was not able to delete a node with specified id.")
	}
	//delete node with illegal id in route, should trigger decoder error
	req = createNewRequest(t, "DELETE", "https://localhost:8080/node/ee", nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusBadRequest {
		testingError(t, test, "Test 3: Should not able to delete a node with illegal id.")
	}
	//test versioning in route
	testNode = pbdatabase.PbNode{Name: "NewVersioninhNode", ConfigItems: []pbdatabase.ConfigItem{
		{Key: "NodeXXAttr1", Value: "Attr1Value"},
		{Key: "NodeYYAttr2", Value: "2ndAttrValue"},
	},
	}
	err = testNode.Create(dbHandle)
	if err != nil {
		testingError(t, test, "Test 4-6: Error setting up node in the database")
	}

	req = createNewRequest(t, "DELETE", fmt.Sprintf("https://localhost:8080/v2/node/%d", testNode.Id), nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Test4: Should not get StatusOK, wrong version sent")
	}

	req = createNewRequest(t, "DELETE", fmt.Sprintf("https://localhost:8080/v1/node/%d", testNode.Id), nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "Test 5: Did not get StatusOK, version is now not 1?")
	}

	req = createNewRequest(t, "DELETE", fmt.Sprintf("https://localhost:8080/v1.0/node/%d", testNode.Id), nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Test 6: Should not get StatusOK, version in wrong format sent")
	}
}

/****** Test all the /node/{nodeid}/{cfgkey} handlers *************/
func TestGetConfigElement(t *testing.T) {
	test := "TestGetConfigElement"
	begin(t, test)
	defer end(t, test)

	dbFile := "test_getConfigElem.db"
	dbHandle := setupDB(t, dbFile)
	defer os.Remove(dbFile)
	defer dbHandle.Close()
	muxRouter := setupApiHandler(dbHandle)
	//setup items in the database for testing
	testNode := pbdatabase.PbNode{Id: 3,
		ConfigItems: []pbdatabase.ConfigItem{
			{Key: "NodeAttribute", Value: "Alive"},
			{Key: "Location", Value: "Head of tree"},
		},
	}
	if err := testNode.Create(dbHandle); err != nil {
		testingError(t, test, "test 0: " + err.Error())
	}
	testNodeId := fmt.Sprintf("%v", testNode.Id)
	//now set up and test the route
	req := createNewRequest(t, "GET", "https://localhost:8080/node/"+testNodeId+"/NodeAttribute", nil)
	writer := httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "Failed to GET the specified config item for the node")
	}
	//check the value of the returned config item
	var nodeCItem pbdatabase.PbNodeConfigItem
	decoder := json.NewDecoder(writer.Body)
	err := decoder.Decode(&nodeCItem)
	if err != nil {
		testingError(t, test, "Decoder error: " + err.Error())
	}
	if nodeCItem.Value != "Alive" {
		testingError(t, test, "Config Item recovered did not have expected value")
	}
	//try to trigger parseIdFromRoute error
	req = createNewRequest(t, "GET", "https://localhost:8080/node/"+"ee"+"/NodeAttribute", nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusBadRequest {
		testingError(t, test, "Test 1: Failed to GET the specified config item for the node")
	}
	//test3: non existent node request
	req = createNewRequest(t, "GET", "https://localhost:8080/node/"+"5"+"/NodeAttribute", nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusNotFound {
		testingError(t, test, "Test 2: Failed to GET the specified config item for the node")
	}
	//test versioning in route
	req = createNewRequest(t, "GET", "https://localhost:8080/v2/node/"+testNodeId+"/NodeAttribute", nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Test3: Should not get StatusOK, wrong version sent")
	}

	req = createNewRequest(t, "GET", "https://localhost:8080/v1/node/"+testNodeId+"/NodeAttribute", nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "Test 4: Did not get StatusOK, version is now not 1?")
	}

	req = createNewRequest(t, "GET", "https://localhost:8080/v1.1/node/"+testNodeId+"/NodeAttribute", nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Test 5: Should not get StatusOK, version in wrong format sent")
	}
}

func TestDeleteConfigElement(t *testing.T) {
	test := "TestDeleteConfigElement"
	begin(t, test)
	defer end(t, test)

	dbFile := "test_deleteConfigElem.db"
	dbHandle := setupDB(t, dbFile)
	defer os.Remove(dbFile)
	defer dbHandle.Close()
	muxRouter := setupApiHandler(dbHandle)

	//set up our node in the database
	testNode := pbdatabase.PbNode{Name: "root",
		ConfigItems: []pbdatabase.ConfigItem{
			{Key: "NodeAttribute", Value: "Alive"},
			{Key: "Location", Value: "Head of tree"},
			{Key: "VersionDeleteAttr", Value: "something"},
		},
	}
	if err := testNode.Create(dbHandle); err != nil {
		testingError(t, test, "Test 0:" + err.Error())
	}
	testNodeId := fmt.Sprintf("%v", testNode.Id)
	//try to delete nonexistent config item, should fail
	req := createNewRequest(t, "DELETE", "https://localhost:8080/node/"+testNodeId+"/NonExisKey", nil)
	writer := httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Test 1: Function returned StatusOK when it should have failed.")
	}

	//delete existing node, should go through
	req = createNewRequest(t, "DELETE", "https://localhost:8080/node/"+testNodeId+"/NodeAttribute", nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "Test 2: Was not able to delete specified config item.")
	}

	//confirm that the key NodeAttribute was deleted in the database
	cfgItm := pbdatabase.PbNodeConfigItem{NodeId: testNode.Id, ConfigItem: pbdatabase.ConfigItem{Key: "NodeAttribute"}}
	exists, err := cfgItm.Exists(dbHandle)
	if err != nil {
		testingError(t, test, "Checking existence of cfgItem error: " + err.Error())
	}
	if exists {
		testingError(t, test, "Test 2: Config item should not still exist")
	}
	//try to trigger parseIdFromRoute error
	req = createNewRequest(t, "DELETE", "https://localhost:8080/node/"+"ee"+"/Location", nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusBadRequest {
		testingError(t, test, "Test 3: Failed to get correct response from server")
	}
	//test versioning in route
	req = createNewRequest(t, "DELETE", "https://localhost:8080/v2/node/"+testNodeId+"/VersionDeleteAttr", nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Test 4: Should not get StatusOK, wrong version sent")
	}

	req = createNewRequest(t, "DELETE", "https://localhost:8080/v1.1/node/"+testNodeId+"/VersionDeleteAttr", nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Test 5: Should not get StatusOK, version in wrong format sent")
	}
	req = createNewRequest(t, "DELETE", "https://localhost:8080/v1/node/"+testNodeId+"/VersionDeleteAttr", nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "Test 6: Did not get StatusOK, version is now not 1?")
	}
}
