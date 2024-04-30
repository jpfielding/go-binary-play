package main

import (
	"fmt"

	"github.com/talostrading/sonic"
)

func main() {
	// Create an IO object which can execute asynchronous operations on the
	// current goroutine.
	ioc := sonic.MustIO()
	defer ioc.Close()

	// Create 10 connections. Each connection reads a message into it's
	// buffer and then closes.
	for i := 0; i < 10; i++ {
		conn, _ := sonic.Dial(ioc, "tcp", "localhost:8080")

		b := make([]byte, 128)
		conn.AsyncRead(b, func(err error, n int) {
			if err != nil {
				fmt.Printf("could not read from %d err=%v\n", i, err)
			} else {
				b = b[:n]
				fmt.Println("got=", string(b))
				conn.Close()
			}
		})
	}

	// Execute all pending reads scheduled in the for-loop, then exit.
	ioc.RunPending()
}
