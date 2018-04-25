package device

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
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"encoding/json"
	mux "github.com/gorilla/mux"
	change "github.com/iti/pbconf/lib/pbchange"
	database "github.com/iti/pbconf/lib/pbdatabase"
	global "github.com/iti/pbconf/lib/pbglobal"
	internode "github.com/iti/pbconf/lib/pbinternode"
	logging "github.com/iti/pbconf/lib/pblogger"
	ontology "github.com/iti/pbconf/lib/pbontology"
	trans "github.com/iti/pbconf/lib/pbtranslate"
	validator "gopkg.in/validator.v2"
)

type APIHandler struct {
	log     logging.Logger
	db      database.AppDatabase
	Version int
}

var nodeComm *internode.InterNode

func NewAPIHandler(loglevel string, d database.AppDatabase) *APIHandler {
	l, _ := logging.GetLogger("Device API")
	logging.SetLevel(loglevel, "Device API")
	nodeComm = internode.NewInterNodeCommunicator(d)
	return &APIHandler{log: l, db: d, Version: 1}
}

func (a *APIHandler) AddAPIEndpoints(router *mux.Router) {
	a.log.Debug("Registering device endpoints")
	//register with the change management engine
	engine, err := change.GetCMEngine(nil)
	if err != nil {
		a.log.Debug("Could not get an instance of Change management Engine.")
	}
	engine.RegisterCommitListener(change.DEVICE, *a, CfgFinalizedCallback)
	engine.RegisterPackRcvdListener(change.DEVICE, *a, CfgPackRcvdCallback)

	for _, v := range global.ApiUrlVersioning {
		router.HandleFunc(v+"/device", a.handleBaseRoute).Methods("HEAD", "GET", "POST", "PATCH")

		s := router.PathPrefix(v + "/device").Subrouter()
		s.HandleFunc("/", a.handleBaseRoute).Methods("HEAD", "GET", "POST", "PATCH")
		s.HandleFunc("/{devid}", a.handleWIdRoute).Methods("GET", "PATCH", "DELETE")
		s.HandleFunc("/{devid}/config", a.handleConfig).Methods("GET", "PATCH")
		s.HandleFunc("/{devid}/meta", a.handleMeta).Methods("GET", "PATCH", "DELETE")
		s.HandleFunc("/{devid}/{cfgkey}", a.handleWIdRouteCfgItem).Methods("GET", "DELETE")
		// Change management hook
		cme := s.PathPrefix("/cme").Subrouter()
		cme.HandleFunc(`/{rest:[a-zA-Z0-9=\-\/]+}`, engine.RequestHandler())
	}
}

func (a *APIHandler) GetInfo() (string, int) {
	return "device", a.Version
}

// handleBaseRoute handles the GET, PUT, POST routes to "/device" or subroute "/"
func (a *APIHandler) handleBaseRoute(writer http.ResponseWriter, req *http.Request) {
	//deal with any version in the URL or Accept header
	version, err := global.ProcessVersioning(req)
	if err != nil {
		a.log.Debug("Versioning: %s", err.Error())
		writer.WriteHeader(http.StatusNotFound)
		return
	}
	if version != nil && *version > a.Version {
		a.log.Debug("Version %v not implemented, current version is %v", *version, a.Version)
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	resp := &logging.ResponseLogger{
		ResponseWriter: writer,
		Logger:         a.log,
		SrcNodeName:    global.RootNode,
	}
	switch req.Method {
	case "HEAD":
		a.headBaseHandler(writer, req)
	case "GET":
		a.getBaseHandler(resp, req)
	case "POST":
		a.postBaseHandler(resp, req)
	case "PATCH":
		a.patchBaseHandler(resp, req)
	}
}

func (a *APIHandler) headBaseHandler(writer http.ResponseWriter, req *http.Request) {
	timetag, err := a.db.GetDevicesLastModified()
	if err != nil {
		a.log.Warning("HEAD /device::Could not recover last modified tag for Devices Error")
	} else {
		writer.Header().Set("Last-Modified", timetag)
	}
	writer.WriteHeader(http.StatusOK)
}

// getBaseHandler handles the Get route for "/". It returns a list of devices
func (a *APIHandler) getBaseHandler(resp *logging.ResponseLogger, req *http.Request) {
	devList, err := a.db.GetDevices()
	if err != nil {
		resp.WriteLog(http.StatusNotFound, "Error", "GET /device::Could not get devices from the database for the node.")
		return
	}
	jsonStr, err := json.Marshal(devList)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Notice", "GET /device::Could not marshal the devices for the node Error: %s", err.Error())
		return
	}
	if _, err = resp.Write(jsonStr); err != nil {
		resp.WriteLog(http.StatusBadRequest, "Notice", "GET /device::Writing response body Error: %s", err.Error())
		return
	}
}

//postBaseHandler handles the POST route for "/". It reads the request body and creates a new device.
// If in the body, the ParentNode is nil, add the parent node= global.rootnode, pass it upstream
// if in the body, the ParentNode != nil, add to db, replace the ParentNode=global.rootnode pass upstream
//this will ensure each upstream node adds the childnode as the parent of this device
func (a *APIHandler) postBaseHandler(resp *logging.ResponseLogger, req *http.Request) {
	var newDevice database.PbDevice
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&newDevice)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Notice", "POST /device::Decoder error: %s", err.Error())
		return
	}
	node, err := nodeComm.GetRootNode()
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Warning", "POST /device::Could not recover root node from database, cannot proceed")
		return
	}
	//set the parent node for the newly created/updated device if its not already set
	if newDevice.ParentNode == nil {
		newDevice.ParentNode = &node.Id
	}

	if err = validator.Validate(newDevice); err != nil {
		resp.WriteLog(http.StatusBadRequest, "Notice", "POST /device::validation error: %s", err.Error())
		return
	}

	if newDevice.Id > 0 {
		resp.WriteLog(http.StatusNotImplemented, "Notice", "POST /device::Device id specified")
		return
	}

	err = newDevice.Create(a.db)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Notice", "POST /device::Error creating device")
		return
	}
	resp.Header().Set("Location", req.URL.String()+"/"+strconv.FormatInt(newDevice.Id, 10))
	resp.WriteHeader(http.StatusCreated)

	device := newDevice //make a copy
	add_to_q, err := nodeComm.AddDeviceUpstream(device)
	if err != nil {
		a.log.Warning("POST /device::Was unable to add the device to the upstream node. OUT OF SYNC")
	}
	if add_to_q {
		internode.Qmutex.Lock()
		global.Queue.PushBack(internode.Qelement{
			Fn_name: "AddDeviceUpstream",
			Device:  &newDevice,
			Param:   "",
			Param2:  "",
			Cfg:     nil,
		})
		internode.Qmutex.Unlock()
	}
}

