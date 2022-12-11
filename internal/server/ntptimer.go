package server

import (
	"time"

	"github.com/donsprallo/gots/internal/ntp"
)

type Timer interface {
	Package() *ntp.NtpPackage
	Increment()
	Set(t time.Time)
	Get() time.Time
}

func PackageFromTimer(
	dst *ntp.NtpPackage,
	src *ntp.NtpPackage,
	timer Timer,
) (*ntp.NtpPackage, error) {
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

type SystemTimer struct {
	NTPPackage ntp.NtpPackage
}

func (timer SystemTimer) Package() *ntp.NtpPackage {
	return &timer.NTPPackage
}

func (timer SystemTimer) Increment() {
	// Do nothing here
}

func (timer SystemTimer) Set(t time.Time) {
	// Do nothing here
}

func (timer SystemTimer) Get() time.Time {
	return time.Now()
}
