package main

import (
	pb "coptest/proto"
	"fmt"
	"io"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)


var Bridge *BridgeServerData

type BridgeServerData struct {
	pb.UnimplementedAsyncServer
	StateModule     *StateModule
	EmergencyModule *EmergencyModule
}

func NewBridgeServer(sm *StateModule, em *EmergencyModule) *BridgeServerData {
	return &BridgeServerData{
		StateModule:     sm,
		EmergencyModule: em,
	}
}


func LaunchServer(port int, stop chan struct{}, sm *StateModule, em *EmergencyModule) error {

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	Bridge = NewBridgeServer(sm, em)
	pb.RegisterAsyncServer(grpcServer, Bridge)

	go func() {
		<-stop
		log.Println("Shutting down gRPC server...")
		grpcServer.GracefulStop()
	}()

	log.Printf("gRPC Server on port %d", port)
	return grpcServer.Serve(lis)
}


func (s *BridgeServerData) GetClientIdentity(stream grpc.ServerStream) string {

	md, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		return "unknown"
	}

	// Try to get client identity from metadata
	if clientID := md.Get("client-id"); len(clientID) > 0 {
		return clientID[0]
	}

	// Fallback to user-agent or other metadata
	if userAgent := md.Get("user-agent"); len(userAgent) > 0 {
		return userAgent[0]
	}

	return "unknown"
}


func (s *BridgeServerData) StatusUpdate(stream pb.Async_StatusUpdateServer) error {

	clientID := s.GetClientIdentity(stream)
	log.Printf("Started StatusUpdate stream for client: %s\n", clientID)

	for {

		status, err := stream.Recv()

		if err == io.EOF {
			return stream.SendAndClose(&emptypb.Empty{})
		}
		if err != nil {
			log.Printf("Error receiving Status from %s: %v", clientID, err)
			return err
		}

		if s.StateModule != nil {
			s.StateModule.StatusChan <- ClientStatus{
				ClientID: clientID,
				Status:   status,
			}
		}
	}
}

func (s *BridgeServerData) EmergencyUpdate(stream pb.Async_EmergencyUpdateServer) error {

	clientID := s.GetClientIdentity(stream)
	log.Printf("Started Emergency Update stream for client: %s\n", clientID)

	for {
		emergency, err := stream.Recv()

		if err == io.EOF {
			return stream.SendAndClose(&emptypb.Empty{})
		}
		if err != nil {
			log.Printf("Error receiving Emergency: %v", err)
			return err
		}

		if s.EmergencyModule != nil {
			s.EmergencyModule.EmergencyChan <- ClientEmergency{
				ClientID:  clientID,
				Emergency: emergency,
			}
		}
	}
}
