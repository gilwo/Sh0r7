package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Printf("starting...")
	processChan = make(chan *somethingToProcess, PROCESS_QUEUE_MAX)
	quit = make(chan any)
	done = make(chan any)
	// queueCount = 0
	for i := 0; i < 1000; i++ {
		b := SomethingToProcess(time.Duration(time.Second), fmt.Sprintf("%d..", i))
		// fmt.Printf("%d (%s) queueing: %s  -- in queue (%d)\n", i, time.Now(), b, queueCount)
		<-time.After(time.Millisecond * 10)
		if !process(b) {
			// fmt.Printf("%d (%s) queueing FAILED: %s  -- in queue (%d)\n", i, time.Now(), b, queueCount)
			skippedCount += 1
		}
	}
	time.Sleep(8 * time.Second)
	close(quit)
	select {
	case <-done:
		fmt.Println("done")
	case <-time.After(2 * time.Second):
		fmt.Println("timeout")
	}
	fmt.Printf("processed: %d\nskipped: %d\n", processedCount, skippedCount)
}

var (
	processChan    chan (*somethingToProcess)
	quit           chan (any)
	done           chan (any)
	queueCount     int
	processedCount int
	skippedCount   int
)

const (
	PROCESS_QUEUE_MAX = 20
)

type somethingToProcess struct {
	when time.Duration
	what string
}

func SomethingToProcess(when time.Duration, what string) *somethingToProcess {
	return &somethingToProcess{
		when: when,
		what: what,
	}
}
func (s somethingToProcess) String() string {
	return fmt.Sprintf("want to process <%s> in <%s>", s.what, s.when)
}

func processor() {
	b := <-processChan
	select {
	case <-time.After(b.when):
	case <-quit:
		defer close(done)
		return
	}
	fmt.Printf("(count: %d) (%s) processing %s\n", queueCount, time.Now(), b.what)
	queueCount -= 1
	processedCount += 1
	if queueCount > 0 {
		go processor()
	}
}

func process(a *somethingToProcess) bool {

	// if queueCount+1 > PROCESS_QUEUE_MAX {
	// 	return false
	// }

	// processChan <- a
	// queueCount += 1

	// go processor()
	// return true
	if queueCount < PROCESS_QUEUE_MAX {
		if queueCount == 0 {
			go processor()
		}
		processChan <- a
		queueCount += 1
		return true
	}
	return false
}
