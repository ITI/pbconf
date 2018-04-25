package reports

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
	"errors"
	"net/http"
	"strings"
	"time"

	mux "github.com/gorilla/mux"
	change "github.com/iti/pbconf/lib/pbchange"
	database "github.com/iti/pbconf/lib/pbdatabase"
	global "github.com/iti/pbconf/lib/pbglobal"
	logging "github.com/iti/pbconf/lib/pblogger"
	validator "gopkg.in/validator.v2"
)

type APIHandler struct {
	log     logging.Logger
	db      database.AppDatabase
	Version int
}

func NewAPIHandler(loglevel string, d database.AppDatabase) *APIHandler {
	l, _ := logging.GetLogger("Report API")
	logging.SetLevel(loglevel, "Report API")
	return &APIHandler{log: l, db: d, Version: 1}
}

func (a *APIHandler) AddAPIEndpoints(router *mux.Router) {
	a.log.Info("Registering reports endpoints")

	for _, v := range global.ApiUrlVersioning {
		router.HandleFunc(v+"/reports", a.handleBaseRoute).Methods("GET", "POST", "HEAD")

		s := router.PathPrefix(v + "/reports").Subrouter()
		s.HandleFunc("/", a.handleBaseRoute).Methods("GET", "POST", "HEAD")
		s.HandleFunc("/{reportid}", a.handleWIdRoute).Methods("GET", "PUT", "DELETE")
		s.HandleFunc("/report/{reportid}", a.handleWIdReportContent).Methods("GET")
	}
	a.startPeriodicReportTimers()
}

func (a *APIHandler) GetInfo() (string, int) {
	return "reports", a.Version
}

func (a *APIHandler) startPeriodicReportTimers() {
	engine, err := change.GetCMEngine(nil)
	if err != nil {
		a.log.Debug("startPeriodicReportTimers::Could not get an instance of Change management Engine.")
	}
	queryList, err := engine.ListObjects(change.REPORT)
	if err != nil {
		a.log.Error("startPeriodicReportTimers::Could not get list of queries.Error:%s", err.Error())
		return
	}
	for _, queryName := range queryList {
		chData, err := engine.GetObject(change.REPORT, queryName)
		if err != nil {
			a.log.Debug("startPeriodicReportTimers::CME Getting object Error: %s", err.Error())
			return
		}
		var savedQuery PbReportQuery
		decoder := json.NewDecoder(bytes.NewBuffer(chData.Content.Files["queryFile"]))
		if err = decoder.Decode(&savedQuery); err != nil {
			a.log.Debug("startPeriodicReportTimers::Report Decoder Error:", err.Error())
			return
		}
		isQueryPeriodic := savedQuery.Period != "-1"
		if isQueryPeriodic {
			duration, err := time.ParseDuration(savedQuery.Period)
			if err != nil {
				a.log.Debug("startPeriodicReportTimers:: Could not recover duration Error:%s", err.Error())
				return
			}
			a.periodicFunc(duration, savedQuery)
		}
	}
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
	switch req.Method {
	case "HEAD":
		a.headBaseHandler(resp, req)
	case "GET":
		a.getBaseHandler(resp, req)
	case "POST":
		a.postBaseHandler(resp, req)
	}
}

func (a *APIHandler) headBaseHandler(resp *logging.ResponseLogger, req *http.Request) {
	changeEng, err := change.GetCMEngine(nil)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Notice", "Could not get handle to change management engine. Error:%s", err.Error())
	}
	commitId := changeEng.GetLatestCommitID(change.REPORT)
	resp.Header().Set("X-Pbconf-Report-LastCommitId", commitId)
	resp.WriteHeader(http.StatusOK)
}

