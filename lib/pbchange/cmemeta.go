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
	"bufio"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func (engine *CMEngine) saveMeta(oname string, metadata map[string]string) error {

	// Metatdata is always going to be a device
	otype := DEVICE

	engine.guard.Unlock()
	engine.reset(otype)
	engine.guard.Lock()

	metapath := filepath.Join(engine.Repopath, otype.String(), oname, "meta/")
	metafile := filepath.Join(metapath, "meta.db")

	// If path is already a directory, MkdirAll does nothing and returns nil.
	if err := os.MkdirAll(metapath, os.ModePerm); err != nil {
		log.Debug("mkdir failed")
		return err
	}

	file, err := os.OpenFile(metafile, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	// Line oriented ops
	rwf := bufio.NewReadWriter(bufio.NewReader(file), bufio.NewWriter(file))

	out, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	file.Truncate(0)
	rwf.WriteString(string(out))
	rwf.Flush()
	return nil
}

// Wraps loadMeta to handle locking
func (engine *CMEngine) LoadMeta(oname string) (map[string]string, error) {
	engine.guard.Lock()
	defer engine.guard.Unlock()

	return engine.loadMeta(oname)
}

func (engine *CMEngine) loadMeta(oname string) (map[string]string, error) {
	// Metatdata is always going to be a device
	otype := DEVICE

	// Reset locks, so unlock, then relock
	engine.guard.Unlock()
	engine.reset(otype)
	engine.guard.Lock()

	metapath := filepath.Join(engine.Repopath, otype.String(), oname, "meta")
	metafile := filepath.Join(metapath, "meta.db")

	// If path is already a directory, MkdirAll does nothing and returns nil.
	if err := os.MkdirAll(metapath, os.ModePerm); err != nil {
		log.Debug("mkdir failed")
		return nil, err
	}

	// Can not just open the meta file.  Need to know if create was used
	var create bool
	if _, err := os.Stat(metafile); os.IsNotExist(err) {
		create = true
	}

	file, err := os.OpenFile(metafile, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return nil, err
	}
	defer func() {
		file.Close()
		if create {
			engine.run(DEVICE, "add", metapath)
			engine.run(DEVICE, "commit", "-m", "updated metadata")
		}
	}()

	// Line oriented ops
	rwf := bufio.NewReadWriter(bufio.NewReader(file), bufio.NewWriter(file))

	// Read in existing metadata KV pairs
	in, err := rwf.ReadBytes('\n')
	if err != nil && err != io.EOF {
		// A read error occured
		return nil, err
	}

	var metadata map[string]string
	if len(in) > 0 {
		// Read data
		err := json.Unmarshal(in, &metadata)
		if err != nil {
			return nil, err
		}
	} else {
		metadata = make(map[string]string)
	}

	return metadata, nil
}

func (engine *CMEngine) GetMeta(oname string, key string) (string, error) {
	engine.guard.Lock()
	defer engine.guard.Unlock()

	if _, err := engine.run(DEVICE, "checkout", "master"); err != nil {
		return "", err
	}

	metadata, err := engine.loadMeta(oname)
	if err != nil {
		return "", err
	}

	if val, ok := metadata[key]; ok {
		return val, nil
	}

	return "", NewCMMetaNoKeyError(key)
}
func (engine *CMEngine) DeleteMeta(oname, key string) error {
	if err := engine.MakeRepo(DEVICE); err != nil {
		return err
	}
	metapath := filepath.Join(engine.Repopath, DEVICE.String(), oname,
		"meta/meta.db")

	engine.guard.Lock()
	defer engine.guard.Unlock()
	if _, err := engine.run(DEVICE, "checkout", "master"); err != nil {
		return err
	}

	metadata, err := engine.loadMeta(oname)
	if err != nil {
		log.Debug("Failed to load meta file")
		return err
	}
	delete(metadata, key)

	err = engine.saveMeta(oname, metadata)
	if err != nil {
		log.Debug("Error %s saving metafile, resetting repo")
		engine.reset(DEVICE)
		return err
	}

	if _, err := engine.run(DEVICE, "add", metapath); err != nil {
		return err
	}
	if o, err := engine.run(DEVICE, "commit", "-m", "updated metadata"); err != nil {
		if !strings.Contains(o, "nothing to commit, working directory clean") {
			log.Debug("Commit Failed:\n%s", o)
			return err
		}
	}
	return nil
}

func (engine *CMEngine) VersionMeta(oname, key, val string) error {
	log.Debug("Versioning metadata: %s:%s", key, val)

	log.Debug("Ensuring repo exists")
	if err := engine.MakeRepo(DEVICE); err != nil {
		return err
	}
	log.Debug("Repo exists")

	metapath := filepath.Join(engine.Repopath, DEVICE.String(), oname,
		"meta/meta.db")

	log.Debug("Metapath: %s", metapath)

	engine.guard.Lock()
	defer engine.guard.Unlock()

	log.Debug("checking out master")
	if _, err := engine.run(DEVICE, "checkout", "master"); err != nil {
		return err
	}
	log.Debug("done: checking out master")

	metadata, err := engine.loadMeta(oname)
	if err != nil {
		log.Debug("Failed to load meta file")
		return err
	}
	log.Debug("Done: loading metafile: %s", metadata)

	metadata[key] = val
	log.Debug("Saving metadata: %s", metadata)
	err = engine.saveMeta(oname, metadata)
	if err != nil {
		log.Debug("Error %s saving metafile, resetting repo")
		engine.reset(DEVICE)
		log.Debug("Done resetting")
		return err
	}

	log.Debug("Adding metafile")
	if _, err := engine.run(DEVICE, "add", metapath); err != nil {
		return err
	}
	log.Debug("done: Adding metafile - Committing")
	if o, err := engine.run(DEVICE, "commit", "-m", "updated metadata"); err != nil {
		if !strings.Contains(o, "nothing to commit, working directory clean") {
			log.Debug("Commit Failed:\n%s", o)
			return err
		}
	}
	log.Debug("done: Committing")

	return nil
}