func (a *APIHandler) patchBaseHandler(resp *logging.ResponseLogger, req *http.Request) {
	//recover id from the request Body
	var device database.PbDevice
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&device)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Info", "PATCH /device::Decoder error: %s", err.Error())
		return
	}
	a.patchDeviceHierarchy(device, resp, req)
}

/******************** START /device/{id} route ******************************************/

// handleWIdRoute handles the GET, PUT, PATCH, DELETE for the "/{devid}" route
func (a *APIHandler) handleWIdRoute(writer http.ResponseWriter, req *http.Request) {
	//deal with any version in the URL or Accept header
	version, err := global.ProcessVersioning(req)
	if err != nil {
		a.log.Debug("Versioning: %s", err.Error())
		writer.WriteHeader(http.StatusNotFound)
		return
	}
	if version != nil && *version > a.Version {
		a.log.Debug("Version %v not implemented, current version is %v", *version, a.Version)
		writer.WriteHeader(http.StatusNotFound)
		return
	}
	resp := &logging.ResponseLogger{
		ResponseWriter: writer,
		Logger:         a.log,
		SrcNodeName:    global.RootNode,
	}
	switch req.Method {
	case "GET":
		a.getWIdHandler(resp, req)
	case "PATCH":
		a.patchWIdHandler(resp, req)
	case "DELETE":
		a.deleteWIdHandler(resp, req)
	}
}

// getWIdHandler handles the GET route for "/{devid}" route. It returns complete device details
func (a *APIHandler) getWIdHandler(resp *logging.ResponseLogger, req *http.Request) {
	params := mux.Vars(req)
	deviceId, err := a.parseIdFromRoute(params["devid"]) // string to int64
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Info", "GET /device/{id}::Could not recover device id from route.")
		return
	}
	dbDev := database.PbDevice{Id: deviceId}
	err = dbDev.Get(a.db)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Info", err.Error())
		return
	}
	jsonStr, err := json.Marshal(dbDev)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Notice", "GET /device/{id}::Could not marshal the requested device, error: %s", err.Error())
		return
	}
	if _, err = resp.Write(jsonStr); err != nil {
		resp.WriteLog(http.StatusBadRequest, "Notice", "GET /device/{id}::Writing response body Error: %s", err.Error())
	}
}

// patchWIdHandler handles the PATCH request on "/{devid}" route.
// Allows partial update.
func (a *APIHandler) patchWIdHandler(resp *logging.ResponseLogger, req *http.Request) {
	params := mux.Vars(req)
	deviceId, err := a.parseIdFromRoute(params["devid"]) // string to int64
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Info", "PATCH /device/{id}::Could not recover device id from route.")
		return
	}
	//get device from the request body
	var device database.PbDevice
	decoder := json.NewDecoder(req.Body)
	err = decoder.Decode(&device)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Info", "PATCH /device/{id}::Decoder error: %s", err.Error())
		return
	}
	device.Id = deviceId //route param takes precedence
	a.patchDeviceHierarchy(device, resp, req)
}

// deleteWIdHandler handles the DELETE request for the "/{devid}" route. Deletes the entity at devid.
func (a *APIHandler) deleteWIdHandler(resp *logging.ResponseLogger, req *http.Request) {
	params := mux.Vars(req)
	deviceId, err := a.parseIdFromRoute(params["devid"]) // string to int64
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Info", "DELETE /device/{id}::Could not recover device id from route.")
		return
	}

	// get this devices' info
	device := database.PbDevice{Id: deviceId}
	if err := device.Get(a.db); err != nil {
		resp.WriteLog(http.StatusBadRequest, "Info", "DELETE /device/{id}::Could not recover the device from the database")
		return
	}

	a.deleteDeviceHierarachy(device, resp, req)
}

/*********************** /device/{id}/{cfgkey} routes **********************************/

func (a *APIHandler) handleWIdRouteCfgItem(writer http.ResponseWriter, req *http.Request) {
	//deal with any version in the URL or Accept header
	version, err := global.ProcessVersioning(req)
	if err != nil {
		a.log.Debug("Versioning: %s", err.Error())
		writer.WriteHeader(http.StatusNotFound)
		return
	}
	if version != nil && *version > a.Version {
		a.log.Debug("Version %v not implemented, current version is %v", *version, a.Version)
		writer.WriteHeader(http.StatusNotFound)
		return
	}
	resp := &logging.ResponseLogger{
		ResponseWriter: writer,
		Logger:         a.log,
		SrcNodeName:    global.RootNode,
	}
	switch req.Method {
	case "GET":
		a.getConfigItemHandler(resp, req)
	case "DELETE":
		a.deleteConfigItemHandler(resp, req)
	}
}

func (a *APIHandler) getConfigItemHandler(resp *logging.ResponseLogger, req *http.Request) {
	params := mux.Vars(req)
	deviceId, err := a.parseIdFromRoute(params["devid"]) // string to int64
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Info", "GET /device/{id}/{cfgkey}::Could not recover device id from route.")
		return
	}
	cfgItem := database.PbDeviceConfigItem{DeviceId: deviceId, ConfigItem: database.ConfigItem{Key: params["cfgkey"]}}
	exists, err := cfgItem.Exists(a.db)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Info", "GET /device/{id}/{cfgkey}::Error checking existence of device cfgitem")
		return
	}
	if !exists {
		resp.WriteLog(http.StatusNotFound, "Debug", "GET /device/{id}/{cfgkey}::device cfgitem does not seem to exist in the database!")
		return
	}
	err = cfgItem.Get(a.db)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Notice", "GET /device/{id}/{cfgkey}::Was not able to get the config item from the database")
		return
	}
	jsonStr, err := json.Marshal(cfgItem)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Notice", "GET /device/{id}/{cfgkey}::Could not marshal the requested config item of the device, Error: %s", err.Error())
		return
	}
	if _, err = resp.Write(jsonStr); err != nil {
		resp.WriteLog(http.StatusBadRequest, "Notice", "GET /device/{id}/{cfgkey}::Writing response body Error: %s", err.Error())
	}
}