func (a *APIHandler) getBaseHandler(resp *logging.ResponseLogger, req *http.Request) {
	engine, err := change.GetCMEngine(nil)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Debug", "GET /reports::Could not get an instance of Change management Engine.")
	}

	queryList, err := engine.ListObjects(change.REPORT)
	if err != nil {
		resp.WriteLog(http.StatusNotFound, "Error", "GET /reports::Could not get list of queries.")
		return
	}
	jsonStr, err := json.Marshal(queryList)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Notice", "GET /reports::Could not marshal the report queries Error: %s", err.Error())
		return
	}
	if _, err = resp.Write(jsonStr); err != nil {
		resp.WriteLog(http.StatusBadRequest, "Notice", "GET /reports::Writing response body Error: %s", err.Error())
		return
	}
}

// postBaseHandler is the api handler to create a new query in the cme
func (a *APIHandler) postBaseHandler(resp *logging.ResponseLogger, req *http.Request) {
	var query PbReportQueryHttp
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&query)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Notice", "POST /reports::Decoder error: %s", err.Error())
		return
	}
	//validate the query
	if err = validator.Validate(query); err != nil {
		resp.WriteLog(http.StatusBadRequest, "Notice", "POST /reports::validation error: %s", err.Error())
		return
	}
	//check if new query can be run successfully
	if _, err = ParseQuery(strings.NewReader(query.Query)); err != nil {
		resp.WriteLog(http.StatusBadRequest, "Debug", "POST /reports:: Could not parse the query successfully, Error:%s", err.Error())
		return
	}

	engine, err := change.GetCMEngine(nil)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Debug", "POST /reports:: Could not get an instance of Change management Engine.")
	}

	queryList, err := engine.ListObjects(change.REPORT)
	if err != nil {
		_, ok := err.(change.CMNoRepoError)
		if !ok {
			resp.WriteLog(http.StatusNotFound, "Error", "POST /reports::Could not get list of queries.")
			return
		}
	}
	for _, q := range queryList {
		if q == query.Name {
			resp.WriteLog(http.StatusBadRequest, "Debug", "POST /reports:: %s already exists. Use PUT request to update", q)
		}
	}
	content := change.NewCMContent(query.Name)
	jsonStr, err := json.Marshal(query.PbReportQuery)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Notice", "POST /reports::Could not marshal the PbReportQuery, Error: %s", err.Error())
		return
	}
	content.Files["queryFile"] = []byte(jsonStr)
	chData := &change.ChangeData{ObjectType: change.REPORT, Content: content, Author: &change.CMAuthor{Name: query.Author}}
	if err = a.completeChangeDataStruct(chData); err != nil {
		resp.WriteLog(http.StatusBadRequest, "Debug", "POST /reports:: Error composing the change.ChangeData structure, Error:%s", err.Error())
		return
	}
	_, err = engine.VersionObject(chData, query.CommitMessage)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Debug", "POST /reports:: Error while trying to save the change.ChangeData structure, Error:%s", err.Error())
		return
	}
	resp.WriteHeader(http.StatusCreated)
}

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
		a.getWIdRouteHandler(resp, req)
	case "PUT":
		a.putWIdRouteHandler(resp, req)
	case "DELETE":
		a.deleteWIdRouteHandler(resp, req)
	}
}

func (a *APIHandler) getWIdRouteHandler(resp *logging.ResponseLogger, req *http.Request) {
	params := mux.Vars(req)
	queryName := params["reportid"]

	engine, err := change.GetCMEngine(nil)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Debug", "Could not get an instance of Change management Engine.")
	}
	chData, err := engine.GetObject(change.REPORT, queryName)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Debug", "GET /reports/{reportid}::CME Getting object Error: %s", err.Error())
		return
	}
	jsonStr := chData.Content.Files["queryFile"]
	if _, err = resp.Write(jsonStr); err != nil {
		resp.WriteLog(http.StatusBadRequest, "Notice", "GET /reports/{reportid}::Writing response body Error: %s", err.Error())
		return
	}
}

func (a *APIHandler) deleteWIdRouteHandler(resp *logging.ResponseLogger, req *http.Request) {
	params := mux.Vars(req)
	queryName := params["reportid"]

	engine, err := change.GetCMEngine(nil)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Debug", "Could not get an instance of Change management Engine.")
	}
	err = engine.RemoveObject(change.REPORT, queryName, &change.CMAuthor{Name: "reportEngine"})
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Debug", "DELETE /reports/{reportid}:: Error while calling RemoveObject, Error:%s", err.Error())
		return
	}
	return
}

