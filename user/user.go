package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	// Handle graceful shutdown in a goroutine
	stopSignal := make(chan os.Signal, 1)
	signal.Notify(stopSignal, os.Interrupt, syscall.SIGTERM)

	// Create and initialize state module with buffer size of 100
	sm := NewStateModule(100)

	stop := make(chan struct{})

	go func() {
		log.Println(" Will call LAUNCH SERVER")
		if err := LaunchServer(50052, stop, sm); err != nil {
			log.Println(" Ooops", err)
		}
	}()

	go func() {
		for {
			time.Sleep(5 * time.Second)
			log.Println(" .. Gettin latests:",sm.GetLatestState("user"))
		}
	}()

	log.Println("User Server is running on port 50052...")
	<-stopSignal

	log.Println("User Server has been stopped.")
}
