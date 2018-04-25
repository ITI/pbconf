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
	"compress/gzip"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	mux "github.com/gorilla/mux"
	change "github.com/iti/pbconf/lib/pbchange"
	database "github.com/iti/pbconf/lib/pbdatabase"
	global "github.com/iti/pbconf/lib/pbglobal"
	internode "github.com/iti/pbconf/lib/pbinternode"
	logging "github.com/iti/pbconf/lib/pblogger"
)

type PolicyHandler struct {
	log     logging.Logger
	db      database.AppDatabase
	Version int
}

var nodeComm *internode.InterNode

func NewAPIHandler(loglevel string, d database.AppDatabase) *PolicyHandler {
	l, _ := logging.GetLogger("Policy API")
	logging.SetLevel(loglevel, "Policy API")
	nodeComm = internode.NewInterNodeCommunicator(d)
	return &PolicyHandler{log: l, db: d, Version: 1}
}

func (a *PolicyHandler) AddAPIEndpoints(router *mux.Router) {
	a.log.Info("Registering policy endpoints")
	// Change management hook
	engine, err := change.GetCMEngine(nil)
	if err != nil {
		a.log.Debug("Could not get an instance of Change management Engine.")
	}

	for _, v := range global.ApiUrlVersioning {
		router.HandleFunc(v+"/policy", a.handleBaseRoute).Methods("GET", "POST", "HEAD")

		s := router.PathPrefix(v + "/policy").Subrouter()
		s.HandleFunc("/", a.handleBaseRoute).Methods("GET", "POST", "HEAD")
		s.HandleFunc("/all", a.handleGetAll).Methods("GET")
		s.HandleFunc("/default/{nodename}", a.handleDefaultRoute).Methods("HEAD")
		s.HandleFunc("/{policyname}", a.handleWIdRoute).Methods("GET", "PUT", "DELETE")
		s.HandleFunc("/{devid}/validate", a.handleWIdValidateRoute)
		s.HandleFunc("/{devid}/validate/{valid}", a.handleWIdValidateStatusRoute)

		cme := s.PathPrefix("/cme").Subrouter()
		cme.HandleFunc(`/{rest:[a-zA-Z0-9=\-\/]+}`, engine.RequestHandler())
	}
}

func (a *PolicyHandler) GetInfo() (string, int) {
	return "policy", a.Version
}

func (a *PolicyHandler) handleBaseRoute(writer http.ResponseWriter, req *http.Request) {
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
		a.getBaseHandler(resp, req)
	case "HEAD":
		a.headBaseHandler(resp, req)
	case "POST":
		a.postBaseHandler(resp, req)
	}
}

func (a *PolicyHandler) getBaseHandler(resp *logging.ResponseLogger, req *http.Request) {
	changeEng, err := change.GetCMEngine(nil)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Notice", "Could not get handle to change management engine. Error:%s", err.Error())
	}

	polList, err := changeEng.ListObjects(change.POLICY)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Notice", err.Error())
		return
	}
	jsonStr, err := json.Marshal(polList)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Notice", "GET /policy::Could not marshal the requested policy list, Error: %s", err.Error())
		return
	}
	if _, err = resp.Write(jsonStr); err != nil {
		resp.WriteLog(http.StatusBadRequest, "Notice", "GET /policy::Writing response body Error: %s", err.Error())
	}
}

func (a *PolicyHandler) headBaseHandler(resp *logging.ResponseLogger, req *http.Request) {
	changeEng, err := change.GetCMEngine(nil)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Notice", "Could not get handle to change management engine. Error:%s", err.Error())
	}
	commitId := changeEng.GetLatestCommitID(change.POLICY)
	resp.Header().Set("X-Pbconf-Policy-LastCommitId", commitId)
	resp.WriteHeader(http.StatusOK)
}

func (a *PolicyHandler) postBaseHandler(resp *logging.ResponseLogger, req *http.Request) {
	var policy change.ChangeData
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&policy)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Info", "POST /policy::Decoder error: %s", err.Error())
		return
	}
	a.putPolicyHierarchy(&policy, resp, req)
}

