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
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	mux "github.com/gorilla/mux"
	change "github.com/iti/pbconf/lib/pbchange"
	config "github.com/iti/pbconf/lib/pbconfig"
	"github.com/iti/pbconf/lib/pbdatabase"
	global "github.com/iti/pbconf/lib/pbglobal"
	logging "github.com/iti/pbconf/lib/pblogger"

	"github.com/iti/pbconf/lib/pbpolicy"
)

var logLevel = "DEBUG"
var rcfg *config.Config
var repoLocker sync.Mutex

func begin(t *testing.T, name string) {
	fmt.Printf("##################### Begin Policy::%s #####################\n", name)
}

func end(t *testing.T, name string) {
	fmt.Printf("###################### End Policy::%s ######################\n", name)
}

func testingError(t *testing.T, test string, format string, a ...interface{}) {
	t.Errorf(test+":: "+format, a...)
}

func setupRepo() *config.Config {
	testrepopath, err := ioutil.TempDir("", "cmengine")
	if err != nil {
		panic(err)
	}
	repoLocker.Lock()
	rcfg = new(config.Config)
	rcfg.ChMgmt.RepoPath = testrepopath
	rcfg.ChMgmt.LogLevel = "DEBUG"
	return rcfg
}

func getCME(t *testing.T) *change.CMEngine {
	//set up CME, to AddAPIEndpoints, CME must exist
	var engine *change.CMEngine
	var err error
	if rcfg == nil {
		setupRepo()
		engine, err = change.GetCMEngine(rcfg)
	} else {
		engine, err = change.GetCMEngine(rcfg)
	}
	if err != nil {
		t.Error("Failed to get CME ref")
	}
	return engine
}

func cleanupCME(e *change.CMEngine) {
	err := os.RemoveAll(rcfg.ChMgmt.RepoPath)
	if err != nil {
		panic(err)
	}
	e.Free()
	rcfg = nil
	repoLocker.Unlock()
}

func setupDB(t *testing.T, dbFile string) pbdatabase.AppDatabase {
	logging.InitLogger(logLevel, &config.Config{}, "") //this is what makes all the module names show up in logged stmts
	os.Remove(dbFile)
	dbHandle := pbdatabase.Open(dbFile, logLevel)
	if dbHandle.Ping() != nil {
		t.Error("Could not create test database file")
	}
	dbHandle.LoadSchema()
	return dbHandle
}

func setupApiHandler(dbHandle pbdatabase.AppDatabase) *mux.Router {
	apiHandler := policy.NewAPIHandler(logLevel, dbHandle)
	muxRouter := mux.NewRouter()
	apiHandler.AddAPIEndpoints(muxRouter)
	return muxRouter
}

func setupGlobal(t *testing.T, rootNode string) {
	//sets up global.RootNode = rootnode and initialized the git repo
	cfgweb := &config.CfgWebAPI{Listen: ":8080"}
	global.Start(rootNode, cfgweb)
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

/********************** Test all the /policy handlers *********************/
func TestHeadPoliciesHandler(t *testing.T) {
	test := "TestHeadPoliciesHandler"
	begin(t, test)
	defer end(t, test)

	dbFile := "test_headpolicies.db"
	dbHandle := setupDB(t, dbFile)
	defer os.Remove(dbFile)
	defer dbHandle.Close()

	cmEngine := getCME(t)
	defer cleanupCME(cmEngine)
	muxRouter := setupApiHandler(dbHandle)
	//setup signature for change management
	sig := &change.CMAuthor{
		Name:  "tester",
		Email: "tester@iti.com",
		When:  time.Now(),
	}

	//create our multiple policies, multiple policy files so that we can retrieve it
	commit := change.NewCMContent("Policy A")
	commit.Files["Rules"] = []byte(Cases[0].C)

	data := &change.ChangeData{
		ObjectType: change.POLICY,
		Content:    commit,
		Author:     sig,
	}
	commitId, err := cmEngine.VersionObject(data, "test message")
	if err != nil {
		testingError(t, test, err.Error())
	}
	//test 1: Test the route Now
	req := createNewRequest(t, "HEAD", "https://localhost:8080/policy", nil)
	writer := httptest.NewRecorder()

	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "Could not GET policy list. got code: %v", writer.Code)
	}
	if writer.Header().Get("X-Pbconf-Policy-LastCommitId") != commitId {
		testingError(t, test, "Test 1: Did not get the expected last commit id.")
	}
	//test Versioning
	req = createNewRequest(t, "HEAD", "https://localhost:8080/v2/policy", nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Should not get StatusOK, wrong version sent")
	}
	//this doesnt even reach our global.ProcessVersioning, it is not a route that is registered (only v[0-9]+ routes are registered)
	req = createNewRequest(t, "HEAD", "https://localhost:8080/v1.0/policy", nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Should not get StatusOK, version in wrong format sent")
	}
	req = createNewRequest(t, "HEAD", "https://localhost:8080/v1/policy", nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "Did not get StatusOK, version is now not 1?")
	}
}

