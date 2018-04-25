package pbchange

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
	"strings"
	"time"
)

var devMonLog logging.Logger
var sweepLog logging.Logger
var cleanLog logging.Logger

func (engine *CMEngine) Start(timers ...int) chan bool {
	ll := logging.GetLevel("CME:Main")
	devMonLog, _ = logging.GetLogger("CME:Dev Monitor")
	logging.SetLevel(ll, "CME:Dev Monitor")
	sweepLog, _ = logging.GetLogger("CME:Sweeper")
	logging.SetLevel(ll, "CME:Sweeper")
	cleanLog, _ = logging.GetLogger("CME:Cleaner")
	logging.SetLevel(ll, "CME:Cleaner")

	doneChan := make(chan bool)

	var devMonTime int
	var sweepTime int
	var cleanTime int

	switch {
	case len(timers) == 1:
		devMonTime = timers[0]
		cleanTime = 10
		sweepTime = 60
	case len(timers) == 2:
		devMonTime = timers[0]
		cleanTime = timers[1]
		sweepTime = 60
	case len(timers) >= 3:
		devMonTime = timers[0]
		cleanTime = timers[1]
		sweepTime = timers[2]
	default:
		devMonTime = 10
		cleanTime = 10
		sweepTime = 60
	}

	devMonTicker := time.NewTicker(time.Second * time.Duration(devMonTime))
	sweepTicker := time.NewTicker(time.Second * time.Duration(sweepTime))
	cleanTicker := time.NewTicker(time.Second * time.Duration(cleanTime))

	go func() {
		go func() {
			for t := range devMonTicker.C {
				engine.checkDevs(t)
			}
		}()

		go func() {
			for t := range sweepTicker.C {
				engine.sweepComplete(t)
			}
		}()

		go func() {
			for t := range cleanTicker.C {
				engine.cleanBranches(t)
			}
		}()

		switch {
		case <-doneChan:
			devMonTicker.Stop()
			sweepTicker.Stop()
			cleanTicker.Stop()
		}
	}()

	return doneChan

}

func (engine *CMEngine) checkDevs(t time.Time) {
	// Check for config changes on devices
	devMonLog.Debug("checkDevs(%v)", t)
}

func (engine *CMEngine) cleanBranches(t time.Time) {
	cleanLog.Debug("cleanBranches(%v)", t)
	tracker := make(map[string][]string, 0)

	// Remove dead refs
	engine.guard.Lock()
	for id, trans := range activeTransactions {
		if _, ok := tracker[trans.Ctype.String()]; !ok {
			l, err := engine.getBranches(trans.Ctype)
			if err != nil {
				cleanLog.Error("Encountered Error listing branches: ", err.Error())
				continue
			}

			tracker[trans.Ctype.String()] = l
		}

		indx := func(s string, list []string) int {
			for i, v := range list {
				if s == v {
					return i
				}
			}
			return -1
		}(id, tracker[trans.Ctype.String()])

		if indx == -1 {
			// There is no branch for the transactionID
			delete(activeTransactions, id)
		}
	}
	engine.guard.Unlock()

	// At this point, all refs should point to something and
	// all branches should be loaded
	//
	// From here on out, we act like a modified mark and sweep GC

	// Mark
	dead := make([]string, 0)
	engine.guard.Lock()
	for _, list := range tracker {
		for _, branch := range list {
			if _, ok := activeTransactions[branch]; !ok {
				dead = append(dead, branch)
			}
		}
	}
	engine.guard.Unlock()

	// Sweep
	engine.guard.Lock()
	for objtype, _ := range tracker {
		engine.run(StringToCMType(objtype), "checkout", "master")
		for _, branch := range dead {
			engine.run(StringToCMType(objtype), "branch", "-D", branch)
		}
	}
	engine.guard.Unlock()
}

func (engine *CMEngine) getBranches(ctype CMType) ([]string, error) {
	o, e := engine.run(ctype, "branch", "--list")
	if e != nil {
		return nil, e
	}

	r := make([]string, 0)
	lines := strings.Split(o, "\n")
	for _, line := range lines {
		line = strings.TrimPrefix(line, "*")
		line = strings.TrimSpace(line)
		if line == "master" {
			continue
		}
		r = append(r, line)
	}

	return r, nil
}

func (engine *CMEngine) sweepComplete(t time.Time) {
	// Clear Completed transactions
	sweepLog.Debug("sweepComplete(%v)\n", t)

	for id, status := range activeTransactions {
		if status.Status == COMPLETE {
			sweepLog.Debug("Removing completed transaction %s", id)
			engine.guard.Lock()
			activeTransactions[id].Status = CLEANED
			engine.guard.Unlock()

			engine.removeBranch(id, status)

			engine.guard.Lock()
			delete(activeTransactions, id)
			engine.guard.Unlock()
		}
	}
}

func (e *CMEngine) removeBranch(id string, t *Transaction) {
	e.guard.Lock()
	defer e.guard.Unlock()

	// Make sure we're on master
	e.run(t.Ctype, "checkout", "master")

	// Force remove the branch
	e.run(t.Ctype, "branch", "-D", id)

}
