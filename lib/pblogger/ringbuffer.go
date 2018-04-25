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

import ()

type RingBuffer struct {
	Ring
}

func (r *RingBuffer) PeekHead() interface{} {
	r.checkInit()
	if r.head == -1 {
		return nil
	}
	return r.get(r.head)
}

func (r *RingBuffer) PeekTail() interface{} {
	return r.Peek()
}