func (a *PolicyHandler) handleGetAll(writer http.ResponseWriter, req *http.Request) {
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
	changeEng, err := change.GetCMEngine(nil)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Notice", "Could not get handle to change management engine. Error:%s", err.Error())
		return
	}

	polList, err := changeEng.ListObjects(change.POLICY)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Notice", "CME ListObjects Error:%s", err.Error())
		return
	}

	policies := make([]change.ChangeData, 0)
	for _, policyName := range polList {
		pol, err := changeEng.GetObject(change.POLICY, policyName)
		if err != nil {
		}
		policies = append(policies, *pol)
	}
	jsonStr, err := json.Marshal(policies)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Notice", "GET /policy/all::Could not marshal the requested policy list, Error: %s", err.Error())
		return
	}
	gzip_writer := gzip.NewWriter(writer)
	defer gzip_writer.Close()
	writer.Header().Set("Content-Encoding", "gzip")
	if _, err = gzip_writer.Write(jsonStr); err != nil {
		resp.WriteLog(http.StatusBadRequest, "Notice", "GET /policy::Writing response body Error: %s", err.Error())
	}
}

func (a *PolicyHandler) handleDefaultRoute(writer http.ResponseWriter, req *http.Request) {
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
		a.headDefaultHandler(resp, req)
	}
}

// headDefaultHandler receives the heartbeat from node specified in the route. If this node does
// not exist in the database, it is added, and its IP address is extracted from the http request.
// At this point, the time at which we received this HEAD request is logged as a config item for the node.
func (a *PolicyHandler) headDefaultHandler(resp *logging.ResponseLogger, req *http.Request) {
	params := mux.Vars(req)
	nodeName := params["nodename"]
	childnode := database.PbNode{Name: nodeName}
	exists, err := childnode.ExistsByName(a.db)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Info", "HEAD /policy/default/{nodename}::Error determining if node exists in the database.")
		return
	}
	if !exists { //if node does not exist, add the node and its IP address
		if err = childnode.Create(a.db); err != nil {
			resp.WriteLog(http.StatusBadRequest, "Info", "HEAD /policy/default/{nodename}::Error adding new node to the database via heartbeat.")
			return
		}
		ip := strings.Split(req.RemoteAddr, ":")[0]
		downstreamIPItem := database.PbNodeConfigItem{NodeId: childnode.Id,
			ConfigItem: database.ConfigItem{Key: "IP Address", Value: ip}}
		if err = downstreamIPItem.Create(a.db); err != nil {
			resp.WriteLog(http.StatusBadRequest, "Info", "HEAD /policy/default/{nodename}::Could not add IP address for newly reported node %s", nodeName)
			return
		}
	}
	err = childnode.GetByName(a.db)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Info", "HEAD /policy/default/{nodename}::Could not retrieve the child node by name from the database.")
		return
	}
	chCheckinItem := database.PbNodeConfigItem{NodeId: childnode.Id,
		ConfigItem: database.ConfigItem{Key: "Checkin Timestamp"},
	}

	chCheckinItem.ConfigItem.Value = time.Now().Format(time.RFC1123)
	err = chCheckinItem.CreateOrUpdate(a.db)
	if err != nil {
		a.log.Info("HEAD /policy/default/{nodename}:: Could not create or update the checkin timestamp for node %s", nodeName)
	}

	a.log.Info("**Received heartbeat from childnode %s at time %s", childnode.Name, chCheckinItem.Value)
	changeEng, err := change.GetCMEngine(nil)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Notice", "Could not get handle to change management engine. Error:%s", err.Error())
	}
	commitId := changeEng.GetLatestCommitID(change.POLICY)
	resp.Header().Set("X-Pbconf-Policy-LastCommitId", commitId)
	resp.WriteHeader(http.StatusOK)
}

func (a *PolicyHandler) handleWIdRoute(writer http.ResponseWriter, req *http.Request) {
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
	case "PUT":
		a.putWIdHandler(resp, req)
	case "DELETE":
		a.deleteWIdHandler(resp, req)
	}
}

func (a *PolicyHandler) getWIdHandler(resp *logging.ResponseLogger, req *http.Request) {
	params := mux.Vars(req)
	policyName := params["policyname"]

	policy, err := a.getPolicyContentFromRepo(policyName)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Info", "Error retrieving policy from change management error: %s", err.Error())
		return
	}
	jsonStr, err := json.Marshal(policy)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Info", "JSON Marshalling error: %s", err.Error())
		return
	}
	if _, err = resp.Write(jsonStr); err != nil {
		resp.WriteLog(http.StatusBadRequest, "Info", err.Error())
	}
}

