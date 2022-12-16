package server

import (
	"errors"
	"time"

	"github.com/donsprallo/gots/internal/ntp"
)

// Timer represents a ntp timer. A timer generates a time value and can be
// updated and incremented. The timestamp is needed to generate a ntp packet
// from different sources.
type Timer interface {

	// Package get the internal ntp.Package from Timer.
	Package() *ntp.Package

	// Update the Timer for example by increment one second. Therefore,
	// this method must be called in an interval of one second.
	Update()

	// Set the timer to time.Time.
	Set(t time.Time)

	// Get the Timer time.Time.
	Get() time.Time
}

type TimerCollectionEntry struct {
	Id    int   // Index of the Timer
	Timer Timer // Timer of the entry
}

// TimerCollection is a collection of Timer instances.
type TimerCollection struct {
	idx     int                    // Index value of the next Timer
	entries []TimerCollectionEntry // A slice of Timer
}

// NewTimerCollection creates a new TimerCollection with a predefined size.
// The size of the collection is not fixed to size.
func NewTimerCollection(size int) *TimerCollection {
	return &TimerCollection{
		idx:     0,
		entries: make([]TimerCollectionEntry, 0, size),
	}
}

// Add append a Timer to the collection. Here each Timer get a unique entry
// to identify the Timer.
func (c *TimerCollection) Add(timer Timer) int {
	id := c.idx
	c.idx++
	c.entries = append(c.entries, TimerCollectionEntry{
		Id:    id,
		Timer: timer,
	})
	return id
}

// Get the TimerCollectionEntry by id.
func (c *TimerCollection) Get(id int) TimerCollectionEntry {
	// Iterate all timers until id is found.
	for _, entry := range c.entries {
		if entry.Id == id {
			return entry
		}
	}
	// No timer found.
	return TimerCollectionEntry{}
}

// Delete a Timer from collection by id.
func (c *TimerCollection) Delete(id int) error {
	// Iterate all timers until id is found.
	for idx, entry := range c.entries {
		if entry.Id == id {
			c.Remove(idx)
			return nil
		}
	}
	// Deletion failed
	return errors.New(
		"can not delete timer by id")
}

// Remove a Timer from collection by index.
func (c *TimerCollection) Remove(index int) {
	length := len(c.entries) - 1
	entries := make([]TimerCollectionEntry, 0, length)
	entries = append(entries, c.entries[:index]...)
	c.entries = append(entries, c.entries[index+1:]...)
}

// All return all TimerCollectionEntry instances added to collection.
func (c *TimerCollection) All() []TimerCollectionEntry {
	return c.entries
}

// AllUpdate updates all Timer instances added to collection.
func (c *TimerCollection) AllUpdate() {
	for _, entry := range c.entries {
		entry.Timer.Update()
	}
}

// Length return the collection entry length.
func (c *TimerCollection) Length() int {
	return len(c.entries)
}

// NtpTimer implements the Timer interface. A NtpTimer generates time values
// from the remote ntp server as source. The timer can be used to generate
// ntp.Package.
type NtpTimer struct {
	NTPPackage ntp.Package
}

// Package implements Timer.Package interface.
func (timer *NtpTimer) Package() *ntp.Package {
	return &timer.NTPPackage
}

// Update implements Timer.Update interface.
func (timer *NtpTimer) Update() {
	// Do nothing here
}

// Set implements Timer.Set interface.
func (timer *NtpTimer) Set(_ time.Time) {
	// Do nothing here
}

// Get implements Timer.Get interface.
func (timer *NtpTimer) Get() time.Time {
	return time.Now()
}

// SystemTimer implements the Timer interface. A SystemTimer generates time
// values from the system time as source. The source can be used to generate
// ntp.Package.
type SystemTimer struct {
	NTPPackage ntp.Package
}

// Package implements Timer.Package interface.
func (timer *SystemTimer) Package() *ntp.Package {
	return &timer.NTPPackage
}

// Update implements Timer.Update interface.
func (timer *SystemTimer) Update() {
	// Do nothing here
}

// Set implements Timer.Set interface.
func (timer *SystemTimer) Set(_ time.Time) {
	// Do nothing here
}

// Get implements Timer.Get interface.
func (timer *SystemTimer) Get() time.Time {
	return time.Now()
}

// ModifyTimer implements the Timer interface. A ModifyTimer generates time
// values from free settable timestamp as source. The timer can be used to#
// generate ntp.Package.
type ModifyTimer struct {
	NTPPackage ntp.Package
	Time       time.Time
}

// Package implements Timer.Package interface.
func (timer *ModifyTimer) Package() *ntp.Package {
	return &timer.NTPPackage
}

// Update implements Timer.Update interface.
func (timer *ModifyTimer) Update() {
	// Increment timer by one
	timer.Time = timer.Time.Add(1 * time.Second)
}

// Set implements Timer.Set interface.
func (timer *ModifyTimer) Set(t time.Time) {
	timer.Time = t
}

// Get implements Timer.Get interface.
func (timer *ModifyTimer) Get() time.Time {
	return timer.Time
}

// PackageFromTimer convert a ntp.Package from dst ntp.Package to
// src ntp.Package with timestamp from Timer instance.
func PackageFromTimer(
	dst *ntp.Package,
	src *ntp.Package,
	timer Timer,
) (*ntp.Package, error) {
	// Set header
	dst.SetLeap(src.GetLeap())
	dst.SetVersion(src.GetVersion())
	dst.SetMode(src.GetMode())
	dst.SetStratum(src.GetStratum())
	dst.SetPoll(src.GetPoll())
	dst.SetPrecision(src.GetPrecision())

	// Set package data
	dst.SetRootDelay(src.GetRootDelay())
	dst.SetRootDispersion(src.GetRootDispersion())
	dst.SetReferenceClockId(src.GetReferenceClockId())
	dst.SetReferenceTimestamp(timer.Get())
	dst.SetOriginateTimestamp(timer.Get())
	// Set transmit timestamp at least before sent
	//dst.SetTransmitTimestamp(timer.Get())

	return dst, nil
}

// TimerName map a Timer instance to corresponding string representation.
func TimerName(timer Timer) string {
	switch timer.(type) {
	case *NtpTimer:
		return "NtpTimer"
	case *SystemTimer:
		return "SystemTimer"
	case *ModifyTimer:
		return "ModifyTimer"
	default:
		return "UnknownTimer"
	}
}
