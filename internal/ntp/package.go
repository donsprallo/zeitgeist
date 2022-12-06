package ntp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"time"
)

/*
+-----------+------------+-----------------------+
| Name      | Formula    | Description           |
+-----------+------------+-----------------------+
| leap      | leap       | leap indicator (LI)   |
| version   | version    | version number (VN)   |
| mode      | mode       | mode                  |
| stratum   | stratum    | stratum               |
| poll      | poll       | poll exponent         |
| precision | rho        | precision exponent    |
| rootdelay | delta_r    | root delay            |
| rootdisp  | epsilon_r  | root dispersion       |
| refid     | refid      | reference ID          |
| reftime   | reftime    | reference timestamp   |
| org       | T1         | origin timestamp      |
| rec       | T2         | receive timestamp     |
| xmt       | T3         | transmit timestamp    |
| dst       | T4         | destination timestamp |
| keyid     | keyid      | key ID                |
| dgst      | dgst       | message digest        |
+-----------+------------+-----------------------+
              tbl. of ntp package header - 32 Bit
*/

// Constants for the ntp package.
const (
	NTP_PACKAGE_SIZE int = 48
)

// Constants for the ntp package header leap indicator field.
const (
	NTP_LI_MASK    int = 0x0000_0003
	NTP_LI_NOT_SET int = 0x0000_0000
	NTP_LI_SUB_SEC int = 0x0000_0001
	NTP_LI_ADD_SEC int = 0x0000_0002
	NTP_LI_NOT_SYN int = 0x0000_0003
)

// Constants for the ntp package header version field.
const (
	NTP_VN_MASK int = 0x0000_001C
	NTP_VN_V3   int = 0x0000_0003
	NTP_VN_V4   int = 0x0000_0004
)

// Constants for the ntp package header mode field.
const (
	NTP_MODE_MASK        int = 0x0000_00E0
	NTP_MODE_RESERVED    int = 0x0000_0000
	NTP_MODE_SYM_ACTIVE  int = 0x0000_0001
	NTP_MODE_SYM_PASSIVE int = 0x0000_0002
	NTP_MODE_CLIENT      int = 0x0000_0003
	NTP_MODE_SERVER      int = 0x0000_0004
	NTP_MODE_BROADCAST   int = 0x0000_0005
	NTP_MODE_CONTROL     int = 0x0000_0006
	NTP_MODE_PRIVATE     int = 0x0000_0007
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
	return uint32((pkg.Header & uint32(NTP_LI_MASK)))
}

// Set the ntp leap indicator. The value can be one of the constants
// NTP_LI_NOT_SET, NTP_LI_SUB_SEC, NTP_LI_ADD_SEC or NTP_LI_NOT_SYN.
func (pkg *NtpPackage) SetLeap(value uint32) {
	pkg.Header |= (uint32(NTP_LI_MASK) & value)
}

// Get the ntp version number.
func (pkg *NtpPackage) GetVersion() uint32 {
	return uint32((pkg.Header & uint32(NTP_VN_MASK)) >> 2)
}

// Set the ntp version number.
func (pkg *NtpPackage) SetVersion(value uint32) {
	pkg.Header |= uint32(NTP_VN_MASK) & (value << 2)
}

// Get the ntp mode.
func (pkg *NtpPackage) GetMode() uint32 {
	return uint32((pkg.Header & uint32(NTP_MODE_MASK)) >> 5)
}

// Set the ntp mode.
func (pkg *NtpPackage) SetMode(value uint32) {
	pkg.Header |= uint32(NTP_MODE_MASK) & (value << 5)
}

func (pkg *NtpPackage) GetStratum() uint32 {
	return uint32((pkg.Header & 0x0000_FF00) >> 7)
}

func (pkg *NtpPackage) SetStratum(val uint32) {

}

func (pkg *NtpPackage) GetPoll() uint32 {
	return uint32((pkg.Header & 0x00FF_0000) >> 16)
}

func (pkg *NtpPackage) SetPoll(val uint32) {

}

func (pkg *NtpPackage) GetPrecision() uint32 {
	return uint32((pkg.Header & 0xFF00_0000) >> 24)
}

func (pkg *NtpPackage) SetPrecision(val uint32) {

}

func getNtpSeconds(t time.Time) (secs, fracs int64) {
	secs = t.Unix() + int64(getNtpDelta())
	fracs = int64(t.Nanosecond())
	return
}

// Calculate the ntp time delta. This is the ntp epoche (1900-01-01)
// substracted from unix epoche (1970-01-01).
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
