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

	mv := NewMover(100, 1.0)

	go func() {
		LaunchCallsServer(50054, stop)
	}()

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
					emergencyOptions := []pb.MyEmergency{
						pb.MyEmergency_ALLOK,
						pb.MyEmergency_FIRE,
						pb.MyEmergency_EVAC,
					}
					emergency := emergencyOptions[rand.Intn(len(emergencyOptions))]
					if err := SendEmergencyUpdate(emergency); err != nil {
						log.Printf("Failed to send 'emergency update': %v", err)
					}

					log.Println(mv)
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

	mv.MoveTo(50)
	select {}
}