func TestGetPoliciesHandler(t *testing.T) {
	test := "TestGetPoliciesHandler"
	begin(t, test)
	defer end(t, test)

	dbFile := "test_getpolicies.db"
	dbHandle := setupDB(t, dbFile)
	defer os.Remove(dbFile)
	defer dbHandle.Close()

	cmEngine := getCME(t)
	defer cleanupCME(cmEngine)
	muxRouter := setupApiHandler(dbHandle)

	//setup signature for change management
	sig := &change.CMAuthor{
		Name:  "tester",
		Email: "tester@iti.com",
		When:  time.Now(),
	}

	//create our multiple policies, multiple policy files so that we can retrieve it
	commit := change.NewCMContent("Policy A")
	commit.Files["Rules"] = []byte(Cases[0].C)

	data := &change.ChangeData{
		ObjectType: change.POLICY,
		Content:    commit,
		Author:     sig,
	}

	_, err := cmEngine.VersionObject(data, "test message")
	if err != nil {
		testingError(t, test, err.Error())
	}

	commit2 := change.NewCMContent("Policy B")
	commit2.Files["Rules"] = []byte(Cases[1].C)

	data2 := &change.ChangeData{
		ObjectType: change.POLICY,
		Content:    commit2,
		Author:     sig,
	}
	_, err = cmEngine.VersionObject(data2, "test message 2")
	if err != nil {
		testingError(t, test, err.Error())
	}

	//test1: now test our route
	req := createNewRequest(t, "GET", "https://localhost:8080/policy", nil)
	writer := httptest.NewRecorder()

	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "Could not GET policy list. got code: %v", writer.Code)
	}

	//confirm names of Policies
	var polList []string
	decoder := json.NewDecoder(writer.Body)
	err = decoder.Decode(&polList)
	if err != nil {
		testingError(t, test, "Decoder error: %s", err.Error())
	}
	if len(polList) != 2 {
		testingError(t, test, "Did not get back two policies")
	}
	for _, pol := range polList {
		if pol != "Policy A" && pol != "Policy B" {
			testingError(t, test, "Did not get back expected policies")
		}
	}
	//test Versioning
	req = createNewRequest(t, "GET", "https://localhost:8080/v2/policy", nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Should not get StatusOK, wrong version sent")
	}
	//this doesnt even reach our global.ProcessVersioning, it is not a route that is registered (only v[0-9]+ routes are registered)
	req = createNewRequest(t, "GET", "https://localhost:8080/v1.0/policy", nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Should not get StatusOK, version in wrong format sent")
	}
	req = createNewRequest(t, "GET", "https://localhost:8080/v1/policy", nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "Did not get StatusOK, version is now not 1?")
	}
}

