package pblogger

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
	"fmt"
	"net/http"
)

type ResponseLogger struct {
	http.ResponseWriter
	Logger
	SrcNodeName string
}
type LogMsg struct {
	Level   string
	Message string
	SrcNode string
}

func (resp *ResponseLogger) WriteLog(httpErrorCode int, logLevel string, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	resp.Log(logLevel, msg)
	resp.WriteHeader(httpErrorCode)
	logmsg := LogMsg{logLevel, msg, resp.SrcNodeName}
	jsonStr, err := json.Marshal(logmsg)
	if err != nil {
		resp.Info("Marshal log message Error: %s", err.Error())
		return
	}
	if _, err = resp.Write(jsonStr); err != nil {
		resp.Notice("Writing response body Error: %s", err.Error())
	}
}
