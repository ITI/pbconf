package node

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
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	database "github.com/iti/pbconf/lib/pbdatabase"
	global "github.com/iti/pbconf/lib/pbglobal"
	logging "github.com/iti/pbconf/lib/pblogger"
	validator "gopkg.in/validator.v2"

	"encoding/json"
)

type APIHandler struct {
	log     logging.Logger
	db      database.AppDatabase
	Version int
}

func NewAPIHandler(loglevel string, d database.AppDatabase) *APIHandler {
	l, _ := logging.GetLogger("Node API")
	logging.SetLevel(loglevel, "Node API")
	return &APIHandler{log: l, db: d, Version: 1}
}

func (a *APIHandler) AddAPIEndpoints(router *mux.Router) {
	a.log.Info("Registering node endpoints")

	for _, v := range global.ApiUrlVersioning {
		router.HandleFunc(v+"/node", a.handleBaseRoute).Methods("HEAD", "GET", "POST", "PATCH")

		s := router.PathPrefix(v + "/node").Subrouter()
		s.HandleFunc("/", a.handleBaseRoute).Methods("HEAD", "GET", "POST", "PATCH")
		s.HandleFunc("/{nodeid}", a.handleWIdRoute).Methods("GET", "PATCH", "DELETE")
		s.HandleFunc("/{nodeid}/{cfgkey}", a.handleConfigItem).Methods("GET", "DELETE")
	}
}

func (a *APIHandler) GetInfo() (string, int) {
	return "node", a.Version
}

// handleBaseRoute handles the GET, PUT, POST routes to "/node" or subroute "/"
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
	timetag, err := a.db.GetNodesLastModified()
	if err != nil {
		a.log.Warning("HEAD /node::Could not recover last modified tag for Nodes")
	} else {
		writer.Header().Set("Last-Modified", timetag)
	}
	writer.WriteHeader(http.StatusOK)
}

// getBaseHandler handles the Get route for "/". It returns a list of nodes
func (a *APIHandler) getBaseHandler(resp *logging.ResponseLogger, req *http.Request) {
	nodeList, err := a.db.GetNodes()
	if err != nil {
		resp.WriteLog(http.StatusNotFound, "Error", "GET /node::Could not get the nodes downstream of this node from the database.")
		return
	}
	jsonStr, err := json.Marshal(nodeList)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Notice", "GET /node::Could not marshal the nodes Error: %s", err.Error())
		return
	}
	if _, err = resp.Write(jsonStr); err != nil {
		resp.WriteLog(http.StatusBadRequest, "Notice", "GET /node::Writing response body Error: %s", err.Error())
		return
	}
}

//postBaseHandler handles the POST route for "/". It reads the request body and creates a new Node.
func (a *APIHandler) postBaseHandler(resp *logging.ResponseLogger, req *http.Request) {
	var newNode database.PbNode
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&newNode)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Info", "POST /node::Decoder error: %s", err.Error())
		return
	}

	if err = validator.Validate(newNode); err != nil {
		resp.WriteLog(http.StatusBadRequest, "Debug", "POST /node::validation error: %s", err.Error())
		return
	}
	if newNode.Id > 0 {
		resp.WriteLog(http.StatusNotImplemented, "Debug", "POST /node:: Node id specified.")
		return
	}

	err = newNode.Create(a.db)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Notice", "POST /node::Error creating node")
		return
	}
	resp.Header().Set("Location", req.URL.String()+"/"+strconv.FormatInt(newNode.Id, 10))
	resp.WriteHeader(http.StatusCreated)
}

// patchBaseHandler handles the PATCH route for "/". It reads the request body and
// updates the node. If the config item does not exist, creates it. Otherwise updates the node name or config item
func (a *APIHandler) patchBaseHandler(resp *logging.ResponseLogger, req *http.Request) {
	var node database.PbNode
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&node)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Info", "PATCH /node::Decoder error: %s", err.Error())
		return
	}
	a.patchNode(node, resp, req)
}

// handleWIdRoute handles the GET, PUT, PATCH, DELETE for the "/{nodeid}" route
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

// getWIdHandler handles the GET route for "/{nodeid}" route. It returns all entity information.
func (a *APIHandler) getWIdHandler(resp *logging.ResponseLogger, req *http.Request) {
	params := mux.Vars(req)
	nodeId, err := a.parseIdFromRoute(params["nodeid"]) // string to int64
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Info", "GET /node/{id}::Could not recover node id from route.")
		return
	}
	node := database.PbNode{Id: nodeId}
	err = node.Get(a.db)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Info", err.Error())
		return
	}
	jsonStr, err := json.Marshal(node)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Notice", "GET /node/{id}::Could not marshal the requested node, error: %s", err.Error())
		return
	}
	if _, err = resp.Write(jsonStr); err != nil {
		resp.WriteLog(http.StatusBadRequest, "Notice", "GET /node/{id}::Writing response body Error: %s", err.Error())
	}
}

