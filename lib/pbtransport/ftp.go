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
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/dutchcoders/goftp"

	logging "github.com/iti/pbconf/lib/pblogger"
)

func init() {
	log, _ = logging.GetLogger("FTP Transport")
}

type FTP struct {
	client *goftp.FTP
	authcb CredentialFn
	id     int64
	dst    string
}

func NewFTP(drvSrvName string) ClientTransport {
	ll := logging.GetLevel(drvSrvName)
	log, _ = logging.GetLogger(drvSrvName + ":FTP Transport")
	logging.SetLevel(ll, drvSrvName+":FTP Transport")
	return new(FTP)
}

func (f *FTP) Dial(id int64, dst string) error {
	f.id = id
	f.dst = dst

	return nil
}

func (f *FTP) dial() error {
	id := f.id
	dst := f.dst

	if f.authcb == nil {
		return errors.New("Missing Credential Function, can't continue")
	}

	u, p, err := f.authcb(id)
	if err != nil {
		return err
	}

	var ftp *goftp.FTP
	if ftp, err = goftp.Connect(dst); err != nil {
		return err
	}
	f.client = ftp

	if err = ftp.Login(u, p); err != nil {
		return err
	}

	return nil
}

func (f *FTP) Read(c []byte) (int, error) {
	f.dial()
	defer f.client.Close()

	cmd := string(c)

	switch cmd {
	case "list":
		if files, err := f.client.List(""); err != nil {
			return 0, err
		} else {
			return copyToBuf(c, strings.Join(files, "\n"))
		}
	case "pwd":
		if curpath, err := f.client.Pwd(); err != nil {
			return 0, err
		} else {
			return copyToBuf(c, curpath)
		}
	}
	return 0, NotImplemented(fmt.Sprintf("%s command not implemented", cmd))
}

func (f *FTP) Write(c []byte) (int, error) {
	f.dial()
	defer f.client.Close()

	return 0, f.client.Cwd(string(c))
}

func (f *FTP) RecvFile(file string) ([]byte, error) {
	f.dial()
	defer f.client.Close()

	var c []byte
	_, err := f.client.Retr(file, func(r io.Reader) error {
		var e error
		c, e = newBufCopy(r)
		return e
	})

	return c, err
}

func (f *FTP) SendFile(name string, data []byte) error {
	f.dial()
	defer f.client.Close()

	buf := bytes.NewBuffer(data)
	if err := f.client.Stor(name, buf); err != nil {
		return err
	}
	return nil
}

func (f *FTP) Interact(r io.ReadWriter) {
}

func (f *FTP) SetCredentialFn(fn CredentialFn) {
	f.authcb = fn
}

func (f *FTP) InternalAuth() bool {
	return true
}

func (f *FTP) Close() error {
	return nil
}

func copyToBuf(buf []byte, str string) (int, error) {
	i := 0
	for _, r := range str {
		if i >= len(buf) {
			break
		}

		buf[i] = byte(r)
		i++
	}
	return i, nil
}

func newBufCopy(r io.Reader) (buf []byte, err error) {
	b := make([]byte, 1024)

	var l int
	for {
		l, err = r.Read(b)
		if l > 0 {
			buf = append(buf, b[:l]...)
		} else {
			return
		}
	}
}