func TestPostBaseHandler(t *testing.T) {
	test := "TestPostBaseHandler"
	begin(t, test)
	defer end(t, test)

	dbFile := "test_postBasePolicy.db"
	dbHandle := setupDB(t, dbFile)
	defer os.Remove(dbFile)
	defer dbHandle.Close()
	cmEngine := getCME(t)
	defer cleanupCME(cmEngine)
	muxRouter := setupApiHandler(dbHandle)

	//setup root node, the post check to see if it has an upstream node for policy validation against ontology
	setupGlobal(t, "root")
	node := pbdatabase.PbNode{Name: "root"}
	if err := node.Create(dbHandle); err != nil {
		testingError(t, test, "setup root: "+err.Error())
	}

	//need to set up a user too.
	user := pbdatabase.PbUser{Name: "tester", Email: "tester@iti.com", Password: "notreally"}
	if err := user.Create(dbHandle); err != nil {
		testingError(t, test, err.Error())
	}
	//test1: now test our route without email, should be retrieved from database
	//create our multiple policies, multiple policy files so that we can retrieve it
	polAContents := change.CMContent{Object: "Policy A"}
	polAContents.Files = make(map[string][]byte)

	// Grab a test case from the parser tests
	polAContents.Files["Rules"] = []byte(Cases[0].C)

	policy := change.ChangeData{ObjectType: change.POLICY,
		Content: &polAContents,
		Author: &change.CMAuthor{
			Name:  "tester",
			Email: "tester@iti.com",
			When:  time.Now(),
		},
		Log: &change.LogLine{Message: "test commit"},
	}

	jsonStr, err := json.Marshal(policy)
	if err != nil {
		testingError(t, test, "Error marshalling policy")
	}

	req := createNewRequest(t, "POST", "https://localhost:8080/policy", bytes.NewBuffer(jsonStr))
	writer := httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "Could not save policy in the repository.")
	}
	time.Sleep(1000 * time.Millisecond) //give time for the spawned push to repo to succeed
	//confirm contents of the policy file
	retrvdPol, err := cmEngine.GetObject(change.POLICY, "Policy A")
	if err != nil {
		testingError(t, test, err.Error())
	}

	if bytes.Compare(retrvdPol.Content.Files["Rules"], polAContents.Files["Rules"]) != 0 {
		testingError(t, test, "contents of policy file does not match")
	}

	//test2 try updating the contents of the policy file
	// again, borrow from the DSL cases
	policy.Content.Files["Rules"] = []byte(Cases[1].C)
	policy.Log.Message = "Blippity commit message"
	policy.Author.When = time.Now()

	jsonStr, err = json.Marshal(policy)
	if err != nil {
		testingError(t, test, "test2: Error marshaling config contents")
	}

	req = createNewRequest(t, "POST", "https://localhost:8080/policy", bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "test 2: Could not update policy information.")
	}
	time.Sleep(1000 * time.Millisecond) //give time for the spawned push to repo to succeed

	//confirm contents of the policy file
	retrvdPol, err = cmEngine.GetObject(change.POLICY, "Policy A")
	if err != nil {
		testingError(t, test, "test 2: "+err.Error())
	}
	if bytes.Compare(retrvdPol.Content.Files["Rules"], policy.Content.Files["Rules"]) != 0 {
		testingError(t, test, "contents of policy file does not match")
	}
	//test3 try to mock decoder Error
	jsonStr = []byte(``)
	req = createNewRequest(t, "POST", "https://localhost:8080/policy", bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusBadRequest {
		testingError(t, test, "test 3: Should not be able to update policy information with decoder error.")
	}
	//test4 send just the user name, and see if email is retrieved from the Database
	policy = change.ChangeData{ObjectType: change.POLICY,
		Content: &polAContents,
		Author: &change.CMAuthor{
			Name: "tester",
		},
		Log: &change.LogLine{Message: "test commit for user without email"},
	}
	policy.Content.Files["Rules"] = []byte(Cases[0].C)
	policy.Log.Message = "tester commited message"
	jsonStr, err = json.Marshal(policy)
	if err != nil {
		testingError(t, test, "test4: Error marshaling config contents")
	}

	req = createNewRequest(t, "POST", "https://localhost:8080/policy", bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "test 4: Could not update policy information.")
	}
	time.Sleep(1000 * time.Millisecond) //give time for the spawned push to repo to succeed

	//test Versioning
	policy.Content.Files["Rules"] = []byte(Cases[0].C)
	policy.Log.Message = "Versioning test commit message"
	jsonStr, err = json.Marshal(policy)
	if err != nil {
		testingError(t, test, "test5-7: Error marshaling config contents")
	}
	req = createNewRequest(t, "POST", "https://localhost:8080/v2/policy", bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Test 5: Should not get StatusOK, wrong version sent")
	}
	//this doesnt even reach our global.ProcessVersioning, it is not a route that is registered (only v[0-9]+ routes are registered)
	req = createNewRequest(t, "POST", "https://localhost:8080/v1.2/policy", bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Test 6: Should not get StatusOK, version in wrong format sent")
	}
	req = createNewRequest(t, "POST", "https://localhost:8080/v1/policy", bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "Test 7:Did not get StatusOK, version is now not 1?")
	}
	time.Sleep(1000 * time.Millisecond) //give time for the spawned push to repo to succeed
}

