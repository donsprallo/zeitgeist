package ntp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"time"
)

// Constants for the ntp package.
const (
	NTP_PACKAGE_SIZE   int    = 48
	NTP_STRATUM_MASK   uint32 = 0x0000_FF00
	NTP_POLL_MASK      uint32 = 0x00FF_0000
	NTP_PRECISION_MASK uint32 = 0xFF00_0000
)

// Constants for the ntp package header leap indicator field.
const (
	NTP_LI_MASK    uint32 = 0x0000_0003
	NTP_LI_NOT_SET uint32 = 0x0000_0000
	NTP_LI_SUB_SEC uint32 = 0x0000_0001
	NTP_LI_ADD_SEC uint32 = 0x0000_0002
	NTP_LI_NOT_SYN uint32 = 0x0000_0003
)

// Constants for the ntp package header version field.
const (
	NTP_VN_MASK uint32 = 0x0000_001C
	NTP_VN_V3   uint32 = 0x0000_0003
	NTP_VN_V4   uint32 = 0x0000_0004
)

// Constants for the ntp package header mode field.
const (
	NTP_MODE_MASK        uint32 = 0x0000_00E0
	NTP_MODE_RESERVED    uint32 = 0x0000_0000
	NTP_MODE_SYM_ACTIVE  uint32 = 0x0000_0001
	NTP_MODE_SYM_PASSIVE uint32 = 0x0000_0002
	NTP_MODE_CLIENT      uint32 = 0x0000_0003
	NTP_MODE_SERVER      uint32 = 0x0000_0004
	NTP_MODE_BROADCAST   uint32 = 0x0000_0005
	NTP_MODE_CONTROL     uint32 = 0x0000_0006
	NTP_MODE_PRIVATE     uint32 = 0x0000_0007
)

// This is the ntp package representation. Its received from
// clients and sent to clients as server response.
type NtpPackage struct {
	Header             uint32
	RootDelay          uint32
	RootDispersion     uint32
	ReferenceClockId   uint32
	ReferenceTimestamp uint64
	OriginateTimestamp uint64
	ReceiveTimestamp   uint64
	TransmitTimestamp  uint64
}

// Get the ntp leap indicator. The value can be one of the constants
// NTP_LI_NOT_SET, NTP_LI_SUB_SEC, NTP_LI_ADD_SEC or NTP_LI_NOT_SYN.
func (pkg *NtpPackage) GetLeap() uint32 {
	return (pkg.Header & NTP_LI_MASK)
}

// Set the ntp leap indicator. The value can be one of the constants
// NTP_LI_NOT_SET, NTP_LI_SUB_SEC, NTP_LI_ADD_SEC or NTP_LI_NOT_SYN.
func (pkg *NtpPackage) SetLeap(value uint32) {
	pkg.Header |= (NTP_LI_MASK & value)
}

// Get the ntp version number.
func (pkg *NtpPackage) GetVersion() uint32 {
	return ((pkg.Header & NTP_VN_MASK) >> 2)
}

// Set the ntp version number.
func (pkg *NtpPackage) SetVersion(value uint32) {
	pkg.Header |= (NTP_VN_MASK & (value << 2))
}

// Get the ntp mode.
func (pkg *NtpPackage) GetMode() uint32 {
	return ((pkg.Header & NTP_MODE_MASK) >> 5)
}

// Set the ntp mode.
func (pkg *NtpPackage) SetMode(value uint32) {
	pkg.Header |= (NTP_MODE_MASK & (value << 5))
}

// Get the ntp stratum value.
func (pkg *NtpPackage) GetStratum() uint32 {
	return ((pkg.Header & NTP_STRATUM_MASK) >> 8)
}

// Set the ntp stratum value.
func (pkg *NtpPackage) SetStratum(value uint32) {
	pkg.Header |= (NTP_STRATUM_MASK & (value << 8))
}

// Get the ntp poll interval.
func (pkg *NtpPackage) GetPoll() uint32 {
	return ((pkg.Header & NTP_POLL_MASK) >> 16)
}

// Set the ntp poll interval.
func (pkg *NtpPackage) SetPoll(value uint32) {
	pkg.Header |= (NTP_POLL_MASK & (value << 16))
}

// Get the ntp precision value.
func (pkg *NtpPackage) GetPrecision() uint32 {
	return ((pkg.Header & NTP_PRECISION_MASK) >> 24)
}

// Set the ntp precision value.
func (pkg *NtpPackage) SetPrecision(value uint32) {
	pkg.Header |= (NTP_PRECISION_MASK & (value << 24))
}

func getNtpSeconds(t time.Time) (secs, fracs int64) {
	secs = t.Unix() + int64(getNtpDelta())
	fracs = int64(t.Nanosecond())
	return
}

// Calculate the ntp time delta in seconds. The time delta is the ntp
// epoch (1900-01-01) substracted from unix epoche (1970-01-01). The
// returned value is represented by the Universal Coordinated Time (UTC).
func getNtpDelta() float64 {
	// Cache calculation
	if ntp_time_delta == 0.0 {
		ntpEpoch := time.Date(
			1900, 1, 1, 0, 0, 0, 0, time.UTC)
		unixEpoch := time.Date(
			1970, 1, 1, 0, 0, 0, 0, time.UTC)
		ntp_time_delta = unixEpoch.Sub(ntpEpoch).Seconds()
	}
	return ntp_time_delta
}

// Stringer interface.
func (pkg *NtpPackage) String() string {
	return fmt.Sprintf("<NtpPackage(tx=%d, stratum=%d)>",
		pkg.TransmitTimestamp, pkg.GetStratum())
}

// Return ntp package bytes.
func (pkg *NtpPackage) ToBytes() ([]byte, error) {
	// Create buffer with ntp package size
	buffer := bytes.NewBuffer(
		make([]byte, 0, NTP_PACKAGE_SIZE))
	// Write ntp package to buffer
	err := binary.Write(buffer, binary.BigEndian, pkg)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// Parse ntp package from bytes.
func PackageFromBytes(data []byte) (*NtpPackage, error) {
	// Validate package data size
	if len(data) < NTP_PACKAGE_SIZE {
		return nil, errors.New(
			"invalid ntp package size")
	}
	// Package data to result
	pkg := &NtpPackage{}
	// Read data with byte reader into package
	buffer := bytes.NewReader(data)
	binary.Read(buffer, binary.BigEndian, pkg)
	return pkg, nil
}
