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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

func (engine *CMEngine) VersionRaw(data *ChangeData, message string, branch ...string) (string, error) {

	var useBranch string

	if len(branch) == 0 {
		useBranch = "master"
	} else {
		useBranch = branch[0]
	}

	var werr error
	filesAdded := false

	defer func() {
		if werr != nil || !filesAdded {
			engine.reset(data.ObjectType)
		}
	}()

	log.Debug("Ensuring repo exists")
	if err := engine.MakeRepo(data.ObjectType); err != nil {
		return "", err
	}
	log.Debug("Repo exists")

	if message == "" {
		message = "Automatically added by Change Management Engine"
	}

	// Make sure we're on the correct branch
	log.Info("Resetting repo")
	engine.reset(data.ObjectType)

	engine.guard.Lock()
	if _, err := engine.run(data.ObjectType, "checkout", useBranch); err != nil {
		log.Warning("Got Error: %v\n", err)
		engine.guard.Unlock()
		return "", err
	}

	// Write the file out
	var fsObjPath string
	if dir, err := engine.getGitDir(data.ObjectType); err != nil {
		engine.guard.Unlock()
		return "", err
	} else {
		fsObjPath = filepath.Join(dir, data.Content.Object)
	}
	fsFilePath := filepath.Join(fsObjPath, "raw")
	repoFileBase := filepath.Join(data.Content.Object, "raw")

	os.MkdirAll(fsFilePath, os.ModeDir|0700)

	for fname, cont := range data.Content.Files {

		// Create creates the named file with mode 0600 (before umask),
		// truncating it if it already exists.
		var f *os.File
		f, werr = os.Create(filepath.Join(fsFilePath, fname))
		if werr != nil {
			log.Warning("Error: %v\n", werr)
			engine.guard.Unlock()
			return "", werr
		} else {
			// we don't actually need the open file here
			f.Close()
		}

		werr = ioutil.WriteFile(filepath.Join(fsFilePath, fname), cont, os.ModePerm)
		if werr != nil {
			log.Warning("Error: %v\n", werr)
			engine.guard.Unlock()
			return "", werr
		}

		// Ad the file to the index
		_, werr = engine.run(data.ObjectType, "add", filepath.Join(repoFileBase, fname))
		if werr != nil {
			log.Warning("Error: %v\n", werr)
			engine.guard.Unlock()
			return "", werr
		}
	}

	if o, werr := engine.run(data.ObjectType, "status", "--porcelain"); werr != nil {
		log.Warning("Could not determine repo status: %v", werr)
		return "", werr
	} else {
		if o != "" {
			filesAdded = true
		}
	}

	// Commit

	if filesAdded {
		_, werr = engine.run(data.ObjectType, "commit", "-m", message, "--author",
			fmt.Sprintf("%s <%s>", data.Author.Name, data.Author.Email))
		if werr != nil {
			log.Warning("Error: %v\n", werr)
			engine.guard.Unlock()
			return "", werr
		}
	}

	// GetLatestCommitID() does its own locking
	engine.guard.Unlock()

	var rval string
	if filesAdded {
		rval = engine.GetLatestCommitID(data.ObjectType)
	} else {
		rval = ""
	}

	engine.guard.Lock()
	engine.run(data.ObjectType, "checkout", "master")
	engine.guard.Unlock()

	return rval, nil
}

func (engine *CMEngine) GetRawObject(otype CMType, oname string) (*ChangeData, error) {

	cd := &ChangeData{ObjectType: otype}

	// Git reset
	engine.reset(otype)

	ll, err := engine.Log(otype, oname, 1)
	if err != nil {
		return nil, err
	}

	cd.Log = &ll[0]
	cd.CommitID = ll[0].Id
	cd.Author = &CMAuthor{
		Name:  ll[0].Author,
		Email: ll[0].AuthorEmail,
		When:  ll[0].Time,
	}

	objectPath := filepath.Join(otype.String(), oname, "raw")
	basePath := filepath.Join(engine.Repopath, objectPath)

	content := &CMContent{Object: oname}
	content.Files = make(map[string][]byte, 1)

	filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if path == basePath {
			return err
		}

		f, e := os.Open(path)
		if e != nil {
			return e
		}

		buf, e := ioutil.ReadAll(f)
		if e != nil {
			return e
		}
		content.Files[info.Name()] = buf
		return err
	})

	cd.Content = content

	return cd, nil
}