func (a *APIHandler) deleteConfigItemHandler(resp *logging.ResponseLogger, req *http.Request) {
	params := mux.Vars(req)
	deviceId, err := a.parseIdFromRoute(params["devid"]) // string to int64
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Info", "DELETE /device/{id}/{cfgkey}::Could not recover device id from route.")
		return
	}

	device := database.PbDevice{Id: deviceId}
	if err := device.Get(a.db); err != nil {
		resp.WriteLog(http.StatusBadRequest, "Debug", "deleteConfigItemHandler thisNode error: %s", err.Error())
		return
	}
	a.deleteConfigItemHierarchy(device, params["cfgkey"], resp, req)
}

/******************************Device Meta data routes ******************************/
func (a *APIHandler) handleMeta(writer http.ResponseWriter, req *http.Request) {
	//deal with any version in the URL or Accept header
	version, err := global.ProcessVersioning(req)
	if err != nil {
		a.log.Debug("Versioning: %s", err.Error())
		writer.WriteHeader(http.StatusNotFound)
		return
	}
	if version != nil && *version > a.Version {
		a.log.Debug("Version %v not implemented, current version is %v", *version, a.Version)
		writer.WriteHeader(http.StatusNotFound)
		return
	}
	resp := &logging.ResponseLogger{
		ResponseWriter: writer,
		Logger:         a.log,
		SrcNodeName:    global.RootNode,
	}
	switch req.Method {
	case "GET":
		a.getMetaHandler(resp, req)
	case "PATCH":
		a.patchMetaHandler(resp, req)
	case "DELETE":
		a.deleteMetaHandler(resp, req)
	}
}

//getMetaHandler returns the meta data key value pair list for the device
func (a *APIHandler) getMetaHandler(resp *logging.ResponseLogger, req *http.Request) {
	params := mux.Vars(req)
	deviceId, err := a.parseIdFromRoute(params["devid"]) // string to int64
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Info", "GET /device/{id}/meta::Could not recover device id from route.")
		return
	}
	device := database.PbDevice{Id: deviceId}
	if err := device.Get(a.db); err != nil {
		resp.WriteLog(http.StatusBadRequest, "Debug", "getMetaHandler thisNode error: %s", err.Error())
		return
	}
	changeEng, err := change.GetCMEngine(nil)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Notice", "GET /device/{id}/meta::Could not get instance of CME, error: %s", err.Error())
		return
	}
	kvlist, err := changeEng.LoadMeta(device.Name)
	jsonStr, err := json.Marshal(kvlist)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Notice", "GET /device/{id}/meta::Could not marshal the requested device meta data, error: %s", err.Error())
		return
	}
	if _, err = resp.Write(jsonStr); err != nil {
		resp.WriteLog(http.StatusBadRequest, "Notice", "GET /device/{id}/meta::Writing response body Error: %s", err.Error())
	}

}

func (a *APIHandler) patchMetaHandler(resp *logging.ResponseLogger, req *http.Request) {
	params := mux.Vars(req)
	deviceId, err := a.parseIdFromRoute(params["devid"]) // string to int64
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Info", "PATCH /device/{id}/meta::Could not recover device id from route.")
		return
	}

	device := database.PbDevice{Id: deviceId}
	exists, err := device.ExistsById(a.db)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Info", "PATCH /device/{id}/meta:: Error checking existence of device in the database.")
		return
	}
	if !exists {
		resp.WriteLog(http.StatusNotFound, "Info", "PATCH /device/{id}/meta:: Could not find device in the database")
		return
	}
	if err := device.Get(a.db); err != nil {
		resp.WriteLog(http.StatusBadRequest, "Debug", "PATCH /device/{id}/meta:: thisNode error: %s", err.Error())
		return
	}

	var metaItem database.ConfigItem
	decoder := json.NewDecoder(req.Body)
	err = decoder.Decode(&metaItem)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Notice", "PATCH /device/{id}/meta::Decoder error: %s", err.Error())
		return
	}
	a.patchDeviceMetadataHierarchy(device, strings.ToLower(metaItem.Key), metaItem.Value, resp, req)
}

func (a *APIHandler) deleteMetaHandler(resp *logging.ResponseLogger, req *http.Request) {
	params := mux.Vars(req)
	deviceId, err := a.parseIdFromRoute(params["devid"]) // string to int64
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Info", "DELETE /device/{id}/meta::Could not recover device id from route.")
		return
	}

	device := database.PbDevice{Id: deviceId}
	exists, err := device.ExistsById(a.db)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Info", "DELETE /device/{id}/meta:: Error checking existence of device in the database.")
		return
	}
	if !exists {
		resp.WriteLog(http.StatusNotFound, "Info", "DELETE /device/{id}/meta:: Could not find device in the database")
		return
	}
	if err := device.Get(a.db); err != nil {
		resp.WriteLog(http.StatusBadRequest, "Debug", "DELETE /device/{id}/meta:: thisNode error: %s", err.Error())
		return
	}

	var metaItem database.ConfigItem
	decoder := json.NewDecoder(req.Body)
	err = decoder.Decode(&metaItem)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Notice", "DELETE /device/{id}/meta::Decoder error: %s", err.Error())
		return
	}
	metaItemKey := strings.ToLower(metaItem.Key)
	//confirm the metadata key exists in device
	changeEng, err := change.GetCMEngine(nil)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Notice", "DELETE /device/{id}/meta::Could not get instance of CME, error: %s", err.Error())
		return
	}

	kvlist, err := changeEng.LoadMeta(device.Name)
	if _, ok := kvlist[metaItemKey]; ok {
		a.deleteDeviceMetadataHierarchy(device, metaItemKey, metaItem.Value, resp, req)
		return
	}
	resp.WriteLog(http.StatusNotFound, "Info", "DELETE /device/{id}/meta:: Could not find metadata key in the device database")
}

/******************************Device Config File routes ******************************/
func (a *APIHandler) handleConfig(writer http.ResponseWriter, req *http.Request) {
	//deal with any version in the URL or Accept header
	version, err := global.ProcessVersioning(req)
	if err != nil {
		a.log.Debug("Versioning: %s", err.Error())
		writer.WriteHeader(http.StatusNotFound)
		return
	}
	if version != nil && *version > a.Version {
		a.log.Debug("Version %v not implemented, current version is %v", *version, a.Version)
		writer.WriteHeader(http.StatusNotFound)
		return
	}
	resp := &logging.ResponseLogger{
		ResponseWriter: writer,
		Logger:         a.log,
		SrcNodeName:    global.RootNode,
	}
	switch req.Method {
	case "GET":
		a.getConfigHandler(resp, req)
	case "PATCH":
		a.patchConfigHandler(resp, req)
	}
}

