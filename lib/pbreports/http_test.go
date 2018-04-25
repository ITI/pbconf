package reports_test

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
	"github.com/iti/pbconf/lib/pbreports"
)

var logLevel = "DEBUG"
var rcfg *config.Config
var repoLocker sync.Mutex

func begin(t *testing.T, name string) {
	fmt.Printf("##################### Begin Reports::%s #####################\n", name)
}

func end(t *testing.T, name string) {
	fmt.Printf("###################### End Reports::%s ######################\n", name)
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
	var cmEngine *change.CMEngine
	var err error
	if rcfg == nil {
		setupRepo()
		cmEngine, err = change.GetCMEngine(rcfg)
	} else {
		cmEngine, err = change.GetCMEngine(rcfg)
	}
	if err != nil {
		t.Error("Failed to get CME ref")
	}
	return cmEngine
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
	apiHandler := reports.NewAPIHandler(logLevel, dbHandle)
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

func TestHeadBaseHandler(t *testing.T) {
	test := "TestHeadBaseHandler"
	begin(t, test)
	defer end(t, test)

	dbFile := "test_headReports.db"
	dbHandle := setupDB(t, dbFile)
	defer os.Remove(dbFile)
	defer dbHandle.Close()
	cmEngine := getCME(t)
	defer cleanupCME(cmEngine)
	muxRouter := setupApiHandler(dbHandle)
	setupGlobal(t, "root")
	//setup root node in database
	node := pbdatabase.PbNode{Name: "root"}
	if err := node.Create(dbHandle); err != nil {
		testingError(t, test, "TestGetBaseHandler setup root node: %s", err.Error())
	}

	//create some report files so that we can retrieve it
	sig := &change.CMAuthor{
		Name:  "tester",
		Email: "tester@iti.com",
		When:  time.Now(),
	}

	//first report query
	query := reports.PbReportQuery{Name: "List of devices Report", Query: "Dont know yet", TimeStamp: time.Now().Format(time.RFC1123), Period: "-1"}
	commit := change.NewCMContent(query.Name)
	jsonStr, err := json.Marshal(query)
	if err != nil {
		testingError(t, test, "Could not marshal the query, Error: %s", err.Error())
		return
	}
	commit.Files["queryFile"] = []byte(jsonStr)
	data := &change.ChangeData{
		ObjectType: change.REPORT,
		Content:    commit,
		Author:     sig,
	}
	commitId, err := cmEngine.VersionObject(data, "test message")
	//test route
	req := createNewRequest(t, "HEAD", "https://localhost:8080/reports", nil)
	writer := httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "Could not get HEAD reports. got code: %v", writer.Code)
	}
	//check the Header
	if writer.Header().Get("X-Pbconf-Report-LastCommitId") != commitId {
		testingError(t, test, "test 1: Did not get the expected last commit id.")
	}

	//second report
	query2 := reports.PbReportQuery{Name: "Device Compliance Report", Query: "Still Dont know yet", TimeStamp: time.Now().Format(time.RFC1123), Period: "-1"}
	commit2 := change.NewCMContent(query2.Name)
	jsonStr, err = json.Marshal(query2)
	if err != nil {
		testingError(t, test, "Could not marshal the query2, Error: %s", err.Error())
		return
	}
	commit2.Files["queryFile"] = []byte(jsonStr)
	data2 := &change.ChangeData{
		ObjectType: change.REPORT,
		Content:    commit2,
		Author:     sig,
	}
	commit2Id, err := cmEngine.VersionObject(data2, "test message2")
	req = createNewRequest(t, "HEAD", "https://localhost:8080/reports", nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "Could not get HEAD reports. got code: %v", writer.Code)
	}
	//check the Header
	if writer.Header().Get("X-Pbconf-Report-LastCommitId") != commit2Id {
		testingError(t, test, "test 1: Did not get the expected last commit id.")
	}
	//test Versioning
	req = createNewRequest(t, "HEAD", "https://localhost:8080/v2/reports", nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Should not get StatusOK, wrong version sent")
	}
	//this doesnt even reach our global.ProcessVersioning, it is not a route that is registered (only v[0-9]+ routes are registered)
	req = createNewRequest(t, "HEAD", "https://localhost:8080/v1.1/reports", nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Should not get StatusOK, version in wrong format sent")
	}
	req = createNewRequest(t, "HEAD", "https://localhost:8080/v1/reports", nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "Did not get StatusOK, version is now not 1?")
	}

}
func TestGetBaseHandler(t *testing.T) {
	test := "TestGetBaseHandler"
	begin(t, test)
	defer end(t, test)

	dbFile := "test_getReports.db"
	dbHandle := setupDB(t, dbFile)
	defer os.Remove(dbFile)
	defer dbHandle.Close()

	cmEngine := getCME(t)
	defer cleanupCME(cmEngine)
	muxRouter := setupApiHandler(dbHandle)
	setupGlobal(t, "root")
	//setup root node in database
	node := pbdatabase.PbNode{Name: "root"}
	if err := node.Create(dbHandle); err != nil {
		testingError(t, test, "TestGetBaseHandler setup root node: %s", err.Error())
	}

	//create some report files so that we can retrieve it
	sig := &change.CMAuthor{
		Name:  "tester",
		Email: "tester@iti.com",
		When:  time.Now(),
	}

	//first report query
	query := reports.PbReportQuery{Name: "List of devices Report", Query: "Dont know yet", TimeStamp: time.Now().Format(time.RFC1123), Period: "-1"}
	commit := change.NewCMContent(query.Name)
	jsonStr, err := json.Marshal(query)
	if err != nil {
		testingError(t, test, "Could not marshal the query, Error: %s", err.Error())
		return
	}
	commit.Files["queryFile"] = []byte(jsonStr)
	data := &change.ChangeData{
		ObjectType: change.REPORT,
		Content:    commit,
		Author:     sig,
	}
	_, err = cmEngine.VersionObject(data, "test message")
	//second report
	query2 := reports.PbReportQuery{Name: "Device Compliance Report", Query: "Still Dont know yet", TimeStamp: time.Now().Format(time.RFC1123), Period: "-1"}
	commit2 := change.NewCMContent(query2.Name)
	jsonStr, err = json.Marshal(query2)
	if err != nil {
		testingError(t, test, "Could not marshal the query2, Error: %s", err.Error())
		return
	}
	commit2.Files["queryFile"] = []byte(jsonStr)
	data2 := &change.ChangeData{
		ObjectType: change.REPORT,
		Content:    commit2,
		Author:     sig,
	}
	_, err = cmEngine.VersionObject(data2, "test message2")

	//test1: now test our route
	req := createNewRequest(t, "GET", "https://localhost:8080/reports", nil)
	writer := httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "Could not GET reports. got code: %v", writer.Code)
	}
	//confirm list of reports retrieved
	var retrievedReports []string
	decoder := json.NewDecoder(writer.Body)
	err = decoder.Decode(&retrievedReports)
	if err != nil {
		testingError(t, test, "Decoder error: %s", err.Error())
	}
	if len(retrievedReports) != 2 {
		testingError(t, test, "Did not GET correct number of reports")
	}
	for _, reportName := range retrievedReports {
		if !(reportName == "Device Compliance Report" || reportName == "List of devices Report") {
			testingError(t, test, "Got back unexpected report")
		}
	}
	//test Versioning
	req = createNewRequest(t, "GET", "https://localhost:8080/v2/reports", nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Should not get StatusOK, wrong version sent")
	}
	//this doesnt even reach our global.ProcessVersioning, it is not a route that is registered (only v[0-9]+ routes are registered)
	req = createNewRequest(t, "GET", "https://localhost:8080/v1.0/reports", nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Should not get StatusOK, version in wrong format sent")
	}
	req = createNewRequest(t, "GET", "https://localhost:8080/v1/reports", nil)
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

	dbFile := "test_postReports.db"
	dbHandle := setupDB(t, dbFile)
	defer os.Remove(dbFile)
	defer dbHandle.Close()

	cmEngine := getCME(t)
	defer cleanupCME(cmEngine)
	muxRouter := setupApiHandler(dbHandle)
	setupGlobal(t, "root")
	//setup root node in database
	node := pbdatabase.PbNode{Name: "root"}
	if err := node.Create(dbHandle); err != nil {
		testingError(t, test, "TestPostBaseHandler setup root node: %s", err.Error())
	}
	//need to set up a user too.
	user := pbdatabase.PbUser{Name: "tester", Email: "tester@iti.com", Password: "notreally"}
	if err := user.Create(dbHandle); err != nil {
		testingError(t, test, "db user create error:"+err.Error())
	}

	//first sample report file
	query := reports.PbReportQuery{Name: "List of devices Report", Query: "Select Name from db.Devices", TimeStamp: time.Now().Format(time.RFC1123), Period: "-1"}
	commit := change.NewCMContent(query.Name)
	jsonStr, err := json.Marshal(query)
	if err != nil {
		testingError(t, test, "Could not marshal the query, Error: %s", err.Error())
		return
	}
	commit.Files["queryFile"] = []byte(jsonStr)

	payload := reports.PbReportQueryHttp{PbReportQuery: query, Author: "tester", CommitMessage: "Something something"}
	jsonStr, err = json.Marshal(payload)
	if err != nil {
		testingError(t, test, "Could not marshal the message payload, Error: %s", err.Error())
		return
	}
	//test1: test the route
	req := createNewRequest(t, "POST", "https://localhost:8080/reports", bytes.NewBuffer(jsonStr))
	writer := httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK && writer.Code != http.StatusCreated {
		testingError(t, test, "Could not POST to reports. got code: ", writer.Code)
		return
	}
	time.Sleep(1000 * time.Millisecond) //give time for the spawned push to repo to succeed
	//test if it is in the cme
	retrievedReportObj, err := cmEngine.GetObject(change.REPORT, query.Name)
	retrievedReport := retrievedReportObj.Content
	if retrievedReport.Object != query.Name {
		testingError(t, test, "Saved object names don't match")
	}
	var retrievedQuery reports.PbReportQuery
	decoder := json.NewDecoder(bytes.NewBuffer(retrievedReport.Files["queryFile"]))
	err = decoder.Decode(&retrievedQuery)
	if err != nil {
		testingError(t, test, "Decoder error: %s", err.Error())
	}
	if query.Query != retrievedQuery.Query || query.TimeStamp != retrievedQuery.TimeStamp || query.Period != retrievedQuery.Period || query.Name != retrievedQuery.Name {
		testingError(t, test, "POSTed contents don't match")
	}

	//test2: Mock decoder Error
	jsonStr = []byte(``)
	req = createNewRequest(t, "POST", "https://localhost:8080/reports", bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK || writer.Code == http.StatusCreated {
		testingError(t, test, "Test2: Should not be able to POST report with decoder error.")
	}
	//test3: mock validate error
	jsonStr2 := []byte(`{"Name":"List of devices Report","Query":"Select * from Devices", "TimeStamp": "Thu, 17 Dec 2015 10:33:55 CST", "Period" : "-1", "Author": "tester", "CommitMessage":"blech"}`)
	req = createNewRequest(t, "POST", "https://localhost:8080/reports", bytes.NewBuffer(jsonStr2))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK || writer.Code == http.StatusCreated {
		testingError(t, test, "Test3: Should not be able to POST report with validate error.")
	}
	//test4 send with invalid author
	jsonStr2 = []byte(`{"Name":"List of devices Report","Query":"Select * from db.Devices", "TimeStamp": "Thu, 17 Dec 2015 10:33:55 CST", "Period" : "-1", "Author": "bbb", "CommitMessage":"fmv"}`)
	req = createNewRequest(t, "POST", "https://localhost:8080/reports", bytes.NewBuffer(jsonStr2))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK || writer.Code == http.StatusCreated {
		testingError(t, test, "Test4: Should not be able to POST report with validate error.")
	}
	//test Versioning
	query2 := reports.PbReportQuery{Name: "Second Report", Query: "Select Id from db.Devices", TimeStamp: time.Now().Format(time.RFC1123), Period: "-1"}
	commit2 := change.NewCMContent(query2.Name)
	jsonStr2, err = json.Marshal(query2)
	if err != nil {
		testingError(t, test, "Test 5-7:Could not marshal the query, Error: %s", err.Error())
		return
	}
	commit2.Files["queryFile"] = []byte(jsonStr2)

	payload2 := reports.PbReportQueryHttp{PbReportQuery: query2, Author: "tester", CommitMessage: "Versioning something"}
	jsonStr2, err = json.Marshal(payload2)
	if err != nil {
		testingError(t, test, "Test 5-7: Could not marshal the message payload, Error: %s", err.Error())
		return
	}
	req = createNewRequest(t, "POST", "https://localhost:8080/v2/reports", bytes.NewBuffer(jsonStr2))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Should not get StatusOK, wrong version sent")
	}
	//this doesnt even reach our global.ProcessVersioning, it is not a route that is registered (only v[0-9]+ routes are registered)
	req = createNewRequest(t, "POST", "https://localhost:8080/v1.0/reports", bytes.NewBuffer(jsonStr2))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Should not get StatusOK, version in wrong format sent")
	}
	req = createNewRequest(t, "POST", "https://localhost:8080/v1/reports", bytes.NewBuffer(jsonStr2))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK && writer.Code != http.StatusCreated {
		testingError(t, test, "Did not get StatusOK, version is now not 1?")
	}
}