func (a *APIHandler) putWIdRouteHandler(resp *logging.ResponseLogger, req *http.Request) {
	params := mux.Vars(req)
	queryName := params["reportid"]

	var query PbReportQueryHttp
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&query)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Notice", "PUT /reports/{reportid}::Decoder error: %s", err.Error())
		return
	}
	query.Name = queryName // Name in the path takes precedence
	if err = validator.Validate(query); err != nil {
		resp.WriteLog(http.StatusBadRequest, "Notice", "PUT /reports/{reportid}::validation error: %s", err.Error())
		return
	}
	//check if new query can be run successfully
	_, err = ParseQuery(strings.NewReader(query.Query))
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Debug", "PUT /reports/{reportid}:: Could not run the query successfully, Error:%s", err.Error())
		return
	}
	//now save and run the query or fire the periodic function
	engine, err := change.GetCMEngine(nil)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Debug", "PUT /reports/{reportid}::Could not get an instance of Change management Engine.")
		return
	}
	//get query from cme if it exists
	savedReportObj, _ := engine.GetObject(change.REPORT, queryName) //cache old query if it existed before saving the new one!
	//save the query retrieved from message payload to the cme
	content := change.NewCMContent(query.Name)
	jsonStr, err := json.Marshal(query.PbReportQuery)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Notice", "PUT /reports/{reportid}::Could not marshal the PbReportQuery, Error: %s", err.Error())
		return
	}
	content.Files["queryFile"] = []byte(jsonStr)
	chData := &change.ChangeData{ObjectType: change.REPORT, Content: content, Author: &change.CMAuthor{Name: query.Author}}
	if err = a.completeChangeDataStruct(chData); err != nil {
		resp.WriteLog(http.StatusBadRequest, "Debug", "PUT /reports/{reportid}:: Error composing the change.ChangeData structure, Error:%s", err.Error())
		return
	}
	_, err = engine.VersionObject(chData, query.CommitMessage)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Debug", "PUT /reports/{reportid}:: Error while trying to save the change.ChangeData structure, Error:%s", err.Error())
		return
	}
	//code @ when to run the report query
	isQueryPeriodic := query.Period != "-1"
	queryTimeStamp, _ := time.Parse(time.RFC1123, query.TimeStamp)

	wasQueryPeriodic := false
	var savedTimeStamp time.Time
	//retrieve things from the saved report if it existed
	if savedReportObj != nil && len(savedReportObj.Content.Files) > 0 {
		queryStr := savedReportObj.Content.Files["queryFile"]
		var savedReport PbReportQuery
		decoder = json.NewDecoder(bytes.NewBuffer(queryStr))
		if err = decoder.Decode(&savedReport); err != nil {
			resp.WriteLog(http.StatusBadRequest, "Notice", "PUT /reports/{reportid}::Decoder error: %s", err.Error())
			return
		}
		savedTimeStamp, _ = time.Parse(time.RFC1123, savedReport.TimeStamp)
		wasQueryPeriodic = savedReport.Period != "-1"
	}
	if !isQueryPeriodic && (savedTimeStamp.IsZero() || (savedTimeStamp != queryTimeStamp)) {
		if err = a.RunQuery(query.PbReportQuery); err != nil {
			resp.WriteLog(http.StatusBadRequest, "Debug", "PUT /reports/{reportid}:: Error while trying to run the query, Error:%s", err.Error())
			return
		}

	}
	//start the periodic function just once. After that it will check the query on timer elapse.
	if isQueryPeriodic && !wasQueryPeriodic {
		duration, err := time.ParseDuration(query.Period)
		if err != nil {
			resp.WriteLog(http.StatusBadRequest, "Debug", "PUT /reports/{reportid}:: Could not recover duration Error:%s", err.Error())
			return
		}
		a.periodicFunc(duration, query.PbReportQuery)
	}
}

