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
	config "github.com/iti/pbconf/lib/pbconfig"
	logging "github.com/iti/pbconf/lib/pblogger"

	"crypto"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"

	_ "crypto/md5"
	_ "crypto/sha1"
	_ "crypto/sha256"
	_ "crypto/sha512"
	_ "golang.org/x/crypto/sha3"
)

type driver_typ struct {
	Name string
	Proc *exec.Cmd
	Done chan bool
}

func loadModules(cfg *config.Config) error {
	log, _ := logging.GetLogger("Driver:Loader")
	logging.SetLevel(cfg.Global.LogLevel, "Driver:Loader")
	var drivers []*driver_typ

	for _, module := range cfg.Translation.TransModules {
		// Check if module exists
		driverpath := path.Join(cfg.Translation.ModuleDir, module)
		opts, ok := cfg.DriverOpts[module]
		if ok {
			if opts.Path != "" {
				driverpath = opts.Path
			}
		}
		if _, err := os.Stat(driverpath); err != nil {
			log.Error(fmt.Sprintf("Driver %s does not exist or is not readable", driverpath))
			continue
		}

		// See if we can load it
		if !checkModSig(cfg.Translation.ForceVerify, driverpath, opts) {
			continue
		}

		proc := &driver_typ{
			Name: module,
			Proc: exec.Command(driverpath, "-c", cfg.Path()),
			Done: make(chan bool, 1),
		}

		drivers = append(drivers, proc)

		go func(drv *driver_typ) {
			log.Debug("Starting %s: %s\n", drv.Name, drv.Proc.Path)
			for {
				drv.Proc.Start()
				drv.Proc.Wait()
				log.Info("%s stopped\n", drv.Name)
				select {
				case <-drv.Done:
					log.Info("Exiting %s\n", drv.Name)
					return
				default:
					log.Info("restarting %s: %s\n", drv.Name, drv.Proc.Path)
					drv.Proc = exec.Command(drv.Proc.Path)
				}
			}
		}(proc)
	}
	return nil
}

func checkModSig(force bool, driverpath string, opts *config.CfgDriverOpts) bool {

	switch {
	case !force && opts == nil:
		return true
	case force && opts == nil:
		log.Warning("Force Verify is on, but no signature defined for %s.  Not loading", driverpath)
		return false
	// Hash not defined and do not force
	case opts.Hash == "" && !force:
		return true
	// Hash not defined and foce is true
	case opts.Hash == "" && force:
		log.Warning("Force Verify is on, but no signature defined for %s.  Not loading", driverpath)
		return false
	}

	// Hash is defined
	var hashFN crypto.Hash
	switch strings.ToUpper(opts.HashType) {
	case "MD5":
		hashFN = crypto.MD5
	case "SHA1":
		hashFN = crypto.SHA1
	case "SHA256":
		hashFN = crypto.SHA256
	case "SHA512":
		hashFN = crypto.SHA512
	case "SHA3_256":
		hashFN = crypto.SHA3_256
	case "SHA3_512":
		hashFN = crypto.SHA3_512
	case "SHA512_256":
		hashFN = crypto.SHA512_256
	default:
		log.Warning("Unknown hash function %s.  Not loading %s", opts.Hash, driverpath)
		return false
	}

	if !hashFN.Available() {
		log.Warning("Hash function %s not available.  Not loading %s", opts.Hash, driverpath)
		return false
	}

	fn := hashFN.New()

	file, err := os.Open(driverpath)
	if err != nil {
		log.Warning("Failed to verify %s: %s", driverpath, err.Error())
		return false
	}
	defer file.Close()

	_, err = io.Copy(fn, file)
	if err != nil {
		log.Warning("Failed to verify %s: %s", driverpath, err.Error())
		return false
	}

	log.Debug("s: %x h: %s", string(fn.Sum(nil)), opts.Hash)
	return fmt.Sprintf("%x", fn.Sum(nil)) == opts.Hash
}
