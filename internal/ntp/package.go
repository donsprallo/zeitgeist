// Copyright 2024 The Zeitgeist Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package ntp

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"time"
)

var (
	// Epoch represent the time.Date for the ntp epoch (1900-01-01).
	Epoch = time.Date(
		1900, time.January, 1, 0, 0, 0, 0, time.UTC)

	// UnixEpoch represent the time.Date for the unix epoch (1970-01-01).
	UnixEpoch = time.Date(
		1970, time.January, 1, 0, 0, 0, 0, time.UTC)

	// TimeDelta is the time-delta in seconds from ntp epoch to unix epoch. The
	// value is calculated by subtracting ntp epoch from unix epoch.
	TimeDelta = uint32(UnixEpoch.Sub(Epoch).Seconds())
)

// Constants for the ntp package.
const (
	PackageSize   int    = 48
	LeapMask      uint32 = 0xC000_0000
	ModeMask      uint32 = 0x0700_0000
	VersionMask   uint32 = 0x3800_0000
	StratumMask   uint32 = 0x00FF_0000
	PollMask      uint32 = 0x0000_FF00
	PrecisionMask uint32 = 0x0000_00FF
)

// Constants for the ntp package header leap indicator field.
const (
	LeapNotSet uint32 = 0x0000_0000
	LeapSubSec uint32 = 0x0000_0001
	LeapAddSec uint32 = 0x0000_0002
	LeapNotSyn uint32 = 0x0000_0003
)

// Constants for the ntp package header version field.
const (
	VersionV3 uint32 = 0x0000_0003
	VersionV4 uint32 = 0x0000_0004
)

// Constants for the ntp package header mode field.
const (
	ModeReserved   uint32 = 0x0000_0000
	ModeSymActive  uint32 = 0x0000_0001
	ModeSymPassive uint32 = 0x0000_0002
	ModeClient     uint32 = 0x0000_0003
	ModeServer     uint32 = 0x0000_0004
	ModeBroadcast  uint32 = 0x0000_0005
	ModeControl    uint32 = 0x0000_0006
	ModePrivate    uint32 = 0x0000_0007
)

type Timestamp struct {
	Seconds  uint32
	Fraction uint32
}

// ToTimestamp convert a unix time.Time to seconds and fractional
// part of a ntp timestamp.
func ToTimestamp(t time.Time) Timestamp {
	var ts Timestamp
	unix := t.Unix()
	ts.Seconds = uint32(unix) + TimeDelta
	ts.Fraction = uint32(float64(t.UnixMicro()) * (1 << 32) * 1.0e-6)
	return ts
}

// ToTime convert seconds and fraction of seconds to time.Time.
func ToTime(ts Timestamp) time.Time {
	if ts.Seconds > 0 {
		ts.Seconds -= TimeDelta
	}
	seconds := time.Duration(ts.Seconds) * time.Second
	nanoseconds := time.Duration(ts.Fraction)
	return UnixEpoch.Add(seconds + nanoseconds)
}

// Package is the ntp package representation. A package is
// received from clients and sent to clients as server response.
type Package struct {
	header             uint32
	rootDelay          uint32
	rootDispersion     uint32
	referenceClockId   uint32
	referenceTimestamp time.Time
	originateTimestamp time.Time
	receiveTimestamp   time.Time
	transmitTimestamp  time.Time
}

// GetLeap get the package leap indicator.
func (pkg *Package) GetLeap() uint32 {
	return (pkg.header & LeapMask) >> 30
}

// SetLeap set the package leap indicator.
func (pkg *Package) SetLeap(value uint32) {
	pkg.header &= ^LeapMask
	pkg.header |= LeapMask & (value << 30)
}

// GetVersion get the package version number.
func (pkg *Package) GetVersion() uint32 {
	return (pkg.header & VersionMask) >> 27
}

// SetVersion set the package version number.
func (pkg *Package) SetVersion(value uint32) {
	pkg.header &= ^VersionMask
	pkg.header |= VersionMask & (value << 27)
}

// GetMode get the package mode.
func (pkg *Package) GetMode() uint32 {
	return (pkg.header & ModeMask) >> 24
}

// SetMode set the package mode.
func (pkg *Package) SetMode(value uint32) {
	pkg.header &= ^ModeMask
	pkg.header |= ModeMask & (value << 24)
}

