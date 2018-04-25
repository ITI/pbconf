package pbwebui

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
	"encoding/json"
	"fmt"
	"golang.org/x/net/http2"
	"html/template"
	"net/http"
	"net/http/httputil"
	"strconv"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	auth "github.com/iti/pbconf/lib/pbauth"
	pbconfig "github.com/iti/pbconf/lib/pbconfig"
	database "github.com/iti/pbconf/lib/pbdatabase"
	"github.com/iti/pbconf/lib/pbglobal"
	"github.com/iti/pbconf/lib/pbhsts"
	logging "github.com/iti/pbconf/lib/pblogger"
)

var log logging.Logger
var webDir string
var proxy *httputil.ReverseProxy
var myAuth *auth.Auth
var cookieStore *sessions.CookieStore
var tokenExpiryTime int
var pbSessionName = "PBCONF-session-name"
var enableLongPolling bool

func StartWebUIhandler(cfg *pbconfig.CfgWebUI, cfgApi *pbconfig.CfgWebAPI, alarmDests []string, db database.AppDatabase) {
	if cfg.GetBoolVar("EnableWebApp") == false {
		return
	}
	log, _ = logging.GetLogger("HTTP UI")
	uiLogLevel := cfg.GetStringVar("LogLevel")
	logging.SetLevel(uiLogLevel, "HTTP UI")
	log.Debug("startwebUIhandler()")
	webDir = cfg.GetStringVar("WebDir")
	//check if longpolling needs activation
	for _, alarm := range alarmDests {
		if alarm == "webui" {
			enableLongPolling = true
		}
	}
	//setup Proxy
	caCertPool := pbglobal.CreateCertPool(cfgApi.GetStringVar("TrustedCerts"))
	tlsConfig := &tls.Config{
		RootCAs: caCertPool, //needed so we accept the api's certificate
	}
	/**** If you have RequireClientCert set to false in the conf file, and do not want to set up a trusted certificate pool
	  on a device, FOR TESTING PURPOSES ONLY, you can use the tlsconfig stmt below to skip any verification of the server certificate
	  tlsConfig := &tls.Config{InsecureSkipVerify:true}
	  *****/
	if cfgApi.GetBoolVar("RequireClientCert") {
		cert, err := tls.LoadX509KeyPair(
			cfg.GetStringVar("ProxyCert"),
			cfg.GetStringVar("ProxyKey"))
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
		log.Debug("*************WAS NOT ABLE TO CONFIGURE TRANSPORT FOR HTTP2********")
	}

	director := func(req *http.Request) {
		req.URL.Scheme = "https"
		req.URL.Host = "localhost" + cfgApi.GetStringVar("Listen")
	}

	proxy = &httputil.ReverseProxy{
		Director:  director,
		Transport: &transport,
	}
	//end of setting up proxy

	cert, err := tls.LoadX509KeyPair(
		cfgApi.GetStringVar("ServerCert"),
		cfgApi.GetStringVar("ServerKey"))
	if err != nil {
		panic(err)
	}

	tlsconfig := tls.Config{
		Certificates:             []tls.Certificate{cert},
		PreferServerCipherSuites: true,
		// Force TLS 1.2 - Shouldn't need Max version as the highest
		// available for both side should be selected
		MinVersion: tls.VersionTLS12,
	}
	server := &http.Server{
		Addr:           cfg.GetStringVar("Listen"),
		MaxHeaderBytes: 1 << 20,
		TLSConfig:      &tlsconfig,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", webuiRequestHandler)

	myAuth = auth.NewAuthHandler(uiLogLevel, db)
	mux.HandleFunc("/ui", rootHandler)
	mux.HandleFunc("/ui/login", loginHandler)
	mux.HandleFunc("/ui/logout", logoutHandler)
	//serve all the resource files (bootstrap.min.js, pbconf.js that are referred to in the index.html file)
	fileserver := http.FileServer(http.Dir(webDir))
	mux.Handle("/resource/", fileserver)

	var hsts http.Handler
	if cfg.GetBoolVar("UseHSTS") {
		hsts = pbhsts.NewHSTS(mux, "HTTP UI")
	} else {
		hsts = mux
	}

	server.Handler = context.ClearHandler(hsts) //not using gorilla/mux, stop memory leaks see gorilla/sessions
	cookieStore = sessions.NewCookieStore([]byte("my-pbconf-supersecret"))
	cookieStore.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 1, // COPIED FROM sessions.NewCookieStore cookie valid for 1 day
		HttpOnly: true,
		Secure:   true,
	}
	tokenExpiryTime, err = strconv.Atoi(cfg.GetStringVar("TokenExpiryTime"))
	if err != nil || tokenExpiryTime > 10 || tokenExpiryTime < 5 {
		tokenExpiryTime = 5
	}

	listener, e := tls.Listen("tcp", cfg.GetStringVar("Listen"), &tlsconfig)
	if e != nil {
		log.Fatal(e.Error())
	}

	server.Serve(listener)
}

