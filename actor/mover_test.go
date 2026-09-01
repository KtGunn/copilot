package main

import (
	"testing"
	"time"
)

func TestNewMover(t *testing.T) {
	m := NewMover(10, 2.0)
	s := m.Status()

	if s.State != Stationary {
		t.Errorf("expected Stationary, got %s", s.State)
	}
	if s.Position != 0 {
		t.Errorf("expected position 0, got %f", s.Position)
	}
	if s.Speed != 2.0 {
		t.Errorf("expected speed 2.0, got %f", s.Speed)
	}
	if s.Target != 0 {
		t.Errorf("expected target 0, got %d", s.Target)
	}
}

func TestMoveToOutOfRange(t *testing.T) {
	m := NewMover(5, 1.0)

	if err := m.MoveTo(-1); err == nil {
		t.Error("expected error for negative spot")
	}
	if err := m.MoveTo(6); err == nil {
		t.Error("expected error for spot above max")
	}
}

func TestMoveToSameSpot(t *testing.T) {
	m := NewMover(5, 1.0)
	err := m.MoveTo(0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := m.Status()
	if s.State != Stationary {
		t.Errorf("expected Stationary when moving to current spot, got %s", s.State)
	}
}

func TestMoveToAndArrive(t *testing.T) {
	m := NewMover(10, 20.0) // fast speed so it arrives quickly
	err := m.MoveTo(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should be moving initially
	time.Sleep(10 * time.Millisecond)
	s := m.Status()
	if s.Target != 1 {
		t.Errorf("expected target 1, got %d", s.Target)
	}

	// Wait for arrival (1 spot at 20 spots/sec = 50ms, give margin)
	time.Sleep(200 * time.Millisecond)
	s = m.Status()
	if s.State != Stationary {
		t.Errorf("expected Stationary after arrival, got %s", s.State)
	}
	if s.Position != 1.0 {
		t.Errorf("expected position 1.0, got %f", s.Position)
	}
}

func TestMoveBackward(t *testing.T) {
	m := NewMover(10, 100.0)
	m.MoveTo(5)
	time.Sleep(300 * time.Millisecond)

	m.MoveTo(2)
	time.Sleep(300 * time.Millisecond)

	s := m.Status()
	if s.State != Stationary {
		t.Errorf("expected Stationary, got %s", s.State)
	}
	if s.Position != 2.0 {
		t.Errorf("expected position 2.0, got %f", s.Position)
	}
}

func TestInterruptMovement(t *testing.T) {
	m := NewMover(100, 1.0) // slow: 1 spot/sec over 100 spots
	m.MoveTo(100)
	time.Sleep(100 * time.Millisecond)

	// Interrupt with new target
	m.MoveTo(0)
	time.Sleep(2 * time.Second)

	s := m.Status()
	if s.Target != 0 {
		t.Errorf("expected target 0, got %d", s.Target)
	}
}

func TestSetSpeed(t *testing.T) {
	m := NewMover(10, 1.0)
	m.SetSpeed(5.0)
	s := m.Status()
	if s.Speed != 5.0 {
		t.Errorf("expected speed 5.0, got %f", s.Speed)
	}
}

func TestMoveToMaxSpot(t *testing.T) {
	m := NewMover(3, 100.0)
	err := m.MoveTo(3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	time.Sleep(300 * time.Millisecond)
	s := m.Status()
	if s.Position != 3.0 {
		t.Errorf("expected position 3.0, got %f", s.Position)
	}
	if s.State != Stationary {
		t.Errorf("expected Stationary, got %s", s.State)
	}
}
