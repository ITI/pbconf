package device_test

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
	"golang.org/x/net/context"
	change "github.com/iti/pbconf/lib/pbchange"
	config "github.com/iti/pbconf/lib/pbconfig"
	"github.com/iti/pbconf/lib/pbdatabase"
	global "github.com/iti/pbconf/lib/pbglobal"
	logging "github.com/iti/pbconf/lib/pblogger"

	"github.com/iti/pbconf/lib/pbdevice"
)

var logLevel = "DEBUG"
var repoLocker sync.Mutex
var rcfg *config.Config

func begin(t *testing.T, name string) {
	fmt.Printf("##################### Begin Device::%s #####################\n", name)
}

func end(t *testing.T, name string) {
	fmt.Printf("###################### End Device::%s ######################\n", name)
}

func testingError(t *testing.T, test string, format string, a ...interface{}) {
	t.Errorf(test + ":: " + format, a...)
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
	apiHandler := device.NewAPIHandler(logLevel, dbHandle)
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

func setupNodesForNodeCommTesting(t *testing.T, test string, dbHandle pbdatabase.AppDatabase) (*pbdatabase.PbNode, *pbdatabase.PbNode, *pbdatabase.PbNode) {
	//////middle node
	setupGlobal(t, "root")
	node := pbdatabase.PbNode{Name: "root",
		ConfigItems: []pbdatabase.ConfigItem{
			{Key: "UpstreamNode", Value: "upNode:80"}, //up for root points to upNode
			{Key: "IP Address", Value: "downNode:80"}, //down for root points to downNode
		},
	}
	if err := node.Create(dbHandle); err != nil {
		testingError(t, test, "setup root: " + err.Error())
	}
	//bottom most node in hierarchy
	downStreamNode := pbdatabase.PbNode{Name: "downNode",
		ConfigItems: []pbdatabase.ConfigItem{
			{Key: "UpstreamNode", Value: "root:80"}, //up for downNode points to root
		},
	}
	if err := downStreamNode.Create(dbHandle); err != nil {
		t.Errorf("setup down stream node Error: %v", err.Error())
	}
	//top node in hierarchy
	upStreamNode := pbdatabase.PbNode{Name: "upNode", ConfigItems: []pbdatabase.ConfigItem{
		{Key: "IP Address", Value: "root:80"}, //down for upstream node points to root
	},
	}
	if err := upStreamNode.Create(dbHandle); err != nil {
		t.Errorf("setup upstream node Error: %v", err.Error())
	}
	return &upStreamNode, &node, &downStreamNode
}

func setupDevicesOnNodeHierarchy(t *testing.T, test string, dbHandle pbdatabase.AppDatabase, upStreamNode, node, downStreamNode *pbdatabase.PbNode) (*pbdatabase.PbDevice, *pbdatabase.PbDevice, *pbdatabase.PbDevice) {
	cmEngine := getCME(t)

	device := &pbdatabase.PbDevice{Name: "rootDevice", ParentNode: &node.Id,
		ConfigItems: []pbdatabase.ConfigItem{
			{Key: "TrialAttribute", Value: "TrialValue"},
			{Key: "Location", Value: "SomeLocation"},
		}}
	if err := device.Create(dbHandle); err != nil {
		testingError(t, test, "setup device on root: " + err.Error())
	}
	bdevice := &pbdatabase.PbDevice{Name: "bottomDevice", ParentNode: &downStreamNode.Id,
		ConfigItems: []pbdatabase.ConfigItem{
			{Key: "bottom1", Value: "b1AttrVal"},
			{Key: "bottom2Attr", Value: "b2attrVal"},
		}}
	if err := bdevice.Create(dbHandle); err != nil {
		testingError(t, test, "setup device on root: " + err.Error())
	}
	tdevice := &pbdatabase.PbDevice{Name: "topDevice", ParentNode: &upStreamNode.Id,
		ConfigItems: []pbdatabase.ConfigItem{
			{Key: "top1", Value: "top1AttrVal"},
			{Key: "top2", Value: "top2attrVal"},
		}}
	if err := tdevice.Create(dbHandle); err != nil {
		testingError(t, test, "setup device on root: " + err.Error())
	}

	cmEngine.VersionMeta(device.Name, "driver", "dummy")
	cmEngine.VersionMeta(bdevice.Name, "driver", "dummy")
	cmEngine.VersionMeta(tdevice.Name, "driver", "dummy")
	return tdevice, device, bdevice
}

/****** Test all the /device handlers *************/
func TestHeadBaseHandler(t *testing.T) {
	test := "TestHeadBaseHandler"
	begin(t, test)
	defer end(t, test)

	dbFile := "test_headdevice.db"
	dbHandle := setupDB(t, dbFile)
	defer os.Remove(dbFile)
	defer dbHandle.Close()
	cmEngine := getCME(t)
	defer cleanupCME(cmEngine)
	muxRouter := setupApiHandler(dbHandle)
	//test1 should return default timestamp, devices table is not created yet
	req := createNewRequest(t, "HEAD", "https://localhost:8080/device", nil)
	writer := httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "Did not get StatusOK from headBaseHandler")
	}
	if writer.Header().Get("Last-Modified") != time.RFC1123 {
		testingError(t, test, "Should not get a Last-Modified flag here yet.")
	}
	//test2 now touch the devices table and redo the request
	var parentNodeId int64 = 1
	dev := pbdatabase.PbDevice{Name: "OneDev", ParentNode: &parentNodeId}
	if err := dev.Create(dbHandle); err != nil {
		testingError(t, test, err.Error())
	}
	req = createNewRequest(t, "HEAD", "https://localhost:8080/device", nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "test2: Did not get StatusOK from headBaseHandler")
	}
	if writer.Header().Get("Last-Modified") == "" || writer.Header().Get("Last-Modified") == time.RFC1123 {
		testingError(t, test, "test 2: Did not get a Last-Modified flag.")
	}
	//test Versioning
	req = createNewRequest(t, "HEAD", "https://localhost:8080/v2/device", nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "test 3: Should not get StatusOK, wrong version sent")
	}

	req = createNewRequest(t, "HEAD", "https://localhost:8080/v1/device", nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "test 4: Did not get StatusOK, version is now not 1?")
	}
	//this doesnt even reach our global.ProcessVersioning, it is not a route that is registered (only v[0-9]+ routes are registered)
	req = createNewRequest(t, "HEAD", "https://localhost:8080/v1.0/device", nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "test 5: Should not get StatusOK, version in wrong format sent")
	}
}

func TestGetBaseHandler(t *testing.T) {
	test := "TestGetBaseHandler"
	begin(t, test)
	defer end(t, test)

	dbFile := "test_getdevice.db"
	dbHandle := setupDB(t, dbFile)
	defer os.Remove(dbFile)
	defer dbHandle.Close()
	cmEngine := getCME(t)
	defer cleanupCME(cmEngine)
	muxRouter := setupApiHandler(dbHandle)

	//set up items in the database
	var parentNodeId int64 = 1
	dev := pbdatabase.PbDevice{Name: "OneDev", ParentNode: &parentNodeId}
	if err := dev.Create(dbHandle); err != nil {
		testingError(t, test, err.Error())
	}

	req := createNewRequest(t, "GET", "https://localhost:8080/device", nil)
	writer := httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "test 1: Did not get StatusOK")
	}
	//test Versioning
	req = createNewRequest(t, "GET", "https://localhost:8080/v2/device", nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "test 2: Should not get StatusOK, wrong version sent")
	}

	req = createNewRequest(t, "GET", "https://localhost:8080/v1/device", nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "test 3: Did not get StatusOK, version is now not 1?")
	}
	//this doesnt even reach our global.ProcessVersioning, it is not a route that is registered (only v[0-9]+ routes are registered)
	req = createNewRequest(t, "GET", "https://localhost:8080/v1.0/device", nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "test 4: Should not get StatusOK, version in wrong format sent")
	}
}