/******************** Test all the /policy/all handlers *********************/
func TestGetAllPoliciesHandler(t *testing.T) {
	test := "TestGetAllPoliciesHandler"
	begin(t, test)
	defer end(t, test)

	dbFile := "test_getallpolicies.db"
	dbHandle := setupDB(t, dbFile)
	defer os.Remove(dbFile)
	defer dbHandle.Close()

	cmEngine := getCME(t)
	defer cleanupCME(cmEngine)
	muxRouter := setupApiHandler(dbHandle)

	//setup signature for change management
	sig := &change.CMAuthor{
		Name:  "tester",
		Email: "tester@iti.com",
		When:  time.Now(),
	}

	//create our multiple policies, multiple policy files so that we can retrieve it
	commit := change.NewCMContent("Policy A")
	commit.Files["Rules"] = []byte(Cases[0].C)

	data := &change.ChangeData{
		ObjectType: change.POLICY,
		Content:    commit,
		Author:     sig,
	}

	_, err := cmEngine.VersionObject(data, "test message")
	if err != nil {
		testingError(t, test, err.Error())
	}

	commit2 := change.NewCMContent("Policy B")
	commit2.Files["Rules"] = []byte(Cases[1].C)

	data2 := &change.ChangeData{
		ObjectType: change.POLICY,
		Content:    commit2,
		Author:     sig,
	}
	_, err = cmEngine.VersionObject(data2, "test message 2")
	if err != nil {
		testingError(t, test, err.Error())
	}

	//test1: now test our route
	req := createNewRequest(t, "GET", "https://localhost:8080/policy/all", nil)
	writer := httptest.NewRecorder()

	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "Could not GET policy list. got code: %v", writer.Code)
	}
	//confirm names of Policies
	polList := make([]change.ChangeData, 0)
	gzipReader, err := gzip.NewReader(writer.Body) //we expect a gzip compressed response
	if err != nil {
		testingError(t, test, err.Error())
	}
	decoder := json.NewDecoder(gzipReader)
	err = decoder.Decode(&polList)
	if err != nil {
		testingError(t, test, "Decoder error: %s", err.Error())
	}
	if len(polList) != 2 {
		testingError(t, test, "Did not get back two policies")
	}
	for _, pol := range polList {
		if pol.Content.Object != "Policy A" && pol.Content.Object != "Policy B" {
			testingError(t, test, "Did not get back expected policies")
		}
		if pol.Content.Object == "Policy A" && bytes.Compare(pol.Content.Files["Rules"], commit.Files["Rules"]) != 0 {
			testingError(t, test, "Contents of Policy A file does not match what was set")
		}
		if pol.Content.Object == "Policy B" && bytes.Compare(pol.Content.Files["Rules"], commit2.Files["Rules"]) != 0 {
			testingError(t, test, "Contents of Policy B file does not match what was set")
		}
	}
	//test Versioning
	req = createNewRequest(t, "GET", "https://localhost:8080/v2/policy/all", nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Should not get StatusOK, wrong version sent")
	}
	//this doesnt even reach our global.ProcessVersioning, it is not a route that is registered (only v[0-9]+ routes are registered)
	req = createNewRequest(t, "GET", "https://localhost:8080/v1.0/policy/all", nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Should not get StatusOK, version in wrong format sent")
	}
	req = createNewRequest(t, "GET", "https://localhost:8080/v1/policy/all", nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "Did not get StatusOK, version is now not 1?")
	}
}

