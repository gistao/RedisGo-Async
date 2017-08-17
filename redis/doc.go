// Copyright 2012 Gary Burd
// Copyright 2017 gistao
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

// Package redis is a client for the Redis database.both asynchronous and synchronous modes
// are supported,its API is fully compatible with redigo.
//
// The RedisGo-Async  FAQ (https://github.com/gistao/RedisGo-Async/wiki/FAQ) contains more
// documentation about this package.
//
// Connections
//
// In synchronous mode, this library creates a connection pool,
// and then you can test to determine a maximum number of connections,
// such as 100.
//
// In asynchronous mode, this library will only create a connection,
// and you don't have to worry about performance issues,
// nor do you have to spend a lot of time testing the number of connections.
//
// Executing Commands
//
// The Conn interface has a generic method for executing Redis commands:
//
//  Do(commandName string, args ...interface{}) (reply interface{}, err error)
//
// The Redis command reference (http://redis.io/commands) lists the available
// commands. An example of using the Redis APPEND command is:
//
//  n, err := conn.Do("APPEND", "key", "value")
//
// The Do method converts command arguments to binary strings for transmission
// to the server as follows:
//
//  Go Type                 Conversion
//  []byte                  Sent as is
//  string                  Sent as is
//  int, int64              strconv.FormatInt(v)
//  float64                 strconv.FormatFloat(v, 'g', -1, 64)
//  bool                    true -> "1", false -> "0"
//  nil                     ""
//  all other types         fmt.Print(v)
//
// Redis command reply types are represented using the following Go types:
//
//  Redis type              Go type
//  error                   redis.Error
//  integer                 int64
//  simple string           string
//  bulk string             []byte or nil if value not present.
//  array                   []interface{} or nil if value not present.
//
// Use type assertions or the reply helper functions to convert from
// interface{} to the specific Go type for the command result.
//
// Pipelining
//
// In synchronous mode, Connections support pipelining using the Send, Flush and Receive methods.
//
//  Send(commandName string, args ...interface{}) error
//  Flush() error
//  Receive() (reply interface{}, err error)
//
// Send writes the command to the connection's output buffer. Flush flushes the
// connection's output buffer to the server. Receive reads a single reply from
// the server. The following example shows a simple pipeline.
//
//  c.Send("SET", "foo", "bar")
//  c.Send("GET", "foo")
//  c.Flush()
//  c.Receive() // reply from SET
//  v, err = c.Receive() // reply from GET
//
// In asynchronous mode, Connections support pipelining using the Do
//
//  AsyncDo(commandName string, args ...interface{}) (reply interface{}, err error)
//
// Above example
//
//  c.AsyncDo("SET", "foo", "bar")
//  ret := c.AsyncDo("Get", "foo")
//  v, err := ret.Get()
//
// Concurrency
//
// Connections support one concurrent caller to the Receive method and one
// concurrent caller to the Send and Flush methods. No other concurrency is
// supported including concurrent calls to the Do method.
//
// For full concurrent access to Redis, use the thread-safe Pool to get.
//
// Publish and Subscribe
//
// Use the Send, Flush and Receive methods to implement Pub/Sub subscribers.
//
//  c.Send("SUBSCRIBE", "example")
//  c.Flush()
//  for {
//      reply, err := c.Receive()
//      if err != nil {
//          return err
//      }
//      // process pushed message
//  }
//
// The PubSubConn type wraps a Conn with convenience methods for implementing
// subscribers. The Subscribe, PSubscribe, Unsubscribe and PUnsubscribe methods
// send and flush a subscription management command. The receive method
// converts a pushed message to convenient types for use in a type switch.
//
//  psc := redis.PubSubConn{Conn: c}
//  psc.Subscribe("example")
//  for {
//      switch v := psc.Receive().(type) {
//      case redis.Message:
//          fmt.Printf("%s: message: %s\n", v.Channel, v.Data)
//      case redis.Subscription:
//          fmt.Printf("%s: %s %d\n", v.Channel, v.Kind, v.Count)
//      case error:
//          return v
//      }
//  }
//
// NOTE asynchronous mode does not support PUB/SUB
//
// Reply Helpers
//
// The Bool, Int, Bytes, String, Strings and Values functions convert a reply
// to a value of a specific type. To allow convenient wrapping of calls to the
// connection Do and Receive methods, the functions take a second argument of
// type error.  If the error is non-nil, then the helper function returns the
// error. If the error is nil, the function converts the reply to the specified
// type:
//
//  exists, err := redis.Bool(c.Do("EXISTS", "foo"))
//  if err != nil {
//      // handle error return from c.Do or type conversion error.
//  }
//
// The Scan function converts elements of a array reply to Go types:
//
//  var value1 int
//  var value2 string
//  reply, err := redis.Values(c.Do("MGET", "key1", "key2"))
//  if err != nil {
//      // handle error
//  }
//   if _, err := redis.Scan(reply, &value1, &value2); err != nil {
//      // handle error
//  }
//
// Errors
//
// Connection methods return error replies from the server as type redis.Error.
//
// Call the connection Err() method to determine if the connection encountered
// non-recoverable error such as a network error or protocol parsing error. If
// Err() returns a non-nil value, then the connection is not usable and should
// be closed.
package redis