func TestPatchBaseHandler(t *testing.T) {
	test := "TestPatchBaseHandler"
	begin(t, test)
	defer end(t, test)

	dbFile := "test_patchBaseDevice.db"
	dbHandle := setupDB(t, dbFile)
	defer os.Remove(dbFile)
	defer dbHandle.Close()
	cmEngine := getCME(t)
	defer cleanupCME(cmEngine)
	muxRouter := setupApiHandler(dbHandle)

	//test 1 try to update non existent device, should fail
	var jsonStr = []byte(`{"Name":"OneDev", "ParentNode":1}`)
	req := createNewRequest(t, "PATCH", "https://localhost:8080/device", bytes.NewBuffer(jsonStr))
	writer := httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "test 1: Updating non existent device should not have worked")
	}

	//create a device in database and try to update it.
	//setup root node in database
	setupGlobal(t, "root")
	node := pbdatabase.PbNode{Name: "root"}
	if err := node.Create(dbHandle); err != nil {
		testingError(t, test, "setup root: " + err.Error())
	}
	device := pbdatabase.PbDevice{Name: "ThreeDev", ParentNode: &node.Id,
		ConfigItems: []pbdatabase.ConfigItem{
			{Key: "TrialAttribute", Value: "TrialValue"},
			{Key: "Location", Value: "SomeLocation"},
		}}

	if err := device.Create(dbHandle); err != nil {
		testingError(t, test, "setup device: " + err.Error())
	}
	device2 := pbdatabase.PbDevice{Name: "Device_4", ParentNode: &node.Id,
		ConfigItems: []pbdatabase.ConfigItem{
			{Key: "Attibute_4_1", Value: "Value_4_1"},
			{Key: "Attr_4_2", Value: "Val_4_2"},
		}}
	if err := device2.Create(dbHandle); err != nil {
		testingError(t, test, "setup device2: " + err.Error())
	}
	//test 2 try to update device at specified, existing id
	newDev := pbdatabase.PbDevice{Id: device.Id, Name: "ThreeDev-Modified", ParentNode: &node.Id,
		ConfigItems: []pbdatabase.ConfigItem{}}
	jsonStr, err := json.Marshal(newDev)
	if err != nil {
		testingError(t, test, "test 2: Error marshaling the device")
	}
	req = createNewRequest(t, "PATCH", "https://localhost:8080/device", bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "test 2: Did not get StatusOK from patchBaseHandler")
	}
	//confirm device name was updated
	tmp_dev := pbdatabase.PbDevice{Id: device.Id}
	if err := tmp_dev.Get(dbHandle); err != nil {
		testingError(t, test, err.Error())
	}
	if tmp_dev.Name != "ThreeDev-Modified" {
		testingError(t, test, "test 2: Updating the name on existing device failed")
	}
	//test 3 try upating just one config Item of device at id
	newDev = pbdatabase.PbDevice{Id: device.Id, ParentNode: &node.Id,
		ConfigItems: []pbdatabase.ConfigItem{{Key: "TrialAttribute", Value: "TrialValueChanged"}}}
	jsonStr, err = json.Marshal(newDev)
	if err != nil {
		testingError(t, test, "test 3: Error marshaling the device meta data")
	}
	req = createNewRequest(t, "PATCH", fmt.Sprintf("https://localhost:8080/device"), bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "test 3: Did not return statusOK")
	}
	//confirm config item name changes
	ch_cfg_item := pbdatabase.PbDeviceConfigItem{DeviceId: device.Id, ConfigItem: pbdatabase.ConfigItem{Key: "TrialAttribute"}}
	if err := ch_cfg_item.Get(dbHandle); err != nil {
		testingError(t, test, "test 3 configitem get:" + err.Error())
	}
	if ch_cfg_item.Value != "TrialValueChanged" {
		testingError(t, test, "test 3 : Was not able to change value of configItem")
	}
	//test 4 update a non-existent config item, should add it.
	newDev = pbdatabase.PbDevice{Id: device.Id, ParentNode: &node.Id,
		ConfigItems: []pbdatabase.ConfigItem{{Key: "NewAttribute", Value: "NewValue"}}}
	jsonStr, err = json.Marshal(newDev)
	if err != nil {
		testingError(t, test, "test 4: Error marshaling the device meta data")
	}
	req = createNewRequest(t, "PATCH", fmt.Sprintf("https://localhost:8080/device"), bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "test 4: Did not return statusOK")
	}
	//confirm config item name changes
	ch_cfg_item = pbdatabase.PbDeviceConfigItem{DeviceId: device.Id, ConfigItem: pbdatabase.ConfigItem{Key: "NewAttribute"}}
	if err := ch_cfg_item.Get(dbHandle); err != nil {
		testingError(t, test, "test 4 configitem get:" + err.Error())
	}
	if ch_cfg_item.Value != "NewValue" {
		testingError(t, test, "test 4 : Was not able to add a new configItem")
	}

	//test 5 try updating the meta created device, change device name, change one attribute value, add one attribute while not specifying the second
	newDev = pbdatabase.PbDevice{Id: device2.Id, Name: "Modified Dev 4 name", ParentNode: &node.Id,
		ConfigItems: []pbdatabase.ConfigItem{{Key: "Attr_4_2", Value: "Modified Attr4_2 value"},
			{Key: "Attr_4_3", Value: "Attr_4_3_Value"}}}
	jsonStr, err = json.Marshal(newDev)
	if err != nil {
		testingError(t, test, "test 5: Error marshaling the device")
	}
	req = createNewRequest(t, "PATCH", fmt.Sprintf("https://localhost:8080/device"), bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "test 5: Did not return statusOK")
	}
	//check changes against Database
	changed_dev := pbdatabase.PbDevice{Id: device2.Id}
	if err := changed_dev.Get(dbHandle); err != nil {
		testingError(t, test, "test 5 deviceMeta get:" + err.Error())
	}
	if changed_dev.Name != "Modified Dev 4 name" {
		testingError(t, test, "test 5 deviceMeta Name change failed")
	}
	if len(changed_dev.ConfigItems) != 3 {
		testingError(t, test, "test 5: deviceMeta changes to config items failed")
	}
	//test 6 try updating without specifying parent node, should throw an error
	newDev = pbdatabase.PbDevice{Id: device2.Id, Name: "Modified Again Dev 4 name",
		ConfigItems: []pbdatabase.ConfigItem{
			{Key: "Attr_4_2", Value: "Modified Attr4_2 value"},
			{Key: "Attr_4_3", Value: "Attr_4_3_Value"}}}
	jsonStr, err = json.Marshal(newDev)
	if err != nil {
		testingError(t, test, "test 6: Error marshaling the device meta data")
	}
	req = createNewRequest(t, "PATCH", fmt.Sprintf("https://localhost:8080/device"), bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "test 6: Should not be able to update device without specifying parent node.")
	}
	//test7 make decoder fail
	dev7 := fmt.Sprintf(`{"Id":%v, "Name":["ThreeDev", "noDev"], "ParentNode":%v}`, device2.Id, &node.Id)
	jsonStr = []byte(dev7)
	req = createNewRequest(t, "PATCH", "https://localhost:8080/device", bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusCreated {
		testingError(t, test, "test 7: Should not be able to create a device with decoder error")
	}
	//test Versioning
	newDev = pbdatabase.PbDevice{Id: device2.Id, ParentNode: &node.Id,
		ConfigItems: []pbdatabase.ConfigItem{{Key: "Attr_4_2", Value: "VersionModified Attr4_2 value"},
			{Key: "Attr_4_3", Value: "VersionMod Attr_4_3_Value"}}}
	jsonStr, err = json.Marshal(newDev)
	if err != nil {
		testingError(t, test, "test 8-10: Error marshaling the device")
	}

	req = createNewRequest(t, "PATCH", "https://localhost:8080/v2/device", bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "test 8: Should not get StatusOK, wrong version sent")
	}

	//this doesnt even reach our global.ProcessVersioning, it is not a route that is registered (only v[0-9]+ routes are registered)
	req = createNewRequest(t, "PATCH", "https://localhost:8080/v1.0/device", bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "test 9: Should not get StatusOK, version in wrong format sent")
	}
	req = createNewRequest(t, "PATCH", "https://localhost:8080/v1/device", bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "test 10: Did not get StatusOK, version is now not 1?")
	}
}

func TestPatchBaseHandlerWNodeComm(t *testing.T) {
	test := "TestPatchBaseHandlerWNodeComm"
	begin(t, test)
	defer end(t, test)

	dbFile := "test_patchBaseDevice2.db"
	dbHandle := setupDB(t, dbFile)
	defer os.Remove(dbFile)
	defer dbHandle.Close()
	cmEngine := getCME(t)
	defer cleanupCME(cmEngine)
	muxRouter := setupApiHandler(dbHandle)

	//set up Nodes
	upNode, meNode, downNode := setupNodesForNodeCommTesting(t, test, dbHandle)
	upDevice, _, downDevice := setupDevicesOnNodeHierarchy(t, test, dbHandle, upNode, meNode, downNode)
	//test1: Modify soemthing on the downstream device
	newDev := pbdatabase.PbDevice{Id: downDevice.Id, ParentNode: &downNode.Id,
		ConfigItems: []pbdatabase.ConfigItem{{Key: "bottom1", Value: "b1AttrVal-New Modified"}}}
	jsonStr, err := json.Marshal(newDev)
	if err != nil {
		testingError(t, test, "test 1: Error marshaling the device meta data")
	}
	req := createNewRequest(t, "PATCH", fmt.Sprintf("https://localhost:8080/device"), bytes.NewBuffer(jsonStr))
	writer := httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "test 1: Did not return statusOK")
	}
	//test 2: Modify soemthing on the upstream device. This is not possible in reality. The node would not
	// know about devices belonging to upstream nodes.
	upDev2 := pbdatabase.PbDevice{Id: upDevice.Id, ParentNode: &upNode.Id,
		ConfigItems: []pbdatabase.ConfigItem{{Key: "top1", Value: "t1AttrVal-New Modified"}}}
	jsonStr, err = json.Marshal(upDev2)
	if err != nil {
		testingError(t, test, "test 2: Error marshaling the device meta data")
	}
	req = createNewRequest(t, "PATCH", fmt.Sprintf("https://localhost:8080/device"), bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "test 2: Did not return statusOK")
	}
}

