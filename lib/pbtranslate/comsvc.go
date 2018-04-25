package pbtranslate

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
	cme "github.com/iti/pbconf/lib/pbchange"
	config "github.com/iti/pbconf/lib/pbconfig"
	logging "github.com/iti/pbconf/lib/pblogger"
	driver "github.com/iti/pbconf/lib/pbtranslate/driver"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"errors"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

var engineService *EngineService

func init() {
	engineService = &EngineService{
		Clients: make(map[string]*PBDriverClient, 0),
	}
}

type PBDriverClient struct {
	Client     driver.DriverClient
	Connection *grpc.ClientConn
}

type EngineService struct {
	Clients map[string]*PBDriverClient
	mx      sync.Mutex
	CME     *cme.CMEngine
}

func (s *EngineService) dialer(addr string, t time.Duration) (net.Conn, error) {

	_, err := os.Stat(addr)
	if err == nil {
		return net.Dial("unix", addr)
	}
	dname := strings.Split(filepath.Base(addr), ".")[0]
	if _, ok := s.Clients[dname]; !ok {
		// initial connection
		return nil, err
	}

	s.Clients[dname].Connection.Close()
	s.mx.Lock()
	delete(s.Clients, dname)
	s.mx.Unlock()
	return nil, err
}

func (s *EngineService) Register(ctx context.Context, req *driver.RegRequest) (*driver.BoolReply, error) {

	log.Debug("Registering %s at %s", req.Name, req.Socket)

	s.mx.Lock()
	if _, ok := s.Clients[req.Name]; ok {
		s.mx.Unlock()
		return &driver.BoolReply{Ok: false}, errors.New("Driver Already Registered")
	}
	s.mx.Unlock()

	client, err := grpc.Dial(req.Socket, grpc.WithInsecure(),
		grpc.WithDialer(s.dialer))
	if err != nil {
		return &driver.BoolReply{Ok: false}, err
	}

	c := driver.NewDriverClient(client)
	s.mx.Lock()
	s.Clients[req.Name] = &PBDriverClient{Client: c, Connection: client}
	s.mx.Unlock()

	return &driver.BoolReply{Ok: true}, nil
}

func (s *EngineService) GetMeta(ctx context.Context, req *driver.KVRequest) (*driver.KVPair, error) {
	log.Debug("GetMeta()")

	log.Debug("Lookup up device %d", req.Devid.Id)
	dev, err := driver.GetDevice(req.Devid.Id)
	log.Debug("Got Device %d", req.Devid.Id)
	if err != nil {
		return nil, err
	}

	log.Debug("Getting metatdata from CME: %s", req.Key)
	val, err := s.CME.GetMeta(dev.Name, req.Key)
	if err != nil {
		return nil, err
	}

	log.Debug("Returning metadata")
	return &driver.KVPair{
		Devid: &driver.DeviceID{Id: req.Devid.Id},
		Key:   req.Key,
		Value: val,
	}, nil
}

func (s *EngineService) SaveMeta(ctx context.Context, req *driver.KVPair) (*driver.BoolReply, error) {

	dev, err := driver.GetDevice(req.Devid.Id)
	if err != nil {
		return driver.ReplyFalse(err)
	}

	err = s.CME.VersionMeta(dev.Name, req.Key, req.Value)
	if err != nil {
		return driver.ReplyFalse(err)
	}

	return driver.ReplyTrue(nil)

}

var log logging.Logger

func Start(cfg *config.Config) (chan bool, error) {
	c := make(chan bool)

	ll := cfg.Translation.LogLevel
	if ll == "" {
		ll = cfg.Global.LogLevel
	}

	logging.SetLevel(ll, "Translation")
	log, _ = logging.GetLogger("Translation")
	_ = log

	socket := filepath.Join(cfg.Translation.SocketDir, "engine.sock")
	if _, ok := os.Stat(socket); os.IsExist(ok) {
		return nil, errors.New("Can Not Start, socket already exists")
	}

	eng, err := cme.GetCMEngine(cfg)
	if err != nil {
		return nil, err
	}

	engineService.CME = eng

	signalschan := make(chan os.Signal, 1)
	signal.Notify(signalschan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signalschan
		os.Remove(socket)
		os.Exit(0)
	}()

	l, e := net.Listen("unix", socket)
	if e != nil {
		return nil, e
	}

	s := grpc.NewServer()
	driver.RegisterEngineServer(s, engineService)

	go func() {
		s.Serve(l)

		l.Close()
		os.Remove(socket)
	}()

	return c, nil
}
