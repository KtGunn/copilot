package main

import (
	pb "coptest/proto"
	"log"
	"sync"
)

type ClientEmergency struct {
	ClientID  string
	Emergency *pb.Emergency
}

type EmergencyModule struct {
	latest map[string]*pb.Emergency

	mu            sync.RWMutex
	EmergencyChan chan ClientEmergency
}

// NewEmergencyModule creates a EmergencyModule with a fixed size buffer for each client
func NewEmergencyModule() *EmergencyModule {
	em := &EmergencyModule{
		latest:        make(map[string]*pb.Emergency),
		EmergencyChan: make(chan ClientEmergency, 100),
	}
	go em.listen()
	return em
}

func (em *EmergencyModule) listen() {
	for ce := range em.EmergencyChan {
		log.Println("lis EM", ce.ClientID, ce.Emergency.GetWassup())
		em.mu.Lock()

		em.latest[ce.ClientID] = ce.Emergency
		em.mu.Unlock()
	}
}

// GetLatestEmergency returns the latest emergency message received for a client ID
func (em *EmergencyModule) GetLatestEmergency(clientID string) *pb.Emergency {
	em.mu.RLock()
	defer em.mu.RUnlock()

	return em.latest[clientID]
}
