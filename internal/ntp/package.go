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

const (
	NTP_PACKAGE_SIZE int = 48
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

func (pkg *NtpPackage) GetLeap() uint32 {
	return uint32((pkg.Header & 0x0000_0003))
}

func (pkg *NtpPackage) SetLeap(val uint32) {

}

func (pkg *NtpPackage) GetVersion() uint32 {
	return uint32((pkg.Header & 0x0000_001C) >> 2)
}

func (pkg *NtpPackage) SetVersion(val uint32) {

}

func (pkg *NtpPackage) GetMode() uint32 {
	return uint32((pkg.Header & 0x0000_E000) >> 5)
}

func (pkg *NtpPackage) SetMode(val uint32) {

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
