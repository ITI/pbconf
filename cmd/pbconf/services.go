package main

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
	"bufio"
	"bytes"
	"io"
	"os/exec"

	logging "github.com/iti/pbconf/lib/pblogger"
)

var log logging.Logger

func init() {
	log, _ = logging.GetLogger("Main")

}

type service struct {
	Name      string
	Proc      *exec.Cmd
	isRunning bool
	pipe      io.Reader
}

func (s *service) Exited() bool {
	if s.isRunning {
		b := bufio.NewReader(s.pipe)
		_, _, err := b.ReadLine()
		if err == io.EOF {
			return false
		}
		if err != nil {
			return true
		}
	}
	return false
}

func (s *service) Restart() error {
	var buf bytes.Buffer

	for _, arg := range s.Proc.Args {
		buf.WriteString(arg + " ")
	}

	log.Debug(buf.String())
	s.Proc = exec.Command(s.Proc.Path, buf.String())
	e := s.Start()
	if e != nil {
		log.Debug("Restart Error: " + e.Error())
	}
	return e
}
