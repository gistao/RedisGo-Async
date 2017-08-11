package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/gistao/RedisGo-Async/client"
	"log"
	"runtime"
	"time"
)

// redis task function type def
type redisTask func(id string, signal chan int, cli *client.AsynClient, iterations int)

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
	taskSpec{doDel, "DEL"},
}

// workers option.  default is equiv to -w=10000 on command line
var workers = flag.Int("w", 10000, "number of concurrent workers")

// opcnt option.  default is equiv to -n=20 on command line
var opcnt = flag.Int("n", 1000, "number of task iterations per worker")

// ----------------------------------------------------------------------------
// benchmarker
// ----------------------------------------------------------------------------

func main() {
	// DEBUG
	runtime.GOMAXPROCS(2)

	log.SetPrefix("[RedisGo-Async|async bench] ")
	flag.Parse()

	fmt.Printf("\n\n== Bench RedisGo-Async == %d goroutines 1 AsyncClient  -- %d opts each --- \n\n", *workers, *opcnt)

	for _, task := range tasks {
		benchTask(task, *opcnt, *workers, true)
	}

}

func benchTask(taskspec taskSpec, iterations int, workers int, printReport bool) (delta time.Duration, err error) {

	signal := make(chan int, workers)

	rdc := client.GetAsynClient("127.0.0.1:6379")
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

func doSet(id string, signal chan int, cli *client.AsynClient, cnt int) {
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

func doGet(id string, signal chan int, cli *client.AsynClient, cnt int) {
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

func doDel(id string, signal chan int, cli *client.AsynClient, cnt int) {
	key := "set-" + id
	for i := 0; i < cnt; i++ {
		_, err := cli.Del(key)
		if err != nil {
			log.Println(err)
		}
	}
	signal <- 1
}
