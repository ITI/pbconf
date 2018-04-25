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
	"errors"
	"io"
)

type CredentialFn func(id int64) (username, password string, err error)

type ClientTransport interface {
	Dial(int64, string) error
	Read([]byte) (int, error)
	Write([]byte) (int, error)
	Interact(io.ReadWriter)
	SendFile(string, []byte) error
	RecvFile(string) ([]byte, error)
	SetCredentialFn(CredentialFn)
	InternalAuth() bool
	Close() error
}

type ErrNotImplemented struct {
	error
}

func NotImplemented(msg string) ErrNotImplemented {
	return ErrNotImplemented{error: errors.New(msg)}
}
