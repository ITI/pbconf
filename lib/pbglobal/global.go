package pbglobal

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
	"container/list"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/gorilla/mux"
	"golang.org/x/net/context"
	"golang.org/x/net/http2"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/iti/pbconf/lib/pbconfig"
)

var Client *http.Client

var UpstreamConnectedStatus bool

var RootNode string
var Queue *list.List
var AlarmChan = make(chan chan string)

var ApiPort = ":443"
var NoConnection = "No Connection established"
var NoHttpOK = "Did not get http.StatusOK"
var ApiUrlVersioning = []string{"/v{version:[0-9]+}", ""} //ORDER MATTERS!!! DO NOT CHANGE
var ApiServerTimeout = 60.0                               //seconds need the float .0 THIS VARIABLE DECIDES WHEN THE ALARM LONG POLL TIMES OUT

var CTX context.Context

func Start(nodeName string, cfgWebApi *pbconfig.CfgWebAPI) {
	caCertPool := CreateCertPool(cfgWebApi.GetStringVar("TrustedCerts"))
	tlsConfig := &tls.Config{RootCAs: caCertPool}
	/**** If you have RequireClientCert set to false in the conf file, and do not want to set up a trusted certificate pool
	  on a device, FOR TESTING PURPOSES ONLY, you can use the tlsconfig stmt below to skip any verification of the server certificate
	  tlsConfig := &tls.Config{InsecureSkipVerify:true}
	  *****/

	if cfgWebApi.GetBoolVar("RequireClientCert") {
		cert, err := tls.LoadX509KeyPair(cfgWebApi.GetStringVar("ClientCert"), cfgWebApi.GetStringVar("ClientKey"))
		if err != nil {
			panic(err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}
	tlsConfig.BuildNameToCertificate()
	transport := http.Transport{TLSClientConfig: tlsConfig}
	//FORCE CLIENT TO CONNECT TO API SERVER USING HTTP/2
	err := http2.ConfigureTransport(&transport)
	if err != nil {
		fmt.Println("*************WAS NOT ABLE TO CONFIGURE HTTP2 FOR NODE WHEN IT ACTS AS CLIENT ********")
	}

	Client = &http.Client{Transport: &transport}

	RootNode = nodeName
	ApiPort = cfgWebApi.GetStringVar("Listen")
	Queue = list.New()

	CTX = context.Background()
}

func CreateCertPool(clientdir string) *x509.CertPool {
	caCertPool := x509.NewCertPool()

	if clientdir == "" {
		return caCertPool
	}

	fileList, err := ioutil.ReadDir(clientdir)
	if err != nil {
		fmt.Printf("Could not read client certificate dir: %s as specified by the config file", clientdir)
		return caCertPool
	}
	for i := 0; i < len(fileList); i++ {
		fileName := fileList[i].Name()
		if filepath.Ext(fileName) == ".pem" || filepath.Ext(fileName) == ".crt" {
			cl_cert, err := ioutil.ReadFile(clientdir + "/" + fileName)
			if err != nil {
				fmt.Printf("Could not read certificate file %v, got error %v\n", fileName, err.Error())
			} else {
				ok := caCertPool.AppendCertsFromPEM(cl_cert)
				if !ok {
					fmt.Printf("******COULD NOT ADD certificate for %v in certpool\n", clientdir+"/"+fileName)
				}
			}
		}
	}
	return caCertPool
}

func ProcessVersioning(req *http.Request) (*int, error) {
	processVersionStr := func(verStr string) (*int, error) {
		version, err := strconv.Atoi(verStr)
		if err != nil {
			return nil, err
		}
		return &version, nil
	}
	//check if version is in the url
	params := mux.Vars(req)
	versionParam := params["version"]
	if versionParam != "" {
		return processVersionStr(versionParam)
	}
	//check if version is sent as part of Accept header in the request
	acceptHeader := req.Header.Get("Accept")
	if acceptHeader != "" {
		re := regexp.MustCompile("pbconfversion=([0-9]+)")
		matchStr := re.FindStringSubmatch(acceptHeader)
		if matchStr == nil || len(matchStr) < 2 {
			return nil, nil //not part of accept header
		}
		return processVersionStr(matchStr[1])
	}
	return nil, nil //not part of accept header or url
}