//getConfigHandler returns the configuration file location for the device
func (a *APIHandler) getConfigHandler(resp *logging.ResponseLogger, req *http.Request) {
	params := mux.Vars(req)
	deviceId, err := a.parseIdFromRoute(params["devid"]) // string to int64
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Info", "GET /device/{id}/config::Could not recover device id from route.")
		return
	}

	dbDev := database.PbDevice{Id: deviceId}
	exists, err := dbDev.ExistsById(a.db)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Info", "GET /device/{id}/config::getConfigHandler error checking existence of device.")
		return
	}
	if !exists {
		resp.WriteLog(http.StatusNotFound, "Info", "GET /device/{id}/config::Could not find device in Devices database table")
		return
	}
	err = dbDev.Get(a.db)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Info", "GET /device/{id}/config:: Error getting device from the database.")
		return
	}
	content, err := a.getDeviceConfigurationContentFromRepo(dbDev.Name)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Info", "GET /device/{id}/config:: Error getting config from the repository. Error:%s", err.Error())
	}
	if _, err = resp.Write(content); err != nil {
		resp.WriteLog(http.StatusBadRequest, "Notice", "GET /device/{id}/config::Writing response body Error: %s", err.Error())
	}
}

// patchConfigHandler handles the PATCH route for "/device/{devid}/config". It reads the request body and
// 1) If configuration does not exist for the device, create a new configuration for that device Id
// 2) If configuration already exists for the device, update the configuration field
func (a *APIHandler) patchConfigHandler(resp *logging.ResponseLogger, req *http.Request) {
	params := mux.Vars(req)
	deviceId, err := a.parseIdFromRoute(params["devid"]) // string to int64
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Info", "PATCH /device/{id}/config::Could not recover device id from route.")
		return
	}

	dbDev := database.PbDevice{Id: deviceId}
	exists, err := dbDev.ExistsById(a.db)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Info", "PATCH /device/{id}/config:: Error checking existence of device in the database.")
		return
	}
	if !exists {
		resp.WriteLog(http.StatusNotFound, "Info", "PATCH /device/{id}/config:: Could not find device in the database")
		return
	}
	err = dbDev.Get(a.db)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Info", "PATCH /device/{id}/config::Error getting device from the database")
		return
	}
	//get the updated device config from the request body
	var cfg change.ChangeData
	decoder := json.NewDecoder(req.Body)
	err = decoder.Decode(&cfg)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Notice", "PATCH /device/{id}/config::Decoder error: %s", err.Error())
		return
	}
	cfg.ObjectType = change.DEVICE
	a.patchDeviceConfigHierarchy(dbDev, &cfg, resp, req)
}

/********************Non route helper functions *******************/
func (a *APIHandler) patchDeviceHierarchy(device database.PbDevice, resp *logging.ResponseLogger, req *http.Request) {
	if device.ParentNode == nil { //can't do anything without a parent node specified
		resp.WriteLog(http.StatusBadRequest, "Debug", "PATCH /device:: device does not have any parent node. Error in forming the request.")
		return
	}
	rootnode, err := nodeComm.GetRootNode()
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Warning", "PATCH /device::Could not recover root node from database, cannot proceed")
		return
	}
	originatingIP := strings.Split(req.RemoteAddr, ":")[0] //where did the request come from
	downstreamIP := "Something not matching originatingIP"
	downstreamIPInfo := ""
	if rootnode.Id != *device.ParentNode { //if root node is not the parent of the device, some other node is, recover that nodes IP address
		downstreamIPInfo, err = nodeComm.GetDownstreamNodeIP(*device.ParentNode)
		if err != nil {
			resp.WriteLog(http.StatusBadRequest, "Warning", "PATCH /device::IP Address of the parent node for the device could not be recovered. Cannot proceed")
			return
		}
		downstreamIP = strings.Split(downstreamIPInfo, ":")[0]
	}
	//retrieve the device name from the database(in case name is changed, the Update upstream/downstream will fail)
	dbDev := database.PbDevice{Id: device.Id}
	if err = dbDev.Get(a.db); err != nil {
		resp.WriteLog(http.StatusBadRequest, "Info", "PATCH /device:: Could not get existing device details from the database")
		return
	}
	// now the logic
	if rootnode.Id == *device.ParentNode || originatingIP == downstreamIP {
		//save to db
		err2 := device.Update(a.db)
		//update Repo
		if dbDev.Name != device.Name {
			err = a.renameDeviceConfigurationInRepo(dbDev.Name, device.Name)
			if err != nil {
				a.log.Warning("PATCH /device:: Was not able to rename the device configuration in the repository. Error:%s", err.Error())
			}
		}
		//send it up
		upDevice := device                                                   //make a copy, UpdateDeviceUpstream could modify parentNode pointer, etc
		add_to_q, err := nodeComm.UpdateDeviceUpstream(upDevice, dbDev.Name) //if error here put in queue
		if err != nil {
			a.log.Warning("PATCH /device::Was unable to update the device on the upstream node. OUT OF SYNC")
		}
		if add_to_q {
			internode.Qmutex.Lock()
			global.Queue.PushBack(internode.Qelement{
				Fn_name: "UpdateDeviceUpstream",
				Device:  &device,
				Param:   dbDev.Name,
				Param2:  "",
				Cfg:     nil,
			})
			internode.Qmutex.Unlock()
		}
		if err2 != nil { //db error at this level now what?
			resp.WriteLog(http.StatusBadRequest, "Warning", "PATCH /device::Was unable to update the device in the database on this node. OUT OF SYNC")
			return
		}
	} else {
		//send it down
		downDevice := device
		success := nodeComm.UpdateDeviceDownstream(downDevice, dbDev.Name, downstreamIPInfo) //if error here tell user no can do
		if !success {
			a.log.Debug("Try again later !!!!!!")
			resp.WriteLog(http.StatusBadRequest, "Notice", "PATCH /device::Could not send the device change to the parent node of the device.")
			return
		}
	}
	resp.WriteHeader(http.StatusOK)
}

