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
		go func(chint chan int, ii int) {
			defer close(chint)
			<-time.After(time.Millisecond * time.Duration(rand.Intn(1000)))
			chint <- ii
		}(chint, i)
	}
	close(all)
	for chint := range all {
		fmt.Printf("found %d\n", <-chint)
	}
}