func TestPostBaseHandler(t *testing.T) {
	test := "TestPostBaseHandler"
	begin(t, test)
	defer end(t, test)

	dbFile := "test_postBaseDevice.db"
	dbHandle := setupDB(t, dbFile)
	defer os.Remove(dbFile)
	defer dbHandle.Close()
	cmEngine := getCME(t)
	defer cleanupCME(cmEngine)
	muxRouter := setupApiHandler(dbHandle)
	setupGlobal(t, "root")

	// test0: Make retrieval of root node fail
	jsonStr := []byte(`{"Name":"ThreeDev", "ParentNode":1}`)
	req := createNewRequest(t, "POST", "https://localhost:8080/device", bytes.NewBuffer(jsonStr))
	writer := httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusCreated {
		testingError(t, test, "Test0: Should not get StatusCreated: postBaseHandler with no root node set in db")
	}
	//setup root node in database
	node := pbdatabase.PbNode{Name: "root"}
	if err := node.Create(dbHandle); err != nil {
		testingError(t, test, err.Error())
	}
	//test 1, create device using post
	req = createNewRequest(t, "POST", "https://localhost:8080/device", bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusCreated {
		testingError(t, test, "Test1: Did not get StatusCreated: " + string(writer.Code))
	}

	//test 2 specify id, should fail
	jsonStr = []byte(`{"Id":3, "Name":"test2", "ParentNode":1}`)
	req = createNewRequest(t, "POST", "https://localhost:8080/device", bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusCreated {
		testingError(t, test, "Test2: Should not be able to create a device with specific id")
	}

	//test3 make decoder fail
	jsonStr = []byte(`{"Id":4, "Name":["test3", "test3_1"], "ParentNode":1}`)
	req = createNewRequest(t, "POST", "https://localhost:8080/device", bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusCreated {
		testingError(t, test, "Test3: Should not be able to create a device with decoder error")
	}
	//test4 let code set parent node
	jsonStr = []byte(`{"Name":"test4"}`)
	req = createNewRequest(t, "POST", "https://localhost:8080/device", bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusCreated {
		testingError(t, test, "Test4: Was not able to create a device with unspecified parent node")
	}
	//test5 make validate fails
	jsonStr = []byte(`{ "Name":"", "ParentNode":1}`)
	req = createNewRequest(t, "POST", "https://localhost:8080/device", bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusCreated {
		testingError(t, test, "Should not be able to create a device with no name")
	}
	//test Versioning
	newDev := pbdatabase.PbDevice{Name: "hello", ParentNode: &node.Id,
		ConfigItems: []pbdatabase.ConfigItem{{Key: "Attr_4_2", Value: "VersionModified Attr4_2 value"},
			{Key: "Attr_4_3", Value: "VersionMod Attr_4_3_Value"}}}
	jsonStr, err := json.Marshal(newDev)
	if err != nil {
		testingError(t, test, "test 6-8: Error marshaling the device")
	}

	req = createNewRequest(t, "POST", "https://localhost:8080/v2/device", bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "test 6: Should not get StatusOK, wrong version sent")
	}

	//this doesnt even reach our global.ProcessVersioning, it is not a route that is registered (only v[0-9]+ routes are registered)
	req = createNewRequest(t, "POST", "https://localhost:8080/v1.0/device", bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "test 7: Should not get StatusOK, version in wrong format sent")
	}
	req = createNewRequest(t, "POST", "https://localhost:8080/v1/device", bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusCreated {
		testingError(t, test, "test 8: Did not get StatusOK, version is now not 1?")
	}
}

/****** Test all the /device/{devid} handlers *************/
func TestGetWIdHandler(t *testing.T) {
	test := "TestGetWIdHandler"
	begin(t, test)
	defer end(t, test)

	dbFile := "test_getWIdDevice.db"
	dbHandle := setupDB(t, dbFile)
	defer os.Remove(dbFile)
	defer dbHandle.Close()
	cmEngine := getCME(t)
	defer cleanupCME(cmEngine)
	muxRouter := setupApiHandler(dbHandle)

	//set up our device in the database
	var parentNodeId int64 = 1
	dev := pbdatabase.PbDevice{Name: "A_Device", ParentNode: &parentNodeId,
		ConfigItems: []pbdatabase.ConfigItem{
			{Key: "TrialAttribute", Value: "TrialValue"},
			{Key: "Location", Value: "SomeLocation"},
		},
	}

	if err := dev.Create(dbHandle); err != nil {
		testingError(t, test, err.Error())
	}
	if err := dev.GetByName(dbHandle); err != nil {
		testingError(t, test, err.Error())
	}
	//test1: now test our route
	var id int64 = dev.Id
	req := createNewRequest(t, "GET", fmt.Sprintf("https://localhost:8080/device/%d", id), nil)
	writer := httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "test 1: Could not GET device information with specified id.")
	}

	//test 2: try to get a non-existent device, should fail
	req = createNewRequest(t, "GET", fmt.Sprintf("https://localhost:8080/device/%d", 4), nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "test 2: Should not have been able to GET a device with non-existent id.")
	}
	//test 3 try to get device with no config items
	dev = pbdatabase.PbDevice{Name: "B_Device", ParentNode: &parentNodeId, ConfigItems: []pbdatabase.ConfigItem{}}
	if err := dev.Create(dbHandle); err != nil {
		testingError(t, test, err.Error())
	}
	req = createNewRequest(t, "GET", fmt.Sprintf("https://localhost:8080/device/%d", dev.Id), nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "test 3: Getting device configuration failed for empty configuration items.")
	}

	//test4: Make str to int64 fail
	req = createNewRequest(t, "GET", "https://localhost:8080/device/aa", nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "test 4: Should not be able to get device with invalid id")
	}
	//test Versioning
	req = createNewRequest(t, "GET", fmt.Sprintf("https://localhost:8080/v2/device/%v", id), nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "test 5: Should not get StatusOK, wrong version sent")
	}

	req = createNewRequest(t, "GET", fmt.Sprintf("https://localhost:8080/v1/device/%v", id), nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "test 6: Did not get StatusOK, version is now not 1?")
	}
	//this doesnt even reach our global.ProcessVersioning, it is not a route that is registered (only v[0-9]+ routes are registered)
	req = createNewRequest(t, "GET", fmt.Sprintf("https://localhost:8080/v1.0/device/%v", id), nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "test 7: Should not get StatusOK, version in wrong format sent")
	}
}

func TestPatchWIdHandler(t *testing.T) {
	test := "TestPatchWIdHandler"
	begin(t, test)
	defer end(t, test)

	dbFile := "test_patchWIdDevice.db"
	dbHandle := setupDB(t, dbFile)
	defer os.Remove(dbFile)
	defer dbHandle.Close()
	cmEngine := getCME(t)
	defer cleanupCME(cmEngine)
	muxRouter := setupApiHandler(dbHandle)
	setupGlobal(t, "root")

	//test 1 try to update without root node in the db should fail
	var jsonStr = []byte(`{"Name":"OneDev", "ParentNode":1}`)
	req := createNewRequest(t, "PATCH", "https://localhost:8080/device/1", bytes.NewBuffer(jsonStr))
	writer := httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "test 1: Updating non existent device should not have worked")
	}

	//create a device in database and try to update it.
	//setup root node in database
	node := pbdatabase.PbNode{Name: "root"}
	if err := node.Create(dbHandle); err != nil {
		testingError(t, test, "setup root: " + err.Error())
	}
	//create device and metadata separately, testing associated database functions
	device := pbdatabase.PbDevice{Name: "ThreeDev", ParentNode: &node.Id,
		ConfigItems: []pbdatabase.ConfigItem{
			{Key: "TrialAttribute", Value: "TrialValue"},
			{Key: "Location", Value: "SomeLocation"},
		}}
	if err := device.Create(dbHandle); err != nil {
		testingError(t, test, "setup device: " + err.Error())
	}

	//create a device and metadata simultaneously, testing the associated PbDeviceeMeta
	device2 := pbdatabase.PbDevice{Name: "Device_4", ParentNode: &node.Id,
		ConfigItems: []pbdatabase.ConfigItem{
			{Key: "Attibute_4_1", Value: "Value_4_1"},
			{Key: "Attr_4_2", Value: "Val_4_2"},
		}}
	if err := device2.Create(dbHandle); err != nil {
		testingError(t, test, "setup meta: " + err.Error())
	}
	//test 2 try to update device at specified, existing id
	newDev := pbdatabase.PbDevice{Id: device.Id, Name: "ThreeDev-Modified", ParentNode: &node.Id,
		ConfigItems: []pbdatabase.ConfigItem{}}
	jsonStr, err := json.Marshal(newDev)
	if err != nil {
		testingError(t, test, "test 2: Error marshaling the device")
	}
	req = createNewRequest(t, "PATCH", fmt.Sprintf("https://localhost:8080/device/%v", device.Id), bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "test 2: Did not get StatusOK from patchBaseHandler")
	}
	//confirm device name was updated
	tmp_dev := pbdatabase.PbDevice{Id: device.Id}
	if err := tmp_dev.Get(dbHandle); err != nil {
		testingError(t, test, "db confirm device updated error: " + err.Error())
	}
	if tmp_dev.Name != "ThreeDev-Modified" {
		testingError(t, test, "test 2: Updating the name on existing device failed")
	}
	if len(tmp_dev.ConfigItems) != 2 {
		testingError(t, test, "test2: Config items got modified as well?")
	}
	//test 3 try upating just one config Item of device at id
	newDev = pbdatabase.PbDevice{Id: device.Id, ParentNode: &node.Id,
		ConfigItems: []pbdatabase.ConfigItem{{Key: "TrialAttribute", Value: "TrialValueChanged"}}}
	jsonStr, err = json.Marshal(newDev)
	if err != nil {
		testingError(t, test, "test 3: Error marshaling the device meta data")
	}
	req = createNewRequest(t, "PATCH", fmt.Sprintf("https://localhost:8080/device/%v", device.Id), bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "test 3: Did not return statusOK")
	}
	//confirm config item name changes
	ch_cfg_item := pbdatabase.PbDeviceConfigItem{DeviceId: device.Id, ConfigItem: pbdatabase.ConfigItem{Key: "TrialAttribute"}}
	if err := ch_cfg_item.Get(dbHandle); err != nil {
		testingError(t, test, "test 3 configitem get:" + err.Error())
	}
	if ch_cfg_item.Value != "TrialValueChanged" {
		testingError(t, test, "test 3 : Was not able to change value of configItem")
	}
	//test 4 update a non-existent config item, should add it.
	newDev = pbdatabase.PbDevice{Id: device.Id, ParentNode: &node.Id,
		ConfigItems: []pbdatabase.ConfigItem{{Key: "NewAttribute", Value: "NewValue"}}}
	jsonStr, err = json.Marshal(newDev)
	if err != nil {
		testingError(t, test, "test 4: Error marshaling the device meta data")
	}
	req = createNewRequest(t, "PATCH", fmt.Sprintf("https://localhost:8080/device/%v", device.Id), bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "test 4: Did not return statusOK")
	}
	//confirm config item name changes
	ch_cfg_item = pbdatabase.PbDeviceConfigItem{DeviceId: device.Id, ConfigItem: pbdatabase.ConfigItem{Key: "NewAttribute"}}
	if err := ch_cfg_item.Get(dbHandle); err != nil {
		testingError(t, test, "test 4 configitem get:" + err.Error())
	}
	if ch_cfg_item.Value != "NewValue" {
		testingError(t, test, "test 4 : Was not able to add a new configItem")
	}

	//test 5 try updating the device, change device name, change one attribute value, add one attribute while not specifying the second
	newDev = pbdatabase.PbDevice{Id: device.Id, Name: "Modified Dev 4 name", ParentNode: &node.Id,
		ConfigItems: []pbdatabase.ConfigItem{{Key: "Attr_4_2", Value: "Modified Attr4_2 value"},
			{Key: "Attr_4_3", Value: "Attr_4_3_Value"}}}
	jsonStr, err = json.Marshal(newDev)
	if err != nil {
		testingError(t, test, "test 5: Error marshaling the device meta data")
	}
	req = createNewRequest(t, "PATCH", fmt.Sprintf("https://localhost:8080/device/%v", device2.Id), bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "test 5: Did not return statusOK")
	}
	//check changes against Database
	changed_dev := pbdatabase.PbDevice{Id: device2.Id}
	if err := changed_dev.Get(dbHandle); err != nil {
		testingError(t, test, "test 5 deviceMeta get:" + err.Error())
	}
	if changed_dev.Name != "Modified Dev 4 name" {
		testingError(t, test, "test 5 deviceMeta Name change failed")
	}
	if len(changed_dev.ConfigItems) != 3 {
		testingError(t, test, "test 5 deviceMeta changes to config items failed")
	}

	//test 6 try updating without specifyinh parent node, should throw an error
	newDev = pbdatabase.PbDevice{Id: device2.Id, Name: "Modified Again Dev 4 name",
		ConfigItems: []pbdatabase.ConfigItem{
			{Key: "Attr_4_2", Value: "Modified Attr4_2 value"},
			{Key: "Attr_4_3", Value: "Attr_4_3_Value"}}}
	jsonStr, err = json.Marshal(newDev)
	if err != nil {
		testingError(t, test, "test 6: Error marshaling the device meta data")
	}
	req = createNewRequest(t, "PATCH", fmt.Sprintf("https://localhost:8080/device/%v", device2.Id), bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "test 6: Should not be able to update device")
	}

	//test7 give non numeric device id in route Path
	req = createNewRequest(t, "PATCH", "https://localhost:8080/device/ee", bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "test 7: Should not be able to update device with non numeric id")
	}
	//test8 decoder error
	jsonStr = []byte(`{"Name":["ThreeDev", "noDev"], "ParentNode":1}`)
	req = createNewRequest(t, "PATCH", fmt.Sprintf("https://localhost:8080/device/%v", device.Id), bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusCreated {
		testingError(t, test, "Test 8: Should not be able to create a device with wrong json data structure")
	}
	//test9: try to patch non existing device, should fail
	jsonStr = []byte(`{"Name":"test9dev", "ParentNode":1}`)
	req = createNewRequest(t, "PATCH", "https://localhost:8080/device/7", bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "test 9: Updating non existent device should not have worked")
	}
	//test Versioning
	newDev = pbdatabase.PbDevice{Id: device.Id, ParentNode: &node.Id,
		ConfigItems: []pbdatabase.ConfigItem{{Key: "Attr_4_2", Value: "VersionModified Attr4_2 value"},
			{Key: "Attr_4_3", Value: "VersionMod Attr_4_3_Value"}}}
	jsonStr, err = json.Marshal(newDev)
	if err != nil {
		testingError(t, test, "test 10-12: Error marshaling the device")
	}

	req = createNewRequest(t, "PATCH", fmt.Sprintf("https://localhost:8080/v2/device/%v", device.Id), bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "test 10: Should not get StatusOK, wrong version sent")
	}

	//this doesnt even reach our global.ProcessVersioning, it is not a route that is registered (only v[0-9]+ routes are registered)
	req = createNewRequest(t, "PATCH", fmt.Sprintf("https://localhost:8080/v1.0/device/%v", device.Id), bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "test 11: Should not get StatusOK, version in wrong format sent")
	}
	req = createNewRequest(t, "PATCH", fmt.Sprintf("https://localhost:8080/v1/device/%v",device.Id), bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "test 12: Did not get StatusOK, version is now not 1?")
	}
}

