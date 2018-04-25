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
	logging "github.com/iti/pbconf/lib/pblogger"
	driver "github.com/iti/pbconf/lib/pbtranslate/driver"
	trans "github.com/iti/pbconf/lib/pbtransport"
	expect "github.com/jamesharr/expect"
	context "golang.org/x/net/context"

	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
)

var passcache = make(map[string]string)

// driver provides a method to get a logging object
var log logging.Logger

// Driver Service hold any driver specific state information, and must
// implement the DriverService interface.
type driverService struct {
	name   string
	client driver.EngineClient
}

/*
Name()
	Returns the drivers official name
*/
func (d *driverService) Name() string {
	return d.name
}

/*
Client()
	Returns the Engine Client object
*/
func (d *driverService) Client() driver.EngineClient {
	return d.client
}

/*
SetClient()
	Set the Engine Client object.  This should be stored internally
*/
func (d *driverService) SetClient(c driver.EngineClient) {
	d.client = c
}

/*
GetConfig()
	Returns all config files as they exist on the device
*/
func (d *driverService) GetConfig(ctx context.Context, id *driver.DeviceID) (*driver.ConfigFiles, error) {
	log.Debug("GetConfig()")

	return nil, nil
}

/*
TranslatePassword()
	Called when a password set is needed

	The meaning of User Name and password is driver specific.  For instance, on
		an SEL421, Username would be "2", for level 2 password, but on a Linux
		computer, this will be an actual user name
*/
func (d *driverService) TranslatePass(ctx context.Context, pass *driver.UserPass) (*driver.CommandSeq, error) {
	log.Debug("TranslatePass()")
	cmdrep := driver.Command{
		Command: fmt.Sprintf("PAS %s %s", pass.Username, pass.Password),
	}

	rep := driver.CommandSeq{
		Devid:    pass.Devid,
		Commands: []*driver.Command{&cmdrep},
	}

	return &rep, nil

}

/*
TranslateService()
	Called to en/dis- able a service on the device

	Should return the commands necessary to immediatly shut the servie off,
	as well as the commands necessary to perminatly disable the service
	such that the service will start up in the given state when the device
	starts
*/
func (d *driverService) TranslateService(ctx context.Context, svc *driver.Service) (*driver.CommandSeq, error) {
	log.Debug("TranslateService()")

	var cmd string
	if svc.State {
		cmd = fmt.Sprintf("\"%s\",\"Y\"", svc.Name)
	} else {
		cmd = fmt.Sprintf("\"%s\",\"N\"", svc.Name)
	}

	rep := driver.CommandSeq{
		Devid: svc.Devid,
		Commands: []*driver.Command{&driver.Command{
			Command: cmd,
		}},
	}

	return &rep, nil
}

/*
TranslateVar()
	Called to set arbitrary configuration parameters

	The meaning of Key and Value are driver specific.  TranslateVar() is a
	catch-all for configuration parameters for which there is not a specific
	required interface.  The driver should "do the right thing" for the device
	in question.  Note, if the setting changes the operating state of the
	device, the setting should be changed such that the the device enters
	the new state immediatly, _and_ that the device will enter the new
	state when the device starts up.
*/
func (d *driverService) TranslateVar(ctx context.Context, variable *driver.Var) (*driver.CommandSeq, error) {
	log.Debug("TranslateVar()")

	// This driver doesn't support variables, but can't return nil
	return &driver.CommandSeq{
		Devid:    variable.Devid,
		Commands: []*driver.Command{},
	}, nil
}

/*
TrasnlateSvcConfig()
	Called to set service configuration options

	This methos is also very device specific.  It would return the commands
	necessary to configure service specific options.  For instance, on an
	SEL421 this sould generate updates to the port 5 config.  However, since
	the 421 does not a a "generic way" of setting service config options,
	this driver must know about all of the specific settable options.

	Initiall only FTP is supported
*/
func (d *driverService) TranslateSvcConfig(ctx context.Context, opt *driver.ServiceConfig) (*driver.CommandSeq, error) {

	var cmd string

	switch opt.Name {
	case "FTP", "ftp":
		cmd = d.transFTP(opt.Key, opt.Value)
	}

	return &driver.CommandSeq{
		Devid: opt.Devid,
		Commands: []*driver.Command{
			&driver.Command{Command: cmd},
		},
	}, nil
}

