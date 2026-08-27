package main

import (
	pb "coptest/proto"
	"log"
	"sync"
)

type ClientStatus struct {
	ClientID string
	Status   *pb.Status
}

type StateModule struct {
	buffers    map[string][]*pb.Status
	bufferSize int
	heads      map[string]int
	mu         sync.RWMutex
	StatusChan chan ClientStatus
}

// NewStateModule creates a StateModule with a fixed size buffer for each client
func NewStateModule(size int) *StateModule {
	sm := &StateModule{
		buffers:    make(map[string][]*pb.Status),
		bufferSize: size,
		heads:      make(map[string]int),
		StatusChan: make(chan ClientStatus, 100),
	}
	go sm.listen()
	return sm
}

func (sm *StateModule) listen() {
	for cs := range sm.StatusChan {
		log.Println("Received status update for client:", cs.ClientID, "Status:", cs.Status.GetState())
		sm.mu.Lock()
		if sm.buffers[cs.ClientID] == nil {
			sm.buffers[cs.ClientID] = make([]*pb.Status, sm.bufferSize)
			sm.heads[cs.ClientID] = 0
		}

		head := sm.heads[cs.ClientID]
		sm.buffers[cs.ClientID][head] = cs.Status
		sm.heads[cs.ClientID] = (head + 1) % sm.bufferSize
		sm.mu.Unlock()
	}
}

// GetLatestState returns the latest status message received for a client ID
func (sm *StateModule) GetLatestState(clientID string) *pb.Status {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if sm.buffers[clientID] == nil {
		return nil
	}

	head := sm.heads[clientID]
	idx := head - 1
	if idx < 0 {
		idx = sm.bufferSize - 1
	}
	return sm.buffers[clientID][idx]
}
