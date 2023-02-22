package main

import (
	"fmt"
	"time"
)

// simple way to emulate ordered responses
func main() {
	now := time.Now()
	defer func() { fmt.Printf("%v\n", time.Now().Sub(now)) }()
	queue := make(chan chan int, 100)
	go func() {
		for i := 0; i < 100; i++ {
			// do this outside the
			chint := make(chan int)
			queue <- chint
			go func(chint chan int, i int) {
				defer close(chint)
				<-time.After(5 * time.Second)
				chint <- i
			}(chint, i)
		}
		close(queue)
	}()
	for chint := range queue {
		now := time.Now()
		i := <-chint
		fmt.Printf("found %d %v\n", i, time.Since(now))
	}
}