func TestGetWIdHandler(t *testing.T) {
	test := "TestGetWIdHandler"
	begin(t, test)
	defer end(t, test)

	dbFile := "test_getWIdReport.db"
	dbHandle := setupDB(t, dbFile)
	defer os.Remove(dbFile)
	defer dbHandle.Close()

	cmEngine := getCME(t)
	defer cleanupCME(cmEngine)
	muxRouter := setupApiHandler(dbHandle)
	setupGlobal(t, "root")
	//setup root node in database
	node := pbdatabase.PbNode{Name: "root"}
	if err := node.Create(dbHandle); err != nil {
		testingError(t, test, "setup root node: %s", err.Error())
	}

	//create some report files so that we can retrieve it
	sig := &change.CMAuthor{
		Name:  "tester",
		Email: "tester@iti.com",
		When:  time.Now(),
	}

	//first report query
	query := reports.PbReportQuery{Name: "List of devices Report", Query: "Dont know yet", TimeStamp: time.Now().Format(time.RFC1123), Period: "-1"}
	commit := change.NewCMContent(query.Name)
	jsonStr, err := json.Marshal(query)
	if err != nil {
		testingError(t, test, "Could not marshal the query, Error: %s", err.Error())
		return
	}
	commit.Files["queryFile"] = []byte(jsonStr)
	data := &change.ChangeData{
		ObjectType: change.REPORT,
		Content:    commit,
		Author:     sig,
	}
	_, err = cmEngine.VersionObject(data, "test message")
	//second report
	query2 := reports.PbReportQuery{Name: "Device Compliance Report", Query: "Still Dont know yet", TimeStamp: time.Now().Format(time.RFC1123), Period: "-1"}
	commit2 := change.NewCMContent(query2.Name)
	jsonStr, err = json.Marshal(query2)
	if err != nil {
		testingError(t, test, "Could not marshal the query2, Error: %s", err.Error())
		return
	}
	commit2.Files["queryFile"] = []byte(jsonStr)
	data2 := &change.ChangeData{
		ObjectType: change.REPORT,
		Content:    commit2,
		Author:     sig,
	}
	_, err = cmEngine.VersionObject(data2, "test message2")
	//test1: Now test our route
	req := createNewRequest(t, "GET", fmt.Sprintf("https://localhost:8080/reports/%s", query.Name), nil)
	writer := httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "Could not GET reports. got code: %v", writer.Code)
	}
	var retrievedQuery reports.PbReportQuery
	decoder := json.NewDecoder(writer.Body)
	err = decoder.Decode(&retrievedQuery)
	if err != nil {
		testingError(t, test, "Decoder error: %s", err.Error())
	}
	if query.Query != retrievedQuery.Query || query.TimeStamp != retrievedQuery.TimeStamp || query.Period != retrievedQuery.Period || query.Name != retrievedQuery.Name {
		testingError(t, test, "GetById contents don't match")
	}
	//test Versioning
	req = createNewRequest(t, "GET", fmt.Sprintf("https://localhost:8080/v2/reports/%s", query.Name), nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Should not get StatusOK, wrong version sent")
	}
	//this doesnt even reach our global.ProcessVersioning, it is not a route that is registered (only v[0-9]+ routes are registered)
	req = createNewRequest(t, "GET", fmt.Sprintf("https://localhost:8080/v1.0/reports/%s", query.Name), nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Should not get StatusOK, version in wrong format sent")
	}
	req = createNewRequest(t, "GET", fmt.Sprintf("https://localhost:8080/v1/reports/%s", query.Name), nil)
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

	dbFile := "test_putWIdReport.db"
	dbHandle := setupDB(t, dbFile)
	defer os.Remove(dbFile)
	defer dbHandle.Close()

	cmEngine := getCME(t)
	defer cleanupCME(cmEngine)
	muxRouter := setupApiHandler(dbHandle)
	setupGlobal(t, "root")
	//setup root node in database
	node := pbdatabase.PbNode{Name: "root"}
	if err := node.Create(dbHandle); err != nil {
		testingError(t, test, "setup root node: %s", err.Error())
	}
	//need to set up a user too.
	user := pbdatabase.PbUser{Name: "tester", Email: "tester@iti.com", Password: "notreally"}
	if err := user.Create(dbHandle); err != nil {
		testingError(t, test, "db User create error: "+err.Error())
	}

	//first sample report file
	query := reports.PbReportQuery{Name: "List of devices Report", Query: "SELECT Name FROM db.Devices", TimeStamp: time.Now().Format(time.RFC1123), Period: "-1"}
	payload := reports.PbReportQueryHttp{PbReportQuery: query, Author: "tester", CommitMessage: "first message"}
	jsonStr, err := json.Marshal(payload)
	if err != nil {
		testingError(t, test, "Could not marshal the message payload, Error: %s", err.Error())
		return
	}
	//test1: test the route
	req := createNewRequest(t, "PUT", fmt.Sprintf("https://localhost:8080/reports/%s", query.Name), bytes.NewBuffer(jsonStr))
	writer := httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK && writer.Code != http.StatusCreated {
		testingError(t, test, "Could not PUT to reports. got code: %s", writer.Code)
	}
	time.Sleep(1000 * time.Millisecond) //give time for the spawned push to repo to succeed
	//test if it is in the cme
	retrievedReportObj, err := cmEngine.GetObject(change.REPORT, query.Name)
	retrievedReport := retrievedReportObj.Content
	if retrievedReport.Object != query.Name {
		testingError(t, test, "Saved object names don't match")
	}
	var retrievedQuery reports.PbReportQuery
	decoder := json.NewDecoder(bytes.NewBuffer(retrievedReport.Files["queryFile"]))
	err = decoder.Decode(&retrievedQuery)
	if err != nil {
		testingError(t, test, "Decoder error: %s", err.Error())
	}
	if query.Query != retrievedQuery.Query || query.TimeStamp != retrievedQuery.TimeStamp || query.Period != retrievedQuery.Period || query.Name != retrievedQuery.Name {
		testingError(t, test, "contents don't match")
	}

	//test2: Mock decoder Error
	jsonStr = []byte(``)
	req = createNewRequest(t, "PUT", fmt.Sprintf("https://localhost:8080/reports/%s", query.Name), bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK || writer.Code == http.StatusCreated {
		testingError(t, test, "Test2: Should not be able to PUT report with decoder error.")
	}
	//test3: mock validate error
	jsonStr2 := []byte(`{"Name":"List of devices Report","Query":"", "TimeStamp": "Thu, 17 Dec 2015 10:33:55 CST", "Period" : "-1", "Author": "tester", "CommitMessage":"second"}`)
	req = createNewRequest(t, "PUT", fmt.Sprintf("https://localhost:8080/reports/%s", query.Name), bytes.NewBuffer(jsonStr2))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK || writer.Code == http.StatusCreated {
		testingError(t, test, "Test3: Should not be able to PUT report with validate error.")
	}
	//test4 send with invalid author
	jsonStr2 = []byte(`{"Name":"List of devices Report","Query":"select Name from db.Nodes", "TimeStamp": "Thu, 17 Dec 2015 10:33:55 CST", "Period" : "-1", "Author": "bbb", "CommitMessage":"third"}`)
	req = createNewRequest(t, "PUT", fmt.Sprintf("https://localhost:8080/reports/%s", query.Name), bytes.NewBuffer(jsonStr2))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK || writer.Code == http.StatusCreated {
		testingError(t, test, "Test4: Should not be able to PUT report with validate error.")
	}
	//test5 try periodic reports
	query2 := reports.PbReportQuery{Name: "List of devices Report", Query: "SELECT Name FROM db.Devices", TimeStamp: time.Now().Format(time.RFC1123), Period: "10s"}
	payload = reports.PbReportQueryHttp{PbReportQuery: query2, Author: "tester", CommitMessage: "see message"}
	jsonStr, err = json.Marshal(payload)
	if err != nil {
		testingError(t, test, "Could not marshal the message payload, Error: %s", err.Error())
		return
	}
	req = createNewRequest(t, "PUT", fmt.Sprintf("https://localhost:8080/reports/%s", query2.Name), bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK && writer.Code != http.StatusCreated {
		testingError(t, test, "Test5: was not be able to PUT report with periodic report function.")
	}
	time.Sleep(30000 * time.Millisecond) //give time for the periodic function to be called repetitively

	//test6: try to trigger parse duration Error
	query2 = reports.PbReportQuery{Name: "New devices Report", Query: "SELECT Name FROM db.Devices", TimeStamp: time.Now().Format(time.RFC1123), Period: "10ss"}
	payload = reports.PbReportQueryHttp{PbReportQuery: query2, Author: "tester", CommitMessage: "see message"}
	jsonStr, err = json.Marshal(payload)
	if err != nil {
		testingError(t, test, "Test6: Could not marshal the message payload, Error: %s", err.Error())
		return
	}
	req = createNewRequest(t, "PUT", fmt.Sprintf("https://localhost:8080/reports/%s", query2.Name), bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK || writer.Code == http.StatusCreated {
		testingError(t, test, "Test6: Should not be able to PUT report with Illegal period report function.")
	}
	//test7 send with nil author
	jsonStr2 = []byte(`{"Name":"Test7 report","Query":"select Name from db.Nodes", "TimeStamp": "Thu, 17 Dec 2015 10:33:55 CST", "Period" : "-1", "Author": "", "CommitMessage":"third"}`)
	req = createNewRequest(t, "PUT", fmt.Sprintf("https://localhost:8080/reports/%s", query.Name), bytes.NewBuffer(jsonStr2))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK || writer.Code == http.StatusCreated {
		testingError(t, test, "Test7: Should not be able to PUT report with no author.")
	}
	//test Versioning
	query3 := reports.PbReportQuery{Name: "Versioning devices Report", Query: "SELECT * FROM db.Devices", TimeStamp: time.Now().Format(time.RFC1123), Period: "10s"}
	payload = reports.PbReportQueryHttp{PbReportQuery: query3, Author: "tester", CommitMessage: "versioning message"}
	jsonStr, err = json.Marshal(payload)
	if err != nil {
		testingError(t, test, "Test 8-10: Could not marshal the message payload, Error: %s", err.Error())
		return
	}
	req = createNewRequest(t, "PUT", fmt.Sprintf("https://localhost:8080/v2/reports/%s", query.Name), bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Test 8: Should not get StatusOK, wrong version sent")
	}
	//this doesnt even reach our global.ProcessVersioning, it is not a route that is registered (only v[0-9]+ routes are registered)
	req = createNewRequest(t, "PUT", fmt.Sprintf("https://localhost:8080/v1.0/reports/%s", query.Name), bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Test 9: Should not get StatusOK, version in wrong format sent")
	}
	req = createNewRequest(t, "PUT", fmt.Sprintf("https://localhost:8080/v1/reports/%s", query.Name), bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK && writer.Code != http.StatusCreated {
		testingError(t, test, "Test 10: Did not get StatusOK, version is now not 1?")
	}
}

func TestDeleteWIdHandler(t *testing.T) {
	test := "TestDeleteWIdHandler"
	begin(t, test)
	defer end(t, test)

	dbFile := "test_getWIdReport.db"
	dbHandle := setupDB(t, dbFile)
	defer os.Remove(dbFile)
	defer dbHandle.Close()

	cmEngine := getCME(t)
	defer cleanupCME(cmEngine)
	muxRouter := setupApiHandler(dbHandle)
	setupGlobal(t, "root")
	//setup root node in database
	node := pbdatabase.PbNode{Name: "root"}
	if err := node.Create(dbHandle); err != nil {
		testingError(t, test, "setup root node: %s", err.Error())
	}

	//create some report files so that we can retrieve it
	sig := &change.CMAuthor{
		Name:  "tester",
		Email: "tester@iti.com",
		When:  time.Now(),
	}

	//first report query
	query := reports.PbReportQuery{Name: "List of devices Report", Query: "Dont know yet", TimeStamp: time.Now().Format(time.RFC1123), Period: "-1"}
	commit := change.NewCMContent(query.Name)
	jsonStr, err := json.Marshal(query)
	if err != nil {
		testingError(t, test, "Could not marshal the query, Error: %s", err.Error())
		return
	}
	commit.Files["queryFile"] = []byte(jsonStr)
	data := &change.ChangeData{
		ObjectType: change.REPORT,
		Content:    commit,
		Author:     sig,
	}
	_, err = cmEngine.VersionObject(data, "test message")
	if err != nil {
		testingError(t, test, err.Error())
	}

	//second report
	query2 := reports.PbReportQuery{Name: "Device Compliance Report", Query: "Still Dont know yet", TimeStamp: time.Now().Format(time.RFC1123), Period: "-10s"}
	commit2 := change.NewCMContent(query2.Name)
	jsonStr, err = json.Marshal(query2)
	if err != nil {
		testingError(t, test, "Could not marshal the query2, Error: %s", err.Error())
		return
	}
	commit2.Files["queryFile"] = []byte(jsonStr)
	data2 := &change.ChangeData{
		ObjectType: change.REPORT,
		Content:    commit2,
		Author:     sig,
	}
	_, err = cmEngine.VersionObject(data2, "test message2")
	if err != nil {
		testingError(t, test, err.Error())
	}

	//third report
	query3 := reports.PbReportQuery{Name: "Versioning Device Report", Query: "Still Dont know yet", TimeStamp: time.Now().Format(time.RFC1123), Period: "-10s"}
	commit3 := change.NewCMContent(query3.Name)
	jsonStr, err = json.Marshal(query3)
	if err != nil {
		testingError(t, test, "Could not marshal the query3, Error: %s", err.Error())
		return
	}
	commit3.Files["queryFile"] = []byte(jsonStr)
	data3 := &change.ChangeData{
		ObjectType: change.REPORT,
		Content:    commit3,
		Author:     sig,
	}
	_, err = cmEngine.VersionObject(data3, "test message2")
	if err != nil {
		testingError(t, test, err.Error())
	}
	reports, err := cmEngine.ListObjects(change.REPORT)
	if err != nil {
		testingError(t, test, err.Error())
		return
	}
	if len(reports) != 3 {
		testingError(t, test, "Have not managed to set up the reports in the CME correctly to run the test.")
	}
	//test1: Now test our route
	req := createNewRequest(t, "DELETE", fmt.Sprintf("https://localhost:8080/reports/%s", query.Name), nil)
	writer := httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "Could not DELETE reports. got code: %v", writer.Code)
	}
	reports, err = cmEngine.ListObjects(change.REPORT)
	if err != nil {
		testingError(t, test, err.Error())
		return
	}
	if len(reports) != 2 {
		testingError(t, test, "Did not manage to DELETE the reports in the CME correctly.")
	}
	//test Versioning
	req = createNewRequest(t, "DELETE", fmt.Sprintf("https://localhost:8080/v2/reports/%s", query3.Name), nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Test 8: Should not get StatusOK, wrong version sent")
	}
	//this doesnt even reach our global.ProcessVersioning, it is not a route that is registered (only v[0-9]+ routes are registered)
	req = createNewRequest(t, "DELETE", fmt.Sprintf("https://localhost:8080/v1.0/reports/%s", query3.Name), nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Test 9: Should not get StatusOK, version in wrong format sent")
	}
	req = createNewRequest(t, "DELETE", fmt.Sprintf("https://localhost:8080/v1/reports/%s", query3.Name), nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "Test 10: Did not get StatusOK, version is now not 1?")
	}
}

