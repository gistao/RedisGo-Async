// Copyright 2017 xiaofei, gistao
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
	// "fmt"
	"time"
)

type tResult struct {
	result interface{}
	err    error
}

// send to dorequest
type tRequest struct {
	cmd  string
	args []interface{}
	c    chan *tResult
}

// send to doreply
type tReply struct {
	cmd string
	c   chan *tResult
}

type asyncRet struct {
	c chan *tResult
}

// conn is the low-level implementation of Conn
type asynConn struct {
	*conn
	t            time.Time
	reqChan      chan *tRequest
	repChan      chan *tReply
	closeReqChan chan bool
	closeRepChan chan bool
	closed       bool
}

// AsyncDialTimeout acts like AsyncDial but takes timeouts for establishing the
// connection to the server, writing a command and reading a reply.
//
// Deprecated: Use AsyncDial with options instead.
func AsyncDialTimeout(network, address string, connectTimeout, readTimeout, writeTimeout time.Duration) (AsynConn, error) {
	return AsyncDial(network, address,
		DialConnectTimeout(connectTimeout),
		DialReadTimeout(readTimeout),
		DialWriteTimeout(writeTimeout))
}

// AsyncDial connects to the Redis server at the given network and
// address using the specified options.
func AsyncDial(network, address string, options ...DialOption) (AsynConn, error) {
	tmp, err := Dial(network, address, options...)
	if err != nil {
		return nil, err
	}

	return getAsynConn(tmp.(*conn))
}

// AsyncDialURL connects to a Redis server at the given URL using the Redis
// URI scheme. URLs should follow the draft IANA specification for the
// scheme (https://www.iana.org/assignments/uri-schemes/prov/redis).
func AsyncDialURL(rawurl string, options ...DialOption) (AsynConn, error) {
	tmp, err := DialURL(rawurl, options...)
	if err != nil {
		return nil, err
	}
	return getAsynConn(tmp.(*conn))
}

func getAsynConn(conn *conn) (AsynConn, error) {
	c := &asynConn{
		conn:         conn,
		reqChan:      make(chan *tRequest, 1000),
		repChan:      make(chan *tReply, 1000),
		closeReqChan: make(chan bool),
		closeRepChan: make(chan bool)}

	// request routine
	go c.doRequest()
	// reply routine
	go c.doReply()

	return c, nil
}

// Do command to redis server,the goroutine of caller will be suspended.
func (c *asynConn) Do(cmd string, args ...interface{}) (interface{}, error) {
	if cmd == "" {
		return nil, errors.New("RedisGo-Async: empty command")
	}

	retChan := make(chan *tResult, 2)

	c.reqChan <- &tRequest{cmd: cmd, args: args, c: retChan}

	ret := <-retChan
	if ret.err != nil {
		// fmt.Println(cmd, ret.err)
		return ret.result, ret.err
	}
	ret = <-retChan
	return ret.result, ret.err
}

// Do command to redis server,the goroutine of caller is not suspended.
func (c *asynConn) AsyncDo(cmd string, args ...interface{}) (AsyncRet, error) {
	if cmd == "" {
		return nil, errors.New("RedisGo-Async: empty command")
	}

	retChan := make(chan *tResult, 2)

	c.reqChan <- &tRequest{cmd: cmd, args: args, c: retChan}

	return &asyncRet{c: retChan}, nil
}

func (c *asynConn) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}
	c.closed = true

	go func() {
		time.Sleep(10 * time.Minute)
		c.closeReqChan <- true
	}()

	c.err = errors.New("RedisGo-Async: closed")
	return c.conn.conn.Close()
}

func (c *asynConn) doRequest() {
	reqs := make([]*tRequest, 0, 1000)
	for {
		select {
		case <-c.closeReqChan:
			close(c.reqChan)
			c.closeRepChan <- true
			return
		case req := <-c.reqChan:
			for i, length := 0, len(c.reqChan); ; {
				if c.writeTimeout != 0 {
					c.conn.conn.SetWriteDeadline(time.Now().Add(c.writeTimeout))
				}
				if err := c.writeCommand(req.cmd, req.args); err != nil {
					req.c <- &tResult{nil, err}
					c.fatal(err)
					break
				}
				reqs = append(reqs, req)
				if i++; i > length || c.bw.Buffered() >= 4096 {
					break
				}
				req = <-c.reqChan
			}
		}

		err := c.bw.Flush()
		for i := range reqs {
			reqs[i].c <- &tResult{nil, err}
			if err != nil {
				c.fatal(err)
				continue
			}
			c.repChan <- &tReply{cmd: reqs[i].cmd, c: reqs[i].c}
		}
		reqs = reqs[:0]
	}
}

func (c *asynConn) doReply() {
	for {
		select {
		case <-c.closeRepChan:
			close(c.repChan)
			return
		case rep := <-c.repChan:
			if c.readTimeout != 0 {
				c.conn.conn.SetReadDeadline(time.Now().Add(c.readTimeout))
			}
			reply, err := c.readReply()
			if err != nil {
				rep.c <- &tResult{nil, c.fatal(err)}
				c.fatal(err)
				continue
			} else {
				c.t = nowFunc()
			}
			if e, ok := reply.(Error); ok {
				err = e
			}
			rep.c <- &tResult{reply, err}
		}
	}
}

// Get get command result asynchronously
func (a *asyncRet) Get() (interface{}, error) {
	send := <-a.c
	if send.err != nil {
		return send.result, send.err
	}

	recv := <-a.c
	return recv.result, recv.err
}
