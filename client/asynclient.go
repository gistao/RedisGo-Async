package client

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

func (c *AsynClient) Exists(key string) (int64, error) {
	conn := c.pool.Get()
	defer conn.Close()
	reply, errDo := conn.Do("EXISTS", key)
	if errDo == nil && reply == nil {
		return 0, nil
	}
	val, err := redis.Int64(reply, errDo)
	return val, err
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

func (c *AsynClient) HGet(hashID string, field string) (string, error) {
	conn := c.pool.Get()
	defer conn.Close()
	reply, errDo := conn.Do("HGET", hashID, field)
	if errDo == nil && reply == nil {
		return "", nil
	}
	val, err := redis.String(reply, errDo)
	return val, err
}

func (c *AsynClient) INCR(key string) (int, error) {
	conn := c.pool.Get()
	defer conn.Close()
	reply, errDo := conn.Do("INCR", key)
	if errDo == nil && reply == nil {
		return 0, nil
	}
	val, err := redis.Int(reply, errDo)
	return val, err
}

func (c *AsynClient) DECR(key string) (int, error) {
	conn := c.pool.Get()
	defer conn.Close()
	reply, errDo := conn.Do("DECR", key)
	if errDo == nil && reply == nil {
		return 0, nil
	}
	val, err := redis.Int(reply, errDo)
	return val, err
}

func (c *AsynClient) HGetAll(hashID string) (map[string]string, error) {
	conn := c.pool.Get()
	defer conn.Close()
	reply, err := redis.StringMap(conn.Do("HGetAll", hashID))
	return reply, err
}

func (c *AsynClient) HSet(hashID string, field string, val string) error {
	conn := c.pool.Get()
	defer conn.Close()
	_, err := conn.Do("HSET", hashID, field, val)
	return err
}

func (c *AsynClient) HMSet(args ...interface{}) error {
	conn := c.pool.Get()
	defer conn.Close()
	_, err := conn.Do("HMSET", args...)
	return err
}

func (c *AsynClient) Expire(key string, expire int64) error {
	conn := c.pool.Get()
	defer conn.Close()
	_, err := conn.Do("EXPIRE", key, expire)
	return err
}

func (c *AsynClient) Set(key string, val string) (string, error) {
	conn := c.pool.Get()
	defer conn.Close()
	val, err := redis.String(conn.Do("SET", key, val))
	return val, err
}

func (c *AsynClient) SetWithExpire(key string, val string, timeOutSeconds int) (string, error) {
	conn := c.pool.Get()
	defer conn.Close()
	val, err := redis.String(conn.Do("SET", key, val, "EX", timeOutSeconds))
	return val, err
}

func (c *AsynClient) GetTTL(key string) (int64, error) {
	conn := c.pool.Get()
	defer conn.Close()
	val, err := redis.Int64(conn.Do("TTL", key))
	return val, err
}

func (c *AsynClient) ZAdd(args ...interface{}) error {
	conn := c.pool.Get()
	defer conn.Close()
	_, err := conn.Do("ZADD", args...)
	return err
}

// list操作
func (c *AsynClient) LLen(key string) (int64, error) {
	conn := c.pool.Get()
	defer conn.Close()
	val, err := redis.Int64(conn.Do("LLEN", key))
	return val, err
}

func (c *AsynClient) RPopLPush(src, dst string) (string, error) {
	conn := c.pool.Get()
	defer conn.Close()
	val, err := redis.String(conn.Do("RPOPLPUSH", src, dst))
	return val, err
}

func (c *AsynClient) BRPop(key string) (string, error) {
	conn := c.pool.Get()
	defer conn.Close()
	val, err := redis.StringMap(conn.Do("BRPOP", key, defaultTimeout))
	if err != nil {
		return "", err
	} else {
		return val[key], nil
	}
}

func (c *AsynClient) RPop(key string) (string, error) {
	conn := c.pool.Get()
	defer conn.Close()
	val, err := redis.String(conn.Do("RPOP", key))
	return val, err
}

func (c *AsynClient) LPop(key string) (string, error) {
	conn := c.pool.Get()
	defer conn.Close()
	val, err := redis.String(conn.Do("LPOP", key))
	return val, err
}

func (c *AsynClient) RPush(key string, val string) (int64, error) {
	conn := c.pool.Get()
	defer conn.Close()
	ret, err := redis.Int64(conn.Do("RPUSH", key, val))
	if err != nil {
		return -1, err
	} else {
		return ret, nil
	}
}

func (c *AsynClient) LPush(key string, val string) (int64, error) {
	conn := c.pool.Get()
	defer conn.Close()
	ret, err := redis.Int64(conn.Do("LPUSH", key, val))
	if err != nil {
		return -1, err
	} else {
		return ret, nil
	}
}
