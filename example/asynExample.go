package main

import (
	"common/RedisGo-Async/redis"
	"errors"
	"flag"
	"fmt"
	"log"
	"runtime"
	"sync"
	"time"
)

type redisClient struct {
	//pool *redis.Pool
	pool    *redis.AsyncPool
	Address string
}

var (
	redisMap map[string]*redisClient
	mapMutex *sync.RWMutex
)

// redis task function type def
type redisTask func(id string, signal chan int, cli *redisClient, iterations int)

// task info
type taskSpec struct {
	task redisTask
	name string
}

// array of Tasks to run in sequence
// Add a task to the list to bench to the runner.
// Tasks are run in sequence.
var tasks = []taskSpec{
	taskSpec{doSet, "SET"},
	taskSpec{doGet, "GET"},
}

// workers option.  default is equiv to -w=10000 on command line
var workers = flag.Int("w", 10000, "number of concurrent workers")

// opcnt option.  default is equiv to -n=20 on command line
var opcnt = flag.Int("n", 100, "number of task iterations per worker")

// ----------------------------------------------------------------------------
// benchmarker
// ----------------------------------------------------------------------------

func init() {
	redisMap = make(map[string]*redisClient)
	mapMutex = new(sync.RWMutex)
}

func main() {
	// DEBUG
	runtime.GOMAXPROCS(2)

	log.SetPrefix("[RedisGo-Async|example] ")
	flag.Parse()

	fmt.Printf("\n\n== Bench RedisGo-Async == %d goroutines 1 AsyncClient  -- %d opts each --- \n\n", *workers, *opcnt)

	for _, task := range tasks {
		benchTask(task, *opcnt, *workers, true)
	}

}

func benchTask(taskspec taskSpec, iterations int, workers int, printReport bool) (delta time.Duration, err error) {

	signal := make(chan int, workers)

	rdc := getRedisClient("127.0.0.1:6379")
	if rdc == nil {
		log.Println("Error creating client for worker")
		return -1, errors.New("Error creating client for worker")
	}

	t0 := time.Now()
	for i := 0; i < workers; i++ {
		id := fmt.Sprintf("%d", i)
		go taskspec.task(id, signal, rdc, iterations)
	}

	// wait for completion
	for i := 0; i < workers; i++ {
		<-signal
	}
	delta = time.Since(t0)

	if printReport {
		report(taskspec.name, workers, delta, iterations*workers)
	}

	return
}

func report(cmd string, workers int, delta time.Duration, cnt int) {
	log.Printf("---\n")
	log.Printf("cmd: %s\n", cmd)
	log.Printf("%d goroutines 1 asyncClient %d iterations of %s in %d msecs\n", workers, cnt, cmd, delta/time.Millisecond)
	log.Printf("---\n\n")
}

// ----------------------------------------------------------------------------
// redis tasks
// ----------------------------------------------------------------------------

func doSet(id string, signal chan int, cli *redisClient, cnt int) {
	key := "set-" + id
	value := "foo"
	for i := 0; i < cnt; i++ {
		//log.Println("set", key, value)
		_, err := cli.Set(key, value)
		//log.Println("set over", key, value)
		if err != nil {
			log.Println("set", err)
		}
	}
	signal <- 1
}
func doGet(id string, signal chan int, cli *redisClient, cnt int) {
	key := "set-" + id
	for i := 0; i < cnt; i++ {
		//	log.Println("get", key)
		_, err := cli.Get(key)
		if err != nil {
			log.Println(err)
		}
	}
	signal <- 1
}

/*func newPool(addr string) *redis.Pool {
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
}*/

func newAsyncPool(addr string) *redis.AsyncPool {
	return &redis.AsyncPool{
		Dial: func() (redis.AsynConn, error) {
			c, err := redis.AsyncDial("tcp", addr)
			return c, err
		},
	}
}

func getRedisClient(address string) *redisClient {
	var redis *redisClient
	var mok bool
	mapMutex.RLock()
	redis, mok = redisMap[address]
	mapMutex.RUnlock()
	if !mok {
		mapMutex.Lock()
		redis, mok = redisMap[address]
		if !mok {
			//redis = &redisClient{Address: address, pool: newPool(address)}
			redis = &redisClient{Address: address, pool: newAsyncPool(address)}
			redisMap[address] = redis
		}
		mapMutex.Unlock()
	}
	return redis
}

func (rc *redisClient) Get(key string) (string, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	reply, errDo := conn.Do("GET", key)
	if errDo == nil && reply == nil {
		return "", nil
	}
	val, err := redis.String(reply, errDo)
	return val, err
}

func (rc *redisClient) Set(key string, val string) (string, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	val, err := redis.String(conn.Do("SET", key, val))
	return val, err
}
