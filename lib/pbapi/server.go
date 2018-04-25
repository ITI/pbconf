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
	"net"
	"net/http"

	mux "github.com/gorilla/mux"
)

type APIHandler interface {
	AddAPIEndpoints(*mux.Router)
	GetInfo() (string, int)
}

type APIServer struct {
	http.Server
	Listener net.Listener

	Handlers []APIHandler
}

func (s *APIServer) AddHandler(h APIHandler, r *mux.Router) {
	s.Handlers = append(s.Handlers, h)
	s.Handlers[len(s.Handlers)-1].AddAPIEndpoints(r)
}
