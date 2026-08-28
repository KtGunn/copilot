package main

import (
	pb "coptest/proto"
	"log"
	"sync"
)

type ItineraryModule struct {
	itineraries   map[string]*pb.Journey
	mu            sync.RWMutex
	ItineraryChan chan ClientItinerary
}

type ClientItinerary struct {
	ClientID  string
	Itinerary *pb.Journey
}

// NewItineraryModule creates an ItineraryModule with a fixed size buffer for each client
func NewItineraryModule(size int) *ItineraryModule {
	im := &ItineraryModule{
		itineraries:   make(map[string]*pb.Journey),
		ItineraryChan: make(chan ClientItinerary, 100),
	}
	go im.listen()
	return im
}

func (im *ItineraryModule) listen() {
	for ci := range im.ItineraryChan {
		log.Println("Received itinerary update for client:", ci.ClientID, "Itinerary:", ci.Itinerary)
		im.mu.Lock()
		im.itineraries[ci.ClientID] = ci.Itinerary
		im.mu.Unlock()
	}
}

// GetLatestItinerary returns the latest itinerary message received for a client ID
func (im *ItineraryModule) GetLatestItinerary(clientID string) *pb.Journey {
	im.mu.RLock()
	defer im.mu.RUnlock()

	return im.itineraries[clientID]
}
