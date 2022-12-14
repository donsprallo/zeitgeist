package server

import (
	"fmt"
	"github.com/donsprallo/gots/internal/ntp"
	"testing"
	"time"
)

// Just a dummy to mock response timer.
type DummyTimer struct {
	Message string
}

// Package implements Timer.Package interface.
func (t DummyTimer) Package() *ntp.Package {
	return nil
}

// Update implements Timer.Update interface.
func (t DummyTimer) Update() {
	// Do nothing here
}

// Set implements Timer.Set interface.
func (t DummyTimer) Set(_ time.Time) {
	// Do nothing here
}

// Get implements Timer.Get interface.
func (t DummyTimer) Get() time.Time {
	return time.Time{}
}

// String implements fmt.Stringer interface.
func (t DummyTimer) String() string {
	return fmt.Sprintf(t.Message)
}

// TestTimerCollectionAdd test to add Timer to collection.
func TestTimerCollectionAdd(t *testing.T) {
	timer := DummyTimer{Message: "test"}

	// Create instance to test.
	collection := NewTimerCollection(10)
	if collection.Length() != 0 {
		t.Errorf("invalid length of new collection")
	}

	// Add timers.
	if collection.Add(timer) != 0 {
		t.Errorf("invalid timer id returned")
	}
	if collection.Add(timer) != 1 {
		t.Errorf("invalid timer id returned")
	}

	// Test length of collection.
	if collection.Length() != 2 {
		t.Errorf("no timer added to collection")
	}
}

// TestTimerCollectionIndex test identifier of Timer instances.
func TestTimerCollectionIndex(t *testing.T) {
	timer := DummyTimer{Message: "test"}

	// Create instance to test.
	collection := NewTimerCollection(10)
	collection.Add(timer)
	collection.Add(timer)
	collection.Add(timer)

	// Test identifier.
	for idx, timer := range collection.All() {
		if idx != timer.Id {
			t.Errorf("collection entry invalid index")
		}
	}
}

// TestTimerCollectionIndex test to get Timer by id.
func TestTimerCollectionGet(t *testing.T) {
	timer := DummyTimer{Message: "test"}

	// Create instance to test.
	collection := NewTimerCollection(10)
	collection.Add(timer)
	collection.Add(timer)
	collection.Add(timer)

	// Test identifier.
	if collection.Get(0).Id != 0 {
		t.Errorf("collection entry invalid id")
	}
	if collection.Get(1).Id != 1 {
		t.Errorf("collection entry invalid id")
	}
	if collection.Get(2).Id != 2 {
		t.Errorf("collection entry invalid id")
	}
}

// TestTimerCollectionRemove test to remove Timer from collection.
func TestTimerCollectionRemove(t *testing.T) {
	timer := DummyTimer{Message: "test"}
	// Create instance to test.
	collection := NewTimerCollection(10)
	collection.Add(timer)
	collection.Add(timer)
	collection.Add(timer)

	// Test to remove timer.
	collection.Remove(1)
	if collection.Length() != 2 {
		t.Errorf("no timer removed from collection")
	}

	// Test that timer 0 and 2 are in collection and timer with
	// id 1 is removed.
	if collection.Get(0).Timer == nil {
		t.Errorf("no timer with id 0")
	}
	if collection.Get(1).Timer != nil {
		t.Errorf("no timer with id 1 removed")
	}
	if collection.Get(2).Timer == nil {
		t.Errorf("no timer with id 2")
	}
}
