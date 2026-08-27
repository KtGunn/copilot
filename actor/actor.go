package main

import (
	pb "coptest/proto"
	"log"
	"math/rand"
	"time"
)

func main() {
	log.Println("Hello")

	stop := make(chan struct{})

	StartStateHandler("localhost", 50052, stop)


	go func(stop chan struct{}) {

		ticker := time.NewTicker(2 * time.Second)
		for {
			select {
				
			case <-stop:
				return
				
			case <-ticker.C:
				
				state := pb.MyStatus_IMOK
				if rand.Intn(2) == 1 {
					state = pb.MyStatus_NOTOK
				}
				
				log.Printf("Sending status update: %s", state)
				
				if err := SendUpdate(state); err != nil {
					log.Printf("Failed to send 'update': %v", err)
				}
			}
		}
		
	}(stop)

	select {}
}
