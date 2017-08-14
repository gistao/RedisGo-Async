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
	"fmt"
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
	t       time.Time
	reqChan chan *tRequest
	repChan chan *tReply
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
		conn:    conn,
		reqChan: make(chan *tRequest, 10000),
		repChan: make(chan *tReply, 10000),
	}

	// request routine
	go c.doRequest()
	// reply routine
	go c.doReply()

	return c, nil
}

func (c *asynConn) Do(cmd string, args ...interface{}) (interface{}, error) {
	if cmd == "" {
		return nil, errors.New("RedisGo-Async: empty command")
	}
	retChan := make(chan *tResult, 2)

	c.reqChan <- &tRequest{cmd: cmd, args: args, c: retChan}

	ret := <-retChan
	if ret.err != nil {
		fmt.Println(cmd, ret.err)
		return ret.result, ret.err
	}
	ret = <-retChan
	return ret.result, ret.err
}

func (c *asynConn) AsyncDo(cmd string, args ...interface{}) (AsyncRet, error) {
	if cmd == "" {
		return nil, errors.New("RedisGo-Async: empty command")
	}

	retChan := make(chan *tResult, 2)

	c.reqChan <- &tRequest{cmd: cmd, args: args, c: retChan}

	ret := <-retChan
	if ret.err != nil {
		return nil, ret.err
	}
	return &asyncRet{c: retChan}, nil
}

func (c *asynConn) Close() error {
	close(c.reqChan)
	close(c.repChan)

	err := c.err
	if c.err == nil {
		c.err = errors.New("RedisGo-Async: closed")
		err = c.conn.Close()
	}
	return err
}

func (c *asynConn) doRequest() {
	reqs := make([]*tRequest, 0, 1000)
	for {
		req := <-c.reqChan
		for i, length := 0, len(c.reqChan); ; {
			if c.writeTimeout != 0 {
				c.conn.conn.SetWriteDeadline(time.Now().Add(c.writeTimeout))
			}
			if err := c.writeCommand(req.cmd, req.args); err != nil {
				req.c <- &tResult{nil, err}
				// TODO
				break
			}
			reqs = append(reqs, req)
			if i++; i > length || c.bw.Buffered() >= 4096 {
				break
			}
			req = <-c.reqChan
		}
		err := c.bw.Flush()
		for i := range reqs {
			reqs[i].c <- &tResult{nil, err}
			if err != nil {
				// TODO
				continue
			}
			c.repChan <- &tReply{cmd: reqs[i].cmd, c: reqs[i].c}
		}
		reqs = reqs[:0]
	}
}

func (c *asynConn) doReply() {
	for {
		rep := <-c.repChan
		if c.readTimeout != 0 {
			c.conn.conn.SetReadDeadline(time.Now().Add(c.readTimeout))
		}
		reply, err := c.readReply()
		if err != nil {
			rep.c <- &tResult{nil, c.fatal(err)}
			// TODO
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

// Get get command result asynchronously
func (a *asyncRet) Get(timeout time.Duration) (interface{}, error) {
	select {
	case ret := <-a.c:
		return ret.result, ret.err
	case <-time.After(timeout):
		return nil, errors.New("RedisGo-Async: async get time out")
	}
}