/******************** Test all the /policy/{policyname} handlers *********************/
func TestGetWIdHandler(t *testing.T) {
	test := "TestGetWIdHandler"
	begin(t, test)
	defer end(t, test)

	dbFile := "test_getWIdPolicy.db"
	dbHandle := setupDB(t, dbFile)
	defer os.Remove(dbFile)
	defer dbHandle.Close()
	cmEngine := getCME(t)
	defer cleanupCME(cmEngine)
	muxRouter := setupApiHandler(dbHandle)

	//setup root node, the patch check to see if it has an upstream node for policy validation against ontology
	setupGlobal(t, "root")
	node := pbdatabase.PbNode{Name: "root"}
	if err := node.Create(dbHandle); err != nil {
		testingError(t, test, "setup root: "+err.Error())
	}
	//set up a signature to put into our repo
	sig := &change.CMAuthor{
		Name:  "tester",
		Email: "tester@iti.com",
		When:  time.Now(),
	}
	//set up policies in our repo to recover through the get Request
	//create our multiple policies, multiple policy files so that we can retrieve it
	policy1 := change.NewCMContent("Policy A")
	policy1.Files["Rules"] = []byte(Cases[0].C)

	data := &change.ChangeData{
		ObjectType: change.POLICY,
		Content:    policy1,
		Author:     sig,
	}
	_, err := cmEngine.VersionObject(data, "test message")
	if err != nil {
		testingError(t, test, err.Error())
	}

	pol2 := change.NewCMContent("Policy B")
	pol2.Files["Rules"] = []byte(Cases[1].C)

	data2 := &change.ChangeData{
		ObjectType: change.POLICY,
		Content:    pol2,
		Author:     sig,
	}
	_, err = cmEngine.VersionObject(data2, "test message 2")
	if err != nil {
		testingError(t, test, err.Error())
	}

	//now try to recover a policy through the route
	req := createNewRequest(t, "GET", "https://localhost:8080/policy/Policy B", nil)
	writer := httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "Could get policy from the repository.")
	}
	//confirm the contents of the recovered policy
	//confirm contents of the config file
	var retrievedPol change.CMContent
	decoder := json.NewDecoder(writer.Body)
	err = decoder.Decode(&retrievedPol)
	if err != nil {
		testingError(t, test, "Decoder error: %s", err.Error())
	}
	if bytes.Compare(retrievedPol.Files["Rules"], pol2.Files["Rules"]) != 0 {
		testingError(t, test, "contents of configuration file does not match")
	}
	//test Versioning
	req = createNewRequest(t, "GET", "https://localhost:8080/v2/policy/Policy B", nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Should not get StatusOK, wrong version sent")
	}
	//this doesnt even reach our global.ProcessVersioning, it is not a route that is registered (only v[0-9]+ routes are registered)
	req = createNewRequest(t, "GET", "https://localhost:8080/v1.0/policy/Policy B", nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Should not get StatusOK, version in wrong format sent")
	}
	req = createNewRequest(t, "GET", "https://localhost:8080/v1/policy/Policy B", nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "Did not get StatusOK, version is now not 1?")
	}
}

