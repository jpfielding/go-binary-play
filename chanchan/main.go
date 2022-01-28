package main

import (
	"fmt"
	"time"
)

// simple way to emulate ordered responses
func main() {
	all := make(chan chan int, 100)
	for i := 0; i < 100; i++ {
		// do this outside the
		chint := make(chan int)
		all <- chint
		go func(chint chan int, i int) {
			defer close(chint)
			<-time.After(5 * time.Second)
			chint <- i
		}(chint, i)
	}
	close(all)
	for chint := range all {
		now := time.Now()
		i := <-chint
		fmt.Printf("found %d %v\n", i, time.Since(now))
	}
}