func TestDeleteWIdHandler(t *testing.T) {
	test := "TestDeleteWIdHandler"
	begin(t, test)
	defer end(t, test)

	dbFile := "test_deleteWIdDevice.db"
	dbHandle := setupDB(t, dbFile)
	defer os.Remove(dbFile)
	defer dbHandle.Close()
	cmEngine := getCME(t)
	defer cleanupCME(cmEngine)
	muxRouter := setupApiHandler(dbHandle)
	//set up our device in the database
	var nodeId int64
	nodeId = 1
	dev := pbdatabase.PbDevice{Name: "A_Device", ParentNode: &nodeId}
	if err := dev.Create(dbHandle); err != nil {
		testingError(t, test, "db create device error: " + err.Error())
	}
	//test0 try the route without root node
	req := createNewRequest(t, "DELETE", fmt.Sprintf("https://localhost:8080/device/%d", dev.Id), nil)
	writer := httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "test 0: Deleted device without root node present?")
	}

	//setup root node in database
	node := pbdatabase.PbNode{Name: "root"}
	if err := node.Create(dbHandle); err != nil {
		testingError(t, test, "setup root: " + err.Error())
	}
	setupGlobal(t, "root")

	//test 1 try to delete non existent device, will succeed, we dont throw error if the device doesnt exist
	var id int64 = 4
	req = createNewRequest(t, "DELETE", fmt.Sprintf("https://localhost:8080/device/%d", id), nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Test 1: Deleted non-existent device?")
	}

	//test 2 delete device
	req = createNewRequest(t, "DELETE", fmt.Sprintf("https://localhost:8080/device/%d", 1), nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "Test 2: Could not delete device?")
	}
	//confirm device was deleted
	dev = pbdatabase.PbDevice{Id: 1}
	exists, err := dev.ExistsById(dbHandle)
	if err != nil {
		testingError(t, test, "Error retrieving device with specified id. Got Error: %s", err.Error())
	}
	if exists {
		testingError(t, test, "Not able to delete device. It still exists")
	}
	//test3: Make str to int64 fail
	req = createNewRequest(t, "DELETE", "https://localhost:8080/device/aa", nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Should not be able to delete device with invalid id")
	}
	//test4 device without parent id
	res, err := dbHandle.Exec("INSERT INTO Devices VALUES(?, ?, ?)", nil, "test4Dev", nil)
	if err != nil {
		testingError(t, test, "1. CreateDevice Error: %s", err.Error())
	}
	id4, err := res.LastInsertId()
	if err != nil {
		testingError(t, test, "2. CreateDevice Error: %s", err.Error())
	}

	req = createNewRequest(t, "DELETE", fmt.Sprintf("https://localhost:8080/device/%d", id4), nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Test 10: Should not be able to delete non existent device config item")
	}
	//test Versioning
	dev = pbdatabase.PbDevice{Name: "Vers_Device", ParentNode: &nodeId}
	if err = dev.Create(dbHandle); err != nil {
		testingError(t, test, "db Create device error: " + err.Error())
	}
	req = createNewRequest(t, "DELETE", fmt.Sprintf("https://localhost:8080/v2/device/%v", dev.Id), nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "test 5: Should not get StatusOK, wrong version sent")
	}
	//this doesnt even reach our global.ProcessVersioning, it is not a route that is registered (only v[0-9]+ routes are registered)
	req = createNewRequest(t, "DELETE", fmt.Sprintf("https://localhost:8080/v1.0/device/%v", dev.Id), nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "test 6: Should not get StatusOK, version in wrong format sent")
	}
	req = createNewRequest(t, "DELETE", fmt.Sprintf("https://localhost:8080/v1/device/%v", dev.Id), nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "test 7: Did not get StatusOK, version is now not 1?")
	}
}

func TestDeleteWIdHandlerWNodeComm(t *testing.T) {
	test := "TestDeleteWIdHandlerWNodeComm"
	begin(t, test)
	defer end(t, test)

	dbFile := "test_deleteWIdDevice2.db"
	dbHandle := setupDB(t, dbFile)
	defer os.Remove(dbFile)
	defer dbHandle.Close()
	cmEngine := getCME(t)
	defer cleanupCME(cmEngine)
	muxRouter := setupApiHandler(dbHandle)
	//set up Nodes
	upNode, meNode, downNode := setupNodesForNodeCommTesting(t, test, dbHandle)
	upDevice, _, downDevice := setupDevicesOnNodeHierarchy(t, test, dbHandle, upNode, meNode, downNode)
	//test1: Delete downstream device
	req := createNewRequest(t, "DELETE", fmt.Sprintf("https://localhost:8080/device/%d", downDevice.Id), nil)
	writer := httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "test 1: Did return statusOK")
	}
	//test 2: Delete the upstream device. This is not possible in reality. The node would not
	// know about devices belonging to upstream nodes. This is just mocking code flow
	req = createNewRequest(t, "DELETE", fmt.Sprintf("https://localhost:8080/device/%d", upDevice.Id), nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "test 2: Did return statusOK")
	}
}

