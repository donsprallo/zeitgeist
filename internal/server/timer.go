package server

import (
	"time"

	"github.com/donsprallo/gots/internal/ntp"
)

// Timer represents a ntp timer. A timer generates a time value and can be
// updated and incremented. The timestamp is needed to generate a ntp packet
// from different sources.
type Timer interface {

	// Package get the internal ntp.Package from Timer.
	Package() *ntp.Package

	// Increment the Timer by one second.
	Increment()

	// Set the timer to time.Time.
	Set(t time.Time)

	// Get the Timer time.Time.
	Get() time.Time
}

// SystemTimer implements the Timer interface. A SystemTimer generates time
// values from the system time as source. The source can be used to generate
// ntp packets.
type SystemTimer struct {
	NTPPackage ntp.Package
}

// Package implements Timer.Package.
func (timer SystemTimer) Package() *ntp.Package {
	return &timer.NTPPackage
}

// Increment implements Timer.Increment.
func (timer SystemTimer) Increment() {
	// Do nothing here
}

// Set implements Timer.Set.
func (timer SystemTimer) Set(_ time.Time) {
	// Do nothing here
}

// Get implements Timer.Get.
func (timer SystemTimer) Get() time.Time {
	return time.Now()
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
	dst.SetTransmitTimestamp(timer.Get())

	return dst, nil
}

// TimerName map a Timer instance to corresponding string representation.
func TimerName(timer Timer) string {
	switch timer.(type) {
	case *SystemTimer:
		return "SystemTimer"
	case SystemTimer:
		return "SystemTimer"
	default:
		return "UnknownTimer"
	}
}