func TestDeleteWIdHandler(t *testing.T) {
	test := "TestDeleteWIdHandler"
	begin(t, test)
	defer end(t, test)

	dbFile := "test_delWIdPolicy.db"
	dbHandle := setupDB(t, dbFile)
	defer os.Remove(dbFile)
	defer dbHandle.Close()
	cmEngine := getCME(t)
	defer cleanupCME(cmEngine)
	muxRouter := setupApiHandler(dbHandle)

	//setup root node, the patch check to see if it has an upstream node for policy validation against ontology
	setupGlobal(t, "root")
	node := pbdatabase.PbNode{Name: "root"}
	if err := node.Create(dbHandle); err != nil {
		testingError(t, test, "setup root: %s", err.Error())
	}
	//set up a signature to put into our repo
	sig := &change.CMAuthor{
		Name:  "tester",
		Email: "tester@iti.com",
		When:  time.Now(),
	}
	//set up policies in our repo to delete later through the route
	//create our multiple policies, multiple policy files so that we can retrieve it
	policy1 := change.NewCMContent("Policy A")
	policy1.Files["Rules"] = []byte(Cases[0].C)

	data := &change.ChangeData{
		ObjectType: change.POLICY,
		Content:    policy1,
		Author:     sig,
	}
	_, err := cmEngine.VersionObject(data, "test message")
	if err != nil {
		testingError(t, test, err.Error())
	}

	pol2 := change.NewCMContent("Policy B")
	pol2.Files["Rules"] = []byte(Cases[1].C)

	data2 := &change.ChangeData{
		ObjectType: change.POLICY,
		Content:    pol2,
		Author:     sig,
	}
	_, err = cmEngine.VersionObject(data2, "test message 2")
	if err != nil {
		testingError(t, test, err.Error())
	}

	pol3 := change.NewCMContent("Policy C")
	pol3.Files["Rules"] = []byte(Cases[2].C)

	data3 := &change.ChangeData{
		ObjectType: change.POLICY,
		Content:    pol3,
		Author:     sig,
	}
	_, err = cmEngine.VersionObject(data3, "test message 3")
	if err != nil {
		testingError(t, test, err.Error())
	}
	//confirm three policies are in the cme
	policies, err := cmEngine.ListObjects(change.POLICY)
	if err != nil {
		testingError(t, test, err.Error())
	}
	if len(policies) != 3 {
		testingError(t, test, "Have not managed to set up the policies in the CME correctly to run the test.")
	}

	//now try to delete a policy through the route
	req := createNewRequest(t, "DELETE", "https://localhost:8080/policy/Policy B", nil)
	writer := httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "Could get policy from the repository.")
	}
	policies, err = cmEngine.ListObjects(change.POLICY)
	if err != nil {
		testingError(t, test, err.Error())
	}
	if len(policies) != 2 {
		testingError(t, test, "Did not manage to DELETE the policy in the CME correctly")
	}
	//test Versioning
	req = createNewRequest(t, "DELETE", "https://localhost:8080/v2/policy/Policy C", nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Should not get StatusOK, wrong version sent")
	}
	//this doesnt even reach our global.ProcessVersioning, it is not a route that is registered (only v[0-9]+ routes are registered)
	req = createNewRequest(t, "DELETE", "https://localhost:8080/v1.0/policy/Policy C", nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Should not get StatusOK, version in wrong format sent")
	}
	req = createNewRequest(t, "DELETE", "https://localhost:8080/v1/policy/Policy C", nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "Did not get StatusOK, version is now not 1?")
	}
}

