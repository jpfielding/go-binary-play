package main

import (
	"fmt"
	"time"
)

func main() {
	ch := make(chan struct{})
	start := time.Now()
	go func() {
		defer close(ch)
		<-time.After(time.Second * 5)
	}()
	fmt.Println("waiting")
	<-ch
	fmt.Printf("waited for %v\n", time.Since(start))
}
