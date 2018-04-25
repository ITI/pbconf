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
	"fmt"
	"io"
	"strings"

	pbdb "github.com/iti/pbconf/lib/pbdatabase"
	pbtransport "github.com/iti/pbconf/lib/pbtransport"
)

func direct(name string, srv io.ReadWriter, db pbdb.AppDatabase) {
	name = strings.TrimSpace(name)

	log.Debug(fmt.Sprintf("Looking up -->%+v<--(t)", strings.TrimSpace(name)))
	device := pbdb.PbDevice{Name: name}
	err := device.GetByName(db)
	if err != nil {
		log.Debug(fmt.Sprintf("Device: %v", device))
		log.Error(err.Error())

		return
	}

	log.Debug(fmt.Sprintf("id:%d, Name:%s", device.Id, device.Name))
	transportType, location, err := device.GetConnectionString(db, "broker")

	if err != nil {
		log.Error(err.Error())
		return
	}

	var trans pbtransport.ClientTransport

	switch {
	// Additional transports (which implement the pbtransport.ClientTransport
	// interface) can be added to this switch statement
	case transportType == "telnet":
		trans = pbtransport.NewTelnet("Broker")
	default:
		log.Error(fmt.Sprintf("Unrecognised Transport: %v", transportType))
		return
	}

	err = trans.Dial(device.Id, location)
	if err != nil {
		log.Error("Failed to connect to device")
		return
	}

	trans.Interact(srv)
}
