package main

import (
	"context"
	pb "coptest/proto"
	"log"
	"time"

	"google.golang.org/grpc"
)

type CallerModule struct {
	client   pb.CallsClient
	SendChan chan *pb.StateQuery
}

func NewCallerModule(conn grpc.ClientConnInterface) *CallerModule {
	cm := &CallerModule{
		client:   pb.NewCallsClient(conn),
		SendChan: make(chan *pb.StateQuery, 100),
	}
	go cm.process()
	return cm
}

func (cm *CallerModule) process() {
	for req := range cm.SendChan {
		// Use a timeout for each outbound request
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		res, err := cm.client.Send(ctx, req)
		if err != nil {
			log.Printf("Send failed: %v", err)
		} else {
			log.Printf("Send succeeded: %v", res)
		}

		cancel()
	}
}
