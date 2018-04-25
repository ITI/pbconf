package transport

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
	logging "github.com/iti/pbconf/lib/pblogger"
	telnet "github.com/ziutek/telnet"

	"fmt"
	"io"
	"strings"
)

var log logging.Logger

func init() {
	log, _ = logging.GetLogger("Telnet Transport")
}

type Telnet struct {
	connection *telnet.Conn
}

func NewTelnet(drvSrvName string) ClientTransport {
	ll := logging.GetLevel(drvSrvName)
	log, _ = logging.GetLogger(drvSrvName + ":Telnet Transport")
	logging.SetLevel(ll, drvSrvName+":Telnet Transport")
	return new(Telnet)
}

func (t *Telnet) Dial(id int64, dst string) error {
	// Check for port
	if !strings.Contains(dst, ":") {
		dst = strings.Join([]string{dst, "23"}, ":")
	}

	tcon, e := telnet.Dial("tcp", dst)
	if e != nil {
		return e
	}
	t.connection = tcon

	return nil
}

func (t *Telnet) Read(buf []byte) (int, error) {
	return t.connection.Read(buf)
}

func (t *Telnet) Write(buf []byte) (int, error) {
	return t.connection.Write(buf)
}

func (t *Telnet) Close() error {
	return t.connection.Conn.Close()
}

func (t *Telnet) Interact(srv io.ReadWriter) {
	errchan := make(chan error)
	exitchan := make(chan bool)

	go func() {
		for {
			_, e := io.Copy(t, srv)
			// Catch panic
			defer func() {
				recover()
				exitchan <- true
				return
			}()
			if e.Error() == "short write" {
				continue
			}
			errchan <- e
		}
	}()

	go func() {
		for {
			n, e := io.Copy(srv, t)
			// Catch panic
			defer func() {
				recover()
				exitchan <- true
				return
			}()
			log.Debug(fmt.Sprintf("%v: %v", e.Error(), n))
			errchan <- e
		}
	}()

	for {
		select {
		case cerror := <-errchan:
			log.Error(cerror.Error())
		case <-exitchan:
			return
		default:
		}
	}
}

func (t *Telnet) SetCredentialFn(fn CredentialFn) {}

func (t *Telnet) RecvFile(file string) ([]byte, error) {
	return nil, NotImplemented("RecvFile not implemented")
}

func (t *Telnet) SendFile(file string, b []byte) error {
	return NotImplemented("SenFile not implemented")
}

func (t *Telnet) InternalAuth() bool {
	return false
}