func TestGetContentWIdHandler(t *testing.T) {
	test := "TestGetContentWIdHandler"
	begin(t, test)
	defer end(t, test)

	dbFile := "test_getContentWIdReport.db"
	dbHandle := setupDB(t, dbFile)
	defer os.Remove(dbFile)
	defer dbHandle.Close()

	cmEngine := getCME(t)
	defer cleanupCME(cmEngine)
	muxRouter := setupApiHandler(dbHandle)
	setupGlobal(t, "root")
	//setup root node in database
	node := pbdatabase.PbNode{Name: "root"}
	if err := node.Create(dbHandle); err != nil {
		testingError(t, test, "setup root node: %s", err.Error())
	}

	//create some report files so that we can retrieve it
	sig := &change.CMAuthor{
		Name:  "tester",
		Email: "tester@iti.com",
		When:  time.Now(),
	}

	//first report query
	query := reports.PbReportQuery{Name: "List of devices Report", Query: "Dont know yet", TimeStamp: time.Now().Format(time.RFC1123), Period: "-1"}
	commit := change.NewCMContent(query.Name)
	jsonStr, err := json.Marshal(query)
	if err != nil {
		testingError(t, test, "Could not marshal the query, Error: %s", err.Error())
	}
	commit.Files["queryFile"] = []byte(jsonStr)
	reportContent := `"Content": "Dummy report when the query is run"`
	commit.Files["reportFile"] = []byte(reportContent)
	data := &change.ChangeData{
		ObjectType: change.REPORT,
		Content:    commit,
		Author:     sig,
	}
	_, err = cmEngine.VersionObject(data, "test message")
	//second report
	query2 := reports.PbReportQuery{Name: "Device Compliance Report", Query: "Still Dont know yet", TimeStamp: time.Now().Format(time.RFC1123), Period: "-1"}
	commit2 := change.NewCMContent(query2.Name)
	jsonStr, err = json.Marshal(query2)
	if err != nil {
		testingError(t, test, "Could not marshal the query2, Error: %s", err.Error())
	}
	commit2.Files["queryFile"] = []byte(jsonStr)
	commit2.Files["reportFile"] = []byte("Report corresponding to query2 being run")
	data2 := &change.ChangeData{
		ObjectType: change.REPORT,
		Content:    commit2,
		Author:     sig,
	}
	_, err = cmEngine.VersionObject(data2, "test message2")
	//test1: Now test our route
	req := createNewRequest(t, "GET", fmt.Sprintf("https://localhost:8080/reports/report/%s", query.Name), nil)
	writer := httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "Could not GET reports. got code: %v", writer.Code)
	}
	retrievedReport := writer.Body.String()
	if reportContent != retrievedReport {
		testingError(t, test, "GetById contents don't match")
	}
	//test Versioning
	req = createNewRequest(t, "GET", fmt.Sprintf("https://localhost:8080/v2/reports/report/%s", query.Name), nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Should not get StatusOK, wrong version sent")
	}
	//this doesnt even reach our global.ProcessVersioning, it is not a route that is registered (only v[0-9]+ routes are registered)
	req = createNewRequest(t, "GET", fmt.Sprintf("https://localhost:8080/v1.0/reports/report/%s", query.Name), nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Should not get StatusOK, version in wrong format sent")
	}
	req = createNewRequest(t, "GET", fmt.Sprintf("https://localhost:8080/v1/reports/report/%s", query.Name), nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "Did not get StatusOK, version is now not 1?")
	}
}
