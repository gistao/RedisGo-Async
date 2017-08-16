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
	"time"
)

type SynClient struct {
	pool *redis.Pool
	Addr string
}

var (
	syncMap   map[string]*SynClient
	syncMutex *sync.RWMutex
)

const (
	defaultTimeout = 300
)

func init() {
	syncMap = make(map[string]*SynClient)
	syncMutex = new(sync.RWMutex)
}

func newSyncPool(addr string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     100,
		MaxActive:   100,
		IdleTimeout: time.Minute * 1,
		Wait:        true,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", addr)
			return c, err
		},
	}
}

func GetSynClient(addr string) *SynClient {
	var redis *SynClient
	var mok bool
	syncMutex.RLock()
	redis, mok = syncMap[addr]
	syncMutex.RUnlock()
	if !mok {
		syncMutex.Lock()
		redis, mok = syncMap[addr]
		if !mok {
			redis = &SynClient{pool: newSyncPool(addr), Addr: addr}
			syncMap[addr] = redis
		}
		syncMutex.Unlock()
	}
	return redis
}

func (c *SynClient) Get(key string) (string, error) {
	conn := c.pool.Get()
	defer conn.Close()
	reply, errDo := conn.Do("GET", key)
	if errDo == nil && reply == nil {
		return "", nil
	}
	val, err := redis.String(reply, errDo)
	return val, err
}

func (c *SynClient) Del(key string) (int64, error) {
	conn := c.pool.Get()
	defer conn.Close()
	reply, errDo := conn.Do("DEL", key)
	if errDo == nil && reply == nil {
		return 0, nil
	}
	val, err := redis.Int64(reply, errDo)
	return val, err
}

func (c *SynClient) Set(key string, val string) (string, error) {
	conn := c.pool.Get()
	defer conn.Close()
	val, err := redis.String(conn.Do("SET", key, val))
	return val, err
}
