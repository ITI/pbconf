package pbapi

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
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	mux "github.com/gorilla/mux"

	pbconfig "github.com/iti/pbconf/lib/pbconfig"
	database "github.com/iti/pbconf/lib/pbdatabase"
	devAPI "github.com/iti/pbconf/lib/pbdevice"
	"github.com/iti/pbconf/lib/pbglobal"
	"github.com/iti/pbconf/lib/pbhsts"
	logging "github.com/iti/pbconf/lib/pblogger"
	namespaceAPI "github.com/iti/pbconf/lib/pbnamespace"
	nodeAPI "github.com/iti/pbconf/lib/pbnode"
	policyAPI "github.com/iti/pbconf/lib/pbpolicy"
	reportsAPI "github.com/iti/pbconf/lib/pbreports"
)

var server *APIServer
var Version int

func StartAPIhandler(cfg *pbconfig.CfgWebAPI, db database.AppDatabase) {
	Version = 1
	log, _ := logging.GetLogger("HTTP API")
	apiLogLevel := cfg.GetStringVar("LogLevel")
	logging.SetLevel(apiLogLevel, "HTTP API")
	log.Debug("startAPIhandler()")

	cert, err := tls.LoadX509KeyPair(
		cfg.GetStringVar("ServerCert"),
		cfg.GetStringVar("ServerKey"))
	if err != nil {
		panic(err)
	}
	var caCertPool *x509.CertPool
	clientauth := tls.NoClientCert

	if cfg.GetBoolVar("RequireClientCert") {
		clientauth = tls.RequireAndVerifyClientCert
		//get all certificates that this server is going to trust
		caCertPool = pbglobal.CreateCertPool(cfg.GetStringVar("TrustedCerts"))
	}

	tlsconfig := tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientCAs:    caCertPool,
		// Prefer for the server to select the cipher suite
		PreferServerCipherSuites: true,

		// Force TLS 1.2 - Shouldn't need Max version as the highest
		// available for both side should be selected
		MinVersion: tls.VersionTLS12,

		// Require Client Certificate Auth
		ClientAuth: clientauth,
	}
	tlsconfig.BuildNameToCertificate()
	server = &APIServer{
		Server: http.Server{
			Addr:           cfg.GetStringVar("Listen"),
			MaxHeaderBytes: 1 << 20,
			TLSConfig:      &tlsconfig,
		},
	}

	rootRouter := mux.NewRouter()
	rootRouter.HandleFunc("/version", handleVersion).Methods("GET")
	rootRouter.HandleFunc("/version/{namespace}", handleVersion).Methods("GET")
	rootRouter.HandleFunc("/alarm", handleAlarm).Methods("GET")

	rootRouter.HandleFunc("/", handleRootRoute).Methods("GET")
	rootRouter.NotFoundHandler = http.HandlerFunc(handleCatchAllRoute)
	// API Handlers
	server.AddHandler(devAPI.NewAPIHandler(apiLogLevel, db), rootRouter)
	server.AddHandler(nodeAPI.NewAPIHandler(apiLogLevel, db), rootRouter)
	server.AddHandler(policyAPI.NewAPIHandler(apiLogLevel, db), rootRouter)
	server.AddHandler(reportsAPI.NewAPIHandler(apiLogLevel, db), rootRouter)

	// This route must be last in the list
	server.AddHandler(namespaceAPI.NewAPIHandler(apiLogLevel, db), rootRouter)

	var hsts http.Handler

	if cfg.GetBoolVar("UseHSTS") {
		hsts = pbhsts.NewHSTS(rootRouter, "HTTP API")
	} else {
		hsts = rootRouter
	}

	http.Handle("/", hsts)
	server.Handler = hsts

	l, e := tls.Listen("tcp", cfg.GetStringVar("Listen"), &tlsconfig)
	if e != nil {
		log.Fatal(e.Error())
	}

	server.Listener = l
	server.Serve(server.Listener)
}

func handleVersion(writer http.ResponseWriter, req *http.Request) {
	ns := mux.Vars(req)["namespace"]
	if ns == "" {
		writer.Write([]byte(strconv.Itoa(Version)))
		return
	} else {
		for _, handler := range server.Handlers {
			name, version := handler.GetInfo()
			if name == mux.Vars(req)["namespace"] {
				writer.Write([]byte(strconv.Itoa(version)))
				return
			}
		}
	}
	writer.WriteHeader(http.StatusNotFound)
}

func handleCatchAllRoute(writer http.ResponseWriter, req *http.Request) {
	writer.WriteHeader(http.StatusNotFound)
	writer.Write([]byte("PBCONF endpoint not found!"))
}

func handleRootRoute(writer http.ResponseWriter, req *http.Request) {
	type ApiInfo struct {
		Namespace string
		Version   int
	}
	log, _ := logging.GetLogger("HTTP API")

	infoList := []ApiInfo{ApiInfo{"PBCONF", Version}}
	for _, handler := range server.Handlers {
		name, version := handler.GetInfo()
		infoList = append(infoList, ApiInfo{name, version})
	}
	jsonStr, err := json.Marshal(infoList)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		log.Debug("GET /::Could not marshal the information list Error: %s", err.Error())
		return
	}
	if _, err = writer.Write(jsonStr); err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		log.Debug("GET /::Writing response body Error: %s", err.Error())
		return
	}
}

func handleAlarm(writer http.ResponseWriter, req *http.Request) {
	reqChan := make(chan string)
	timeout, err := strconv.ParseFloat(req.URL.Query().Get("timeout"), 32)
	if err != nil || timeout > 0.8*pbglobal.ApiServerTimeout || timeout < 0 {
		timeout = 0.80 * pbglobal.ApiServerTimeout
	}
	select {
	case pbglobal.AlarmChan <- reqChan:
	case <-time.After(time.Duration(timeout) * time.Second):
		writer.WriteHeader(http.StatusRequestTimeout)
		return
	}
	writer.Write([]byte(<-reqChan))
}
