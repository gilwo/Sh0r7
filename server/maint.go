package server

import (
	"log"
	"math/rand"
	"sync/atomic"
	"time"

	"github.com/gilwo/Sh0r7/store"
	"github.com/gin-gonic/gin"
)

var (
	processQueueChan chan (*maintQueueElem)
	quit             chan (any)
	done             chan (any)
	queueCountA      atomic.Int32
)

const (
	PROCESS_QUEUE_MAX = 60
)

type maintQueueElem struct {
	when time.Duration
}

func MaintQueueElem(when time.Duration) *maintQueueElem {
	return &maintQueueElem{
		when: when,
	}
}

func init() {
	processQueueChan = make(chan *maintQueueElem, PROCESS_QUEUE_MAX)
	quit = make(chan any)
	done = make(chan any)
}

func QueueMaintWork() {
	var when time.Duration
	if gin.Mode() == gin.DebugMode {
		// when = time.Duration(rand.Intn(5)) * time.Minute
		when = time.Duration(rand.Intn(5)) * time.Second
	} else {
		when = time.Duration(rand.Intn(5))*time.Hour + time.Duration(rand.Intn(60))*time.Minute
	}
	e := MaintQueueElem(when)

	if !addToQueue(e) {
		log.Printf("work not queued")
	}
}

func addToQueue(e *maintQueueElem) bool {
	if queueCountA.CompareAndSwap(PROCESS_QUEUE_MAX, PROCESS_QUEUE_MAX) {
		return false
	}
	if queueCountA.CompareAndSwap(0, 0) {
		go workProcessor()
	}
	processQueueChan <- e
	queueCountA.Add(1)
	log.Printf("queue have %d works\n", queueCountA.Load())
	return true
}

func workProcessor() {
	b := <-processQueueChan
	select {
	case <-time.After(b.when):
		log.Printf("maintainence triggered after %s\n", b.when)
		store.Maintainence()
	case <-quit:
		defer close(done)
		return
	case <-mainCtx.Done():
		log.Println("maintainence aborted")
		return
	}

	prev := queueCountA.Swap(queueCountA.Load() - 1)
	if prev > 0 {
		go workProcessor()
	}
}
