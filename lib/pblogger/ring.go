/*
Package ring provides a simple implementation of a ring buffer.
https://github.com/zfjagann/golang-ring

LICENSE
Copyright (c) 2014, Zeal Jagannatha
All rights reserved.

Redistribution and use in source and binary forms, with or without modification, are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright notice, this list of conditions and the following disclaimer.

2. Redistributions in binary form must reproduce the above copyright notice, this list of conditions and the following disclaimer in the documentation and/or other materials provided with the distribution.

3. Neither the name of the copyright holder nor the names of its contributors may be used to endorse or promote products derived from this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

*/
package pblogger

/*
The DefaultCapacity of an uninitialized Ring buffer.

Changing this value only affects ring buffers created after it is changed.
*/
var DefaultCapacity int = 10

/*
Type Ring implements a Circular Buffer.
The default value of the Ring struct is a valid (empty) Ring buffer with capacity DefaultCapacify.
*/
type Ring struct {
	head int // the most recent value written
	tail int // the least recent value written
	buff []interface{}
}

/*
Set the maximum size of the ring buffer.
*/
func (r *Ring) SetCapacity(size int) {
	r.checkInit()
	r.extend(size)
}

/*
Capacity returns the current capacity of the ring buffer.
*/
func (r Ring) Capacity() int {
	return len(r.buff)
}

/*
Enqueue a value into the Ring buffer.
*/
func (r *Ring) Enqueue(i interface{}) {
	r.checkInit()
	r.set(r.head+1, i)
	old := r.head
	r.head = r.mod(r.head + 1)
	if old != -1 && r.head == r.tail {
		r.tail = r.mod(r.tail + 1)
	}
}

/*
Dequeue a value from the Ring buffer.

Returns nil if the ring buffer is empty.
*/
func (r *Ring) Dequeue() interface{} {
	r.checkInit()
	if r.head == -1 {
		return nil
	}
	v := r.get(r.tail)
	if r.tail == r.head {
		r.head = -1
		r.tail = 0
	} else {
		r.tail = r.mod(r.tail + 1)
	}
	return v
}

/*
Read the value that Dequeue would have dequeued without actually dequeuing it.

Returns nil if the ring buffer is empty.
*/
func (r *Ring) Peek() interface{} {
	r.checkInit()
	if r.head == -1 {
		return nil
	}
	return r.get(r.tail)
}

/*
Values returns a slice of all the values in the circular buffer without modifying them at all.
The returned slice can be modified independently of the circular buffer. However, the values inside the slice
are shared between the slice and circular buffer.
*/
func (r *Ring) Values() []interface{} {
	if r.head == -1 {
		return []interface{}{}
	}
	arr := make([]interface{}, 0, r.Capacity())
	for i := 0; i < r.Capacity(); i++ {
		idx := r.mod(i + r.tail)
		arr = append(arr, r.get(idx))
		if idx == r.head {
			break
		}
	}
	return arr
}

/**
*** Unexported methods beyond this point.
**/

// sets a value at the given unmodified index and returns the modified index of the value
func (r *Ring) set(p int, v interface{}) {
	r.buff[r.mod(p)] = v
}

// gets a value based at a given unmodified index
func (r *Ring) get(p int) interface{} {
	return r.buff[r.mod(p)]
}

// returns the modified index of an unmodified index
func (r *Ring) mod(p int) int {
	return p % len(r.buff)
}

func (r *Ring) checkInit() {
	if r.buff == nil {
		r.buff = make([]interface{}, DefaultCapacity)
		for i := range r.buff {
			r.buff[i] = nil
		}
		r.head, r.tail = -1, 0
	}
}

func (r *Ring) extend(size int) {
	if size == len(r.buff) {
		return
	} else if size < len(r.buff) {
		r.buff = r.buff[0:size]
	}
	newb := make([]interface{}, size-len(r.buff))
	for i := range newb {
		newb[i] = nil
	}
	r.buff = append(r.buff, newb...)
}