func TestPutWIdHandler(t *testing.T) {
	test := "TestPutWIdHandler"
	begin(t, test)
	defer end(t, test)

	dbFile := "test_putWIdPolicy.db"
	dbHandle := setupDB(t, dbFile)
	defer os.Remove(dbFile)
	defer dbHandle.Close()
	cmEngine := getCME(t)
	defer cleanupCME(cmEngine)
	muxRouter := setupApiHandler(dbHandle)

	//setup root node, the patch check to see if it has an upstream node for policy validation against ontology
	setupGlobal(t, "root")
	node := pbdatabase.PbNode{Name: "root"}
	if err := node.Create(dbHandle); err != nil {
		testingError(t, test, "setup root: "+err.Error())
	}

	//need to set up a user too.
	user := pbdatabase.PbUser{Name: "tester", Email: "tester@iti.com", Password: "notreally"}
	if err := user.Create(dbHandle); err != nil {
		testingError(t, test, err.Error())
	}
	//test1: now test our route without email, should be retrieved from database
	//create our multiple policies, multiple policy files so that we can retrieve it
	polAContents := change.CMContent{Object: "Policy A"}
	polAContents.Files = make(map[string][]byte)
	polAContents.Files["Rules"] = []byte(Cases[0].C)

	policy := change.ChangeData{ObjectType: change.POLICY,
		Content: &polAContents,
		Author: &change.CMAuthor{
			Name:  "tester",
			Email: "tester@iti.com",
			When:  time.Now(),
		},
		Log: &change.LogLine{Message: "test commit"},
	}

	jsonStr, err := json.Marshal(policy)
	if err != nil {
		testingError(t, test, "Error marshalling policy")
	}

	req := createNewRequest(t, "PUT", fmt.Sprintf("https://localhost:8080/policy/%s", policy.Content.Object), bytes.NewBuffer(jsonStr))
	writer := httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "Could not save policy in the repository.")
	}
	time.Sleep(1000 * time.Millisecond) //give time for the spawned push to repo to succeed
	//confirm contents of the policy file
	retrievedPol, err := cmEngine.GetObject(change.POLICY, "Policy A")
	if err != nil {
		testingError(t, test, err.Error())
	}

	if bytes.Compare(retrievedPol.Content.Files["Rules"], policy.Content.Files["Rules"]) != 0 {
		testingError(t, test, "contents of policy file does not match")
	}

	//test2 try updating the contents of the policy file
	policy.Content.Files["Rules"] = []byte(Cases[1].C)
	policy.Log.Message = "Blippity commit message"
	jsonStr, err = json.Marshal(policy)
	if err != nil {
		testingError(t, test, "test2: Error marshaling config contents")
	}

	req = createNewRequest(t, "PUT", fmt.Sprintf("https://localhost:8080/policy/%s", policy.Content.Object), bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "test 2: Could not update policy information.")
	}
	time.Sleep(1000 * time.Millisecond) //give time for the spawned push to repo to succeed

	//confirm contents of the policy file
	retrievedPol, err = cmEngine.GetObject(change.POLICY, "Policy A")
	if err != nil {
		testingError(t, test, "test 2: "+err.Error())
	}
	if bytes.Compare(retrievedPol.Content.Files["Rules"], policy.Content.Files["Rules"]) != 0 {
		testingError(t, test, "test2: contents of policy file does not match")
	}

	//test 3 mock decoder Error
	jsonStr = []byte(``)
	req = createNewRequest(t, "PUT", fmt.Sprintf("https://localhost:8080/policy/%s", policy.Content.Object), bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusBadRequest {
		testingError(t, test, "test 3: Should not be able to update policy information with decoder error.")
	}
	//test4 send just the user name, and see if email is retrieved from the Database
	policy = change.ChangeData{ObjectType: change.POLICY,
		Content: &polAContents,
		Author: &change.CMAuthor{
			Name: "tester",
		},
		Log: &change.LogLine{Message: "test commit for user without email"},
	}
	policy.Content.Files["Rules"] = []byte(Cases[0].C)
	policy.Log.Message = "tester commited message"
	jsonStr, err = json.Marshal(policy)
	if err != nil {
		testingError(t, test, "test4: Error marshaling config contents")
	}

	req = createNewRequest(t, "PUT", fmt.Sprintf("https://localhost:8080/policy/%s", policy.Content.Object), bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "test 4: Could not update policy information.")
	}
	time.Sleep(1000 * time.Millisecond) //give time for the spawned push to repo to succeed

	//test Versioning
	policy.Content.Files["Rules"] = []byte(Cases[0].C)
	policy.Log.Message = "Versioning test commit message"
	jsonStr, err = json.Marshal(policy)
	if err != nil {
		testingError(t, test, "test5-7: Error marshaling config contents")
	}
	req = createNewRequest(t, "PUT", fmt.Sprintf("https://localhost:8080/v2/policy/%s", policy.Content.Object), bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Test 5: Should not get StatusOK, wrong version sent")
	}
	//this doesnt even reach our global.ProcessVersioning, it is not a route that is registered (only v[0-9]+ routes are registered)
	req = createNewRequest(t, "PUT", fmt.Sprintf("https://localhost:8080/v1.2/policy/%s", policy.Content.Object), bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Test 6: Should not get StatusOK, version in wrong format sent")
	}
	req = createNewRequest(t, "PUT", fmt.Sprintf("https://localhost:8080/v1/policy/%s", policy.Content.Object), bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "Test 7:Did not get StatusOK, version is now not 1?")
	}
	time.Sleep(1000 * time.Millisecond) //give time for the spawned push to repo to succeed
}

