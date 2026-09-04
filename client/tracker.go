package main

import (
	pb "coptest/proto"
	"sync"
)

// StatusTracker manages the incoming stream of status updates.
type StatusTracker struct {
	mu           sync.RWMutex
	latestStatus *pb.Status
	err          error
}

// NewStatusTracker creating a new instance of StatusTracker.
func NewStatusTracker() *StatusTracker {
	return &StatusTracker{}
}

// SetLatest simply records the newest status from the stream.
func (st *StatusTracker) SetLatest(status *pb.Status) {
	st.mu.Lock()
	defer st.mu.Unlock()
	st.latestStatus = status
	st.err = nil
}

// SetError records any stream errors.
func (st *StatusTracker) SetError(err error) {
	st.mu.Lock()
	defer st.mu.Unlock()
	st.err = err
}

// GetLatest returns the latest received status or nil if none have arrived yet.
func (st *StatusTracker) GetLatest() (*pb.Status, error) {
	st.mu.RLock()
	defer st.mu.RUnlock()
	return st.latestStatus, st.err
}
