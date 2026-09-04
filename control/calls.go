package main

import (
	"context"
	pb "coptest/proto"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
)

type CallsServer struct {
	pb.UnimplementedCallsServer
}

func NewCallsServer() *CallsServer {
	return &CallsServer{}
}

// Send receives a StateQuery and returns a StateResponse
func (s *CallsServer) Send(ctx context.Context, req *pb.StateQuery) (*pb.StateResponse, error) {
	log.Printf("Received StateQuery: %s", req.Squery)

	// Respond to the query
	return &pb.StateResponse{
		Sresponse: "Received: " + req.Squery,
	}, nil
}

func LaunchCallsServer(port int, stop chan struct{}) error {
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterCallsServer(grpcServer, NewCallsServer())

	go func() {
		<-stop
		log.Println("Shutting down Calls gRPC server...")
		grpcServer.GracefulStop()
	}()

	log.Printf("Calls gRPC Server listening on port %d", port)
	return grpcServer.Serve(lis)
}