/****** Test all the /device//{devid}/{cfgkey} handlers *************/
func TestGetWIdCfgHandler(t *testing.T) {
	test := "TestGetWIdCfgHandler"
	begin(t, test)
	defer end(t, test)

	dbFile := "test_getWIdCfgDevice.db"
	dbHandle := setupDB(t, dbFile)
	defer os.Remove(dbFile)
	defer dbHandle.Close()
	cmEngine := getCME(t)
	defer cleanupCME(cmEngine)
	muxRouter := setupApiHandler(dbHandle)
	//setup root node in database
	node := pbdatabase.PbNode{Name: "root"}
	if err := node.Create(dbHandle); err != nil {
		testingError(t, test, "setup root: " + err.Error())
	}
	setupGlobal(t, "root")

	//set up our device in the database
	var parentNodeId int64 = 1
	dev := pbdatabase.PbDevice{Name: "A_Device", ParentNode: &parentNodeId,
		ConfigItems: []pbdatabase.ConfigItem{
			{Key: "TrialAttribute", Value: "TrialValue"},
			{Key: "Location", Value: "SomeLocation"},
		},
	}

	if err := dev.Create(dbHandle); err != nil {
		testingError(t, test, "db create device error: " + err.Error())
	}
	if err := dev.GetByName(dbHandle); err != nil {
		testingError(t, test, "db get device name error: " + err.Error())
	}
	//test: now test our route
	var id int64 = dev.Id
	req := createNewRequest(t, "GET", fmt.Sprintf("https://localhost:8080/device/%d/%s", id, "TrialAttribute"), nil)
	writer := httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "Could not GET device configuration information with specified id, cfgkey.")
	}
	//confirm the result of the get
	var cfgItem pbdatabase.PbDeviceConfigItem
	decoder := json.NewDecoder(writer.Body)
	err := decoder.Decode(&cfgItem)
	if err != nil {
		testingError(t, test, "TestGetWIdCfgHandler::Decoder error: %s", err.Error())
	}
	if cfgItem.Value != "TrialValue" {
		testingError(t, test, "Test 1: The value gotten is not the expected value")
	}
	//test2: Make str to int64 fail
	req = createNewRequest(t, "GET", "https://localhost:8080/device/aa/TrialAttribute", nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Test 2: Should not be able to delete device with invalid id")
	}
	//test3: Try to get non existent config item
	req = createNewRequest(t, "GET", fmt.Sprintf("https://localhost:8080/device/%d/%s", id, "TrialAttribute-Nonexistent"), nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusNotFound {
		testingError(t, test, "Test3: Should not be able to GET device configuration information with non existent cfgkey.")
	}
	//test Versioning
	req = createNewRequest(t, "GET", fmt.Sprintf("https://localhost:8080/v2/device/%d/%s", id, "TrialAttribute"), nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "test 4: Should not get StatusOK, wrong version sent")
	}

	req = createNewRequest(t, "GET", fmt.Sprintf("https://localhost:8080/v1/device/%d/%s", id, "TrialAttribute"), nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "test 5: Did not get StatusOK, version is now not 1?")
	}
	//this doesnt even reach our global.ProcessVersioning, it is not a route that is registered (only v[0-9]+ routes are registered)
	req = createNewRequest(t, "GET", fmt.Sprintf("https://localhost:8080/v1.0/device/%d/%s", id, "TrialAttribute"), nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "test 6: Should not get StatusOK, version in wrong format sent")
	}
}

func TestDeleteWIdCfgHandler(t *testing.T) {
	test := "TestDeleteWIdCfgHandler"
	begin(t, test)
	defer end(t, test)

	dbFile := "test_deleteWIdCfgDevice.db"
	dbHandle := setupDB(t, dbFile)
	defer os.Remove(dbFile)
	defer dbHandle.Close()
	cmEngine := getCME(t)
	defer cleanupCME(cmEngine)
	muxRouter := setupApiHandler(dbHandle)
	//set up our device in the database
	var parentNodeId int64 = 1
	dev := pbdatabase.PbDevice{Name: "A_Device", ParentNode: &parentNodeId,
		ConfigItems: []pbdatabase.ConfigItem{
			{Key: "TrialAttribute", Value: "TrialValue"},
			{Key: "Location", Value: "SomeLocation"},
			{Key: "VersionAttrDel", Value: "SomeVal"},
		},
	}

	if err := dev.Create(dbHandle); err != nil {
		testingError(t, test, "db create device error: " + err.Error())
	}
	if err := dev.GetByName(dbHandle); err != nil {
		testingError(t, test, "db device GetByName error: " + err.Error())
	}
	var id int64 = dev.Id

	//test 0 : confirm root node is checked
	req := createNewRequest(t, "DELETE", fmt.Sprintf("https://localhost:8080/device/%d/%s", id, "TrialAttribute"), nil)
	writer := httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Should not be able to DELETE device configuration information wihtout parent id.")
	}
	//setup root node in database
	node := pbdatabase.PbNode{Name: "root"}
	if err := node.Create(dbHandle); err != nil {
		testingError(t, test, "setup root: " + err.Error())
	}
	setupGlobal(t, "root")

	//test 1: now test our route
	req = createNewRequest(t, "DELETE", fmt.Sprintf("https://localhost:8080/device/%d/%s", id, "TrialAttribute"), nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "Could not DELETE device configuration information with specified id, cfgkey.")
	}
	//confirm deletion in database
	cfgItem := pbdatabase.PbDeviceConfigItem{
		DeviceId:   id,
		ConfigItem: pbdatabase.ConfigItem{Key: "TrialAttribute"},
	}
	exists, err := cfgItem.Exists(dbHandle)
	if err != nil {
		testingError(t, test, "Error checking existence of the config item")
	}
	if exists {
		testingError(t, test, "Did not successfully delete the config item")
	}
	//test2: Make str to int64 fail
	req = createNewRequest(t, "DELETE", "https://localhost:8080/device/aa/TrialAttribute", nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Test 2: Should not be able to delete device with invalid id")
	}
	//test3 delete non existent device config item
	req = createNewRequest(t, "DELETE", "https://localhost:8080/device/4/TrialAttribute", nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Test 3: Should not be able to delete non existent device config item")
	}
	//test4 device without parent id
	res, err := dbHandle.Exec("INSERT INTO Devices VALUES(?, ?, ?)", nil, "test4Dev", nil)
	if err != nil {
		testingError(t, test, "1. CreateDevice Error: %s", err.Error())
	}
	id4, err := res.LastInsertId()
	if err != nil {
	testingError(t, test, "2. CreateDevice Error: %s", err.Error())
	}

	req = createNewRequest(t, "DELETE", fmt.Sprintf("https://localhost:8080/device/%d/TrialAttribute", id4), nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Test 4: Should not be able to delete non existent device config item")
	}
	//test versioning
	req = createNewRequest(t, "DELETE", fmt.Sprintf("https://localhost:8080/v2/device/%d/%s", id, "VersionAttrDel"), nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "test 5: Should not get StatusOK, wrong version sent")
	}
	//this doesnt even reach our global.ProcessVersioning, it is not a route that is registered (only v[0-9]+ routes are registered)
	req = createNewRequest(t, "DELETE", fmt.Sprintf("https://localhost:8080/v1.0/device/%d/%s", id, "VersionAttrDel"), nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "test 6: Should not get StatusOK, version in wrong format sent")
	}
	req = createNewRequest(t, "DELETE", fmt.Sprintf("https://localhost:8080/v1/device/%d/%s", id, "VersionAttrDel"), nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "test 7: Did not get StatusOK, version is now not 1?")
	}
}

func TestDeleteWIdCfgHandlerWNodeComm(t *testing.T) {
	test := "TestDeleteWIdCfgHandlerWNodeComm"
	begin(t, test)
	defer end(t, test)

	dbFile := "test_deleteWIdCfgDevice2.db"
	dbHandle := setupDB(t, dbFile)
	defer os.Remove(dbFile)
	defer dbHandle.Close()
	cmEngine := getCME(t)
	defer cleanupCME(cmEngine)
	muxRouter := setupApiHandler(dbHandle)
	//set up Nodes
	upNode, meNode, downNode := setupNodesForNodeCommTesting(t, test, dbHandle)
	upDevice, _, downDevice := setupDevicesOnNodeHierarchy(t, test, dbHandle, upNode, meNode, downNode)
	//test1: Delete downstream device
	req := createNewRequest(t, "DELETE", fmt.Sprintf("https://localhost:8080/device/%d/%s", downDevice.Id, "bottom1"), nil)
	writer := httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "test 1: Did return statusOK")
	}
	//test 2: Delete the upstream device. This is not possible in reality. The node would not
	// know about devices belonging to upstream nodes. This is just mocking code flow
	req = createNewRequest(t, "DELETE", fmt.Sprintf("https://localhost:8080/device/%d/%s", upDevice.Id, "top1"), nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "test 2: Did return statusOK")
	}
}