// GetStratum get the package stratum value.
func (pkg *Package) GetStratum() uint32 {
	return (pkg.header & StratumMask) >> 16
}

// SetStratum set the package stratum value.
func (pkg *Package) SetStratum(value uint32) {
	pkg.header &= ^StratumMask
	pkg.header |= StratumMask & (value << 16)
}

// GetPoll get the package poll interval.
func (pkg *Package) GetPoll() uint32 {
	return (pkg.header & PollMask) >> 8
}

// SetPoll set the package poll interval.
func (pkg *Package) SetPoll(value uint32) {
	pkg.header &= ^PollMask
	pkg.header |= PollMask & (value << 8)
}

// GetPrecision get the package precision value.
func (pkg *Package) GetPrecision() uint32 {
	return (pkg.header & PrecisionMask) >> 0
}

// SetPrecision set the package precision value.
func (pkg *Package) SetPrecision(value uint32) {
	pkg.header &= ^PrecisionMask
	pkg.header |= PrecisionMask & (value << 0)
}

// GetRootDelay get the package root delay.
func (pkg *Package) GetRootDelay() uint32 {
	return pkg.rootDelay
}

// SetRootDelay set the package root delay.
func (pkg *Package) SetRootDelay(value uint32) {
	pkg.rootDelay = value
}

// GetRootDispersion get the package root dispersion.
func (pkg *Package) GetRootDispersion() uint32 {
	return pkg.rootDispersion
}

// SetRootDispersion set the package root dispersion.
func (pkg *Package) SetRootDispersion(value uint32) {
	pkg.rootDispersion = value
}

// GetReferenceClockId get the package reference clock identifier.
func (pkg *Package) GetReferenceClockId() []byte {
	buf := make([]byte, 4)
	return binary.BigEndian.AppendUint32(
		buf, pkg.referenceClockId)
}

// SetReferenceClockId set the package reference clock identifier.
func (pkg *Package) SetReferenceClockId(value []byte) {
	pkg.referenceClockId = binary.BigEndian.Uint32(value)
}

// GetReferenceTimestamp get the package reference timestamp.
func (pkg *Package) GetReferenceTimestamp() time.Time {
	return pkg.referenceTimestamp
}

// SetReferenceTimestamp set the package reference timestamp.
func (pkg *Package) SetReferenceTimestamp(value time.Time) {
	pkg.referenceTimestamp = value
}

// GetOriginateTimestamp get the package originate timestamp.
func (pkg *Package) GetOriginateTimestamp() time.Time {
	return pkg.originateTimestamp
}

// SetOriginateTimestamp set the package originate timestamp.
func (pkg *Package) SetOriginateTimestamp(value time.Time) {
	pkg.originateTimestamp = value
}

// GetReceiveTimestamp get the package receive timestamp.
func (pkg *Package) GetReceiveTimestamp() time.Time {
	return pkg.receiveTimestamp
}

// SetReceiveTimestamp set the package receive timestamp.
func (pkg *Package) SetReceiveTimestamp(value time.Time) {
	pkg.receiveTimestamp = value
}

// GetTransmitTimestamp get the package receive timestamp.
func (pkg *Package) GetTransmitTimestamp() time.Time {
	return pkg.transmitTimestamp
}

// SetTransmitTimestamp set the package receive timestamp.
func (pkg *Package) SetTransmitTimestamp(value time.Time) {
	pkg.transmitTimestamp = value
}

// ToBytes converts package to bytes.
func (pkg *Package) ToBytes() ([]byte, error) {
	return pkg.MarshalBinary()
}

// PackageFromBytes parse package from bytes.
func PackageFromBytes(data []byte) (*Package, error) {
	pkg := Package{}
	err := pkg.UnmarshalBinary(data)
	if err != nil {
		return nil, err
	}
	return &pkg, nil
}

// String implements the fmt.Stringer interface.
func (pkg *Package) String() string {
	return fmt.Sprintf("<NtpPackage(mode=%d, version=%d, stratum=%d)>",
		pkg.GetMode(), pkg.GetVersion(), pkg.GetStratum())
}

