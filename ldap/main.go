package main

import (
	"flag"
	"fmt"
	"net"
	"os"
)

type args struct {
	hostname string
	port     string
}

func (a *args) Parse() {
	flag.StringVar(&a.hostname, "hostname", "localhost", "the bind host")
	flag.StringVar(&a.port, "port", "389", "the bind port")
	flag.Parse()
}

func main() {
	args := &args{}
	args.Parse()
	// listen
	l, err := net.Listen("tcp", fmt.Sprintf("%s:%s", args.hostname, args.port))
	if err != nil {
		panic(err)
	}
	// Close the listener when the application closes.
	defer l.Close()
	fmt.Printf("Listening on %s:%s", args.hostname, args.port)
	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		// Handle connections in a new goroutine.
		go Serve(conn)
	}
}
