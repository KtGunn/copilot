package main

import (
	"context"
	"log"
	"time"
	"fmt"
	
	pb "coptest/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type CallerClient struct {
	client   pb.CallsClient
	conn *grpc.ClientConn

	SendChan chan *pb.StateQuery
	cientCtx context.Context
}


func NewCallerClient(host string, port int) *CallerClient {

	var options[]grpc.DialOption

	creds := grpc.WithTransportCredentials(insecure.NewCredentials())
	options = append(options, creds)
	
	hostPort := fmt.Sprintf("%s:%d", host, port)
	conn, err := grpc.NewClient(hostPort, options...)

	if err != nil {
		return nil
	}
	
	cm := &CallerClient{
		client:   pb.NewCallsClient(conn),
		SendChan: make(chan *pb.StateQuery, 100),
	}

	go cm.process()
	return cm
}


func (cm *CallerClient) process() {
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
