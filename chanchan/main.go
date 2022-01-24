package main

import (
	"fmt"
	"math/rand"
	"time"
)

// simple way to emulate ordered responses
func main() {
	all := make(chan chan int, 100)
	for i := 0; i < 100; i++ {
		// do this outside the
		chint := make(chan int)
		all <- chint
		// min time so we can be sure anything was consumed by concurrency
		min := 10 * time.Millisecond
		// some variation
		maxRnd := rand.Intn(4) * int(time.Second)
		go func(chint chan int, ii int) {
			defer close(chint)
			<-time.After(min + time.Duration(maxRnd))
			chint <- ii
		}(chint, i)
	}
	close(all)
	for chint := range all {
		now := time.Now()
		i := <-chint
		fmt.Printf("found %d %v\n", i, time.Since(now))
	}
}
