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

package main

import (
	"github.com/gistao/RedisGo-Async/redis"
	"sync"
)

type AsynClient struct {
	pool *redis.AsyncPool
	Addr string
}

var (
	asyncMap   map[string]*AsynClient
	asyncMutex *sync.RWMutex
)

const (
	defaultTimeout = 300
)

func init() {
	asyncMap = make(map[string]*AsynClient)
	asyncMutex = new(sync.RWMutex)
}

func newAsyncPool(addr string) *redis.AsyncPool {
	return &redis.AsyncPool{
		Dial: func() (redis.AsynConn, error) {
			c, err := redis.AsyncDial("tcp", addr)
			return c, err
		},
	}
}

func GetAsynClient(addr string) *AsynClient {
	var redis *AsynClient
	var mok bool
	asyncMutex.RLock()
	redis, mok = asyncMap[addr]
	asyncMutex.RUnlock()
	if !mok {
		asyncMutex.Lock()
		redis, mok = asyncMap[addr]
		if !mok {
			redis = &AsynClient{pool: newAsyncPool(addr), Addr: addr}
			asyncMap[addr] = redis
		}
		asyncMutex.Unlock()
	}
	return redis
}

func (c *AsynClient) Get(key string) (string, error) {
	conn := c.pool.Get()
	defer conn.Close()
	reply, errDo := conn.Do("GET", key)
	if errDo == nil && reply == nil {
		return "", nil
	}
	val, err := redis.String(reply, errDo)
	return val, err
}

func (c *AsynClient) Del(key string) (int64, error) {
	conn := c.pool.Get()
	defer conn.Close()
	reply, errDo := conn.Do("DEL", key)
	if errDo == nil && reply == nil {
		return 0, nil
	}
	val, err := redis.Int64(reply, errDo)
	return val, err
}

func (c *AsynClient) Set(key string, val string) (string, error) {
	conn := c.pool.Get()
	defer conn.Close()
	val, err := redis.String(conn.Do("SET", key, val))
	return val, err
}
