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

/*

This is an example in Linux password driver.  It demonstrates how to implement
	a device specific driver in Go, using framework components from the
	translation engine.

Note:  Drivers do not NEED to be written in Go.  Any stand alone executable (
	in any language) will work.  That executable needs to run in the
	forground (so the translation engine can manage the service) and connect
	to the UNIX domain socket for the translation engine (see config file).
	The driver must expect  Protocol Buffers[1] on the client connection.  The
	translation engine will "execute" methods via the protobuf interface of the
	incomming client connection.

	If written in Go, use of the provided driver framework is recommended.  The
	framework provides most of the boilerplate necessary to connect to and
	use the translation engine services.  If using the framework, a
	DriverService object must be passed to the framework Main function (
	driver.Main()).  The expanded interface defination:

	type DriverService interface {
		GetConfig(context.Context, *DeviceID) (*ConfigFiles, error)
		TranslatePass(context.Context, *UserPass) (*CommandSeq, error)
		TranslateService(context.Context, *Service) (*CommandSeq, error)
		TranslateVar(context.Context, *Var) (*CommandSeq, error)
		TranslateSvcConfig(context.Context, *ServiceConfig) (*CommandSeq, error)
		ExecuteConfig(context.Context, *CommandSeq) (*BoolReply, error)
		Name() string
		Client() EngineClient
		SetClient(EngineClient)
	}

	DriverService incorporates the DriverServer interface which is defined
	by the Protocol Buffers implementation.

	If you choose another language, you must still implement a Protocol
	Buffers version 3 service running on top of gRPC to provide to the
	translation engine.  See the protobuf service definations in
	pbtranslate/gen/driver.proto for details.  Also review package
	iti/pbconf/pbtranslate/driver for Go implementation details.

}
*/
package main

import (
	"fmt"
	"strings"

	context "golang.org/x/net/context"

	logging "github.com/iti/pbconf/lib/pblogger"
	"github.com/iti/pbconf/lib/pbtranslate/driver"
)

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
		Command: fmt.Sprintf("echo \"%s:%s\"|chpasswd", pass.Username, pass.Password),
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
		cmd = fmt.Sprintf("rm -f /etc/init/%s.override;service %s start", svc.Name, svc.Name)
	} else {
		cmd = fmt.Sprintf("service %s stop; echo \"manual\" > /etc/init/%s.override", svc.Name, svc.Name)
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
	necessary to configure service specific options.  For instance, this
	driver assumes that any option set should by a var=val pair that would be
	inserted into /etc/default/<service name>
*/
func (d *driverService) TranslateSvcConfig(ctx context.Context, opt *driver.ServiceConfig) (*driver.CommandSeq, error) {

	cmd1 := fmt.Sprintf("sed -i -e 's/^%s[ ]*=.*//g' /etc/default/%s",
		opt.Key, opt.Name)
	cmd2 := fmt.Sprintf("echo %s=%s >> /etc/default/%s", opt.Key, opt.Value,
		opt.Name)

	return &driver.CommandSeq{
		Devid: opt.Devid,
		Commands: []*driver.Command{
			&driver.Command{Command: cmd1},
			&driver.Command{Command: cmd2},
		},
	}, nil
}

/*
EcexuteConfig()
	Called by the translation engine to apply a series of commands to a
	device
*/
func (d *driverService) ExecuteConfig(ctx context.Context, commands *driver.CommandSeq) (*driver.BoolReply, error) {
	log.Debug("ExecuteConfig()")

	transport, err := driver.ConnectToDevice(d.authFn, commands.Devid.Id, d.Name())
	log.Debug("HERE")
	if err != nil {
		log.Info("Failed to connect: %s", err.Error())
		return driver.ReplyFalse(err)
	}
	log.Debug("Got transport: %v", transport)

	for _, cmd := range commands.Commands {
		// Update password in meta first so we don't end up broken
		if err := d.isRoot(commands.Devid.Id, cmd.Command); err != nil {
			return driver.ReplyFalse(err)
		}

		log.Debug("Doing command: %v", cmd.Command)

		// Need to be able to check output from service start
		buf := append([]byte(cmd.Command), make([]byte, 30)...)

		_, err := transport.Read(buf)
		if err != nil {
			// check that service was already running
			log.Debug("returned output: %s", string(buf))
			if !strings.Contains(string(buf), "Job is already running") {
				log.Debug("Something Failed")
				log.Info("Command <<%s>> Failed: %s", cmd.Command, err.Error())
			}
		}
	}

	return driver.ReplyTrue(nil)
}

func (d *driverService) isRoot(id int64, cmd string) error {
	// Fail fast if it's not a password command
	if !strings.Contains(cmd, "chpasswd") {
		return nil
	}

	root, _, err := d.authFn(id)
	if err != nil {
		return err
	}

	// It is a password command, see if it's our root
	if !strings.Contains(cmd, root) {
		return nil
	}

	// so it is likely a password command on our root user, fully parse the cmd
	echopart := strings.Split(cmd, "|")[0]
	fq := strings.LastIndex(cmd, "\"")
	upasspart := echopart[6:fq]
	upass := strings.Split(upasspart, ":")

	if upass[0] == root {
		log.Debug("Need to update")
		kv := &driver.KVPair{
			Devid: &driver.DeviceID{
				Id: id,
			},
			Key:   "password",
			Value: upass[1],
		}

		r, e := d.Client().SaveMeta(context.Background(), kv)
		if r.Ok != true {
			return e
		}
	}

	return nil
}

/*
authFn() [[optional]]

	Credential provider function.  Returns the authentication parameters
	to the transport layer if needed.  Not all transports or devices require
	authentication.  At the time of writing, only the ssh transport uses this
	feature.

	This is NOT a function required by the DriverService interface, and is
	passed into the transport creation routine.  As such naming is not
	important.  This could be, for instance, an inline function passed into
	ConnectToDevice()
*/
func (d *driverService) authFn(id int64) (username, password string, err error) {
	log.Debug("authFn()")
	u, err := d.Client().GetMeta(context.Background(), &driver.KVRequest{
		Devid: &driver.DeviceID{Id: id},
		Key:   "username",
	})
	if err != nil {
		return "", "", err
	}

	p, err := d.Client().GetMeta(context.Background(), &driver.KVRequest{
		Devid: &driver.DeviceID{Id: id},
		Key:   "password",
	})
	if err != nil {
		return "", "", err
	}

	log.Debug("user/pass: %s/%s", u.Value, p.Value)
	return u.Value, p.Value, nil
}

/*
Service configuration and start up are handled by the framework.  There is
	no need to do anything beyond instantiate an instance of the driver
	service object, and pass it to driver.Main()
*/
func main() {
	devdriver := driverService{name: "linux"}
	log = driver.GetLogger(devdriver.Name())
	driver.Main(&devdriver)
}
