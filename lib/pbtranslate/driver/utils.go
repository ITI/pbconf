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
	db "github.com/iti/pbconf/lib/pbdatabase"
	global "github.com/iti/pbconf/lib/pbglobal"
	transport "github.com/iti/pbconf/lib/pbtransport"
)

func ConnectToDevice(fn transport.CredentialFn, id int64, drvSerName string) (transport.ClientTransport, error) {
	trans, err := GetTransportDriver(id, drvSerName)
	if err != nil {
		return nil, err
	}

	// Get the connection string
	location, err := GetConnectionString(id)
	if err != nil {
		return nil, err
	}

	trans.SetCredentialFn(fn)

	// Connect to the device
	err = trans.Dial(id, location)
	if err != nil {
		return nil, NewConnectionError("Failed to connect to device: " + err.Error())
	}

	return trans, nil
}

func GetTransportDriver(id int64, drvSerName string) (transport.ClientTransport, error) {
	dev, err := GetDevice(id)
	if err != nil {
		return nil, NewConnectionError("Failed to acquire transport: " + err.Error())
	}

	// Look up the device transport and connection string
	transportType, _, err := GetDeviceConnectionString(dev)
	if err != nil {
		return nil, NewConnectionError("Failed to connect to device: " + err.Error())
	}

	// Ask the transport layer for the transport layer communication module
	trans, err := transport.GetTransport(transportType, drvSerName)
	if err != nil {
		return nil, NewConnectionError("Failed to get transport: " + err.Error())
	}

	return trans, nil
}

func GetDevice(id int64) (*db.PbDevice, error) {
	cfg := global.CTX.Value("configuration").(*config.Config)
	pdb := db.Open(cfg.Global.Database, cfgLogLevel)
	defer pdb.Close()

	// Look up the device in the database
	dev := db.PbDevice{Id: id}
	if err := dev.Get(pdb); err != nil {
		return nil, err
	}

	return &dev, nil
}

func GetConnectionString(id int64) (string, error) {
	dev, err := GetDevice(id)
	if err != nil {
		return "", err
	}

	_, loc, err := GetDeviceConnectionString(dev)
	return loc, err
}

func GetDeviceConnectionString(dev *db.PbDevice) (string, string, error) {
	pdb := db.Open(cfg.Global.Database, cfgLogLevel)
	defer pdb.Close()

	return dev.GetConnectionString(pdb, "driver")
}

func ReplyFalse(err error) (r *BoolReply, e error) {
	e = err
	r = &BoolReply{Ok: false}
	return
}

func ReplyTrue(err error) (r *BoolReply, e error) {
	e = err
	r = &BoolReply{Ok: true}
	return
}