func (a *PolicyHandler) putWIdHandler(resp *logging.ResponseLogger, req *http.Request) {
	params := mux.Vars(req)
	policyName := params["policyname"]

	var policy change.ChangeData
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&policy)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Debug", "Decoder error: %s", err.Error())
		return
	}
	policy.Content.Object = policyName //route param takes precedence
	a.putPolicyHierarchy(&policy, resp, req)
}

func (a *PolicyHandler) deleteWIdHandler(resp *logging.ResponseLogger, req *http.Request) {
	params := mux.Vars(req)
	policyName := params["policyname"]
	a.deletePolicyHierarchy(policyName, resp, req)
}

// putPolicyHierarchy tries to send the policy upstream for validation against ontology. If the passing
// upstream fails due to connection issues, this node acts as the master and validates against the ontology.
// IT IS NOT PUT ON QUEUE. On resuming of the connection, the policy upstream takes precedence over the polcies downstream.
// We dont need a repository callback set as we rely on the HEAD request informing the nodes below that the policy has changed
// prompting them to request the latest policies.
func (a *PolicyHandler) putPolicyHierarchy(policy *change.ChangeData, resp *logging.ResponseLogger, req *http.Request) {
	if err := ParsePolicy(bytes.NewReader(policy.Content.Files["Rules"])); err != nil {
		resp.WriteLog(http.StatusBadRequest, "Debug", "Parser failed. Check input syntax. Error:%s", err.Error())
		return
	}
	policy.ObjectType = change.POLICY
	if err := a.completePolicyDataStruct(policy); err != nil {
		resp.WriteLog(http.StatusBadRequest, "Notice", "Policy missing fields, error: %s", err.Error())
		return
	}
	engine, err := change.GetCMEngine(nil)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Notice", "PUT /Policy::Could not get instance of change management engine. Cannot proceed")
		return
	}
	policy.TransactionID, err = engine.BeginTransaction(policy, policy.Log.Message)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Notice", "PUT /Policy::BeginTransaction error: %s", err.Error())
		return
	}

	upstreamNodeIP, err := nodeComm.GetUpstreamNodeIP()
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Warning", "PUT /Policy::Error retrieving IP Address of the upstream node. Cannot proceed")
		return
	}
	// now the logic
	if upstreamNodeIP == nil { //if we are the master node
		validated, explanation, err := ValidateAgainstOntology(a.log, policy)
		if err != nil {
			a.log.Debug(explanation)
			resp.WriteLog(http.StatusBadRequest, "Notice", "PUT /Policy::Error trying to validate. Error:%s", err.Error())
			return
		}
		if validated {
			// at this point go ahead and store the modified policy in the git repository. The heartbeat system will handle sending it down
			a.savePolicyToRepo(policy)
			LogOntologyInconsistencies(a.log)
		} else {
			a.log.Debug("Failed to validate policy, not saved to repo. Got explanantion:%s", explanation)
			resp.WriteLog(http.StatusBadRequest, "Notice", "PUT /Policy::Could not validate policy. Explanation:%s", explanation)
			return
		}
	} else { //pass upstream. If any communication error here, start acting as master and validate
		err = nodeComm.UpdatePolicyUpstream(policy, upstreamNodeIP) //if error here tell user no can do
		if err != nil {
			if strings.HasPrefix(err.Error(), global.NoConnection) || strings.HasPrefix(err.Error(), global.NoHttpOK) {
				//connection error, we are now master
				validated, explanation, err := ValidateAgainstOntology(a.log, policy)
				if err != nil {
					a.log.Debug(explanation)
					resp.WriteLog(http.StatusBadRequest, "Notice", "PUT /Policy::Error trying to validate while acting as master. Error:%s", err.Error())
					return
				}
				if validated {
					// at this point go ahead and store the modified policy in the git repository. The heartbeat system will handle sending it down
					a.savePolicyToRepo(policy)
					LogOntologyInconsistencies(a.log)
				} else {
					a.log.Debug("Failed to validate policy while acting as master, not saved to repo. Got explanantion:%s", explanation)
					resp.WriteLog(http.StatusBadRequest, "Notice", "PUT /Policy::Could not validate policy while acting as master. Explanantion:%s", explanation)
					return
				}
			} else {
				resp.WriteLog(http.StatusBadRequest, "Notice", "PUT /policy::Could not send the policy upstream, Error:%s", err.Error())
				return
			}
		}
	}
	resp.WriteHeader(http.StatusOK)
}

