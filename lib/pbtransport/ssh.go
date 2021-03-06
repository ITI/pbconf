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

	"errors"
	"golang.org/x/crypto/ssh"
	"io"
)

func init() {
	log, _ = logging.GetLogger("SSH Transport")
}

type SSH struct {
	connection *ssh.Client
	authcb     CredentialFn
}

func NewSSH(drvSrvName string) ClientTransport {
	ll := logging.GetLevel(drvSrvName)
	log, _ = logging.GetLogger(drvSrvName + ":SSH Transport")
	logging.SetLevel(ll, drvSrvName+":SSH Transport")
	return new(SSH)
}

func (s *SSH) Dial(id int64, dst string) error {

	// Make sure we have username and password
	if s.authcb == nil {
		return errors.New("Missing Credential Function, can't continue")
	}

	u, p, err := s.authcb(id)
	if err != nil {
		return err
	}

	config := &ssh.ClientConfig{
		User: u,
		Auth: []ssh.AuthMethod{
			ssh.Password(p),
		},
	}

	client, err := ssh.Dial("tcp", dst, config)
	if err != nil {
		return err
	}

	s.connection = client
	return nil
}

/*
The semantics of Read and write break with the built in SSH transport
As such, they shall be treated as follows:

Read:  buf shall contain a command to be run, buf will be replaced with
		the out put of the command.  Like the telnet transport, the len
		of buf determines the amount of output received.
		For instance, if buf is passed in as
		byte("echo foo") it will be replaced with byte("foo") and 3, nil
		will be returned

Write: buf shal contain the command to be run.  buf will not be replaced,
		Any output generated by the command will be discarded and the returns
		will be 0|1 (0 if the command exited clean), nil will be returned.
*/
func (s *SSH) Read(buf []byte) (int, error) {
	sess, err := s.connection.NewSession()
	if err != nil {
		return 0, err
	}
	defer sess.Close()

	o, err := sess.CombinedOutput(string(buf))

	log.Debug("got output: %s", string(o))

	var i int
	for i = 0; i < len(buf); i++ {
		if i >= len(o) {
			break
		}
		buf[i] = o[i]
	}

	return i, err
}

func (s *SSH) Write(buf []byte) (int, error) {
	sess, err := s.connection.NewSession()
	if err != nil {
		return 0, err
	}
	defer sess.Close()

	err = sess.Run(string(buf))
	if err != nil {
		return 1, err
	}

	return 0, nil
}

func (s *SSH) Close() error {
	return s.connection.Conn.Close()
}

func (s *SSH) Interact(srv io.ReadWriter) {
}

func (s *SSH) SetCredentialFn(fn CredentialFn) {
	s.authcb = fn
}

func (s *SSH) RecvFile(file string) ([]byte, error) {
	return nil, NotImplemented("RecvFile not implemented")
}

func (s *SSH) SendFile(file string, b []byte) error {
	return NotImplemented("SenFile not implemented")
}

func (s *SSH) InternalAuth() bool {
	return true
}
