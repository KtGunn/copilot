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
	"google.golang.org/grpc/metadata"
)

// /////////////////////////////////////////////////////////////
// STATE CLIENT
var StateClient *StateClientData

type StateClientData struct {
	client pb.AsyncClient
	conn   *grpc.ClientConn

	stream pb.Async_StatusUpdateClient
	mu     sync.RWMutex

	streamCtx    context.Context
	streamCancel context.CancelFunc
}

func NewStateClient() *StateClientData {
	return &StateClientData{}
}

func StartStateHandler(host string, port int, stop chan struct{}) error {

	var options []grpc.DialOption

	creds := grpc.WithTransportCredentials(insecure.NewCredentials())
	options = append(options, creds)

	var err error
	StateClient, err = CreateStateClient(host, port, options)

	if err != nil {
		return fmt.Errorf("could not initialize state client: %v", err)
	}

	// Start a background routine to ensure the stream stays alive
	go maintainStream(stop)

	return nil
}

/*
The strategy to maintain an active gRPC stream connection in state.go relies on
a continuous background monitoring loop combined with error-triggered stream invalidation:

1.  **Background Monitoring**:
    When initialized via `StartStateHandler`, a background goroutine is launched to
	run the `maintainStream` function.

2.  **Periodic Checks**:
    The `maintainStream` function runs an infinite loop that safely checks
	(using an `RWMutex`)whether the connection stream is `nil` every 1 second.

3.  **Automatic Recovery**:
    If the stream is found to be missing (`nil`), it calls `CreateUpdateStream` to establish
	a new one. In case of failure, it sleeps for 2 seconds as a backoff before trying again.

4.  **Error-Driven Invalidation**:
    When a payload fails to send inside `SendUpdate`, the code explicitly sets
	`StateClient.stream = nil` under a write lock. This automatically signals the
	`maintainStream` loop to recreate the connection on its next iteration.
*/

func maintainStream(stop chan struct{}) {
	for {
		select {
		case <-stop:
			return
		default:
		}

		StateClient.mu.RLock()
		hasStream := StateClient.stream != nil
		StateClient.mu.RUnlock()

		if !hasStream {
			if err := CreateUpdateStream(); err != nil {
				log.Printf("Failed to create update stream, retrying... (%v)", err)
				time.Sleep(2 * time.Second) // backoff before retrying
				continue
			}
			log.Println("Server for Update stream is up!")

		time.Sleep(1 * time.Second)
	}

	}
}

func CreateUpdateStream() error {
	stream, err := StateClient.client.StatusUpdate(StateClient.streamCtx)

	if err != nil {
		return err
	}

	StateClient.mu.Lock()
	StateClient.stream = stream
	StateClient.mu.Unlock()

	return nil
}

func CreateStateClient(host string, port int, options []grpc.DialOption) (*StateClientData, error) {

	hostPort := fmt.Sprintf("%s:%d", host, port)
	conn, err := grpc.NewClient(hostPort, options...)

	if err != nil {
		log.Fatalf("could not create a connection to %s; error: %v", hostPort, err)
	}

	client := pb.NewAsyncClient(conn)
	context, cancel := context.WithCancel(context.Background())

	md := metadata.Pairs("client-id", "user")
	context = metadata.NewOutgoingContext(context, md)

	return &StateClientData{
		client:       client,
		conn:         conn,
		streamCtx:    context,
		streamCancel: cancel,
	}, nil

}

func SendUpdate(msg pb.MyStatus) error {

	if StateClient == nil || StateClient.client == nil {
		return fmt.Errorf("StateClient is nil..")
	}

	StateClient.mu.RLock()
	stream := StateClient.stream
	StateClient.mu.RUnlock()

	if stream == nil {
		return fmt.Errorf("StateClient.stream is nil and currently reconnecting")
	}

	update := &pb.Status{
		State: msg,
	}

	if err := stream.Send(update); err != nil {
		log.Printf("Send failed, marking stream for reconnection: %v", err)

		StateClient.mu.Lock()
		StateClient.stream = nil
		StateClient.mu.Unlock()

		return err
	}

	return nil
}
