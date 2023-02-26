package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"sync/atomic"
	"time"

	"github.com/codingconcepts/semaphore"
	"github.com/codingconcepts/throttle"
)

var (
	requestsMade uint64
)

func main() {
	n := flag.Int("n", 200, "number of requests to run, can't be less than c")
	c := flag.Int("c", 1, "number of workers to run concurrently, can't be more than n")
	q := flag.Int64("q", 0, "number of queries to run per second")
	z := flag.Duration("z", 0, "duration of test")
	flag.Parse()

	if *n < *c {
		flag.Usage()
		os.Exit(2)
	}

	thr := throttle.New(*q, time.Second)
	sem := semaphore.New(*c)

	if *z > 0 {
		thr.DoFor(context.Background(), *z, func() { sem.Run(makeRequest) })
	} else {
		thr.Do(context.Background(), *n, func() { sem.Run(makeRequest) })
	}

	sem.Wait()
	fmt.Println(requestsMade)
}

func makeRequest() {
	atomic.AddUint64(&requestsMade, 1)
}