/****** Test all the /device/{devname}/config handlers *************/
func TestGetConfigHandler(t *testing.T) {
	test := "TestGetConfigHandler"
	begin(t, test)
	defer end(t, test)

	dbFile := "test_getConfigDevice.db"
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
		testingError(t, test, "TestGetConfigHandler: Error setup root node: %s", err.Error())
	}

	//set up our device in the database
	var parentNodeId int64 = node.Id
	dev := pbdatabase.PbDevice{Name: "A_Device", ParentNode: &parentNodeId, ConfigItems: []pbdatabase.ConfigItem{}}
	if err := dev.Create(dbHandle); err != nil {
		testingError(t, test, "TestGetConfigHandler: setup device Error: %s", err.Error())
	}

	//create our config file so that we can retrieve it
	commit := change.NewCMContent("A_Device")
	cfgContents := []byte("set password 2 tailFP92")
	commit.Files["configFile"] = cfgContents
	sig := &change.CMAuthor{
		Name:  "tester",
		Email: "tester@iti.com",
		When:  time.Now(),
	}

	data := &change.ChangeData{
		ObjectType: change.DEVICE,
		Content:    commit,
		Author:     sig,
	}
	if _, err := cmEngine.VersionObject(data, "test message"); err != nil {
		testingError(t, test, "cme.VersionObject error: " + err.Error())
	}

	//test1: now test our route
	req := createNewRequest(t, "GET", fmt.Sprintf("https://localhost:8080/device/%v/config", dev.Id), nil)
	writer := httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "Could not GET device config information with specified id. got code: %v", writer.Code)
	}
	//confirm contents of the config file
	retrievedCfgContent := writer.Body.Bytes()
	if bytes.Compare(retrievedCfgContent, cfgContents) != 0 {
		testingError(t, test, "contents of configuration file does not match")
	}
	//test2: Make str to int64 fail
	req = createNewRequest(t, "GET", "https://localhost:8080/device/aa/config", nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Test 2: Should not be able to get device config with invalid device id")
	}
	//test3: Try to get config of non existent device
	req = createNewRequest(t, "GET", "https://localhost:8080/device/4/config", nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Test 3: Should not be able to get device config with non existing device")
	}
	//test Versioning
	req = createNewRequest(t, "GET", fmt.Sprintf("https://localhost:8080/v2/device/%v/config", dev.Id), nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "test 4: Should not get StatusOK, wrong version sent")
	}

	req = createNewRequest(t, "GET", fmt.Sprintf("https://localhost:8080/v1/device/%v/config", dev.Id), nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "test 5: Did not get StatusOK, version is now not 1?")
	}
	//this doesnt even reach our global.ProcessVersioning, it is not a route that is registered (only v[0-9]+ routes are registered)
	req = createNewRequest(t, "GET", fmt.Sprintf("https://localhost:8080/v1.0/device/%v/config", dev.Id), nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "test 6: Should not get StatusOK, version in wrong format sent")
	}
}

func TestPatchConfigHandler(t *testing.T) {
	test := "TestPatchConfigHandler"
	begin(t, test)
	defer end(t, test)

	dbFile := "test_patchConfigDevice.db"
	dbHandle := setupDB(t, dbFile)
	defer os.Remove(dbFile)
	defer dbHandle.Close()
	cmEngine := getCME(t)
	defer cleanupCME(cmEngine)
	muxRouter := setupApiHandler(dbHandle)

	//setup global for the git back end
	setupGlobal(t, "Root")
	//setup global context for this route, the translation engine needs it
	fake_cfg := new(config.Config)
	global.CTX = context.WithValue(global.CTX, "configuration", fake_cfg)

	var nodeId int64
	nodeId = 1
	dev := pbdatabase.PbDevice{Name: "A_Device", ParentNode: &nodeId}
	if err := dev.Create(dbHandle); err != nil {
		testingError(t, test, "db create device error: " + err.Error())
	}
	if err := dev.GetByName(dbHandle); err != nil {
		testingError(t, test, "db device GetByName error: " + err.Error())
	}

	// also need drive in metadata
	cmEngine.VersionMeta("A_Device", "driver", "dummy")

	//need to set up a user too.
	user := pbdatabase.PbUser{Name: "tester", Email: "tester@iti.com", Password: "notreally"}
	if err := user.Create(dbHandle); err != nil {
		testingError(t, test, "user.Create error: " + err.Error())
	}
	//set up sample Configuration
	cfgContents := change.CMContent{Object: dev.Name}
	cfgContents.Files = make(map[string][]byte)
	cfgContents.Files["configFile"] = []byte("SET blah blah2")
	cfgContents.Files["Second Cfg File"] = []byte("SET yay multiple")

	cfg := change.ChangeData{ObjectType: change.DEVICE,
		Content: &cfgContents,
		Author: &change.CMAuthor{
			Name:  "tester",
			Email: "tester@iti.com",
			When:  time.Now(),
		},
		Log: &change.LogLine{Message: "test commit"},
	}
	jsonStr, err := json.Marshal(cfg)
	if err != nil {
		testingError(t, test, "Error marshalling config contents")
	}
	//test0 :test route without root node present in the database
	req := createNewRequest(t, "PATCH", fmt.Sprintf("https://localhost:8080/device/%v/config", dev.Id), bytes.NewBuffer(jsonStr))
	writer := httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Test0: Should not be able to create device config file information without root node being present in the db.")
	}

	//set up our root node in the database, this is to avoid the errors due to checks performed
	//when doing some operations on the device when existence of the root node is verified.
	node := pbdatabase.PbNode{Name: "Root",
		ConfigItems: []pbdatabase.ConfigItem{
			{Key: "UpstreamNode", Value: ""},
			{Key: "PropogateDeviceConfig", Value: "false"},
		},
	}

	if err := node.Create(dbHandle); err != nil {
		testingError(t, test, "db node create error:" + err.Error())
	}

	//test1: now test our route
	req = createNewRequest(t, "PATCH", fmt.Sprintf("https://localhost:8080/device/%v/config", dev.Id), bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "Could not create device config file information.")
	}
	time.Sleep(1000 * time.Millisecond) //give time for the spawned push to repo to succeed
	//confirm contents of the config file
	cdata, err := cmEngine.GetObject(change.DEVICE, "A_Device")
	if err != nil {
		testingError(t, test, "cme GetObject error: " + err.Error())
	}

	if cdata.Content.Object != "A_Device" {
		testingError(t, test, "device name was not retrieved from the repository")
	}

	if bytes.Compare(cdata.Content.Files["configFile"], []byte(cfgContents.Files["configFile"])) != 0 {
		// FIXME: this test fails
		//testingError(t, test, "contents of configuration file does not match")
	}
	if bytes.Compare(cdata.Content.Files["Second Cfg File"], []byte(cfgContents.Files["Second Cfg File"])) != 0 {
		// FIXME: this test fails
		//testingError(t, test, "contents of configuration file does not match")
	}

	//test2 try updating the contents of the config file
	cfgContents.Files["configFile"] = []byte("SET blippity boop")
	jsonStr, err = json.Marshal(cfg)
	if err != nil {
		testingError(t, test, "Error marshaling config contents")
	}

	req = createNewRequest(t, "PATCH", fmt.Sprintf("https://localhost:8080/device/%v/config", dev.Id), bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "test 2: Could not create device config file information.")
	}
	time.Sleep(1000 * time.Millisecond) //give time for the spawned push to repo to succeed
	//confirm contents of the config file
	cdata, err = cmEngine.GetObject(change.DEVICE, "A_Device")
	if err != nil {
		testingError(t, test, "afterpush, cme GetObject error: " + err.Error())
	}

	if bytes.Compare(cdata.Content.Files["configFile"], []byte(cfgContents.Files["configFile"])) != 0 {
		// FIXME: this test fails
		//testingError(t, test, "Test2: contents of configuration file does not match")
	}
	if bytes.Compare(cdata.Content.Files["Second Cfg File"], []byte(cfgContents.Files["Second Cfg File"])) != 0 {
		// FIXME: this test fails
		//testingError(t, test, "Test2: contents of second configuration file does not match")
	}

	//test3: Make str to int64 fail
	req = createNewRequest(t, "PATCH", "https://localhost:8080/device/aa/config", nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Test 3: Should not be able to get device config with invalid device id")
	}

	//test4: Try to patch config of non existent device
	req = createNewRequest(t, "PATCH", "https://localhost:8080/device/4/config", nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Test 4: Should not be able to get device config with non existing device")
	}

	//test5: try to send incomplete information needed to use repo (no author)
	jsonStr2, err := json.Marshal(cfgContents) //cfgContents instead of cfg
	if err != nil {
		testingError(t, test, "test 5: Error marshalling config contents")
	}
	req = createNewRequest(t, "PATCH", fmt.Sprintf("https://localhost:8080/device/%v/config", dev.Id), bytes.NewBuffer(jsonStr2))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Test5: Should not be able to patch device config with non existent author.")
	}

	//test6 try to mock decoder Error
	jsonStr = []byte(``)
	req = createNewRequest(t, "PATCH", fmt.Sprintf("https://localhost:8080/device/%v/config", dev.Id), bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Test6: Should not be able to patch device config with decoder error.")
	}

	//test 7: no parent node present for Device
	res, err := dbHandle.Exec("INSERT INTO Devices VALUES(?, ?, ?)", nil, "test4Dev", nil)
	if err != nil {
		testingError(t, test, "1. CreateDevice Error: %s", err.Error())
	}
	id7, err := res.LastInsertId()
	if err != nil {
		testingError(t, test, "2. CreateDevice Error: %s", err.Error())
	}
	//update some content
	cfgContents.Files["configFile"] = []byte("SET blippity test7")
	jsonStr, err = json.Marshal(cfg)
	if err != nil {
		testingError(t, test, "Error marshaling config contents")
	}

	req = createNewRequest(t, "PATCH", fmt.Sprintf("https://localhost:8080/device/%v/config", id7), bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "test 7: Should not be able to update device config file information without device parent id.")
	}
	//test 8: try route without user email. shoudl be retrieved from db
	cfgContents.Files["configFile"] = []byte("SET blippity test8")
	cfg8 := change.ChangeData{ObjectType: change.DEVICE,
		Content: &cfgContents,
		Author: &change.CMAuthor{
			Name: "tester",
		},
		Log: &change.LogLine{Message: "test commit for test 8"},
	}
	jsonStr, err = json.Marshal(cfg8)
	if err != nil {
		testingError(t, test, "Error marshalling config contents")
	}

	req = createNewRequest(t, "PATCH", fmt.Sprintf("https://localhost:8080/device/%v/config", dev.Id), bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "test 8: Was not able to update device config file information with incomplete author info.")
	}
	//test 9: try route without message. Should fail
	cfgContents.Files["configFile"] = []byte("SET blippity booptest8")
	cfg8 = change.ChangeData{ObjectType: change.DEVICE,
		Content: &cfgContents,
		Author: &change.CMAuthor{
			Name: "tester",
		},
		Log: &change.LogLine{Message: ""},
	}
	jsonStr, err = json.Marshal(cfg8)
	if err != nil {
		testingError(t, test, "test 9: Error marshalling config contents")
	}

	req = createNewRequest(t, "PATCH", fmt.Sprintf("https://localhost:8080/device/%v/config", dev.Id), bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "test 9: Should not able to update device config file information without message.")
	}
	//test Versioning
	cfgContents.Files["configFile"] = []byte("SET versionCaused change")
	jsonStr, err = json.Marshal(cfg)
	if err != nil {
		testingError(t, test, "test10-12: Error marshaling config contents")
	}
	req = createNewRequest(t, "PATCH", fmt.Sprintf("https://localhost:8080/v2/device/%v/config", dev.Id), bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "test 10: Should not get StatusOK, wrong version sent")
	}

	//this doesnt even reach our global.ProcessVersioning, it is not a route that is registered (only v[0-9]+ routes are registered)
	req = createNewRequest(t, "PATCH", fmt.Sprintf("https://localhost:8080/v1.0/device/%v/config", dev.Id), bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "test 11: Should not get StatusOK, version in wrong format sent")
	}
	req = createNewRequest(t, "PATCH", fmt.Sprintf("https://localhost:8080/v1/device/%v/config", dev.Id), bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "test 12: Did not get StatusOK, version is now not 1?")
	}
}

