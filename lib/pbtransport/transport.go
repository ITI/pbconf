package transport

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
	"errors"
	"fmt"
)

func GetTransport(name, drvSrvName string) (ClientTransport, error) {
	switch {
	// Add additional transports as they are defined
	case name == "telnet":
		return NewTelnet(drvSrvName), nil
	case name == "ssh":
		return NewSSH(drvSrvName), nil
	case name == "ftp":
		return NewFTP(drvSrvName), nil
	default:
		return nil, errors.New(fmt.Sprintf("Unrecognised Transport: %v", name))
	}
}
