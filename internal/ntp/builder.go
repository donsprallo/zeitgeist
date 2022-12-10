package ntp

import "time"

type ResponseBuilder interface {
	BuildResponse(pkg *NtpPackage) (*NtpPackage, error)
}

type SystemResponseBuilder struct {
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
func (builder SystemResponseBuilder) BuildResponse(
	pkg *NtpPackage,
) (*NtpPackage, error) {
	// Set header
	pkg.SetLeap(NTP_LI_NOT_SYN)
	pkg.SetVersion(builder.Version)
	pkg.SetMode(builder.Mode)
	pkg.SetStratum(builder.Stratum)
	pkg.SetPoll(builder.poll)
	pkg.SetPrecision(builder.precision)

	// Set package data
	pkg.SetRootDelay(builder.rootDelay)
	pkg.SetRootDispersion(builder.rootDispersion)
	pkg.SetReferenceClockId(builder.Id)
	pkg.SetReferenceTimestamp(time.Now())
	pkg.SetOriginateTimestamp(time.Now())
	// Set transmit timestamp at least before sent
	pkg.SetTransmitTimestamp(time.Now())

	return pkg, nil
}