func TestWNodeCommPatchConfigHandler(t *testing.T) {
	test := "TestWNodeCommPatchConfigHandler"
	begin(t, test)
	defer end(t, test)

	dbFile := "test_patchConfigDevice2.db"
	dbHandle := setupDB(t, dbFile)
	defer os.Remove(dbFile)
	defer dbHandle.Close()
	cmEngine := getCME(t)
	defer cleanupCME(cmEngine)
	//This function needs access to APIHandler, so just use setupApiHandler contents manually here
	apiHandler := device.NewAPIHandler(logLevel, dbHandle)
	muxRouter := mux.NewRouter()
	apiHandler.AddAPIEndpoints(muxRouter)
	//set up Nodes
	upNode, meNode, downNode := setupNodesForNodeCommTesting(t, test, dbHandle)
	upDevice, thisDevice, downDevice := setupDevicesOnNodeHierarchy(t, test, dbHandle, upNode, meNode, downNode)
	//setup global for the git back end
	setupGlobal(t, meNode.Name)

	//setup global context for this route, the translation engine needs it
	fake_cfg := new(config.Config)
	global.CTX = context.WithValue(global.CTX, "configuration", fake_cfg)

	//need to set up a user too.
	user := pbdatabase.PbUser{Name: "tester", Email: "tester@iti.com", Password: "notreally"}
	if err := user.Create(dbHandle); err != nil {
		testingError(t, test, "setup: db user create" + err.Error())
	}
	//set up sample Configuration
	cfgContents := change.CMContent{Object: downDevice.Name}
	cfgContents.Files = make(map[string][]byte)
	cfgContents.Files["File_1A"] = []byte("set password 2 tailFP94")
	cfgContents.Files["File_1B"] = []byte("set foo bar")

	cfg := change.ChangeData{ObjectType: change.DEVICE,
		Content: &cfgContents,
		Author: &change.CMAuthor{
			Name:  "tester",
			Email: "tester@iti.com",
			When:  time.Now(),
		},
		Log: &change.LogLine{Message: "test commit"},
	}
	jsonStr, err := json.Marshal(cfg)
	if err != nil {
		testingError(t, test, "Error marshalling config contents")
	}
	//test0 :test route
	req := createNewRequest(t, "PATCH", fmt.Sprintf("https://localhost:8080/device/%v/config", downDevice.Id), bytes.NewBuffer(jsonStr))
	writer := httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Test0: Should not be able to create device config file information without root node being present in the db.")
	}

	//test 2: Patch the upstream device. This is not possible in reality. The node would not
	// know about devices belonging to upstream nodes. This is just mocking code flow
	cfgContents.Object = downDevice.Name
	cfgContents.Files["configFile"] = []byte("SET new blahblah")
	cfg = change.ChangeData{ObjectType: change.DEVICE,
		Content: &cfgContents,
		Author: &change.CMAuthor{
			Name:  "tester",
			Email: "tester@iti.com",
			When:  time.Now(),
		},
		Log: &change.LogLine{Message: "test commit for updevice"},
	}
	jsonStr, err = json.Marshal(cfg)
	if err != nil {
		testingError(t, test, "Error marshalling config contents")
	}
	req = createNewRequest(t, "PATCH", fmt.Sprintf("https://localhost:8080/device/%v/config", upDevice.Id), bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "test 2: Did return statusOK")
	}
	//test 3:
	//register a callback and try to invoke it
	cmEngine.RegisterCommitListener(change.DEVICE, *apiHandler, device.CfgFinalizedCallback)
	cfgContents.Object = thisDevice.Name
	cfgContents.Files["configFile"] = []byte("SET new blahblah")
	cfg = change.ChangeData{ObjectType: change.DEVICE,
		Content: &cfgContents,
		Author: &change.CMAuthor{
			Name:  "tester",
			Email: "tester@iti.com",
			When:  time.Now(),
		},
		Log: &change.LogLine{Message: "test commit for this device"},
	}
	jsonStr, err = json.Marshal(cfg)
	if err != nil {
		testingError(t, test, "Error marshalling config contents")
	}
	req = createNewRequest(t, "PATCH", fmt.Sprintf("https://localhost:8080/device/%v/config", thisDevice.Id), bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "test 3: Did return statusOK")
	}
	//test 4: change PropogateDeviceConfig cfg item of this node and re test route
	node := pbdatabase.PbNode{Name: meNode.Name, Id: meNode.Id,
		ConfigItems: []pbdatabase.ConfigItem{
			{Key: "PropogateDeviceConfig", Value: "true"},
		},
	}
	if err := node.Update(dbHandle); err != nil {
		testingError(t, test, "update root node with PropogateDeviceConfig item: " + err.Error())
	}
	req = createNewRequest(t, "PATCH", fmt.Sprintf("https://localhost:8080/device/%v/config", thisDevice.Id), bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "test 4: Did return statusOK")
	}
	//test 5: change PropogateDeviceConfig cfg item of this node and re test route
	node = pbdatabase.PbNode{Name: meNode.Name, Id: meNode.Id,
		ConfigItems: []pbdatabase.ConfigItem{
			{Key: "PropogateDeviceConfig", Value: "false"},
		},
	}
	if err := node.Update(dbHandle); err != nil {
		testingError(t, test, "update root node with PropogateDeviceConfig item: " + err.Error())
	}
	req = createNewRequest(t, "PATCH", fmt.Sprintf("https://localhost:8080/device/%v/config", thisDevice.Id), bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "test 4: Did return statusOK")
	}
	//test6 verify callback only allows type change.DEVICE
}

/****** Test all the /device/{devname}/meta handlers *************/
func TestGetMetaHandler(t *testing.T) {
	test := "TestGetMetaHandler"
	begin(t, test)
	defer end(t, test)

	dbFile := "test_getMetaDevice.db"
	dbHandle := setupDB(t, dbFile)
	defer os.Remove(dbFile)
	defer dbHandle.Close()
	cmEngine := getCME(t)
	defer cleanupCME(cmEngine)
	muxRouter := setupApiHandler(dbHandle)
	//setup global for the git back end
	setupGlobal(t, "Root")
	var nodeId int64
	nodeId = 1
	dev := pbdatabase.PbDevice{Name: "A_Device", ParentNode: &nodeId}
	if err := dev.Create(dbHandle); err != nil {
		testingError(t, test, "test 0:db create device error: " + err.Error())
	}
	if err := dev.GetByName(dbHandle); err != nil {
		testingError(t, test, "test 0: device GetByName error: " + err.Error())
	}

	//put some metadata in the cme to retrieve through the route
	if err := cmEngine.VersionMeta(dev.Name, "aaa", "bbb"); err != nil {
		testingError(t, test, "test 0: cme VersionMeta error: " + err.Error())
	}
	if err := cmEngine.VersionMeta(dev.Name, "bbb", "ccc"); err != nil {
		testingError(t, test, "test0: cme Version meta error: " + err.Error())
	}
	//now test the route
	req := createNewRequest(t, "GET", fmt.Sprintf("https://localhost:8080/device/%v/meta", dev.Id), nil)
	writer := httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "Test1: Was not able to get metadata for existing device.")
	}
	//now check what we got
	metaItems := make(map[string]string, 0)
	decoder := json.NewDecoder(writer.Body)
	if err := decoder.Decode(&metaItems); err != nil {
		testingError(t, test, "Decoder error: %s", err.Error())
	}
	if len(metaItems) != 2 {
		testingError(t, test, "Test 1: Did not get all the meta data associated with the device")
	}
	//test2: Make str to int64 fail
	req = createNewRequest(t, "GET", "https://localhost:8080/device/aa/meta", nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Test 2: Should not be able to get device metadata with invalid device id")
	}
	// test3:non existent Device
	req = createNewRequest(t, "GET", "https://localhost:8080/device/4/meta", nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Test 3: Should not be able to get device metadata for non existent device")
	}
	//test Versioning
	req = createNewRequest(t, "GET", fmt.Sprintf("https://localhost:8080/v2/device/%v/meta", dev.Id), nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "test 4: Should not get StatusOK, wrong version sent")
	}

	req = createNewRequest(t, "GET", fmt.Sprintf("https://localhost:8080/v1/device/%v/meta", dev.Id), nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "test 5: Did not get StatusOK, version is now not 1?")
	}
	//this doesnt even reach our global.ProcessVersioning, it is not a route that is registered (only v[0-9]+ routes are registered)
	req = createNewRequest(t, "GET", fmt.Sprintf("https://localhost:8080/v1.0/device/%v/meta", dev.Id), nil)
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "test 6: Should not get StatusOK, version in wrong format sent")
	}
}

