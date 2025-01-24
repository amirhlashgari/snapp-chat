package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/amirhlashgari/snapp-chat/internal/service"
	store "github.com/amirhlashgari/snapp-chat/pkg/nats"
	pb "github.com/amirhlashgari/snapp-chat/proto"

	"github.com/nats-io/nats.go"
	"google.golang.org/grpc"
)

func main() {
	port := flag.Int("port", 50051, "The server port")
	natsURL := flag.String("nats", nats.DefaultURL, "NATS server URL")
	flag.Parse()

	// Create a listener on TCP (clients connect through grpc)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	nc, err := nats.Connect(*natsURL)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer nc.Close()

	jetStreamStore, err := store.NewJetStreamStore(nc)
	if err != nil {
		log.Fatalf("Failed to create JetStream store: %v", err)
	}

	s := grpc.NewServer()

	chatService := service.NewChatService(jetStreamStore)
	pb.RegisterChatServiceServer(s, chatService)

	log.Printf("Starting gRPC server on port %d", *port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