// MarshalBinary implements encoding.BinaryMarshaler interface.
func (pkg *Package) MarshalBinary() ([]byte, error) {
	// Create encoder with network byte order
	encoder := binary.BigEndian
	// Create ntp package buffer
	enc := make([]byte, 0, PackageSize)

	// Encode package data
	enc = encoder.AppendUint32(enc, pkg.header)
	enc = encoder.AppendUint32(enc, pkg.rootDelay)
	enc = encoder.AppendUint32(enc, pkg.rootDispersion)
	enc = encoder.AppendUint32(enc, pkg.referenceClockId)

	// Encode package data timestamps
	ts := ToTimestamp(pkg.referenceTimestamp)
	enc = encoder.AppendUint32(enc, ts.Seconds)
	enc = encoder.AppendUint32(enc, ts.Fraction)

	ts = ToTimestamp(pkg.originateTimestamp)
	enc = encoder.AppendUint32(enc, ts.Seconds)
	enc = encoder.AppendUint32(enc, ts.Fraction)

	ts = ToTimestamp(pkg.receiveTimestamp)
	enc = encoder.AppendUint32(enc, ts.Seconds)
	enc = encoder.AppendUint32(enc, ts.Fraction)

	ts = ToTimestamp(pkg.transmitTimestamp)
	enc = encoder.AppendUint32(enc, ts.Seconds)
	enc = encoder.AppendUint32(enc, ts.Fraction)

	return enc, nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler interface.
func (pkg *Package) UnmarshalBinary(data []byte) error {
	// Validate package size
	if len(data) < PackageSize {
		return errors.New(
			"ntp package size to short")
	}

	// Create decoder with network byte order
	buf := data
	dec := binary.BigEndian

	// Decode package data
	pkg.header = dec.Uint32(buf)
	pkg.rootDelay = dec.Uint32(buf[4:])
	pkg.rootDispersion = dec.Uint32(buf[8:])
	pkg.referenceClockId = dec.Uint32(buf[12:])

	// Decode package data timestamps
	ts := Timestamp{
		Seconds:  dec.Uint32(buf[16:]),
		Fraction: dec.Uint32(buf[20:]),
	}
	pkg.referenceTimestamp = ToTime(ts)

	ts = Timestamp{
		Seconds:  dec.Uint32(buf[24:]),
		Fraction: dec.Uint32(buf[28:]),
	}
	pkg.originateTimestamp = ToTime(ts)

	ts = Timestamp{
		Seconds:  dec.Uint32(buf[32:]),
		Fraction: dec.Uint32(buf[36:]),
	}
	pkg.receiveTimestamp = ToTime(ts)

	ts = Timestamp{
		Seconds:  dec.Uint32(buf[40:]),
		Fraction: dec.Uint32(buf[44:]),
	}
	pkg.transmitTimestamp = ToTime(ts)

	return nil
}

// Request a Package from remote host.
func Request(host string, port int) (*Package, error) {
	var pkg Package
	pkg.SetMode(ModeClient)
	pkg.SetVersion(VersionV3)
	pkg.SetTransmitTimestamp(time.Now())

	// Convert package to bytes.
	bytesToSent, err := pkg.ToBytes()
	if err != nil {
		return nil, err
	}

	// Create udp connection with read write timeout.
	conn, err := createUdpConn(host, port, 1*time.Second)
	if err != nil {
		return nil, err
	}

	// Write bytes to connection.
	write, err := conn.Write(bytesToSent)
	if err != nil || write != PackageSize {
		return nil, err
	}

	// Read response from connection.
	buffer := make([]byte, PackageSize)
	read, err := conn.Read(buffer)
	if err != nil || read != PackageSize {
		return nil, err
	}

	// Parse package from received bytes.
	err = pkg.UnmarshalBinary(buffer)
	if err != nil {
		return nil, err
	}

	return &pkg, nil
}

func createUdpConn(
	host string, port int, timeout time.Duration,
) (net.Conn, error) {
	addr := fmt.Sprintf("%s:%d", host, port)
	// Dial to remote udp address.
	conn, err := net.Dial("udp", addr)
	if err != nil {
		return nil, err
	}

	// Setup connection read and write timeout. We need to set up
	// timeout to a future time value here.
	deadline := time.Now().Add(timeout)

	err = conn.SetReadDeadline(deadline)
	if err != nil {
		return nil, err
	}

	err = conn.SetWriteDeadline(deadline)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