/****************** Test all the /policy//default/{nodename} handlers ****************/
func TestHeadDefaultHandler(t *testing.T) {
	test := "TestHeadDefaultHandler"
	begin(t, test)
	defer end(t, test)

	dbFile := "test_headDefaultPolicy.db"
	dbHandle := setupDB(t, dbFile)
	defer os.Remove(dbFile)
	defer dbHandle.Close()
	cmEngine := getCME(t)
	defer cleanupCME(cmEngine)
	muxRouter := setupApiHandler(dbHandle)

	//setup root node
	setupGlobal(t, "root")
	node := pbdatabase.PbNode{Name: "root"}
	if err := node.Create(dbHandle); err != nil {
		testingError(t, test, "setup root: "+err.Error())
	}

	//set up a signature to put into our repo
	sig := &change.CMAuthor{
		Name:  "tester",
		Email: "tester@iti.com",
		When:  time.Now(),
	}

	//set up policies in our repo
	commit := change.NewCMContent("Policy-1A")
	commit.Files["Rules"] = []byte(Cases[0].C)

	data := &change.ChangeData{
		ObjectType: change.POLICY,
		Content:    commit,
		Author:     sig,
	}
	_, err := cmEngine.VersionObject(data, "1A test message")
	if err != nil {
		testingError(t, test, err.Error())
	}

	commit2 := change.NewCMContent("Policy-BB")
	commit2.Files["Rules"] = []byte(Cases[1].C)

	data2 := &change.ChangeData{
		ObjectType: change.POLICY,
		Content:    commit2,
		Author:     sig,
	}
	commit2Id, err := cmEngine.VersionObject(data2, "BB test message")
	if err != nil {
		testingError(t, test, err.Error())
	}

	//test 1 setup a new child node through the HEAD request
	req := createNewRequest(t, "HEAD", "https://localhost:8080/policy/default/NewTestNode", nil)
	writer := httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "test 1: Could not add a new node via the HEAD heartbeat request.")
	}
	if writer.Header().Get("X-Pbconf-Policy-LastCommitId") != commit2Id {
		testingError(t, test, "test 1: Did not get the expected last commit id.")
	}

	//test2 modify policy in the repository and see if we get the right commit id through the route
	commit3 := change.NewCMContent("Policy-BB")
	commit3.Files["Rules"] = []byte(Cases[2].C)

	data3 := &change.ChangeData{
		ObjectType: change.POLICY,
		Content:    commit3,
		Author:     sig,
	}
	commit3Id, err := cmEngine.VersionObject(data3, "BB test message-Modified")
	if err != nil {
		testingError(t, test, "test 2: error updating to repo :"+err.Error())
	}
	req = createNewRequest(t, "HEAD", "https://localhost:8080/policy/default/NewTestNode", nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "test 2: Could not add a new node via the HEAD heartbeat request.")
	}

	if writer.Header().Get("X-Pbconf-Policy-LastCommitId") != commit3Id {
		testingError(t, test, "test 2: Did not get the expected last commit id.")
	}
	//test Versioning
	req = createNewRequest(t, "HEAD", "https://localhost:8080/v2/policy/default/NewTestNode", nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Should not get StatusOK, wrong version sent")
	}
	//this doesnt even reach our global.ProcessVersioning, it is not a route that is registered (only v[0-9]+ routes are registered)
	req = createNewRequest(t, "HEAD", "https://localhost:8080/v1.0/policy/default/NewTestNode", nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Should not get StatusOK, version in wrong format sent")
	}
	req = createNewRequest(t, "HEAD", "https://localhost:8080/v1/policy/default/NewTestNode", nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "Did not get StatusOK, version is now not 1?")
	}
}
