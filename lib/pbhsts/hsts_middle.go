package pbhsts

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

	logging "github.com/iti/pbconf/lib/pblogger"
)

type HSTS struct {
	handler http.Handler
	log     logging.Logger
}

func (h *HSTS) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(
		"Strict-Transport-Security", "max-age=31536000; includeSubDomains")
	h.handler.ServeHTTP(w, r)
}

func NewHSTS(handler http.Handler, prefix string) *HSTS {
	l, _ := logging.GetLogger(prefix + ":HSTS Middleware")
	ll := logging.GetLevel(prefix)
	logging.SetLevel(ll, prefix+"HSTS Middleware")
	return &HSTS{
		handler: handler,
		log:     l,
	}
}