// patchWIdHandler handles the PATCH request on "/{nodeid}" route. If node does not exist, returns error. Allows partial update.
func (a *APIHandler) patchWIdHandler(resp *logging.ResponseLogger, req *http.Request) {
	params := mux.Vars(req)
	nodeId, err := a.parseIdFromRoute(params["nodeid"]) // string to int64
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Info", "PATCH /node/{id}::Could not recover node id from route.")
		return
	}
	var node database.PbNode
	decoder := json.NewDecoder(req.Body)
	err = decoder.Decode(&node)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Info", "PATCH /node/{id}::Decoder error: %s", err.Error())
		return
	}
	node.Id = nodeId //route param takes precedence
	a.patchNode(node, resp, req)
}

// deleteWIdHandler handles the DELETE request for the "/{nodeid}" route. Deletes the entity at nodeid.
func (a *APIHandler) deleteWIdHandler(resp *logging.ResponseLogger, req *http.Request) {
	params := mux.Vars(req)
	nodeId, err := a.parseIdFromRoute(params["nodeid"]) // string to int64
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Info", "DELETE /node/{id}::Could not recover node id from route.")
		return
	}

	node := database.PbNode{Id: nodeId}
	err = node.Delete(a.db)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Info", "DELETE /node/{id}::Could not recover the node from the database")
		return
	}
	resp.WriteHeader(http.StatusOK)
}

// handleConfigItem handles the GET, DELETE requests for the "/{nodeid}/{cfgkey}" route
func (a *APIHandler) handleConfigItem(writer http.ResponseWriter, req *http.Request) {
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

// getConfigElement handles the GET request for the "/{nodeid}/{cfgkey}" route. It returns the json representation of PbNodeConfigItem
func (a *APIHandler) getConfigItemHandler(resp *logging.ResponseLogger, req *http.Request) {
	params := mux.Vars(req)
	configKey := params["cfgkey"]
	nodeId, err := a.parseIdFromRoute(params["nodeid"]) // string to int64
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Info", "GET /node/{id}/{cfgkey}::Could not recover node id from route.")
		return
	}
	dbConfigEl := database.PbNodeConfigItem{
		NodeId:     nodeId,
		ConfigItem: database.ConfigItem{Key: configKey},
	}
	exists, err := dbConfigEl.Exists(a.db)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Info", "GET /node/{id}/{cfgkey}::Error checking existence of node cfgitem")
		return
	}
	if !exists {
		resp.WriteLog(http.StatusNotFound, "Debug", "GET /node/{id}/{cfgkey}::node cfgitem does not seem to exist in the database!")
		return
	}
	err = dbConfigEl.Get(a.db)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Notice", "GET /node/{id}/{cfgkey}::Was not able to get the config item from the database")
		return
	}
	jsonStr, err := json.Marshal(dbConfigEl)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Notice", "GET /node/{id}/{cfgkey}::Could not marshal the requested config item of the node, Error: %s", err.Error())
		return
	}
	if _, err = resp.Write(jsonStr); err != nil {
		resp.WriteLog(http.StatusBadRequest, "Notice", "GET /node/{id}/{cfgkey}::Writing response body Error: %s", err.Error())
	}
}

// deleteConfigItem handles the DELETE request for the "/{nodeid}/{cfgkey}" route.
// It deletes the config item corresponding the configkey from the node.
func (a *APIHandler) deleteConfigItemHandler(resp *logging.ResponseLogger, req *http.Request) {
	params := mux.Vars(req)
	configKey := params["cfgkey"]
	nodeId, err := a.parseIdFromRoute(params["nodeid"]) // string to int64
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Info", "DELETE /node/{id}::Could not recover node id from route.")
		return
	}
	dbConfigEl := database.PbNodeConfigItem{
		NodeId:     nodeId,
		ConfigItem: database.ConfigItem{Key: configKey},
	}

	err = dbConfigEl.Delete(a.db)
	if err != nil {
		resp.WriteLog(http.StatusNotFound, "Info", "DELETE /device/{id}/{cfgkey}::Was unable to delete the config item in the database on this node.")
		return
	}
	resp.WriteHeader(http.StatusOK)
}

/************************ Non route helper functions **************************/
func (a *APIHandler) patchNode(node database.PbNode, resp *logging.ResponseLogger, req *http.Request) {
	err := node.Update(a.db)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Debug", "PATCH /node:: Update error::%s", err)
		return
	}
	resp.WriteHeader(http.StatusOK)
}

/************* Utility Common functions ***************************/
func (a *APIHandler) parseIdFromRoute(paramId string) (int64, error) {
	nodeId, err := strconv.ParseInt(paramId, 10, 64) // string to int64
	if err != nil {
		a.log.Debug("Strconv error: %s", err.Error())
	}
	return nodeId, err
}
