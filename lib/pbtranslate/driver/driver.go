package driver

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
	global "github.com/iti/pbconf/lib/pbglobal"
	logging "github.com/iti/pbconf/lib/pblogger"

	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var cfgLogLevel string
var cfg *config.Config

type DriverService interface {
	DriverServer
	Name() string
	Client() EngineClient
	SetClient(EngineClient)
}

func Main(driver DriverService) {
	var cfgFile string

	flag.StringVar(&cfgFile, "c", "pbconf.conf",
		"PBCONF config file")
	flag.StringVar(&cfgLogLevel, "l", "",
		"Log level ([CRITICAL|ERROR|WARNING|NOTICE|INFO|DEBUG)")
	flag.Parse()

	var err error
	cfg, err = config.NewConfig(cfgFile)
	if err != nil {
		panic(err)
	}

	global.CTX = context.WithValue(global.CTX, "configuration", cfg)

	if cfgLogLevel == "" {
		cfgLogLevel = cfg.Translation.LogLevel
	}
	if cfgLogLevel == "" {
		cfgLogLevel = cfg.Global.LogLevel
	}

	logging.InitLogger(cfgLogLevel, cfg, "Driver")

	log := GetLogger(driver.Name())

	engine := filepath.Join(cfg.Translation.SocketDir, "engine.sock")

	log.Debug("Connecting to engine")
	client, err := grpc.Dial(engine, grpc.WithInsecure(), grpc.WithDialer(
		func(addr string, t time.Duration) (net.Conn, error) {
			return net.Dial("unix", addr)
		}),
	)
	if err != nil {
		panic(err)
	}

	driver.SetClient(NewEngineClient(client))

	socket := filepath.Join(cfg.Translation.SocketDir, fmt.Sprintf(
		"%s.sock", driver.Name()))

	log.Debug("setting up signal handler")

	// Also trap ctrl-c
	signalschan := make(chan os.Signal, 1)
	signal.Notify(signalschan, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGSTOP)
	go func() {
		<-signalschan
		os.Remove(socket)
		os.Exit(0)
	}()

	listener, err := net.Listen("unix", socket)
	if err != nil {
		panic(err)
	}
	defer func() {
		listener.Close()
		os.Remove(socket)
	}()

	service := grpc.NewServer()
	RegisterDriverServer(service, driver)

	req := RegRequest{
		Name:   driver.Name(),
		Socket: socket,
	}

	go func() {

		r, err := driver.Client().Register(context.Background(), &req)
		if err != nil {
			panic(err)
		}

		if r.Ok != true {
			panic(errors.New("Failed to register with engine"))
		}
	}()

	service.Serve(listener)

}

func GetLogger(name string) logging.Logger {
	logging.SetLevel(cfgLogLevel, fmt.Sprintf("Driver:%s", name))
	log, _ := logging.GetLogger(fmt.Sprintf("Driver:%s", name))
	return log
}