func (a *APIHandler) deleteDeviceHierarachy(device database.PbDevice, resp *logging.ResponseLogger, req *http.Request) {
	if device.ParentNode == nil { //can't do anything without a parent node specified
		resp.WriteLog(http.StatusBadRequest, "Debug", "DELETE /device:: device does not have any parent node. Error in forming the request.")
		return
	}
	rootnode, err := nodeComm.GetRootNode()
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Warning", "DELETE /device::Could not recover root node from database, cannot proceed")
		return
	}
	originatingIP := strings.Split(req.RemoteAddr, ":")[0] //where did the request come from
	downstreamIP := "Something not matching originatingIP"
	downstreamIPInfo := ""
	if rootnode.Id != *device.ParentNode { //if root node is not the parent of the device, some other node is, recover that nodes IP address
		downstreamIPInfo, err = nodeComm.GetDownstreamNodeIP(*device.ParentNode)
		if err != nil {
			resp.WriteLog(http.StatusBadRequest, "Warning", "DELETE /device::IP Address of the parent node for the device could not be recovered. Cannot proceed")
			return
		}
		downstreamIP = strings.Split(downstreamIPInfo, ":")[0]
	}

	// now the logic
	if rootnode.Id == *device.ParentNode || originatingIP == downstreamIP {
		upDevice := device //make a copy
		cacheDevice := device
		//save to db
		err = device.Delete(a.db)
		//delete from Repo
		rerr := a.deleteDeviceConfigurationFromRepo(cacheDevice.Name)
		if rerr != nil {
			a.log.Info("DELETE /device::Couldnt delete config files from repository, Error:%s", rerr.Error())
		}
		//send it up
		add_to_q, err2 := nodeComm.DeleteDeviceUpstream(upDevice) //if error here put in queue
		if err2 != nil {
			a.log.Warning("DELETE /device::Was unable to delete the device on the upstream node. OUT OF SYNC")
		}
		if add_to_q {
			internode.Qmutex.Lock()
			global.Queue.PushBack(internode.Qelement{
				Fn_name: "DeleteDeviceUpstream",
				Device:  &cacheDevice,
				Param:   "",
				Param2:  "",
				Cfg:     nil,
			})
			internode.Qmutex.Unlock()
		}

		if err != nil { //db error at this level now what?
			resp.WriteLog(http.StatusBadRequest, "Warning", "DELETE /device::Was unable to delete the device in the database on this node. OUT OF SYNC")
			return
		}
	} else {
		//send it down
		downDevice := device
		success := nodeComm.DeleteDeviceDownstream(downDevice, downstreamIPInfo) //if error here tell user no can do
		if !success {
			a.log.Debug("Try again later !!!!!!")
			resp.WriteLog(http.StatusBadRequest, "Notice", "DELETE /device::Could not send the device deletion to the parent node of the device.")
			return
		}
	}
	resp.WriteHeader(http.StatusOK)
}

func (a *APIHandler) deleteConfigItemHierarchy(device database.PbDevice, cfgKey string, resp *logging.ResponseLogger, req *http.Request) {
	if device.ParentNode == nil { //can't do anything without a parent node specified
		resp.WriteLog(http.StatusBadRequest, "Debug", "DELETE /device/{id}/{cfgkey}:: device does not have any parent node. Error in forming the request.")
		return
	}
	rootnode, err := nodeComm.GetRootNode()
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Warning", "DELETE /device/{id}/{cfgkey}::Could not recover root node from database, cannot proceed")
		return
	}
	originatingIP := strings.Split(req.RemoteAddr, ":")[0] //where did the request come from
	downstreamIP := "Something not matching originatingIP"
	downstreamIPInfo := ""
	if rootnode.Id != *device.ParentNode { //if root node is not the parent of the device, some other node is, recover that nodes IP address
		downstreamIPInfo, err = nodeComm.GetDownstreamNodeIP(*device.ParentNode)
		if err != nil {
			resp.WriteLog(http.StatusBadRequest, "Warning", "DELETE /device/{id}/{cfgkey}::IP Address of the parent node for the device could not be recovered. Cannot proceed")
			return
		}
		downstreamIP = strings.Split(downstreamIPInfo, ":")[0]
	}

	// now the logic
	if rootnode.Id == *device.ParentNode || originatingIP == downstreamIP {
		//save to db
		dbConfigEl := database.PbDeviceConfigItem{DeviceId: device.Id, ConfigItem: database.ConfigItem{Key: cfgKey}}
		err = dbConfigEl.Delete(a.db)
		//send it up
		upDevice := device                                                          //make a copy
		add_to_q, err2 := nodeComm.DeleteDeviceConfigItemUpstream(upDevice, cfgKey) //if error here put in queue
		if err2 != nil {
			a.log.Warning("DELETE /device/{id}/{cfgkey}::Was unable to delete the device config item on the upstream node. OUT OF SYNC")
		}
		if add_to_q {
			internode.Qmutex.Lock()
			global.Queue.PushBack(internode.Qelement{
				Fn_name: "DeleteDeviceConfigItemUpstream",
				Device:  &device,
				Param:   cfgKey,
				Param2:  "",
				Cfg:     nil,
			})
			internode.Qmutex.Unlock()
		}

		if err != nil { //db error at this level now what?
			resp.WriteLog(http.StatusBadRequest, "Warning", "DELETE /device/{id}/{cfgkey}::Was unable to delete the device config item in the database on this node. OUT OF SYNC")
			return
		}
	} else {
		//send it down
		downDevice := device
		success := nodeComm.DeleteDeviceConfigItemDownstream(downDevice, cfgKey, downstreamIPInfo) //if error here tell user no can do
		if !success {
			resp.WriteLog(http.StatusBadRequest, "Notice", "DELETE /device/{id}/{cfgkey}::Could not send the device config item deletion to the parent node of the device.")
			return
		}
	}
	resp.WriteHeader(http.StatusOK)
}

