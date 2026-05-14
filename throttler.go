package main

import (
	"fmt"
	"math/rand"
	"time"
)

type Throttler struct {
	inputCH chan *ReqMsg
	output  chan *ReqMsg
	rate    time.Duration
}

func NewThrottler(msgPerSecond int, exit chan struct{}) *Throttler {
	rate := time.Second / time.Duration(msgPerSecond)
	t := &Throttler{
		inputCH: make(chan *ReqMsg),
		output:  make(chan *ReqMsg),
		rate:    rate,
	}
	go t.leak(exit)
	return t
}

func (t *Throttler) leak(exit chan struct{}) {
	ticker := time.NewTicker(t.rate)
	for {
		select {
		case <-exit:
			return
		case <-ticker.C:
			select {
			case msg := <-t.inputCH:
				t.output <- msg
			default:
				if rand.Intn(10) < 1 {
					fmt.Println("no msg in throttler --> skipped")
				}
			}
		}
	}
}
