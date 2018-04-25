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
	"io"
)

type NullTransport struct {
}

func NewNullTransport() ClientTransport {
	return new(NullTransport)
}

func (t *NullTransport) Dial(id int64, dst string) error {
	return NotImplemented("Dial not implemented")
}

func (t *NullTransport) Read(buf []byte) (int, error) {
	return -1, NotImplemented("Read not implemented")
}

func (t *NullTransport) Write(buf []byte) (int, error) {
	return -1, NotImplemented("Write not implemented")
}

func (t *NullTransport) Close() error {
	return nil
}

func (t *NullTransport) Interact(srv io.ReadWriter) {
	_ = srv
}

func (t *NullTransport) RecvFile(file string) ([]byte, error) {
	return nil, NotImplemented("RecvFile not implemented")
}

func (t *NullTransport) SendFile(name string, data []byte) error {
	return NotImplemented("SendFile not implemented")
}

func (t *NullTransport) SetCredentialFn(fn CredentialFn) {}

func (t *NullTransport) InternalAuth() bool {
	return true
}
