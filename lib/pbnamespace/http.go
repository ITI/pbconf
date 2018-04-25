package namespace

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
	"net/http"
	"regexp"
	"strings"

	mux "github.com/gorilla/mux"
	database "github.com/iti/pbconf/lib/pbdatabase"
	global "github.com/iti/pbconf/lib/pbglobal"
	logging "github.com/iti/pbconf/lib/pblogger"
)

var apiNamespaceRoutes map[string]struct{}

type APIHandler struct {
	log                logging.Logger
	db                 database.AppDatabase
	Version            int
	namespaceRouteList []string
}

func NewAPIHandler(loglevel string, d database.AppDatabase) *APIHandler {
	l, _ := logging.GetLogger("Namespace API")
	logging.SetLevel(loglevel, "Namespace API")
	return &APIHandler{log: l, db: d, Version: 1}
}

func (a *APIHandler) AddAPIEndpoints(router *mux.Router) {
	a.log.Debug("Registering namespace endpoints")

	for _, v := range global.ApiUrlVersioning {
		router.HandleFunc(v+"/namespace", a.handleBaseRoute).Methods("GET")

		s := router.PathPrefix(v + "/namespace").Subrouter()
		s.HandleFunc("/", a.handleBaseRoute).Methods("GET")
	}
	a.walkallroutes(router)
}

func (a *APIHandler) GetInfo() (string, int) {
	return "namespace", a.Version
}

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

	jsonStr, err := json.Marshal(a.namespaceRouteList)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Notice", "GET /namespace::Could not marshal the list of routes Error: %s", err.Error())
		return
	}
	if _, err = resp.Write(jsonStr); err != nil {
		resp.WriteLog(http.StatusBadRequest, "Notice", "GET /namespace::Writing response body Error: %s", err.Error())
		return
	}

}

func filterRouteToNamespace(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
	rPath, err := route.GetPathTemplate()
	if err != nil {
		return err
	}
	matched, err := regexp.MatchString("version:", rPath)
	if matched {
		return nil
	}
	trimmedPath := strings.TrimPrefix(rPath, "/")
	namespacePath := strings.Split(trimmedPath, "/")
	if len(namespacePath) < 2 {
		return nil
	}
	_, ok := apiNamespaceRoutes[namespacePath[0]]
	if !ok {
		apiNamespaceRoutes[namespacePath[0]] = struct{}{}
	}
	return nil
}
func (a *APIHandler) walkallroutes(router *mux.Router) {
	apiNamespaceRoutes = make(map[string]struct{})
	router.Walk(filterRouteToNamespace)
	//convert map to list for easier access to list of namespaces
	a.namespaceRouteList = make([]string, len(apiNamespaceRoutes))
	i := 0
	for k := range apiNamespaceRoutes {
		a.namespaceRouteList[i] = k
		i++
	}

}
