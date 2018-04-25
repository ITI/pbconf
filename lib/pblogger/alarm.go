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
	"sync"

	logging "github.com/iti/go-logging"
	config "github.com/iti/pbconf/lib/pbconfig"
)

type newAlarmFN func(*config.Config) AlarmDest

var alarmDests map[string]newAlarmFN
var alarmLocker sync.RWMutex

func AlarmRegister(name string, newFN newAlarmFN) {
	alarmLocker.Lock()
	if alarmDests == nil {
		alarmDests = make(map[string]newAlarmFN)
	}
	if _, ok := alarmDests[name]; !ok {
		alarmDests[name] = newFN
	}
	alarmLocker.Unlock()
}

type AlarmDest interface {
	Emit(*logging.Record, *sync.WaitGroup) error
}

type AlarmBackend struct {
	alarmThreshold logging.Level
	handlers       []AlarmDest
	handlerLock    sync.Mutex
}

func (b *AlarmBackend) Log(level logging.Level, calldepth int, rec *logging.Record) error {
	if b.IsEnabledFor(level, rec.Module) {
		var wg sync.WaitGroup
		for _, a := range b.handlers {
			wg.Add(1)
			go a.Emit(rec, &wg)
		}
		wg.Wait()
	}
	return nil
}

func (b *AlarmBackend) GetLevel(module string) logging.Level {
	return b.alarmThreshold
}

func (b *AlarmBackend) IsEnabledFor(level logging.Level, module string) bool {
	alarmLocker.RLock()
	defer alarmLocker.RUnlock()
	return level <= b.alarmThreshold
}

func NewAlarmBackend(threshold logging.Level, cfg *config.Config) *AlarmBackend {
	be := AlarmBackend{alarmThreshold: threshold}

	for _, alarm := range cfg.Global.AlarmDest {
		if newFN, ok := alarmDests[alarm]; ok {
			if alarm == "webui" && cfg.WebUI.EnableWebApp == false {
				continue //skip adding the channel alarm if webapp is not enabled.
			}
			be.handlerLock.Lock()
			be.handlers = append(be.handlers, newFN(cfg))
			be.handlerLock.Unlock()
		}
	}

	return &be
}
