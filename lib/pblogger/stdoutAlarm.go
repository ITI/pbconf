package pblogger

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
	"sync"

	logging "github.com/iti/go-logging"
	config "github.com/iti/pbconf/lib/pbconfig"
)

func init() {
	AlarmRegister("stdout", NewStdoutAlarm)
}

type StdoutAlarm struct{}

func (a *StdoutAlarm) Emit(rec *logging.Record, wg *sync.WaitGroup) error {
	fmt.Println("ALARM: ", rec)
	wg.Done()
	return nil
}

func NewStdoutAlarm(cfg *config.Config) AlarmDest {
	return &StdoutAlarm{}
}