func (d *driverService) transFTP(k, v string) string {
	switch k {
	case "anonFTP", "anonymousftp", "FTPANMS":
		return fmt.Sprintf("FTPANMS=%s", v)
	case "FTPCBAN", "banner":
		return fmt.Sprintf("FTPCBAN=%s", v)
	case "FTPIDLE", "idletimeout":
		return fmt.Sprintf("FTPIDLE=%s", v)
	case "FTPAUSER", "userlevel":
		return fmt.Sprintf("FTPAUSER=%s", v)
	}

	return ""
}

func (d *driverService) ExecuteConfig(ctx context.Context, commands *driver.CommandSeq) (*driver.BoolReply, error) {
	log.Debug("ExecuteConfig()")

	transport, err := driver.ConnectToDevice(d.authFn, commands.Devid.Id, d.Name())
	if err != nil {
		log.Info("Failed to connect: %s", err.Error())
		return driver.ReplyFalse(err)
	}

	// This driver only really supports telnet/serial/ftp transports
	passtrans := transport
	configtrans := transport
	var filePrefix string
	var e error

	_ = configtrans
	_ = passtrans

	if _, ok := transport.(*trans.FTP); ok {
		// Get alt transport and set passtrans
		passtrans, e = d.altTransport(commands.Devid.Id)
		if e != nil {
			return driver.ReplyFalse(err)
		}

		filePrefix = "/SEL-421-1/"
	}

	if _, ok := transport.(*trans.Telnet); ok {
		// get alt transport and set configtrans
		configtrans, e = d.altTransport(commands.Devid.Id)
		if e != nil {
			return driver.ReplyFalse(err)
		}
	}

	cfgfile, err := configtrans.RecvFile(fmt.Sprintf("%sSETTINGS/SET_P5.TXT", filePrefix))
	if err != nil {
		if err != io.EOF {
			return driver.ReplyFalse(err)
		}
	}

	cfgScanner := bufio.NewScanner(bytes.NewBuffer(cfgfile))
	cfgLines := make([]string, 0)
	for cfgScanner.Scan() {
		cfgLines = append(cfgLines, cfgScanner.Text())
	}
	if err := cfgScanner.Err(); err != nil {
		return driver.ReplyFalse(err)
	}

	var wg sync.WaitGroup
	var mux sync.Mutex
	for _, cmd := range commands.Commands {
		wg.Add(1)
		go func(cmd *driver.Command) {
			defer wg.Done()

			if strings.HasPrefix(cmd.Command, "PAS ") {
				log.Debug("PAS command")
				if err := d.isRoot(commands.Devid.Id, cmd.Command); err != nil {
					return
				}

				log.Debug("isRoot returned")

				// Set password
				//re auth
				// isRoot() should have set the current passwords in cache
				var l1, l2 string
				var ok bool

				if l1, ok = passcache["1"]; !ok {
					log.Debug("Password 1 missing: %s", passcache)
					d.resetMeta(commands.Devid.Id, err)
					return
				}
				if l2, ok = passcache["2"]; !ok {
					log.Debug("Password 2 missing: %s", passcache)
					d.resetMeta(commands.Devid.Id, err)
					return
				}

				log.Debug("Auth for password change: %T", passtrans)
				exp := expect.Create(passtrans, func() {})
				exp.SetLogger(expect.FileLogger("/tmp/pblog.log"))
				exp.SetTimeout(5 * time.Second)

				atL1 := false
				atL2 := false

				for atL1 == false || atL2 == false {
					m, err := exp.Expect("=>>|=>|=")
					if err != nil {
						log.Error("1 %s", err.Error())
						return
					}
					log.Debug("Got groups: %v", m.Groups)
					switch m.Groups[0] {
					case "=":
						log.Debug("sending")
						defer func() {
							r := recover()
							if r != nil {
								log.Error("PANIC!!!: %v", r)
							}
						}()

						if err := exp.Send("acc\r\n"); err != nil {
							log.Error("2 %s", err.Error())
							return
						}
						log.Debug("Sent")
						if _, err := exp.Expect("Password: ?"); err != nil {
							log.Error("3 %s", err.Error())
							return
						}
						log.Debug("Sending L1 password")
						if err := exp.Send(fmt.Sprintf("%s\r\n", l1)); err != nil {
							log.Error("4 %s", err.Error())
							return
						}
						atL1 = true
					case "=>":
						atL1 = true
						if err := exp.Send("2ac\r\n"); err != nil {
							log.Error("5 %s", err.Error())
							return
						}
						if _, err := exp.Expect("Password: ?"); err != nil {
							log.Error("6 %s", err.Error())
							return
						}
						if err := exp.Send(fmt.Sprintf("%s\r\n", l2)); err != nil {
							log.Error("7 %s", err.Error())
							return
						}
						atL2 = true
					case "=>>":
						atL1 = true
						atL2 = true
					}
				}

				// We should be at L2
				if _, err := exp.Expect("=>>"); err != nil {
					log.Error("6 %s", err.Error())
					return
				}

				if err := exp.Send(fmt.Sprintf("%s\r\n", cmd.Command)); err != nil {
					log.Error("8 %s", err.Error())
					return
				}

				passcache["1"] = ""
				passcache["2"] = ""

				return
			}

			log.Debug("Not password command: %s", cmd.Command)
			for i, l := range cfgLines {
				if strings.HasPrefix(l, strings.Split(cmd.Command, ",")[0]) {
					mux.Lock()
					cfgLines[i] = fmt.Sprintf(cmd.Command)
					mux.Unlock()
					return
				}
			}
		}(cmd)
	}

	wg.Wait()

	obuf := make([]byte, 0)
	for _, line := range cfgLines {
		obuf = append(obuf, fmt.Sprintf("%s\r\n", line)...)
	}
	_ = obuf

	if err := configtrans.SendFile(
		fmt.Sprintf("%sSETTINGS/SET_P5.TXT", filePrefix), obuf); err != nil {
		log.Error(err.Error())
		return driver.ReplyFalse(err)
	}

	return driver.ReplyTrue(nil)
}