func (a *APIHandler) patchDeviceConfigHierarchy(device database.PbDevice, cfg *change.ChangeData, resp *logging.ResponseLogger, req *http.Request) {
	if device.ParentNode == nil { //can't do anything without a parent node specified
		resp.WriteLog(http.StatusBadRequest, "Debug", "PATCH /device/{id}/config:: device does not have any parent node. Error in forming the request.")
		return
	}
	rootnode, err := nodeComm.GetRootNode()
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Warning", "PATCH /device/{id}/config::Could not recover root node from database, cannot proceed")
		return
	}

	originatingIP := strings.Split(req.RemoteAddr, ":")[0] //where did the request come from
	downstreamIP := "Something not matching originatingIP"
	downstreamIPInfo := ""
	if rootnode.Id != *device.ParentNode { //if root node is not the parent of the device, some other node is, recover that nodes IP address
		downstreamIPInfo, err = nodeComm.GetDownstreamNodeIP(*device.ParentNode)
		if err != nil {
			resp.WriteLog(http.StatusBadRequest, "Warning", "PATCH /device/{id}/config::IP Address of the parent node for the device could not be recovered. Cannot proceed")
			return
		}
		downstreamIP = strings.Split(downstreamIPInfo, ":")[0]
	}

	if err = a.completeConfigurationDataStruct(cfg); err != nil {
		resp.WriteLog(http.StatusBadRequest, "Notice", "Config missing fields, error: %s", err.Error())
		return
	}

	changeEng, err := change.GetCMEngine(nil)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Notice", "PATCH /device/{id}/config::Could not get instance of CME, error: %s", err.Error())
		return
	}

	transID, err := changeEng.BeginTransaction(cfg, cfg.Log.Message)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Debug", "PATCH /device/{id}/config:::BeginTransaction error:%s", err.Error())
		return
	}
	cfg.TransactionID = transID

	// now the logic
	if rootnode.Id == *device.ParentNode || originatingIP == downstreamIP {
		device_ok := false
		var buf *bytes.Buffer
		for _, v := range cfg.Content.Files {
			buf = bytes.NewBuffer(v)
			break
		}

		err := a.checkDeviceConfigWOntology(device, buf, changeEng)
		if err != nil {
			resp.WriteLog(http.StatusBadRequest, "Info", "PATCH /device/{id}/config::%s", err)
			return
		}

		if rootnode.Id == *device.ParentNode {
			device_ok, err = a.updatePhysicalDeviceWithCfg(device, buf)
			if err != nil {
				resp.WriteLog(http.StatusBadRequest, "Info", "PATCH /device/{id}/config::Was not able to update the configuration on the device. Cannot proceed")
				return
			}
		}

		if device_ok == true || originatingIP == downstreamIP {
			// at this point go ahead and store the modified config in the git repository. The callback from the repo will handle sending it up
			saveToRepoErr := a.saveDeviceConfigurationToRepo(cfg)
			if saveToRepoErr != nil {
				resp.WriteLog(http.StatusBadRequest, "Warning", "PATCH /device/{id}/config::Was unable to save the device config in the repository on this node. OUT OF SYNC")
				return
			}
		}
	} else {
		//send it down
		downDevice := device
		success := nodeComm.UpdateDeviceConfigDownstream(downDevice, cfg, downstreamIPInfo) //if error here tell user no can do
		if !success {
			resp.WriteLog(http.StatusBadRequest, "Notice", "PATCH /device/{id}/config::Could not send the device config to the parent node of the device.")
			return
		}
	}
	resp.WriteHeader(http.StatusOK)
}

func (a *APIHandler) patchDeviceMetadataHierarchy(device database.PbDevice, metaKey string, metaValue string, resp *logging.ResponseLogger, req *http.Request) {
	if device.ParentNode == nil { //can't do anything without a parent node specified
		resp.WriteLog(http.StatusBadRequest, "Debug", "PATCH /device/{id}/meta:: device does not have any parent node. Error in forming the request.")
		return
	}
	rootnode, err := nodeComm.GetRootNode()
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Warning", "PATCH /device/{id}/meta::Could not recover root node from database, cannot proceed")
		return
	}
	originatingIP := strings.Split(req.RemoteAddr, ":")[0] //where did the request come from
	downstreamIP := "Something not matching originatingIP"
	downstreamIPInfo := ""
	if rootnode.Id != *device.ParentNode { //if root node is not the parent of the device, some other node is, recover that nodes IP address
		downstreamIPInfo, err = nodeComm.GetDownstreamNodeIP(*device.ParentNode)
		if err != nil {
			resp.WriteLog(http.StatusBadRequest, "Warning", "PATCH /device/{id}/meta::IP Address of the parent node for the device could not be recovered. Cannot proceed")
			return
		}
		downstreamIP = strings.Split(downstreamIPInfo, ":")[0]
	}

	// now the logic
	if rootnode.Id == *device.ParentNode || originatingIP == downstreamIP {
		changeEng, err := change.GetCMEngine(nil)
		if err != nil {
			resp.WriteLog(http.StatusBadRequest, "Notice", "PATCH /device/{id}/meta::Could not get instance of CME, error: %s", err.Error())
			return
		}
		a.log.Debug("Sending device %s, key %s, value %s to Version Meta", device.Name, metaKey, metaValue)
		err = changeEng.VersionMeta(device.Name, metaKey, metaValue)
		//send it up
		upDevice := device                                                                //make a copy
		add_to_q, err2 := nodeComm.UpdateDeviceMetaUpstream(upDevice, metaKey, metaValue) //if error here put in queue
		if err2 != nil {
			a.log.Warning("PATCH /device/{id}/meta::Was unable to update the device metadata on the upstream node. OUT OF SYNC")
		}
		if add_to_q {
			internode.Qmutex.Lock()
			global.Queue.PushBack(internode.Qelement{
				Fn_name: "UpdateDeviceMetaUpstream",
				Device:  &device,
				Param:   metaKey,
				Param2:  metaValue,
				Cfg:     nil,
			})

			internode.Qmutex.Unlock()
		}

		if err != nil {
			resp.WriteLog(http.StatusBadRequest, "Warning", "PATCH /device/{id}/meta::Was unable to update the device meta data on this node. OUT OF SYNC")
			return
		}
	} else {
		//send it down
		downDevice := device
		success := nodeComm.UpdateDeviceMetaDownstream(downDevice, metaKey, metaValue, downstreamIPInfo) //if error here tell user no can do
		if !success {
			resp.WriteLog(http.StatusBadRequest, "Notice", "PATCH /device/{id}/meta::Could not send the device metadata to the parent node of the device.")
			return
		}
	}
	resp.WriteHeader(http.StatusOK)
}