// periodicFunction calls the function f after duration time. The function f runs the query and then
// retrieves the query object from repository to make sure that it is still periodic. If it is, it spins up another
// periodic function call.
func (a *APIHandler) periodicFunc(duration time.Duration, query PbReportQuery) {
	f := func() {
		engine, err := change.GetCMEngine(nil)
		if err != nil {
			a.log.Debug("Timed report could not be generated, GetCMEngine Error:", err.Error())
			return
		}
		chData, err := engine.GetObject(change.REPORT, query.Name)
		if err != nil {
			a.log.Debug("Timed report could not be generated, GetObject Error:", err.Error())
			return
		}
		if len(chData.Content.Files["queryFile"]) == 0 { //No content, object was deleted.
			return
		}
		//run the query after  verifying it still exists in the cme
		if err = a.RunQuery(query); err != nil {
			a.log.Debug("Timed report: Could not run query, parser returned error:%s", err.Error())
			return
		}

		var savedQuery PbReportQuery
		decoder := json.NewDecoder(bytes.NewBuffer(chData.Content.Files["queryFile"]))
		if err = decoder.Decode(&savedQuery); err != nil {
			a.log.Debug("Timed report could not be generated, Decoder Error:", err.Error())
			return
		}
		isQueryPeriodic := savedQuery.Period != "-1"
		if !isQueryPeriodic {
			return
		}
		savedDuration, err := time.ParseDuration(savedQuery.Period)
		if err != nil {
			a.log.Debug("Timed report could not be generated, ParseDuration Error:", err.Error())
			return
		}
		a.periodicFunc(savedDuration, savedQuery)
	}
	time.AfterFunc(duration, f)
}

func (a *APIHandler) RunQuery(query PbReportQuery) error {
	engine, err := change.GetCMEngine(nil)
	if err != nil {
		return err
	}
	//run the query
	answer, err := RunReport(a.db, a.log, strings.NewReader(query.Query))
	if err != nil {
		return err
	}
	ranQueryAt := time.Now().Format(time.RFC1123)
	a.log.Debug("RAN THE QUERY %s at time %v", query.Name, ranQueryAt)
	//now save the results of running the query
	reportObj, _ := engine.GetObject(change.REPORT, query.Name)
	reportObj.Content.Files["reportFile"] = []byte(answer)
	_, err = engine.VersionObject(reportObj, "Updated at timestamp "+ranQueryAt)
	if err != nil {
		return err
	}
	return nil
}

func (a *APIHandler) handleWIdReportContent(writer http.ResponseWriter, req *http.Request) {
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
	params := mux.Vars(req)
	queryName := params["reportid"]

	engine, err := change.GetCMEngine(nil)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Debug", "GET /reports/report/{reportid}::Could not get an instance of Change management Engine.")
	}
	chData, err := engine.GetObject(change.REPORT, queryName)
	if chData == nil || chData.Content == nil {
		writer.WriteHeader(http.StatusNotFound)
		return
	}
	jsonStr := chData.Content.Files["reportFile"]
	if _, err = resp.Write(jsonStr); err != nil {
		resp.WriteLog(http.StatusBadRequest, "Notice", "GET /reports/report/{reportid}::Writing response body Error: %s", err.Error())
		return
	}
}

/********************************* utility functions *****************************************/
func (a *APIHandler) completeChangeDataStruct(chData *change.ChangeData) error {
	if chData.Author == nil || chData.Author.Name == "" {
		a.log.Debug("completeChangeDataStruct: No Author found. Cannot proceed")
		return errors.New("No Author found!")
	}
	user := database.PbUser{Name: chData.Author.Name, Email: chData.Author.Email}
	if chData.Author.Email == "" {
		err := user.GetByName(a.db)
		if err != nil {
			a.log.Debug("completeChangeDataStruct: Could not recover user details to store to git repository")
			return err
		}
		chData.Author.Email = user.Email
	}

	if chData.Author.When.IsZero() {
		chData.Author.When = time.Now()
	}
	return nil
}
