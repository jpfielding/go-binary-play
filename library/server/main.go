// Package main implements a server for Library service.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"

	pb "go-binary-play/library/proto"

	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 50051, "The server port")
)

// taste test the local implementation and ensure it meets the interface
var _ pb.LibraryServer = &server{}

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedLibraryServer
}

// CHeckout implements library.LiberServer
func (s *server) Checkout(ctx context.Context, in *pb.CheckoutRequest) (*pb.CheckoutReply, error) {
	log.Printf("Received Request: %v", in.GetTitle())
	rnd := rand.Intn(10)
	threeDays := time.Now().Add(time.Hour * time.Duration(24*rnd))
	return &pb.CheckoutReply{AvailableOnDate: threeDays.Unix()}, nil
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterLibraryServer(s, &server{})
	log.Printf("library server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
