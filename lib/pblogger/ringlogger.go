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

type RingBackend struct {
	ring *RingBuffer

	lock sync.RWMutex
}

var ringbackend *RingBackend

func NewRingBackend(threshold logging.Level, cfg *config.Config) *RingBackend {
	if ringbackend != nil {
		return GetRing()
	}

	r := RingBackend{lock: sync.RWMutex{},
		ring: &RingBuffer{},
	}

	if cfg.Global.RingMemLimit < 1024 {
		r.ring.SetCapacity(1024)
	} else {
		r.ring.SetCapacity(cfg.Global.RingMemLimit)
	}
	ringbackend = &r
	return ringbackend
}

func GetRing() *RingBackend {
	return ringbackend
}

func (b *RingBackend) GetLimits() (*logging.Record, *logging.Record) {
	b.lock.RLock()
	defer b.lock.RUnlock()

	return b.ring.PeekTail().(*logging.Record), b.ring.PeekHead().(*logging.Record)
}

func (b *RingBackend) GetValues() []*logging.Record {
	b.lock.Lock()
	defer b.lock.Unlock()

	r := b.ring.Values()
	rec := make([]*logging.Record, 0)
	for _, i := range r {
		if i != nil {
			rec = append(rec, i.(*logging.Record))
		}
	}

	return rec
}

func (b *RingBackend) Log(level logging.Level, calldepth int, rec *logging.Record) error {
	b.lock.Lock()
	b.ring.Enqueue(rec)
	b.lock.Unlock()

	return nil
}
