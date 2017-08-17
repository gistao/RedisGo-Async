RedisGo-Async
======

[![GoDoc](http://godoc.org/github.com/gistao/RedisGo-Async/redis?status.svg)](http://godoc.org/github.com/gistao/RedisGo-Async/redis)

Features
------------

* ### asynchronous(only one connection)
  1.A Print-like API with support for all Redis commands.

  2.Pipelining
  
  3.Script helper type with optimistic use of EVALSHA.
  
  4.Transactions limited support.

  5.Helper functions for working with command replies.


* ### synchronous(connection pool)
  1.A Print-like API with support for all Redis commands.

  2.Pipelining
  
  3.Publish/Subscribe 
  
  4.Script helper type with optimistic use of EVALSHA.
  
  5.Helper functions for working with command replies.
  

Installation
------------

Install RedisGo-Async using the "go get" command:

    go get github.com/gistao/RedisGo-Async/redis

The Go distribution is RedisGo-Async's only dependency.


Benchmark Test
------------
![bench report](https://github.com/gistao/RedisGo-Async/blob/master/bench/bench.png "Title")


Related Projects
----------------

- [garyburd/redigo](https://github.com/garyburd/redigo) - A client library for Redis.


Contributing
------------

* garyburd(https://github.com/garyburd).
* xiaofei(https://github.com/feibulei).


License
-------

RedisGo-Async is available under the [Apache License, Version 2.0](http://www.apache.org/licenses/LICENSE-2.0.html).