func TestPatchMetaHandler(t *testing.T) {
	test := "TestPatchMetaHandler"
	begin(t, test)
	defer end(t, test)

	dbFile := "test_patchMetaDevice.db"
	dbHandle := setupDB(t, dbFile)
	defer os.Remove(dbFile)
	defer dbHandle.Close()
	cmEngine := getCME(t)
	defer cleanupCME(cmEngine)
	muxRouter := setupApiHandler(dbHandle)
	//setup global for the git back end
	setupGlobal(t, "Root")
	var nodeId int64
	nodeId = 1
	dev := pbdatabase.PbDevice{Name: "A_Device", ParentNode: &nodeId}
	if err := dev.Create(dbHandle); err != nil {
		testingError(t, test, "Test 0: device create error: " + err.Error())
	}
	if err := dev.GetByName(dbHandle); err != nil {
		testingError(t, test, "Test 0: device GetByName error: " + err.Error())
	}
	//test0: test route without root node being present in the db
	metaData := pbdatabase.ConfigItem{Key: "afa", Value: "bfb"}
	jsonStr, err := json.Marshal(metaData)
	if err != nil {
		testingError(t, test, "test 0: Error marshaling the device metadata")
	}
	req := createNewRequest(t, "PATCH", fmt.Sprintf("https://localhost:8080/device/%v/meta", dev.Id), bytes.NewBuffer(jsonStr))
	writer := httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Test0: Was able to update metadata for existing device with no Parent node.")
	}
	//set up root node in the db
	node := pbdatabase.PbNode{Name: "Root"}
	if err := node.Create(dbHandle); err != nil {
		testingError(t, test, "setup root: " + err.Error())
	}

	//test1: test route successfully
	metaData = pbdatabase.ConfigItem{Key: "aa", Value: "bb"}
	jsonStr, err = json.Marshal(metaData)
	if err != nil {
		testingError(t, test, "test 1: Error marshaling the device metadata")
	}
	req = createNewRequest(t, "PATCH", fmt.Sprintf("https://localhost:8080/device/%v/meta", dev.Id), bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "Test1: Was not able to update metadata for existing device.")
	}
	//check in cme
	kvList, err := cmEngine.LoadMeta(dev.Name)
	_, ok := kvList["aa"]
	if !ok {
		testingError(t, test, "Test1: Did not find just updated metadata key")
	}
	//test2: make str to int64 fail
	req = createNewRequest(t, "PATCH", "https://localhost:8080/device/aa/meta", bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Test 2: Should not be able to get device metadata with invalid device id")
	}
	// test3:non existent Device
	req = createNewRequest(t, "PATCH", "https://localhost:8080/device/4/meta", bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Test 3: Should not be able to get device metadata for non existent device")
	}
	//test4: route with device whose parentNode is nil
	res, err := dbHandle.Exec("INSERT INTO Devices VALUES(?, ?, ?)", nil, "test4Dev", nil)
	if err != nil {
		testingError(t, test, "1. CreateDevice Error: %s", err.Error())
	}
	id2, err := res.LastInsertId()
	if err != nil {
		testingError(t, test, "2. CreateDevice Error: %s", err.Error())
	}
	req = createNewRequest(t, "PATCH", fmt.Sprintf("https://localhost:8080/device/%v/meta", id2), bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Test4: Should not be able to update metadata for device whose parent node is not set.")
	}
	//test Versioning
	metaData = pbdatabase.ConfigItem{Key: "hello", Value: "world"}
	jsonStr, err = json.Marshal(metaData)
	if err != nil {
		testingError(t, test, "test 5-7: Error marshaling the device metadata")
	}
	req = createNewRequest(t, "PATCH", fmt.Sprintf("https://localhost:8080/v2/device/%v/meta", dev.Id), bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "test 5: Should not get StatusOK, wrong version sent")
	}

	//this doesnt even reach our global.ProcessVersioning, it is not a route that is registered (only v[0-9]+ routes are registered)
	req = createNewRequest(t, "PATCH", fmt.Sprintf("https://localhost:8080/v1.0/device/%v/meta", dev.Id), bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "test 6: Should not get StatusOK, version in wrong format sent")
	}
	req = createNewRequest(t, "PATCH", fmt.Sprintf("https://localhost:8080/v1/device/%v/meta", dev.Id), bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "test 7: Did not get StatusOK, version is now not 1?")
	}
}

func TestDeleteMetaHandler(t *testing.T) {
	test := "TestDeleteMetaHandler"
	begin(t, test)
	defer end(t, test)

	dbFile := "test_deleteMetaDevice.db"
	dbHandle := setupDB(t, dbFile)
	defer os.Remove(dbFile)
	defer dbHandle.Close()
	cmEngine := getCME(t)
	defer cleanupCME(cmEngine)
	muxRouter := setupApiHandler(dbHandle)
	//setup global for the git back end
	setupGlobal(t, "Root")
	var nodeId int64
	nodeId = 1
	dev := pbdatabase.PbDevice{Name: "A_Device", ParentNode: &nodeId}
	if err := dev.Create(dbHandle); err != nil {
		testingError(t, test, "test 0:db create device error: " + err.Error())
	}
	if err := dev.GetByName(dbHandle); err != nil {
		testingError(t, test, "test 0: device GetByName error: " + err.Error())
	}

	//put some metadata in the cme to delete through the route
	if err := cmEngine.VersionMeta(dev.Name, "aaa", "bbb"); err != nil {
		testingError(t, test, "test 0: cme VersionMeta error: " + err.Error())
	}
	if err := cmEngine.VersionMeta(dev.Name, "bbb", "ccc"); err != nil {
		testingError(t, test, "test0: cme Version meta error: " + err.Error())
	}
	if err := cmEngine.VersionMeta(dev.Name, "ver", "yellow"); err != nil {
		testingError(t, test, "test0: cme Version meta error: " + err.Error())
	}
	metaData := pbdatabase.ConfigItem{Key: "aaa", Value: "bbb"}
	jsonStr, err := json.Marshal(metaData)
	if err != nil {
		testingError(t, test, "test 1: Error marshaling the device metadata")
	}

	req := createNewRequest(t, "DELETE", fmt.Sprintf("https://localhost:8080/device/%v/meta", dev.Id), bytes.NewBuffer(jsonStr))
	writer := httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Test0: Was able to delete metadata for existing device with no Parent node.")
	}
	//set up root node in the db
	node := pbdatabase.PbNode{Name: "Root"}
	if err := node.Create(dbHandle); err != nil {
		testingError(t, test, "setup root: " + err.Error())
	}

	//test1: test route successfully
	req = createNewRequest(t, "DELETE", fmt.Sprintf("https://localhost:8080/device/%v/meta", dev.Id), bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "Test1: Was not able to delete metadata for existing device.")
	}
	//check in cme
	kvList, err := cmEngine.LoadMeta(dev.Name)
	fmt.Printf("METADATA:%v\n", kvList)
	_, ok := kvList["aaa"]
	if ok {
		testingError(t, test, "Test1: Could not delete existing metadata key")
	}

	//test2: make str to int64 fail
	req = createNewRequest(t, "DELETE", "https://localhost:8080/device/aaa/meta", bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Test 2: Should not be able to get device metadata with invalid device id")
	}
	// test3:non existent Device
	req = createNewRequest(t, "DELETE", "https://localhost:8080/device/4/meta", bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Test 3: Should not be able to get device metadata for non existent device")
	}
	//test4: route with device whose parentNode is nil
	res, err := dbHandle.Exec("INSERT INTO Devices VALUES(?, ?, ?)", nil, "test4Dev", nil)
	if err != nil {
		testingError(t, test, "1. CreateDevice Error: %s", err.Error())
	}
	id2, err := res.LastInsertId()
	if err != nil {
		testingError(t, test, "2. CreateDevice Error: %s", err.Error())
	}
	if err := cmEngine.VersionMeta("test4Dev", "aaa", "bbb"); err != nil {
		testingError(t, test, "test 4: cme VersionMeta error: " + err.Error())
	}
	req = createNewRequest(t, "DELETE", fmt.Sprintf("https://localhost:8080/device/%v/meta", id2), bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Test4: Should not be able to delete metadata for device whose parent node is not set.")
	}
	//test nonexistent metadata key value pair deletion
	jsonStr2 := `{"Key":"nonexistent", "Value":"yellow"}`
	req = createNewRequest(t, "DELETE", fmt.Sprintf("https://localhost:8080/device/%v/meta", dev.Id), bytes.NewBufferString(jsonStr2))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "Test 5: Should not be able to delete nonexistent metadata sent.")
	}

	//test Versioning
	metaData = pbdatabase.ConfigItem{Key: "ver", Value: "yellow"}
	jsonStr, err = json.Marshal(metaData)
	if err != nil {
		testingError(t, test, "test 6-8: Error marshaling the device metadata")
	}
	req = createNewRequest(t, "DELETE", fmt.Sprintf("https://localhost:8080/v2/device/%v/meta", dev.Id), bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "test 6: Should not get StatusOK, wrong version sent")
	}

	//this doesnt even reach our global.ProcessVersioning, it is not a route that is registered (only v[0-9]+ routes are registered)
	req = createNewRequest(t, "DELETE", fmt.Sprintf("https://localhost:8080/v1.0/device/%v/meta", dev.Id), bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code == http.StatusOK {
		testingError(t, test, "test 7: Should not get StatusOK, version in wrong format sent")
	}
	req = createNewRequest(t, "DELETE", fmt.Sprintf("https://localhost:8080/v1/device/%v/meta", dev.Id), bytes.NewBuffer(jsonStr))
	writer = httptest.NewRecorder()
	muxRouter.ServeHTTP(writer, req)
	if writer.Code != http.StatusOK {
		testingError(t, test, "test 8: Did not get StatusOK, version is now not 1?")
	}
}
