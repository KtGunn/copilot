package main

import (
	"fmt"
	"math"
	"sync"
	"time"
)

type MoverState string

const (
	Moving     MoverState = "moving"
	Stationary MoverState = "stationary"
)

type MoverStatus struct {
	State    MoverState `json:"state"`
	Speed    float64    `json:"speed"`
	Position float64    `json:"position"`
	Target   int        `json:"target"`
}

type Mover struct {
	mu       sync.RWMutex
	maxSpot  int
	speed    float64 // spots per second
	position float64
	target   int
	state    MoverState
	stopCh   chan struct{}
}

// NewMover creates a mover that travels between spots 0..maxSpot at the given speed (spots/sec).
func NewMover(maxSpot int, speed float64) *Mover {
	return &Mover{
		maxSpot:  maxSpot,
		speed:    speed,
		position: 0,
		target:   0,
		state:    Stationary,
	}
}

// MoveTo commands the mover to travel to the given spot.
func (m *Mover) MoveTo(spot int) error {
	if spot < 0 || spot > m.maxSpot {
		return fmt.Errorf("spot %d out of range [0, %d]", spot, m.maxSpot)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Stop any current movement
	if m.stopCh != nil {
		close(m.stopCh)
	}

	m.target = spot
	if math.Abs(float64(spot)-m.position) < 0.001 {
		m.state = Stationary
		return nil
	}

	m.state = Moving
	m.stopCh = make(chan struct{})
	go m.simulate(m.stopCh)

	return nil
}

func (m *Mover) simulate(stop chan struct{}) {
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			m.mu.Lock()
			target := float64(m.target)
			dist := target - m.position
			step := m.speed * 0.05 // distance per tick

			if math.Abs(dist) <= step {
				m.position = target
				m.state = Stationary
				m.stopCh = nil
				m.mu.Unlock()
				return
			}

			if dist > 0 {
				m.position += step
			} else {
				m.position -= step
			}
			m.mu.Unlock()
		}
	}
}

// Status returns the current state of the mover.
func (m *Mover) Status() MoverStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return MoverStatus{
		State:    m.state,
		Speed:    m.speed,
		Position: m.position,
		Target:   m.target,
	}
}

// [ktg] Added String method to Mover for better logging and debugging.
func (m *Mover) String() string {
	s := m.Status()
	if s.State == Stationary {
		return fmt.Sprintf("'%s' at %v", s.State, int(s.Position))
	} else {
		return fmt.Sprintf("'%s' through %v, @spd %v, going to %v",
			s.State, int(s.Position), s.Speed, s.Target)
	}
}

// SetSpeed updates the movement speed (spots per second).
func (m *Mover) SetSpeed(speed float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.speed = speed
}
