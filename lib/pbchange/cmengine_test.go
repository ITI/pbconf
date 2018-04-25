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
	"runtime"
	"testing"
	"time"

	"fmt"
	"io/ioutil"
	"os"

	config "github.com/iti/pbconf/lib/pbconfig"
)

func begin(t *testing.T, name string) {
	fmt.Printf("##################### Begin %s #####################\n", name)
}

func end(t *testing.T, name string) {
	fmt.Printf("###################### End %s ######################\n", name)
}

func setup() *config.Config {
	testrepopath, err := ioutil.TempDir("", "cmengine")
	if err != nil {
		panic(err)
	}

	cfg := new(config.Config)
	cfg.ChMgmt.RepoPath = testrepopath
	cfg.ChMgmt.LogLevel = "DEBUG"

	return cfg
}

func cleanup(cfg *config.Config, e *CMEngine) {
	err := os.RemoveAll(cfg.ChMgmt.RepoPath)
	if err != nil {
		panic(err)
	}

	e.Free()
}

// borrowed from from go2git
func checkFatal(t *testing.T, err error) {
	if err == nil {
		return
	}

	// The failure happens at wherever we were called, not here
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		t.Fatalf("Unable to get caller")
	}
	t.Fatalf("Fail at %v:%v; %v", file, line, err)
}

func TestNewRepo(t *testing.T) {
	begin(t, "TestNewRepo")
	defer end(t, "TestNewRepo")

	cfg := setup()
	engine, err := GetCMEngine(cfg)
	defer cleanup(cfg, engine)

	_ = engine
	checkFatal(t, err)
}

func TestAddFile(t *testing.T) {
	begin(t, "TestAddFile")
	defer end(t, "TestAddFile")

	cfg := setup()
	engine, err := GetCMEngine(cfg)
	defer cleanup(cfg, engine)

	checkFatal(t, err)

	commit := NewCMContent("test")
	commit.Files["file1"] = []byte("This is the content of a test file")

	sig := &CMAuthor{
		Name:  "Larry Bird",
		Email: "tootall@celtics.net",
		When:  time.Now(),
	}

	data := &ChangeData{
		ObjectType: DEVICE,
		Content:    commit,
		Author:     sig,
	}

	s, e := engine.VersionObject(data, "")
	checkFatal(t, e)
	t.Log("Commit ID: " + s)
}

func TestLog(t *testing.T) {
	begin(t, "TestLog")
	defer end(t, "TestLog")

	cfg := setup()
	engine, err := GetCMEngine(cfg)
	defer cleanup(cfg, engine)

	checkFatal(t, err)

	var ids []string
	for _, file := range []string{"file1", "file2"} {
		t.Log(fmt.Sprintf("Adding %s", file))
		commit := NewCMContent(file)
		commit.Files[file] = []byte("This is the content of a test file")

		sig := &CMAuthor{
			Name:  "Michael Jordan",
			Email: "larrysajoke@bulls.net",
			When:  time.Now(),
		}

		data := &ChangeData{
			ObjectType: DEVICE,
			Content:    commit,
			Author:     sig,
		}

		id, e := engine.VersionObject(data, "testing")
		checkFatal(t, e)
		t.Log("Commit ID : " + id)
		ids = append(ids, id)
	}

	loglines, err := engine.Log(DEVICE, "")
	checkFatal(t, err)

	// Need 1 extra line to account for the init commit
	if len(loglines) != 3 {
		t.Error("Expecting length 3 got", len(loglines))
	}

	for _, l := range loglines[:len(loglines)-1] {
		if l.Author != "Michael Jordan" {
			t.Error("Expecting string 'Michael Jordan' got", l.Author)
		}
		// Comitter should be the CME
		if l.Committer != "Change Management Engine" {
			t.Error("Expecting string 'Change Management Engine' got", l.Committer)
		}
		if l.AuthorEmail != "larrysajoke@bulls.net" {
			t.Error("Expecting string 'larrysajoke@bulls.net' got", l.AuthorEmail)
		}
		// Comitter should be the CME
		if l.CommitterEmail != "ignore@ignore" {
			t.Error("Expecting string 'ignore@ignore' got", l.CommitterEmail)
		}
		if l.Message != "testing" {
			t.Error("Expecting string 'testing' got", l.Message)
		}

	}
}