func rootHandler(writer http.ResponseWriter, req *http.Request) {
	session, err := cookieStore.Get(req, pbSessionName)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		log.Debug("Error Getting session from cookie store, Err:%s", err.Error())
		return
	}
	if session.IsNew {
		showLoginPage(writer, req)
		return
	}
	token, ok := parseTokenValidity(session.Values["token"].(string), writer)
	if ok {
		handleAuthenticatedHomePage(writer, token)
	} else { //can enter this branch because we dynamically generate new RSA keys for tokens on application restart.
		//Any old logins will be invalid even if there is a time-valid token. the decrypting will fail.
		showLoginPage(writer, req)
	}
}

func loginHandler(writer http.ResponseWriter, req *http.Request) {
	usr := struct {
		User string
		Pass string
	}{}
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&usr)
	if err != nil {
		log.Debug("POST /ui/login::Decoder error: %s", err.Error())
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	user := usr.User
	pass := usr.Pass

	dbUser, err := myAuth.AuthenticateUser(user, pass)
	if err != nil {
		writer.WriteHeader(http.StatusUnauthorized)
		log.Debug("Error authenticating user, Error:%s", err.Error())
		return
	}
	tokenString, err := myAuth.GenerateToken(dbUser, tokenExpiryTime)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		log.Debug("Error generating token string, %s", err.Error())
		return
	}
	err = updateSessionToken(tokenString, writer, req)
	if err != nil {
		return
	}
	writer.WriteHeader(http.StatusOK)
}

func logoutHandler(writer http.ResponseWriter, req *http.Request) {
	session, err := cookieStore.Get(req, pbSessionName)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		log.Debug("Error Getting session from cookie store, Err:%s", err.Error())
		return
	}
	session.Options.MaxAge = -1
	err = session.Save(req, writer)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		log.Debug("Error Saving session, Err:", err.Error())
		return
	}
	//redirect to login page? httpStatus?
	writer.WriteHeader(http.StatusUnauthorized)
}

func webuiRequestHandler(writer http.ResponseWriter, req *http.Request) {
	session, err := cookieStore.Get(req, pbSessionName)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		log.Debug("Error Getting session from cookie store, Err:%s", err.Error())
		return
	}
	tokenstr, found := session.Values["token"].(string)
	if !found {
		writer.WriteHeader(http.StatusUnauthorized)
		return
	}
	token, ok := parseTokenValidity(tokenstr, writer)
	if ok {
		if req.Method != "HEAD" && req.URL.Path != "/alarm" {
			//renew the token for x more time, if for some reason we can't, token will eventually expire
			renewedTokenString, _ := myAuth.RenewToken(token, tokenExpiryTime)
			updateSessionToken(renewedTokenString, writer, req)
		}
		handleAuthenticatedRequest(writer, req, token)
		return
	} else {
		//return a code that redirects the js file to login page
		writer.WriteHeader(http.StatusUnauthorized)
	}
}

