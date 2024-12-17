package main

import (
	"log"
	"net"

	pb "llm-qa-system/backend-service/proto"

	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedMedicalQAServiceServer
}

func main() {
	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterMedicalQAServiceServer(s, &server{})

	log.Printf("Backend server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
