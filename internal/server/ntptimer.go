package server

import (
	"time"

	"github.com/donsprallo/gots/internal/ntp"
)

type NtpTimer interface {
	Package(pkg *ntp.NtpPackage) (*ntp.NtpPackage, error)
	Increment()
	Set(t time.Time)
}

type SystemNtpTimer struct {
	Version        uint32
	Mode           uint32
	Stratum        uint32
	Id             []byte
	poll           uint32
	precision      uint32
	rootDelay      uint32
	rootDispersion uint32
}

// Build a ntp package from response and current time.
func (timer SystemNtpTimer) Package(
	pkg *ntp.NtpPackage,
) (*ntp.NtpPackage, error) {
	// Set header
	pkg.SetLeap(ntp.NTP_LI_NOT_SYN)
	pkg.SetVersion(timer.Version)
	pkg.SetMode(timer.Mode)
	pkg.SetStratum(timer.Stratum)
	pkg.SetPoll(timer.poll)
	pkg.SetPrecision(timer.precision)

	// Set package data
	pkg.SetRootDelay(timer.rootDelay)
	pkg.SetRootDispersion(timer.rootDispersion)
	pkg.SetReferenceClockId(timer.Id)
	pkg.SetReferenceTimestamp(time.Now())
	pkg.SetOriginateTimestamp(time.Now())
	// Set transmit timestamp at least before sent
	pkg.SetTransmitTimestamp(time.Now())

	return pkg, nil
}

func (timer SystemNtpTimer) Increment() {
	// Do nothing here
}

func (timer SystemNtpTimer) Set(t time.Time) {
	// Do nothing here
}