func (a *APIHandler) deleteDeviceMetadataHierarchy(device database.PbDevice, metaKey string, metaValue string, resp *logging.ResponseLogger, req *http.Request) {
	if device.ParentNode == nil { //can't do anything without a parent node specified
		resp.WriteLog(http.StatusBadRequest, "Debug", "DELETE /device/{id}/meta:: device does not have any parent node. Error in forming the request.")
		return
	}
	rootnode, err := nodeComm.GetRootNode()
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Warning", "DELETE /device/{id}/meta::Could not recover root node from database, cannot proceed")
		return
	}
	originatingIP := strings.Split(req.RemoteAddr, ":")[0] //where did the request come from
	downstreamIP := "Something not matching originatingIP"
	downstreamIPInfo := ""
	if rootnode.Id != *device.ParentNode { //if root node is not the parent of the device, some other node is, recover that nodes IP address
		downstreamIPInfo, err = nodeComm.GetDownstreamNodeIP(*device.ParentNode)
		if err != nil {
			resp.WriteLog(http.StatusBadRequest, "Warning", "DELETE /device/{id}/meta::IP Address of the parent node for the device could not be recovered. Cannot proceed")
			return
		}
		downstreamIP = strings.Split(downstreamIPInfo, ":")[0]
	}

	// now the logic
	if rootnode.Id == *device.ParentNode || originatingIP == downstreamIP {
		changeEng, err := change.GetCMEngine(nil)
		if err != nil {
			resp.WriteLog(http.StatusBadRequest, "Notice", "DELETE /device/{id}/meta::Could not get instance of CME, error: %s", err.Error())
			return
		}
		a.log.Debug("SEnding %s, %s to DeleteMeta", device.Name, metaKey)
		err = changeEng.DeleteMeta(device.Name, metaKey)
		//send it up
		upDevice := device                                                                //make a copy
		add_to_q, err2 := nodeComm.DeleteDeviceMetaUpstream(upDevice, metaKey, metaValue) //if error here put in queue
		if err2 != nil {
			a.log.Warning("DELETE /device/{id}/meta::Was unable to update the device metadata on the upstream node. OUT OF SYNC")
		}
		if add_to_q {
			internode.Qmutex.Lock()
			global.Queue.PushBack(internode.Qelement{
				Fn_name: "DeleteDeviceMetaUpstream",
				Device:  &device,
				Param:   metaKey,
				Param2:  metaValue,
				Cfg:     nil,
			})

			internode.Qmutex.Unlock()
		}

		if err != nil {
			resp.WriteLog(http.StatusBadRequest, "Warning", "DELETE /device/{id}/meta::Was unable to update the device meta data on this node. OUT OF SYNC")
			return
		}
	} else {
		//send it down
		downDevice := device
		success := nodeComm.DeleteDeviceMetaDownstream(downDevice, metaKey, metaValue, downstreamIPInfo) //if error here tell user no can do
		if !success {
			resp.WriteLog(http.StatusBadRequest, "Notice", "DELETE /device/{id}/meta::Could not send the device metadata to the parent node of the device.")
			return
		}
	}
	resp.WriteHeader(http.StatusOK)
}

/************* Device configuration repository, physical device related functions ***************************/
func (a *APIHandler) checkDeviceConfigWOntology(device database.PbDevice, buf *bytes.Buffer, changeEng *change.CMEngine) error {
	jsonstr, _ := trans.GetMarshalledCfg(buf)

	dvrtype, err := changeEng.GetMeta(device.Name, "driver")
	a.log.Debug("Got metatdata: %s<<", dvrtype)
	if err != nil {
		a.log.Debug("Error getting meta")
		return errors.New(fmt.Sprintf("Was not able to get the metadata for the device. Cannot proceed.Error:%s", err))
	}

	ontology_ok, explanation, err := a.validateCfgWOntology(device, dvrtype, jsonstr)
	if err != nil {
		return errors.New(fmt.Sprintf("Was not able to validate the configuration against ontology. Error: %s", err))
	}
	if ontology_ok == false {
		return errors.New(fmt.Sprintf("Could not validate against ontology, explanation = %s", explanation))
	}
	return nil
}

// CfgFinalizedCallback is called when the
func CfgFinalizedCallback(handler interface{}, cmData *change.ChangeData) {
	fmt.Println("*******CfgFinalizedCallback CALLED***************")
	if cmData.ObjectType != change.DEVICE {
		return
	}

	var device database.PbDevice

	apiHandler, ok := handler.(APIHandler)
	if !ok { //is not of apihandler type
		fmt.Println("***ERROR: UNEXPECTED TYPE IN CALLBACK")
		return
	}

	rootnode, err := nodeComm.GetRootNode()
	if err != nil {
		apiHandler.log.Warning("Repository callback function::Could not recover root node from database, cannot proceed")
		return
	}

	upstreamNodeIPInfo, err := nodeComm.GetUpstreamNodeIP()
	if err != nil {
		apiHandler.log.Warning("Repository callback function::Could not recover upstream node information from database, cannot proceed")
		return
	}

	device = database.PbDevice{Name: cmData.Content.Object}
	if err = device.GetByName(apiHandler.db); err != nil {
		apiHandler.log.Debug("Repository callback function, could not recover device specified in the ChangeData")
		return
	}

	if upstreamNodeIPInfo != nil && apiHandler.db.GetPropogationStatus(rootnode.Id) {
		apiHandler.log.Debug("one")
		put_on_q, err := nodeComm.UpdateDeviceConfigUpstream(device, cmData)
		if err != nil { //errors other than communication errors!
			apiHandler.log.Debug("Changed Repository, callback function: Error other than connection issues occurred when trying to send config upstream. Error:%s", err.Error())
			apiHandler.log.Warning("Changed Repository, callback function: Was unable to send the device config to the upstream node. OUT OF SYNC")
		} else {
			if put_on_q { //communication error pushing upstream,
				internode.Qmutex.Lock()
				global.Queue.PushBack(internode.Qelement{
					Fn_name: "UpdateDeviceConfigUpstream",
					Device:  &device,
					Param:   "",
					Param2:  "",
					Cfg:     cmData,
				})
				internode.Qmutex.Unlock()
			}
		}
	}
}
func CfgPackRcvdCallback(handler interface{}, cmData *change.ChangeData) {
	fmt.Println("*******CfgPackRcvdCallback CALLED***************")
	if cmData.ObjectType != change.DEVICE {
		return
	}
	fmt.Printf("Received change data pushed:%v\n", cmData)

	apiHandler, ok := handler.(APIHandler)
	if !ok { //is not of apihandler type
		fmt.Println("***ERROR: UNEXPECTED TYPE IN CALLBACK")
		return
	}
	rootnode, err := nodeComm.GetRootNode()
	if err != nil {
		apiHandler.log.Warning("Repository callback function::Could not recover root node from database, cannot proceed")
		return
	}

	upstreamNodeIPInfo, err := nodeComm.GetUpstreamNodeIP()
	if err != nil {
		apiHandler.log.Warning("Repository callback function::Could not recover upstream node information from database, cannot proceed")
		return
	}

	engine, err := change.GetCMEngine(nil)
	if err != nil {
		fmt.Println("***ERROR:Could not get change management engine in receive-pack callback")
		return
	}
	node := database.PbNode{Name: cmData.SrcNode}
	if err := node.GetByName(apiHandler.db); err != nil {
		fmt.Printf("Could not retrieve Id for node %s from database\n", cmData.SrcNode)
	}
	deviceList, err := apiHandler.db.GetNodeDevices(node.Id)
	if err != nil {
		fmt.Printf("Error trying to retrieve devices from database. Error:%s\n", err)
	}
	for _, dev := range deviceList {
		cfg, err := engine.GetObject(cmData.ObjectType, dev.Name)
		if err != nil {
			fmt.Println("Could not get any config for device " + dev.Name)
			continue
		}
		var buf *bytes.Buffer
		for _, v := range cfg.Content.Files {
			buf = bytes.NewBuffer(v)
			break
		}
		if err := apiHandler.checkDeviceConfigWOntology(dev, buf, engine); err != nil {
			fmt.Printf("Now what? Error:%s\n", err)
		}
	}
	//send the pack up again
	if len(deviceList) > 0 && apiHandler.db.GetPropogationStatus(rootnode.Id) {
		upInfo := internode.UpstreamTransInfo{
			IpInfo:        *upstreamNodeIPInfo,
			TransactionId: "",
		}
		engine.Push(change.DEVICE, upInfo, global.RootNode)
	}
}

