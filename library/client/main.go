// Package main implements a client for Library service.
package main

import (
	"context"
	"flag"
	"log"
	"time"

	pb "github.com/jpfielding/go-binary-play/library/proto"

	"google.golang.org/grpc"
)

var (
	addr  = flag.String("addr", "localhost:50051", "the address to connect to")
	title = flag.String("title", "Huckleberry Finn", "Book request")
)

func main() {
	flag.Parse()
	// Set up a connection to the server.
	conn, err := grpc.Dial(*addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewLibraryClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.Checkout(ctx, &pb.CheckoutRequest{Title: *title})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	availableOn := time.Unix(r.GetAvailableOnDate(), 0)
	log.Printf("Available in: %v", time.Until(availableOn))
}
