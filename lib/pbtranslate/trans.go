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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	change "github.com/iti/pbconf/lib/pbchange"
	config "github.com/iti/pbconf/lib/pbconfig"
	global "github.com/iti/pbconf/lib/pbglobal"
	driver "github.com/iti/pbconf/lib/pbtranslate/driver"

	"golang.org/x/net/context"
)

func configure(cfg *config.Config, deviceID int64, b io.Reader) ([]*driver.Command, error) {
	log.Debug("Configure()")
	execmds := make([]*driver.Command, 0)

	if cfg == (*config.Config)(nil) {
		cfg = global.CTX.Value("configuration").(*config.Config)
	}

	dev, err := driver.GetDevice(int64(deviceID))
	if err != nil {
		return nil, err
	}

	var cs *driver.CommandSeq
	var e error
	parsed_stmts, err := parseCfg(b)
	if err != nil {
		return nil, err
	}
	for _, op := range parsed_stmts {
		switch op.Op {
		case "service":
			cs, e = translateService(deviceID, dev.Name, op.Key, op.Val, cfg)
			if e != nil {
				continue
			}
			for _, cmd := range cs.Commands {
				execmds = append(execmds, cmd)
			}
		case "password":
			cs, e = translatePassword(deviceID, dev.Name, op.Key, op.Val, cfg)
			if e != nil {
				continue
			}
			for _, cmd := range cs.Commands {
				execmds = append(execmds, cmd)
			}
		case "variable":
			cs, e = translateVar(deviceID, dev.Name, op.Key, op.Val, cfg)
			if e != nil {
				continue
			}
			for _, cmd := range cs.Commands {
				execmds = append(execmds, cmd)
			}
		case "service_option":
			cs, e = translateSvcConfig(deviceID, dev.Name, op.Svc, op.Key, op.Val, cfg)
			if e != nil {
				continue
			}
			for _, cmd := range cs.Commands {
				execmds = append(execmds, cmd)
			}
		default:
			continue
		}
	}

	buf := bytes.Buffer{}
	for _, cmd := range execmds {
		buf.WriteString(fmt.Sprintf("%s\n", cmd.Command))
	}

	content := change.NewCMContent(dev.Name)
	content.Files["config.raw"] = buf.Bytes()

	cd := &change.ChangeData{
		ObjectType: change.DEVICE,
		Content:    content,
		Author: &change.CMAuthor{
			Name:  "Translation Engine",
			Email: "none@none.com",
			When:  time.Now(),
		},
	}

	cme, err := change.GetCMEngine(cfg)
	if err != nil {
		return nil, err
	}

	_, err = cme.VersionRaw(cd, "Translation Engine Raw Config Update")
	if err != nil {
		return nil, err
	}

	return execmds, nil

}
func parseCfg(b io.Reader) ([]op, error) {
	log.Debug("ConfigureParsedStmt()")

	parsed_stmts, err := Parse(b)
	if err != nil {
		// ABORT further code. Error parsing the input
		return nil, err
	}
	return parsed_stmts, nil
}

func GetMarshalledCfg(b io.Reader) ([]byte, error) {
	parsed_stmts, err := parseCfg(b)
	if err != nil {
		return nil, err
	}
	jsonstr, err := json.Marshal(parsed_stmts)
	return jsonstr, nil
}

func ExecuteConfig(cfg *config.Config, devID int64, b io.Reader) error {
	dev, err := driver.GetDevice(int64(devID))
	if err != nil {
		return err
	}
	execmds, err := configure(cfg, devID, b)
	if err != nil {
		return err
	}

	log.Debug("Available clients: %v", engineService.Clients)
	client, ok := engineService.Clients[getDriver(dev.Name, cfg)]
	if !ok {
		log.Error("Something wrong")
		return errors.New("no clients")
	}

	cmds := driver.CommandSeq{
		Devid: &driver.DeviceID{
			Id: int64(devID),
		},
	}
	cmds.Commands = execmds

	_, e := client.Client.ExecuteConfig(context.Background(), &cmds)
	if e != nil {
		log.Error(e.Error())
		return e
	}

	return nil
}

func translateService(id int64, dev, name, state string, cfg *config.Config) (*driver.CommandSeq, error) {

	client, ok := engineService.Clients[getDriver(dev, cfg)]
	if !ok {
		return nil, errors.New(
			fmt.Sprintf("No Driver for device ID: %s", dev))
	}

	var bstate bool
	switch state {
	case "on", "ON":
		bstate = true
	case "off", "OFF":
		bstate = false
	default:
		return nil, errors.New("Unknown service State")
	}

	rpcid := driver.DeviceID{
		Id: int64(id),
	}

	s := driver.Service{
		Devid: &rpcid,
		Name:  name,
		State: bstate,
	}

	r, err := client.Client.TranslateService(context.Background(), &s)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func translatePassword(id int64, dev, name, pass string, cfg *config.Config) (*driver.CommandSeq, error) {
	client, ok := engineService.Clients[getDriver(dev, cfg)]
	if !ok {
		return nil, errors.New(
			fmt.Sprintf("No Driver for device ID: %d", id))
	}

	rpcid := driver.DeviceID{
		Id: int64(id),
	}

	p := driver.UserPass{
		Devid:    &rpcid,
		Username: name,
		Password: pass,
	}

	r, err := client.Client.TranslatePass(context.Background(), &p)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func translateSvcConfig(id int64, dev, svc, variable, opt string, cfg *config.Config) (*driver.CommandSeq, error) {

	client, ok := engineService.Clients[getDriver(dev, cfg)]
	if !ok {
		return nil, errors.New(
			fmt.Sprintf("No Driver for device ID: %s", dev))
	}

	rpcid := driver.DeviceID{
		Id: int64(id),
	}

	v := driver.ServiceConfig{
		Devid: &rpcid,
		Name:  svc,
		Key:   variable,
		Value: opt,
	}

	r, err := client.Client.TranslateSvcConfig(context.Background(), &v)

	if err != nil {
		return nil, err
	}

	return r, nil
}

func translateVar(id int64, dev, key, val string, cfg *config.Config) (*driver.CommandSeq, error) {
	client, ok := engineService.Clients[getDriver(dev, cfg)]
	if !ok {
		return nil, errors.New(
			fmt.Sprintf("No Driver for device ID: %d", id))
	}

	rpcid := driver.DeviceID{
		Id: int64(id),
	}

	v := driver.Var{
		Devid: &rpcid,
		Key:   key,
		Value: val,
	}

	r, err := client.Client.TranslateVar(context.Background(), &v)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func getDriver(dev string, cfg *config.Config) string {
	cme, err := change.GetCMEngine(cfg)
	if err != nil {
		log.Error(err.Error())
		return ""
	}

	// Look up driver for device
	d, e := cme.GetMeta(dev, "driver")
	if e != nil {
		log.Error(e.Error())
		return ""
	}

	return d
}
