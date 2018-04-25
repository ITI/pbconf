package pbchange

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
	"github.com/twinj/uuid"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
)

func (e *CMEngine) getUUID(iters int, fn checkUUID) (string, error) {
	uuid.SwitchFormat(uuid.Clean)

	var id string

	for ctr := 0; ctr < iters; ctr++ {
		id = uuid.NewV4().String()
		if fn(id) {
			return id, nil
		}
	}
	return "", NewCMIDError()
}

func (e *CMEngine) getGitDir(ctype CMType) (string, error) {
	if !e.checkBaseDir() {
		return "", NewCMNoRepoError(ctype.String())
	}

	f := path.Join(e.Repopath, ctype.String())

	if _, err := os.Stat(f); os.IsNotExist(err) {
		return "", err
	}

	return f, nil
}

func (e *CMEngine) checkBaseDir() bool {
	root := e.Repopath

	if root == "" {
		cwd, err := os.Getwd()

		if err != nil {
			log.Warning(err.Error())
			return false
		}

		root = cwd
	}

	f := root
	if _, err := os.Stat(f); os.IsNotExist(err) {
		return false
	}

	return true
}

func (e *CMEngine) MakeRepo(ctype CMType) error {
	if !e.checkBaseDir() {
		return NewCMNoRepoError(ctype.String())
	}

	fullpath := path.Join(e.Repopath, ctype.String())
	if _, err := os.Stat(fullpath); os.IsNotExist(err) {
		os.Mkdir(fullpath, os.ModeDir|0700)
	}

	e.guard.Lock()
	defer e.guard.Unlock()
	if _, err := e.run(ctype, "status"); err == nil {
		// Repo initialized
		return nil
	}

	if _, err := e.run(ctype, "init"); err != nil {
		return err
	}

	if _, err := e.run(ctype, "config", "receive.denyCurrentBranch", "ignore"); err != nil {
		return err
	}

	// Comits will fail if a name and email are not set
	if _, err := e.run(ctype, "config", "user.name", "Change Management Engine"); err != nil {
		return err
	}
	if _, err := e.run(ctype, "config", "user.email", "ignore@ignore"); err != nil {
		return err
	}

	if err := ioutil.WriteFile(path.Join(fullpath, "README"),
		[]byte("PBCONF Repository"), os.ModePerm); err != nil {
		return err
	}

	if _, err := e.run(ctype, "add", "README"); err != nil {
		return err
	}
	if _, err := e.run(ctype, "commit", "-m", "Initializa Repo"); err != nil {
		return err
	}

	return nil
}

// run() runs the command and returns the stdout
func (e *CMEngine) run(cmtype CMType, opts ...string) (string, error) {
	cmd, err := e.cmd(cmtype, opts...)
	if err != nil {
		return "", err
	}
	o, err := cmd.Output()
	return string(o), err
}

// runC() runs the command and returns stdout and stderr
func (e *CMEngine) runC(cmtype CMType, opts ...string) (string, string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd, err := e.cmd(cmtype, opts...)
	if err != nil {
		return "", "", err
	}

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Start()
	if err != nil {
		return stdout.String(), stderr.String(), err
	}

	log.Debug("Waiting")
	err = cmd.Wait()
	if err != nil {
		return stdout.String(), stderr.String(), err
	}

	return stdout.String(), stderr.String(), nil
}

// cmdP returns in/out/err pipes, does not run command
func (e *CMEngine) cmdP(ctype CMType, opts ...string) (*exec.Cmd, io.WriteCloser, io.ReadCloser, io.ReadCloser, error) {

	cmd, err := e.cmd(ctype, opts...)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Error(err.Error())
		return cmd, nil, nil, nil, err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Error(err.Error())
		return cmd, nil, nil, nil, err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Error(err.Error())
		return cmd, nil, nil, nil, err
	}

	return cmd, stdin, stdout, stderr, nil
}

// cmd builds the command
func (e *CMEngine) cmd(cmtype CMType, opts ...string) (*exec.Cmd, error) {
	cmd := exec.Command(e.binpath, opts...)

	dir, err := e.getGitDir(cmtype)
	if err != nil {
		log.Warning("Command Error: %v\n", err)
		return nil, NewCMError(err.Error())
	}
	cmd.Dir = dir
	return cmd, nil
}

func (e *CMEngine) pathExists(path string) bool {
	if _, err := os.Stat("./conf/app.ini"); err != nil {
		return os.IsNotExist(err)
	}

	return true
}
