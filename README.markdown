## RedisGo-Async

[![GoDoc](http://godoc.org/github.com/gistao/RedisGo-Async/redis?status.svg)](http://godoc.org/github.com/gistao/RedisGo-Async/redis)

## Features
### asynchronous(only one connection)
  * A [Print-like](http://godoc.org/github.com/gistao/RedisGo-Async/redis#hdr-Executing_Commands) API with support for all Redis commands.

  * [Pipelining](http://godoc.org/github.com/gistao/RedisGo-Async/redis#hdr-Pipelining)
  
  * [Script helper type](http://godoc.org/github.com/gistao/RedisGo-Async/redis#Script) with optimistic use of EVALSHA.
  
  * Transactions limited support.

  * [Helper functions](http://godoc.org/github.com/gistao/RedisGo-Async/redis#hdr-Reply_Helpers) for working with command replies.


### synchronous(connection pool)
  * [Print-like](http://godoc.org/github.com/gistao/RedisGo-Async/redis#hdr-Executing_Commands) API with support for all Redis commands.

  * [Pipelining](http://godoc.org/github.com/gistao/RedisGo-Async/redis#hdr-Pipelining)
  
  * [Publish/Subscribe](http://godoc.org/github.com/gistao/RedisGo-Async/redis#hdr-Publish_and_Subscribe)
  
  * [Script helper type](http://godoc.org/github.com/gistao/RedisGo-Async/redis#Script) with optimistic use of EVALSHA.
  
  * [Helper functions](http://godoc.org/github.com/gistao/RedisGo-Async/redis#hdr-Reply_Helpers) for working with command replies.
  

## Installation

Install RedisGo-Async using the "go get" command:

    go get github.com/gistao/RedisGo-Async/redis

The Go distribution is RedisGo-Async's only dependency.


## Benchmark Test
![bench report](https://github.com/gistao/RedisGo-Async/blob/master/bench/bench.png "Title")


## Related Projects
- [garyburd/redigo](https://github.com/garyburd/redigo) - A client library for Redis.


## Contributing
* garyburd(https://github.com/garyburd).
* xiaofei(https://github.com/feibulei).


License
-------

RedisGo-Async is available under the [Apache License, Version 2.0](http://www.apache.org/licenses/LICENSE-2.0.html).
