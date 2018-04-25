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
	"io/ioutil"
	"os"

	ssh "golang.org/x/crypto/ssh"

	pbconfig "github.com/iti/pbconf/lib/pbconfig"
	logging "github.com/iti/pbconf/lib/pblogger"

	database "github.com/iti/pbconf/lib/pbdatabase"
)

// Need this elsewhere
var cfg *pbconfig.Config
var log logging.Logger

func main() {

	var cfgFile string
	var cfgLogLevel string

	flag.StringVar(&cfgFile, "c", "pbconf.conf",
		"PBCONF config file")
	flag.StringVar(&cfgLogLevel, "l", "",
		"Log level ([CRITICAL|ERROR|WARNING|NOTICE|INFO|DEBUG)")
	flag.Parse()

	cfg, _ = pbconfig.NewConfig(cfgFile)

	if cfgLogLevel == "" {
		cfgLogLevel = cfg.Broker.LogLevel
	}

	if cfgLogLevel == "" {
		cfgLogLevel = cfg.Global.LogLevel
	}

	logging.InitLogger(cfgLogLevel, cfg, "")
	var err error
	log, err = logging.GetLogger("Broker")
	if err != nil {
		panic(err)
	}

	db := database.Open(cfg.Global.Database, cfgLogLevel)
	defer db.Close()

	sshconfig := &ssh.ServerConfig{
		NoClientAuth: true,
		AuthLogCallback: func(conn ssh.ConnMetadata, method string, err error) {
			log.Info(conn.User() + " logged in")
		},
		ServerVersion: "SSH-2.0-PBCONF_1.0.0",
	}

	keyfile, err := os.Open(cfg.Broker.PrivKey)
	if err != nil {
		log.Error("Failed to open ssh key file")
		panic(err)
	}
	defer keyfile.Close()

	privkey, err := ioutil.ReadAll(keyfile)
	if err != nil {
		log.Error("Failed to read keyfile")
		panic(err)
	}

	log.Debug("Got Priv: " + cfg.Broker.PrivKey)

	pkey, err := ssh.ParsePrivateKey(privkey)
	if err != nil {
		panic(err.Error())
	}
	sshconfig.AddHostKey(pkey)

	ssh_server(sshconfig, db)

}
