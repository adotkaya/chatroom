package main

import (
	"fmt"
	"math/rand"
	"time"
)

type Throttler struct {
	inputCH  chan *ReqMsg
	outputCH chan *ReqMsg
	rate     time.Duration
}

func NewThrottler(msgPerSecond int, exit chan struct{}) *Throttler {
	rate := time.Second / time.Duration(msgPerSecond)
	t := &Throttler{
		inputCH:  make(chan *ReqMsg),
		outputCH: make(chan *ReqMsg),
		rate:     rate,
	}
	go t.leak(exit)
	return t
}

func (t *Throttler) leak(exit chan struct{}) {
	ticker := time.NewTicker(t.rate)
	defer func() {
		ticker.Stop()
		close(t.outputCH)
	}()
	for {
		select {
		case <-exit:
			close(t.inputCH)
			for msg := range t.inputCH {
				fmt.Printf("draining msg = %v\n", msg)
			}
			fmt.Println("draining done")
			return
		case <-ticker.C:
			select {
			case msg := <-t.inputCH:
				t.outputCH <- msg
			default:
				if rand.Intn(10) < 1 {
					fmt.Println("no msg in throttler --> skipped")
				}
			}
		}
	}
}