func (d *driverService) altTransport(id int64) (trans.ClientTransport, error) {
	var transport trans.ClientTransport

	n, err := d.Client().GetMeta(context.Background(), &driver.KVRequest{
		Devid: &driver.DeviceID{Id: id},
		Key:   "alttransport",
	})
	if err != nil {
		return nil, err
	}

	l, err := d.Client().GetMeta(context.Background(), &driver.KVRequest{
		Devid: &driver.DeviceID{Id: id},
		Key:   "altlocation",
	})
	if err != nil {
		return nil, err
	}

	transport, err = trans.GetTransport(n.Value, d.Name())
	if err != nil {
		return nil, err
	}

	err = transport.Dial(id, l.Value)
	if err != nil {
		return nil, err
	}

	return transport, nil
}

func (d *driverService) authFn(id int64) (username, password string, err error) {
	return d.authget(id, "2")
}

func (d *driverService) auth2Fn(id int64) (username, password string, err error) {
	return d.authget(id, "1")
}

func (d *driverService) authget(id int64, level string) (username, password string, err error) {

	var u *driver.KVPair

	switch level {
	case "1":
		username = "1AC"
	case "2":
		username = "2AC"
	default:
		err = errors.New("Unknown username")
		return
	}

	u, err = d.Client().GetMeta(context.Background(), &driver.KVRequest{
		Devid: &driver.DeviceID{Id: id},
		Key:   fmt.Sprintf("l%spassword", level),
	})
	if err != nil {
		return
	}

	password = u.Value
	return

}

func (d *driverService) isRoot(id int64, cmd string) (err error) {
	if !strings.Contains(cmd, "PAS") {
		return
	}

	var oldpass string

	// PAS <level> <password>
	passcmd := strings.Split(cmd, " ")

	if _, oldpass, err = d.authget(id, "1"); err != nil {
		return
	}
	passcache["1"] = oldpass

	if _, oldpass, err = d.authget(id, "2"); err != nil {
		return
	}
	passcache["2"] = oldpass

	switch passcmd[1] {
	case "1":
		if err = d.updateMeta(id, "1", passcmd[2]); err != nil {
			return
		}
	case "2":
		if err = d.updateMeta(id, "2", passcmd[2]); err != nil {
			return
		}
	default:
		return
	}

	return
}

func (d *driverService) updateMeta(id int64, level, newpass string) error {

	kv := &driver.KVPair{
		Devid: &driver.DeviceID{
			Id: id,
		},
		Key:   fmt.Sprintf("l%spassword", level),
		Value: newpass,
	}

	r, e := d.Client().SaveMeta(context.Background(), kv)
	if r.Ok != true {
		return e
	}

	return nil
}

func (d *driverService) resetMeta(id int64, err error) (*driver.BoolReply, error) {

	for _, level := range []string{"1", "2"} {
		if passcache[level] != "" {
			d.updateMeta(id, level, passcache[level])
		}
		passcache[level] = ""
	}

	return driver.ReplyFalse(err)
}

/*
Service configuration and start up are handled by the framework.  There is
	no need to do anything beyond instantiate an instance of the driver
	service object, and pass it to driver.Main()
*/
func main() {
	devdriver := driverService{name: "sel421"}
	log = driver.GetLogger(devdriver.Name())
	driver.Main(&devdriver)
}