func showLoginPage(writer http.ResponseWriter, req *http.Request) {
	t, err := template.ParseFiles(webDir + "/login.html")
	if err != nil {
		log.Debug("Error parsing login.html file Error:%s", err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = t.Execute(writer, req)
	if err != nil {
		log.Debug("Error executing template Error:%s", err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func handleAuthenticatedHomePage(writer http.ResponseWriter, token *jwt.Token) {
	userInfo, ok := getTokenClaimsInfo(writer, token)
	if !ok {
		return
	}
	webData := struct {
		auth.UserClaim
		EnableLongPoll bool
	}{*userInfo, enableLongPolling}

	writer.Header().Set("Content-Type", "text/html")
	writer.WriteHeader(http.StatusOK)
	t, err := template.ParseFiles(webDir + "/index.html")
	if err != nil {
		log.Debug("Error parsing template file index.html, ", err)
	}
	err = t.Execute(writer, webData)
	if err != nil {
		log.Debug("Error template.Execute, ", err)
	}
}

func handleAuthenticatedRequest(writer http.ResponseWriter, req *http.Request, token *jwt.Token) {
	userInfo, ok := getTokenClaimsInfo(writer, token)
	if !ok {
		return
	}
	allowed := checkPermissions(req, userInfo.Role)
	if !allowed {
		writer.WriteHeader(http.StatusForbidden)
		return
	}
	proxy.ServeHTTP(writer, req)
}

func checkPermissions(req *http.Request, role string) bool {
	if role == "user" {
		if req.Method == "GET" || req.Method == "HEAD" {
			return true
		}
		return false
	}
	return true
}

func parseTokenValidity(tokenString string, writer http.ResponseWriter) (*jwt.Token, bool) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return myAuth.RsaPublicKey(), nil
	})

	// branch out into the possible error from signing
	switch err.(type) {
	case nil: // no error
		if !token.Valid { // but may still be invalid
			log.Debug("WHAT? Invalid Token? Now what!")
			return token, false
		}
		// see stdout and watch for the CustomUserInfo, nicely unmarshalled
		return token, true
	case *jwt.ValidationError: // something was wrong during the validation
		vErr := err.(*jwt.ValidationError)
		switch vErr.Errors {
		case jwt.ValidationErrorExpired:
			log.Debug("Token Expired, get a new one.")
			return token, false
		default:
			writer.WriteHeader(http.StatusInternalServerError)
			log.Debug("ValidationError error: %+v\n", vErr.Errors)
			return token, false
		}
	default: // something else went wrong
		writer.WriteHeader(http.StatusInternalServerError)
		log.Debug("Token parse error: %v\n", err)
		return token, false
	}
	return token, false //should never reach here
}

func getTokenClaimsInfo(writer http.ResponseWriter, token *jwt.Token) (*auth.UserClaim, bool) {
	u, ok := token.Claims["UserInfo"].(map[string]interface{})
	if !ok {
		writer.WriteHeader(http.StatusUnauthorized)
		log.Debug("No User claim found in token, contact developer.")
		return nil, ok
	}
	role, ok := u["Role"].(string)
	if !ok {
		writer.WriteHeader(http.StatusUnauthorized)
		log.Debug("No user role found in token claims, contact developer.")
		return nil, ok
	}
	name, ok := u["Name"].(string)
	if !ok {
		writer.WriteHeader(http.StatusUnauthorized)
		log.Debug("No user name found in token claims, contact developer.")
		return nil, ok
	}
	return &auth.UserClaim{name, role}, true
}

func updateSessionToken(tokenString string, writer http.ResponseWriter, req *http.Request) error {
	session, err := cookieStore.Get(req, pbSessionName)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		log.Debug("Error Getting session from cookie store, Err: %s", err.Error())
		return err
	}
	session.Values["token"] = tokenString
	err = session.Save(req, writer)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		log.Debug("Error Saving session, Err:", err.Error())
		return err
	}
	return nil
}
