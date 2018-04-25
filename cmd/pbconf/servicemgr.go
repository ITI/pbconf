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
	"flag"
	"fmt"
	"os/exec"
	"time"

	api "github.com/iti/pbconf/lib/pbapi"
	change "github.com/iti/pbconf/lib/pbchange"
	"github.com/iti/pbconf/lib/pbconfig"
	database "github.com/iti/pbconf/lib/pbdatabase"
	global "github.com/iti/pbconf/lib/pbglobal"
	logging "github.com/iti/pbconf/lib/pblogger"
	trans "github.com/iti/pbconf/lib/pbtranslate"
	webui "github.com/iti/pbconf/lib/pbwebui"

	"golang.org/x/net/context"
)

func main() {

	var cfgFile string
	var cfgLogLevel string

	flag.StringVar(&cfgFile, "c", "pbconf.conf",
		"PBCONF config file")
	flag.StringVar(&cfgLogLevel, "l", "",
		"Log level ([CRITICAL|ERROR|WARNING|NOTICE|INFO|DEBUG)")
	flag.Parse()

	cfg, err := pbconfig.NewConfig(cfgFile)
	if err != nil {
		fmt.Println("Loading config file " + cfgFile + ", got error: \n\t" + err.Error())
		return
	}
	if cfgLogLevel == "" {
		cfgLogLevel = cfg.Global.LogLevel
	}
	logging.InitLogger(cfgLogLevel, cfg, "")
	log, err := logging.GetLogger("Main")
	if err != nil {
		panic(err)
	}
	log.Info("Initializing System")

	log.Info("loading Config file")

	log.Info("Opening Database Connection")
	db := database.Open(cfg.Global.Database, cfg.WebAPI.LogLevel)
	defer db.Close()

	log.Info("Loading Change Management Engine")
	var cmEngine *change.CMEngine
	if cfg.SvcManager.ChangeEngine != "" && cfg.SvcManager.ChangeEngine != "internal" {
		changeEngine := exec.Command(cfg.SvcManager.ChangeEngine, "-c "+cfgFile)
		changeEngine.Start()
	} else {

		cmEngine, err = change.GetCMEngine(cfg)
		if err != nil {
			log.Error(err.Error())
		}
	}

	_ = cmEngine

	log.Info("Loading API manager")
	go api.StartAPIhandler(&cfg.WebAPI, db)

	go webui.StartWebUIhandler(&cfg.WebUI, &cfg.WebAPI, cfg.Global.AlarmDest, db)

	log.Info("Loading Policy Engine")
	if cfg.SvcManager.PolicyEngine != "" && cfg.SvcManager.PolicyEngine != "internal" {
		polEngine := exec.Command(cfg.SvcManager.PolicyEngine, "-c "+cfgFile)
		log.Info("Starting policy engine")
		perr := polEngine.Start()

		if perr != nil {
			panic(perr)
		}
	} else {
		log.Info("Using Internal pol engine")
	}

	log.Info("Loading Secure Connection Broker")
	var sshBroker service
	if cfg.SvcManager.SSHBroker != "" && cfg.SvcManager.SSHBroker != "internal" {
		p := exec.Command(cfg.SvcManager.SSHBroker, "-c", cfgFile, "-l", cfgLogLevel)
		sshBroker = service{"SSH", p, false, nil}
		go sshBroker.Start()
	}

	log.Info("Loading Translation Engine")
	if cfg.SvcManager.TranslationEngine != "" && cfg.SvcManager.TranslationEngine != "internal" {
		transEngine := exec.Command(cfg.SvcManager.TranslationEngine, "-c "+cfgFile)
		transEngine.Start()
	} else {
		_, terr := trans.Start(cfg)
		if terr != nil {
			log.Error(terr.Error())
		}
	}

	// Translation modules are always external
	log.Info("Loading translation Modules")
	if err := loadModules(cfg); err != nil {
		log.Error(err.Error())
	}

	log.Info("Starting API Handler")
	global.Start(cfg.Global.NodeName, &cfg.WebAPI)

	// Add the config to the global context
	global.CTX = context.WithValue(global.CTX, "configuration", cfg)
	db.LoadSchema()
	db.LoadRootNode(global.RootNode)

	hbDoneChan := make(chan bool)
	heartbeatComm := NewHeartbeatComm(db, cfg.WebAPI.LogLevel)
	heartbeatComm.Start(hbDoneChan) //signaling true on the hbDoneChan will stop the heartbeat

	log.Info("Loading Translation Modules")

	for {
		for _, proc := range []service{sshBroker} {
			done := proc.Exited()
			if done {
				log.Debug(proc.Name + " Exited")
				proc.Restart()
			}
		}
		time.Sleep(1000 * time.Millisecond)
	}
}
