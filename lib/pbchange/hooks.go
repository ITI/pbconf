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
	"fmt"
	"sync"
)

type cbStore struct {
	context interface{}
	fn      CommitCB
	cbtype  CMType
}

type CommitCB func(interface{}, *ChangeData)

func (engine *CMEngine) RegisterCommitListener(cbtype CMType, context interface{}, fn CommitCB) {
	engine.commitCBs = append(engine.commitCBs, &cbStore{context: context, fn: fn, cbtype: cbtype})
}

func (engine *CMEngine) runCommitCBs(cbtype CMType, cbdata *ChangeData) {

	var wg sync.WaitGroup

	for _, store := range engine.commitCBs {
		if store.cbtype == cbtype {
			wg.Add(1)
			go func(s *cbStore) {
				defer func() {
					if r := recover(); r != nil {
						log.Error(fmt.Sprintf("Callback failed: %v", r))
					}
					wg.Done()
				}()

				s.fn(s.context, cbdata)
			}(store)
		}
	}

	wg.Wait()
}
func (engine *CMEngine) RegisterPackRcvdListener(cbtype CMType, context interface{}, fn CommitCB) {
	engine.packRcvdCBs = append(engine.packRcvdCBs, &cbStore{context: context, fn: fn, cbtype: cbtype})
}

func (engine *CMEngine) runPackRcvCBs(cbtype CMType, cbdata *ChangeData) {

	var wg sync.WaitGroup

	for _, store := range engine.packRcvdCBs {
		if store.cbtype == cbtype {
			wg.Add(1)
			go func(s *cbStore) {
				defer func() {
					if r := recover(); r != nil {
						log.Error(fmt.Sprintf("Callback failed: %v", r))
					}
					wg.Done()
				}()

				s.fn(s.context, cbdata)
			}(store)
		}
	}

	wg.Wait()
}
