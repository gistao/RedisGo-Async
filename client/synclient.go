package client

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

func (c *SynClient) Exists(key string) (int64, error) {
	conn := c.pool.Get()
	defer conn.Close()
	reply, errDo := conn.Do("EXISTS", key)
	if errDo == nil && reply == nil {
		return 0, nil
	}
	val, err := redis.Int64(reply, errDo)
	return val, err
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

func (c *SynClient) HGet(hashID string, field string) (string, error) {
	conn := c.pool.Get()
	defer conn.Close()
	reply, errDo := conn.Do("HGET", hashID, field)
	if errDo == nil && reply == nil {
		return "", nil
	}
	val, err := redis.String(reply, errDo)
	return val, err
}

func (c *SynClient) INCR(key string) (int, error) {
	conn := c.pool.Get()
	defer conn.Close()
	reply, errDo := conn.Do("INCR", key)
	if errDo == nil && reply == nil {
		return 0, nil
	}
	val, err := redis.Int(reply, errDo)
	return val, err
}

func (c *SynClient) DECR(key string) (int, error) {
	conn := c.pool.Get()
	defer conn.Close()
	reply, errDo := conn.Do("DECR", key)
	if errDo == nil && reply == nil {
		return 0, nil
	}
	val, err := redis.Int(reply, errDo)
	return val, err
}

func (c *SynClient) HGetAll(hashID string) (map[string]string, error) {
	conn := c.pool.Get()
	defer conn.Close()
	reply, err := redis.StringMap(conn.Do("HGetAll", hashID))
	return reply, err
}

func (c *SynClient) HSet(hashID string, field string, val string) error {
	conn := c.pool.Get()
	defer conn.Close()
	_, err := conn.Do("HSET", hashID, field, val)
	return err
}

func (c *SynClient) HMSet(args ...interface{}) error {
	conn := c.pool.Get()
	defer conn.Close()
	_, err := conn.Do("HMSET", args...)
	return err
}

func (c *SynClient) Expire(key string, expire int64) error {
	conn := c.pool.Get()
	defer conn.Close()
	_, err := conn.Do("EXPIRE", key, expire)
	return err
}

func (c *SynClient) Set(key string, val string) (string, error) {
	conn := c.pool.Get()
	defer conn.Close()
	val, err := redis.String(conn.Do("SET", key, val))
	return val, err
}

func (c *SynClient) SetWithExpire(key string, val string, timeOutSeconds int) (string, error) {
	conn := c.pool.Get()
	defer conn.Close()
	val, err := redis.String(conn.Do("SET", key, val, "EX", timeOutSeconds))
	return val, err
}

func (c *SynClient) GetTTL(key string) (int64, error) {
	conn := c.pool.Get()
	defer conn.Close()
	val, err := redis.Int64(conn.Do("TTL", key))
	return val, err
}

func (c *SynClient) ZAdd(args ...interface{}) error {
	conn := c.pool.Get()
	defer conn.Close()
	_, err := conn.Do("ZADD", args...)
	return err
}

// list操作
func (c *SynClient) LLen(key string) (int64, error) {
	conn := c.pool.Get()
	defer conn.Close()
	val, err := redis.Int64(conn.Do("LLEN", key))
	return val, err
}

func (c *SynClient) RPopLPush(src, dst string) (string, error) {
	conn := c.pool.Get()
	defer conn.Close()
	val, err := redis.String(conn.Do("RPOPLPUSH", src, dst))
	return val, err
}

func (c *SynClient) BRPop(key string) (string, error) {
	conn := c.pool.Get()
	defer conn.Close()
	val, err := redis.StringMap(conn.Do("BRPOP", key, defaultTimeout))
	if err != nil {
		return "", err
	} else {
		return val[key], nil
	}
}

func (c *SynClient) RPop(key string) (string, error) {
	conn := c.pool.Get()
	defer conn.Close()
	val, err := redis.String(conn.Do("RPOP", key))
	return val, err
}

func (c *SynClient) LPop(key string) (string, error) {
	conn := c.pool.Get()
	defer conn.Close()
	val, err := redis.String(conn.Do("LPOP", key))
	return val, err
}

func (c *SynClient) RPush(key string, val string) (int64, error) {
	conn := c.pool.Get()
	defer conn.Close()
	ret, err := redis.Int64(conn.Do("RPUSH", key, val))
	if err != nil {
		return -1, err
	} else {
		return ret, nil
	}
}

func (c *SynClient) LPush(key string, val string) (int64, error) {
	conn := c.pool.Get()
	defer conn.Close()
	ret, err := redis.Int64(conn.Do("LPUSH", key, val))
	if err != nil {
		return -1, err
	} else {
		return ret, nil
	}
}
