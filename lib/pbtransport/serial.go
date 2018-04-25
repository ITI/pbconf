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

	goserial "github.com/tarm/serial"

	"fmt"
	"io"
	"strconv"
	"strings"
)

func init() {
	log, _ = logging.GetLogger("Serial Transport")
}

type Serial struct {
	port *goserial.Port
}

func NewSerial(drvSrvName string) ClientTransport {
	ll := logging.GetLevel(drvSrvName)
	log, _ = logging.GetLogger(drvSrvName + ":Serial Transport")
	logging.SetLevel(ll, drvSrvName+":Serial Transport")
	return new(Serial)
}

func (s *Serial) Dial(id int64, dst string) error {
	con := strings.Split(dst, ":")
	baud, err := strconv.Atoi(con[1])
	if err != nil {
		return err
	}

	log.Debug(fmt.Sprintf("Opening %v at %v baud", con[0], baud))
	c := &goserial.Config{Name: con[0], Baud: baud}
	s.port, err = goserial.OpenPort(c)
	if err != nil {
		log.Error("Failed to open serial port")
		return err
	}

	return nil
}

func (s *Serial) Read(buf []byte) (int, error) {
	return s.port.Read(buf)
}

func (s *Serial) Write(buf []byte) (int, error) {
	return s.port.Write(buf)
}

func (s *Serial) Close() error {
	return s.port.Close()
}

func (s *Serial) Interact(srv io.ReadWriter) {
	errchan := make(chan error)
	exitchan := make(chan bool)

	go func() {
		for {
			_, e := io.Copy(s, srv)
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
			n, e := io.Copy(srv, s)
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

func (s *Serial) SetCredentialFn(fn CredentialFn) {}

func (s *Serial) RecvFile(file string) ([]byte, error) {
	return nil, NotImplemented("RecvFile not implemented")
}

func (s *Serial) SendFile(name string, data []byte) error {
	return NotImplemented("SendFile not implemented")
}

func (s *Serial) InternalAuth() bool {
	return false
}