//updatePhysicalDeviceWithCfg updates the physical device with all config changes.
//Here device is the stored device on the database, in case the content changes things like password, and we
//need to access the cached stored content before applying the new content
func (a *APIHandler) updatePhysicalDeviceWithCfg(device database.PbDevice, buf *bytes.Buffer) (bool, error) {
	// Op complete, so apply
	a.log.Debug("Configuring device with id %d", device.Id)
	trans.ExecuteConfig(nil, device.Id, buf)
	return true, nil
}

func (a *APIHandler) renameDeviceConfigurationInRepo(deviceName string, newDeviceName string) error {
	engine, err := change.GetCMEngine(nil)
	if err != nil {
		return err
	}
	author := &change.CMAuthor{Name: "jane", Email: "jane@merry.com", When: time.Now()}

	err = engine.RenameObject(change.DEVICE, deviceName, newDeviceName, author)
	return err
}

func (a *APIHandler) deleteDeviceConfigurationFromRepo(deviceName string) error {
	engine, err := change.GetCMEngine(nil)
	if err != nil {
		return err
	}
	author := &change.CMAuthor{Name: "jane", Email: "jane@merry.com", When: time.Now()}

	err = engine.RemoveObject(change.DEVICE, deviceName, author)
	return err
}

func (a *APIHandler) getDeviceConfigurationContentFromRepo(deviceName string) ([]byte, error) {
	eng, err := change.GetCMEngine(nil)
	if err != nil {
		return nil, err
	}

	cdata, err := eng.GetObject(change.DEVICE, deviceName)
	if err != nil {
		return nil, err
	}
	return cdata.Content.Files["configFile"], nil
}

func (a *APIHandler) saveDeviceConfigurationToRepo(cfg *change.ChangeData) error {
	changeEng, err := change.GetCMEngine(nil)
	if err != nil {
		return err
	}

	_, err = changeEng.VersionObject(cfg, cfg.Log.Message)

	if err != nil {
		a.log.Debug("saveDeviceConfigurationToRepo: CersionObject error: " + err.Error())
		return err
	}
	err = changeEng.FinalizeTransaction(cfg)
	if err != nil {
		a.log.Debug("saveDeviceConfigurationToRepo: FinalizeTransaction returned with error: " + err.Error())
	}
	return err
}

func (a *APIHandler) completeConfigurationDataStruct(cfg *change.ChangeData) error {
	if cfg.Author == nil {
		a.log.Debug("saveDeviceConfigurationToRepo: No Author found. Cannot proceed")
		return errors.New("No Author found!")
	}
	user := database.PbUser{Name: cfg.Author.Name, Email: cfg.Author.Email}
	if cfg.Author.Email == "" {
		err := user.GetByName(a.db)
		if err != nil {
			a.log.Debug("saveDeviceConfigurationToRepo: Could not recover user details to store to git repository")
			return err
		}
		cfg.Author.Email = user.Email
	}
	// CHECK cfg.Author.When ZERO Value
	if cfg.Author.When.IsZero() {
		cfg.Author.When = time.Now()
	}
	if cfg.Log.Message == "" {
		a.log.Debug("saveDeviceConfigurationToRepo: Cannot commit with empty message")
		return errors.New("Commit message cannot be empty")
	}
	return nil
}

/************* Utility Common functions ***************************/
func (a *APIHandler) parseIdFromRoute(paramId string) (int64, error) {
	deviceId, err := strconv.ParseInt(paramId, 10, 64) // string to int64
	if err != nil {
		a.log.Debug("Strconv error: %s", err.Error())
	}
	return deviceId, err
}

/************* End Utility Common functions ***************************/
/**
* Validate a device configuration against the ontology server
 */
func (a *APIHandler) validateCfgWOntology(device database.PbDevice, devtype string, cfg []byte) (bool, string, error) {
	var buffer bytes.Buffer

	//Logging configuration to see what's going on.
	a.log.Debug("====Translated configuration====")
	a.log.Debug(string(cfg[:]))

	//If the configuration is being all sent at once, we need to send it in a different format than if it's
	//being sent in pieces.  Sending it all at once will clear it out, sending a single piece will just update things

	buffer.WriteString("{\"ontology\":\"config\",")
	buffer.WriteString("\"ontologizer\":\"" + string(devtype) + "\",")
	buffer.WriteString("\"individual\":\"" + string(device.Name) + "\",")
	buffer.WriteString("\"properties\":")

	//In case we get an empty value for properties, just write an empty array
	if len(string(cfg)) > 0 {
		buffer.WriteString(string(cfg))
	} else {
		buffer.WriteString("[]")
	}
	buffer.WriteString("}")

	a.log.Debug("=== Validating device configuration against ontology server ===")
	a.log.Debug(buffer.String())

	status, explanation, err := ontology.ValidateAgainstOntology(buffer)
	if err != nil {
		a.log.Debug("Error validating : " + err.Error())
	}
	return status, explanation, err
}
