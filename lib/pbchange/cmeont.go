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
	"io/ioutil"
	"os"
	"path/filepath"
)

func (engine *CMEngine) saveOntology(ontology string) error {

	otype := ONTOLOGY

	engine.guard.Unlock()
	engine.reset(otype)
	engine.guard.Lock()

	fspath := filepath.Join(engine.Repopath, otype.String())
	ontfile := filepath.Join(fspath, "ontology")

	// If path is already a directory, MkdirAll does nothing and returns nil.
	if err := os.MkdirAll(fspath, os.ModePerm); err != nil {
		log.Debug("mkdir failed")
		return err
	}

	file, err := os.OpenFile(ontfile, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	// Line oriented ops
	rwf := bufio.NewReadWriter(bufio.NewReader(file), bufio.NewWriter(file))

	file.Truncate(0)
	rwf.WriteString(ontology)
	rwf.Flush()
	return nil
}

func (engine *CMEngine) loadOntology() (string, error) {
	otype := ONTOLOGY

	engine.guard.Unlock()
	engine.reset(otype)
	engine.guard.Lock()

	fspath := filepath.Join(engine.Repopath, otype.String())
	ontfile := filepath.Join(fspath, "ontology")

	// If path is already a directory, MkdirAll does nothing and returns nil.
	if err := os.MkdirAll(fspath, os.ModePerm); err != nil {
		log.Debug("mkdir failed")
		return "", err
	}

	// Can not just open the meta file.  Need to know if create was used
	var create bool
	if _, err := os.Stat(ontfile); os.IsNotExist(err) {
		create = true
	}

	file, err := os.OpenFile(ontfile, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return "", err
	}
	defer func() {
		file.Close()
		if create {
			engine.run(DEVICE, "add", fspath)
			engine.run(DEVICE, "commit", "-m", "updated metadata")
		}
	}()

	content, err := ioutil.ReadAll(file)
	if err != nil {
		return "", err
	}

	return string(content), nil
}
