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
	"strconv"

	pbdb "github.com/iti/pbconf/lib/pbdatabase"
	term "golang.org/x/crypto/ssh/terminal"
)

func shell(connection io.ReadWriter, db pbdb.AppDatabase) {

	term := term.NewTerminal(connection, "==> ")

	deviceList, err := db.GetDevices()
	if err != nil {
		log.Error(err.Error())
		return
	}

	for {
		for dev := range deviceList {
			term.Write([]byte(fmt.Sprintf("%v) %s\r\n", dev, deviceList[dev].Name)))
		}

		term.Write([]byte("Select a device\r\n"))
		selection, err := term.ReadLine()

		if err != nil {
			log.Error(err.Error())
			return
		}

		intSelection, err := strconv.Atoi(selection)
		if err != nil {
			term.Write([]byte("Please select the ID of the device\r\n\r\r"))
			continue
		}

		direct(deviceList[intSelection].Name, connection, db)
		return
	}

}
