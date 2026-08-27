package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "coptest/proto"
)

type server struct {
	pb.UnimplementedAsyncServer
	tracker *StatusTracker
}

func (s *server) StatusUpdate(stream grpc.ClientStreamingServer[pb.Status, emptypb.Empty]) error {
	log.Printf("Client connected to user server. Receiving status updates...")

	for {
		statusMsg, err := stream.Recv()
		if err != nil {
			log.Printf("Client disconnected or error: %v", err)
			s.tracker.SetError(err)
			return stream.SendAndClose(&emptypb.Empty{})
		}

		s.tracker.SetLatest(statusMsg)
		log.Printf("Received status from actor: %v", statusMsg.GetState())
	}
}

func main() {

	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	tracker := NewStatusTracker()
	pb.RegisterAsyncServer(
		s, &server{
			tracker: tracker,
		},
	)

	// Handle graceful shutdown in a goroutine
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-stop
		log.Println("Shutting down user server...")

		// Force shutdown after timeout
		done := make(chan struct{})
		go func() {
			s.GracefulStop()
			close(done)
		}()

		select {
		case <-done:
			log.Println("Server stopped gracefully")
		case <-time.After(2 * time.Second):
			log.Println("Timeout reached, forcing shutdown")
			s.Stop()
		}
	}()
	log.Println("User Server is running on port 50052...")

	if err := s.Serve(lis); err != nil && err != grpc.ErrServerStopped {
		log.Fatalf("failed to serve: %v", err)
	}
}