func (a *PolicyHandler) deletePolicyHierarchy(policyName string, resp *logging.ResponseLogger, req *http.Request) {
	engine, err := change.GetCMEngine(nil)
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Notice", "DELETE /Policy::Could not get instance of change management engine. Cannot proceed")
		return
	}

	upstreamNodeIP, err := nodeComm.GetUpstreamNodeIP()
	if err != nil {
		resp.WriteLog(http.StatusBadRequest, "Warning", "DELETE /Policy::Error retrieving IP Address of the upstream node. Cannot proceed")
		return
	}
	// now the logic
	if upstreamNodeIP == nil { //if we are the master node
		// at this point go ahead and delete the policy in the git repository. The heartbeat system will handle sending it down
		if err = engine.RemoveObject(change.POLICY, policyName, &change.CMAuthor{Name: "Policy Engine"}); err != nil {
			resp.WriteLog(http.StatusBadRequest, "Warning", "DELETE /Policy::Error in CME engine, Error:%s", err.Error())
			return
		}
		status, explanation, err := ValidateAgainstOntology(a.log, nil)
		if err != nil {
			resp.WriteLog(http.StatusBadRequest, "Warning", "DELETE /Policy::Error validating with ontology, Error:%s", err.Error())
			return
		}
		if status == false {
			resp.WriteLog(http.StatusBadRequest, "Warning", "DELETE /Policy::COuld not validate with ontology, Explanation:%s", explanation)
			return
		}
	} else { //pass upstream. If any error here, give up
		err = nodeComm.DeletePolicyUpstream(policyName, upstreamNodeIP) //if error here tell user no can do
		if err != nil {
			resp.WriteLog(http.StatusBadRequest, "Notice", "DELETE /policy::Could not send the policy upstream, Error:%s", err.Error())
			return
		}
	}
	resp.WriteHeader(http.StatusOK)
}

func (a *PolicyHandler) handleWIdValidateRoute(writer http.ResponseWriter, req *http.Request) {
	writer.Write([]byte("policyvalidate"))
}

func (a *PolicyHandler) handleWIdValidateStatusRoute(writer http.ResponseWriter, req *http.Request) {
	writer.Write([]byte("policyvalidatestatus"))
}

/******************************* change management engine related functions ********************************/

func (a *PolicyHandler) completePolicyDataStruct(policy *change.ChangeData) error {
	user := database.PbUser{Name: policy.Author.Name, Email: policy.Author.Email}
	if policy.Author.Email == "" {
		err := user.GetByName(a.db)
		if err != nil {
			a.log.Debug("SaveToRepo: Could not recover user details to store to git repository")
			return err
		}
		policy.Author.Email = user.Email
	}
	if policy.Author.When.IsZero() {
		policy.Author.When = time.Now()
	}
	if policy.Log.Message == "" {
		a.log.Debug("SaveToRepo: Cannot commit with empty message")
		return errors.New("Commit message cannot be empty")
	}
	return nil
}

func (a *PolicyHandler) getPolicyContentFromRepo(policyName string) (*change.CMContent, error) {
	changeEng, err := change.GetCMEngine(nil)
	if err != nil {
		a.log.Notice("Could not get handle to change management engine. Error:%s", err.Error())
		return nil, err
	}
	cdata, err := changeEng.GetObject(change.POLICY, policyName)
	if err != nil {
		return nil, err
	}
	return cdata.Content, nil
}

func (a *PolicyHandler) savePolicyToRepo(policy *change.ChangeData) error {
	changeEng, err := change.GetCMEngine(nil)
	if err != nil {
		a.log.Notice("Could not get handle to change management engine. Error:%s", err.Error())
		return err
	}
	_, err = changeEng.VersionObject(policy, policy.Log.Message)

	if err != nil {
		a.log.Debug("SaveToRepo: ChangeEngine returned with error: " + err.Error())
		return err
	}
	changeEng.FinalizeTransaction(policy)
	return nil
}
