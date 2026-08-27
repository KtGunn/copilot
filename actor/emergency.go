// filepath: /home/rugrat/a_ktg.d/golang.d/copilot/actor/emergency.go
package main

import (
	"context"
	pb "coptest/proto"
	"fmt"
	"log"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var EmergencyClient *EmergencyClientData

type EmergencyClientData struct {
	client pb.AsyncClient
	conn   *grpc.ClientConn

	stream pb.Async_EmergencyUpdateClient
	mu     sync.RWMutex

	streamCtx    context.Context
	streamCancel context.CancelFunc
}

func NewEmergencyClient() *EmergencyClientData {
	return &EmergencyClientData{}
}

func StartEmergencyHandler(host string, port int, stop chan struct{}) error {

	var options []grpc.DialOption

	creds := grpc.WithTransportCredentials(insecure.NewCredentials())
	options = append(options, creds)

	var err error
	EmergencyClient, err = CreateEmergencyClient(host, port, options)

	if err != nil {
		return fmt.Errorf("could not initialize emergency client: %v", err)
	}

	go maintainEmergencyStream(stop)

	return nil
}

func maintainEmergencyStream(stop chan struct{}) {
	for {
		select {
		case <-stop:
			return
		default:
		}

		EmergencyClient.mu.RLock()
		hasStream := EmergencyClient.stream != nil
		EmergencyClient.mu.RUnlock()

		if !hasStream {
			if err := CreateEmergencyUpdateStream(); err != nil {
				log.Printf("Failed to create emergency update stream, retrying... (%v)", err)
				time.Sleep(2 * time.Second) // backoff before retrying
				continue
			}
			log.Println("Server for Emergency update stream is up!")
		}

		time.Sleep(1 * time.Second)
	}
}

func CreateEmergencyUpdateStream() error {
	stream, err := EmergencyClient.client.EmergencyUpdate(EmergencyClient.streamCtx)

	if err != nil {
		return err
	}

	EmergencyClient.mu.Lock()
	EmergencyClient.stream = stream
	EmergencyClient.mu.Unlock()

	return nil
}

func CreateEmergencyClient(host string, port int, options []grpc.DialOption) (*EmergencyClientData, error) {

	hostPort := fmt.Sprintf("%s:%d", host, port)
	conn, err := grpc.NewClient(hostPort, options...)

	if err != nil {
		log.Fatalf("could not create a connection to %s; error: %v", hostPort, err)
	}

	client := pb.NewAsyncClient(conn)
	context, cancel := context.WithCancel(context.Background())

	return &EmergencyClientData{
		client:       client,
		conn:         conn,
		streamCtx:    context,
		streamCancel: cancel,
	}, nil
}

func SendEmergencyUpdate(msg pb.MyEmergency) error {

	if EmergencyClient == nil || EmergencyClient.client == nil {
		return fmt.Errorf("EmergencyClient is nil..")
	}

	EmergencyClient.mu.RLock()
	stream := EmergencyClient.stream
	EmergencyClient.mu.RUnlock()

	if stream == nil {
		return fmt.Errorf("EmergencyClient.stream is nil and currently reconnecting")
	}

	emergency := &pb.Emergency{
		Wassup: msg,
	}

	if err := stream.Send(emergency); err != nil {
		log.Printf("Send failed, marking stream for reconnection: %v", err)

		EmergencyClient.mu.Lock()
		EmergencyClient.stream = nil
		EmergencyClient.mu.Unlock()

		return err
	}

	return nil
}
