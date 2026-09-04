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
	latest     map[string]*pb.Status
	mu         sync.RWMutex
	StatusChan chan ClientStatus
}

// NewStateModule creates a StateModule with a fixed size buffer for each client
func NewStateModule(size int) *StateModule {
	sm := &StateModule{
		latest:     make(map[string]*pb.Status),
		StatusChan: make(chan ClientStatus, 100),
	}
	go sm.listen()
	return sm
}

func (sm *StateModule) listen() {
	for cs := range sm.StatusChan {

		mo := cs.Status.GetMotion()
		spd := cs.Status.GetSpeed()
		spot := cs.Status.GetSpot()
		targ := cs.Status.GetTarget()

		log.Println("Recd status for `", cs.ClientID, "`: ", mo, " Speed", spd, "Spot", spot, "Target", targ)
		sm.mu.Lock()
		sm.latest[cs.ClientID] = cs.Status
		sm.mu.Unlock()
	}
}

// GetLatestState returns the latest status message received for a client ID
func (sm *StateModule) GetLatestState(clientID string) *pb.Status {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return sm.latest[clientID]
}
