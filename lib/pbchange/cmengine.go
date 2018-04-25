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
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	config "github.com/iti/pbconf/lib/pbconfig"
	logging "github.com/iti/pbconf/lib/pblogger"
)

var log logging.Logger

func init() {
	log, _ = logging.GetLogger("CME:Main")
}

var cmEngine *CMEngine

func GetCMEngine(cfg *config.Config) (*CMEngine, error) {
	var err error

	switch {
	case cfg == nil && cmEngine == nil:
		return nil, NewCMInitError()
	case cmEngine == nil && cfg != nil:
		cmEngine, err = initCMEngine(cfg)
		return cmEngine, err
	case cmEngine != nil:
		return cmEngine, nil
	}

	// Shouldn't get here
	return nil, nil
}

func initCMEngine(cfg *config.Config) (*CMEngine, error) {
	path := cfg.ChMgmt.RepoPath
	var binpath string

	if cfg.ChMgmt.BinPath == "" {
		binpath = "/usr/bin/git"
	} else {
		binpath = cfg.ChMgmt.BinPath
	}
	log, _ = logging.GetLogger("CME:Main") //reload the logger which now has ringlogger initialized
	logging.SetLevel(cfg.ChMgmt.LogLevel, "CME:Main")
	err := os.Mkdir(path, os.ModeDir|0700)
	if err != nil {
		if serr, ok := err.(*os.PathError); ok {
			if serr.Err.Error() != "file exists" {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	e := &CMEngine{
		Repopath:    path,
		binpath:     binpath,
		UploadPack:  true,
		ReceivePack: true,
	}

	e.commitCBs = make([]*cbStore, 0)
	e.packRcvdCBs = make([]*cbStore, 0)

	return e, nil
}

// Deallocates engine instance.
//
// Should only be used in special cases
func (engine *CMEngine) Free() {
	cmEngine = nil
}

func (engine *CMEngine) reset(cmtype CMType) {
	if _, err := engine.getGitDir(cmtype); err != nil {
		return
	}

	engine.guard.Lock()
	defer engine.guard.Unlock()

	_, err := engine.run(cmtype, "reset", "--hard")
	if err != nil {
		log.Debug("Reset Error: %v\n", err)
		return
	}
	return
}

func (engine *CMEngine) GetLatestCommitID(cmtype CMType) string {
	engine.guard.Lock()
	defer engine.guard.Unlock()

	id, err := engine.run(cmtype, "log", "--pretty=format:\"%H\"", "-1")
	if err != nil {
		log.Info("Reset Error: %v\n", err)
		return ""
	}

	return string(id)
}

func (engine *CMEngine) HasBranch(cmtype CMType, branch string) bool {
	branches, err := engine.getBranches(cmtype)
	if err != nil {
		log.Warning("Failed to list branches")
		return false
	}

	for _, v := range branches {
		log.Debug("Comparing %s to %s", branch, v)
		if v == branch {
			return true
		}
	}
	return false
}

func (engine *CMEngine) VersionObject(data *ChangeData, message string, branch ...string) (string, error) {

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
	fsFilePath := filepath.Join(fsObjPath, "data")
	repoFileBase := filepath.Join(data.Content.Object, "data")

	os.MkdirAll(fsFilePath, os.ModeDir|0700)

	for fname, cont := range data.Content.Files {
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

func dirExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func (engine *CMEngine) ListObjects(otype CMType) ([]string, error) {
	engine.reset(otype)

	path := filepath.Join(engine.Repopath, otype.String())
	r := make([]string, 0)
	exists, err := dirExists(path)
	if err != nil {
		return r, err
	}
	if !exists {
		return r, NewCMNoRepoError(otype.String())
	}

	objects, walkerr := ioutil.ReadDir(path)
	for _, o := range objects {
		if strings.HasPrefix(o.Name(), ".") {
			continue
		}
		if o.Name() == "README" {
			continue
		}

		r = append(r, o.Name())
	}

	return r, walkerr
}

func (engine *CMEngine) RemoveObject(otype CMType, oname string, author *CMAuthor) error {
	engine.reset(otype)

	basePath := filepath.Join(engine.Repopath, otype.String(), oname)

	engine.guard.Lock()
	defer engine.guard.Unlock()

	_, err := engine.run(otype, "rm", "-r", basePath)
	if err != nil {
		return err
	}
	message := fmt.Sprintf("Removing %s from repository", oname)
	_, err = engine.run(otype, "commit", "-m", message, "--author",
		fmt.Sprintf("%s <%s>", author.Name, author.Email))
	if err != nil {
		log.Warning("Error: %v\n", err)
		return err
	}
	return nil
}

func (engine *CMEngine) RenameObject(otype CMType, oname string, newname string, author *CMAuthor) error {
	engine.reset(otype)

	srcPath := filepath.Join(engine.Repopath, otype.String(), oname)
	destPath := filepath.Join(engine.Repopath, otype.String(), newname)
	engine.guard.Lock()
	defer engine.guard.Unlock()

	_, err := engine.run(otype, "mv", srcPath, destPath)
	if err != nil {
		return err
	}
	message := fmt.Sprintf("Moving %s to %s in the repository", oname, newname)
	_, err = engine.run(otype, "commit", "-m", message, "--author",
		fmt.Sprintf("%s <%s>", author.Name, author.Email))
	if err != nil {
		log.Warning("Error: %v\n", err)
		return err
	}
	return nil
}

func (engine *CMEngine) GetObject(otype CMType, oname string) (*ChangeData, error) {

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

	objectPath := filepath.Join(otype.String(), oname, "data")
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

func (engine *CMEngine) Push(cmtype CMType, upstream Upstream, srcNodeName string) error {

	upurl := fmt.Sprintf("https://%s/%s/cme", upstream.IP(),
		strings.ToLower(cmtype.String()))
	if srcNodeName != "" {
		upurl = fmt.Sprintf("%s/src=%s", upurl, srcNodeName)
	}
	if t := upstream.Transaction(); t != "" {
		upurl = fmt.Sprintf("%s/trans=%s", upurl, t)
	}

	if err := engine.Pull(cmtype, upstream); err != nil {
		return err
	}

	// When using self signed certificates, Git needs to bypass certificate
	// validity checks.
	//stdout, stderr, err := engine.runC(cmtype, "-c", "http.sslVerify=false", "push", upurl, "master:master")

	// When using certificates that the host system can validate, make sure
	// that Git is verifying certificates
	stdout, stderr, err := engine.runC(cmtype, "-c", "http.sslVerify=true", "push", upurl, "master:master")
	log.Debug("Out: %s\nErr: %s", stdout, stderr)
	if err != nil {
		return NewCMCommunicationError(stdout, err)
	}

	return nil
}

func (engine *CMEngine) Pull(cmtype CMType, upstream Upstream) error {
	upurl := fmt.Sprintf("https://%s/%s/cme", upstream.IP(),
		strings.ToLower(cmtype.String()))

	// Should never need a transaction for a pull

	// When using self signed certificates, Git needs to bypass certificate
	// validity checks.
	//stdout, stderr, err := engine.runC(cmtype, "-c", "http.sslVerify=false", "pull", upurl)

	// When using certificates that the host system can validate, make sure
	// that Git is verifying certificates
	stdout, stderr, err := engine.runC(cmtype, "-c", "http.sslVerify=false", "pull", upurl)
	log.Debug("Out: %s\nErr: %s", stdout, stderr)
	if err != nil {
		return NewCMCommunicationError(stdout, err)
	}

	return nil
}

func (engine *CMEngine) Log(cmtype CMType, path string, limit ...int) ([]LogLine, error) {
	format := "--pretty=format:%ct::%H::%s::%an::%ae::%cn::%ce"
	if path != "" {
		format = fmt.Sprintf("%s %s", format, path)
	}

	loglines := make([]LogLine, 0)

	engine.reset(cmtype)

	engine.guard.Lock()
	defer engine.guard.Unlock()

	var o string
	var err error
	if len(limit) == 0 || limit[0] == 0 {
		o, err = engine.run(cmtype, "log", format)
		if err != nil {
			return nil, err
		}
	} else {
		log.Debug("Called with limit >%v<", limit[0])
		o, err = engine.run(cmtype, "log", format,
			fmt.Sprintf("-%d", limit[0]))
		if err != nil {
			return nil, err
		}
	}

	lines := strings.Split(o, "\n")
	for _, line := range lines {
		vals := strings.Split(line, "::")

		t64, _ := strconv.ParseInt(vals[0], 10, 64)

		ll := LogLine{
			Time:           time.Unix(t64, 0),
			Id:             vals[1],
			Message:        vals[2],
			Author:         vals[3],
			AuthorEmail:    vals[4],
			Committer:      vals[5],
			CommitterEmail: vals[6],
		}

		loglines = append(loglines, ll)
	}

	return loglines, nil
}

func (engine *CMEngine) Diff(cmtype CMType, path string, commits ...string) (io.Reader, error) {
	opts := make([]string, 0)
	opts = append(opts, "diff")

	if len(commits) > 2 {
		return nil, NewCMError("To many revisions to diff")
	}

	for _, c := range commits {
		opts = append(opts, c)
	}

	if path != "" {
		opts = append(opts, path)
	}

	o, err := engine.run(cmtype, opts...)
	if err != nil {
		return nil, err
	}

	return bytes.NewBufferString(o), nil
}
