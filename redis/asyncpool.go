// Copyright 2012 Gary Burd
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package redis

import (
	"errors"
	"sync"
	"time"
)

var errorCompatibility = errors.New("RedisGo-Async: should use AsyncDo func")

// AsyncPool maintains one connection.
type AsyncPool struct {
	// Dial is an application supplied function for creating and configuring a
	// connection.
	//
	// The connection returned from Dial must not be in a special state
	// (subscribed to pubsub channel, transaction started, ...).
	Dial func() (AsynConn, error)
	// TestOnBorrow is an optional application supplied function for checking
	// the health of an idle connection before the connection is used again by
	// the application. Argument t is the time that the connection was returned
	// to the pool. If the function returns an error, then the connection is
	// closed.
	TestOnBorrow func(c AsynConn, t time.Time) error
	// MaxGetCount is the maximum value that limits the hang up 'Get()' goroutine.
	// When zero, there is no limit.
	MaxGetCount int
	// MaxGetCount is the maximum value that limits the hang up 'Do()' goroutine.
	// When zero, there is no limit.
	MaxDoCount int

	c        *asyncPoolConnection
	mu       sync.Mutex
	cond     *sync.Cond
	getCount int
	doCount  int
	closed   bool
	blocking bool
}

// NewAsyncPool creates a new async pool.
func NewAsyncPool(newFn func() (AsynConn, error), testFn func(AsynConn, time.Time) error) *AsyncPool {
	return &AsyncPool{Dial: newFn, TestOnBorrow: testFn}
}

// Get gets a connection.
func (p *AsyncPool) Get() AsynConn {
	p.mu.Lock()
	if p.cond == nil {
		p.cond = sync.NewCond(&p.mu)
	}

	p.getCount++
	if p.MaxGetCount != 0 && p.getCount >= p.MaxGetCount {
		p.getCount--
		p.mu.Unlock()
		return errorConnection{ErrPoolExhausted}
	}

	var pc AsynConn
	for {
		if p.closed {
			p.getCount--
			p.mu.Unlock()
			return errorConnection{errPoolClosed}
		}

		if p.blocking {
			p.cond.Wait()
			continue
		}

		if p.c != nil && p.c.Err() == nil {
			if test := p.TestOnBorrow; test != nil {
				p.blocking = true
				ic := p.c.c.(*asynConn)
				p.mu.Unlock()

				err := test(p.c, ic.t)

				p.mu.Lock()
				p.blocking = false
				if err == nil {
					pc = p.c
					p.getCount--
					p.cond.Signal()
					p.mu.Unlock()
					return pc
				}
			} else {
				pc = p.c
				p.getCount--
				p.cond.Signal()
				p.mu.Unlock()
				return pc
			}
		}

		if p.c != nil {
			p.c.c.Close()
		}
		p.blocking = true
		p.mu.Unlock()

		c, err := p.Dial()

		p.mu.Lock()
		p.blocking = false
		if err != nil {
			p.getCount--
			p.cond.Signal()
			p.mu.Unlock()
			return errorConnection{err}
		}

		p.c = &asyncPoolConnection{p: p, c: c}
		pc := p.c
		p.getCount--
		p.cond.Signal()
		p.mu.Unlock()

		return pc
	}
}

// ActiveCount returns the number of client of this pool.
func (p *AsyncPool) ActiveCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.c != nil && p.c.Err() == nil {
		return 1
	}
	return 0
}

// IdleCount returns the number of idle connections in the pool.
func (p *AsyncPool) IdleCount() int {
	return 0
}

// Close releases the resources used by the pool.
func (p *AsyncPool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}
	p.closed = true
	if p.cond != nil {
		p.cond.Broadcast()
	}
	err := p.c.c.Close()
	p.c = nil

	return err
}

type asyncPoolConnection struct {
	p *AsyncPool
	c AsynConn
}

func (pc *asyncPoolConnection) Close() error {
	return nil
}

func (pc *asyncPoolConnection) Err() error {
	return pc.c.Err()
}

func (pc *asyncPoolConnection) Do(commandName string, args ...interface{}) (reply interface{}, err error) {
	if pc.p.MaxDoCount != 0 {
		pc.p.mu.Lock()
		pc.p.doCount++
		if pc.p.doCount >= pc.p.MaxDoCount {
			pc.p.doCount--
			pc.p.mu.Unlock()
			return nil, ErrPoolExhausted
		}
	}

	reply, err = pc.c.Do(commandName, args...)

	if pc.p.MaxDoCount != 0 {
		pc.p.mu.Lock()
		pc.p.doCount--
		pc.p.mu.Unlock()
	}

	return reply, err
}

func (pc *asyncPoolConnection) AsyncDo(commandName string, args ...interface{}) (ret AsyncRet, err error) {
	return pc.c.AsyncDo(commandName, args...)
}

func (pc *asyncPoolConnection) Send(commandName string, args ...interface{}) error {
	return errorCompatibility
}

func (pc *asyncPoolConnection) Flush() error {
	return errorCompatibility
}

func (pc *asyncPoolConnection) Receive() (reply interface{}, err error) {
	return nil, errorCompatibility
}

func (ec errorConnection) AsyncDo(string, ...interface{}) (AsyncRet, error) { return nil, ec.err }
