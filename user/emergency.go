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
	buffers       map[string][]*pb.Emergency
	bufferSize    int
	heads         map[string]int
	mu            sync.RWMutex
	EmergencyChan chan ClientEmergency
}

// NewEmergencyModule creates a EmergencyModule with a fixed size buffer for each client
func NewEmergencyModule(size int) *EmergencyModule {
	em := &EmergencyModule{
		buffers:       make(map[string][]*pb.Emergency),
		bufferSize:    size,
		heads:         make(map[string]int),
		EmergencyChan: make(chan ClientEmergency, 100),
	}
	go em.listen()
	return em
}

func (em *EmergencyModule) listen() {
	for ce := range em.EmergencyChan {
		log.Println("Received emergency update for client:", ce.ClientID, "Emergency:", ce.Emergency.GetWassup())
		em.mu.Lock()
		if em.buffers[ce.ClientID] == nil {
			em.buffers[ce.ClientID] = make([]*pb.Emergency, em.bufferSize)
			em.heads[ce.ClientID] = 0
		}

		head := em.heads[ce.ClientID]
		em.buffers[ce.ClientID][head] = ce.Emergency
		em.heads[ce.ClientID] = (head + 1) % em.bufferSize
		em.mu.Unlock()
	}
}

// GetLatestEmergency returns the latest emergency message received for a client ID
func (em *EmergencyModule) GetLatestEmergency(clientID string) *pb.Emergency {
	em.mu.RLock()
	defer em.mu.RUnlock()

	if em.buffers[clientID] == nil {
		return nil
	}

	head := em.heads[clientID]
	idx := head - 1
	if idx < 0 {
		idx = em.bufferSize - 1
	}
	return em.buffers[clientID][idx]
}
