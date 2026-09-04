package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	pb "coptest/proto"
)

func main() {

	log.SetOutput(os.Stdout)

	// Handle graceful shutdown in a goroutine
	stopSignal := make(chan os.Signal, 1)
	signal.Notify(stopSignal, os.Interrupt, syscall.SIGTERM)

	// Create and initialize state module with buffer size of 100
	sm := NewStateModule(10)
	em := NewEmergencyModule()

	cl := NewCallerClient("localhost", 50054)
	
	stop := make(chan struct{})

	go func() {
		log.Println(" Will call LAUNCH SERVER")
		if err := LaunchServer(50052, stop, sm, em); err != nil {
			log.Println(" Ooops", err)
		}
	}()

	go func() {
		for {
			time.Sleep(7 * time.Second)
			log.Println(" .. Gettin EMG:", em.GetLatestEmergency("user"))
		}
	}()

	go func() {
		for {
			time.Sleep(5 * time.Second)
			log.Println(" .. Gettin STATUS:", sm.GetLatestState("user"))

			time.Sleep(5 * time.Second)
			log.Println(" .. __ Calling Squery:")
			cl.SendChan <- &pb.StateQuery{
				Squery: "Hello! It's me.",
			}
		}
	}()

	log.Println("User Server is running on port 50052...")
	<-stopSignal

	log.Println("User Server has been stopped.")
}
