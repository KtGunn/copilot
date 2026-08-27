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
	StartEmergencyHandler("localhost", 50052, stop)

	go func(stop chan struct{}) {

		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		everyFifth := 0

		for {
			select {

			case <-stop:
				return

			case <-ticker.C:

				everyFifth++
				if everyFifth%5 == 0 {
					log.Println("Sending emergency update")
					if err := SendEmergencyUpdate(pb.MyEmergency_ALLOK); err != nil {
						log.Printf("Failed to send 'emergency update': %v", err)
					}
					continue
				}

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
