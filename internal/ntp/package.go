package ntp

import (
	"encoding/binary"
	"errors"
	"fmt"
	"time"
)

var (
	NtpEpoch  time.Time
	UnixEpoch time.Time
	NtpDelta  float64
)

func init() {
	// Calculate the ntp time delta in seconds. The time delta is the ntp
	// epoch (1900-01-01) substracted from unix epoche (1970-01-01). The
	// returned value is represented by the Universal Coordinated Time (UTC).
	NtpEpoch = time.Date(
		1900, 1, 1, 0, 0, 0, 0, time.UTC)
	UnixEpoch = time.Date(
		1970, 1, 1, 0, 0, 0, 0, time.UTC)
	NtpDelta = UnixEpoch.Sub(NtpEpoch).Seconds()
}

// Constants for the ntp package.
const (
	NTP_PACKAGE_SIZE   int    = 48
	NTP_STRATUM_MASK   uint32 = 0x00FF_0000
	NTP_POLL_MASK      uint32 = 0x0000_FF00
	NTP_PRECISION_MASK uint32 = 0x0000_00FF
)

// Constants for the ntp package header leap indicator field.
const (
	NTP_LI_MASK    uint32 = 0xC000_0000
	NTP_LI_NOT_SET uint32 = 0x0000_0000
	NTP_LI_SUB_SEC uint32 = 0x0000_0001
	NTP_LI_ADD_SEC uint32 = 0x0000_0002
	NTP_LI_NOT_SYN uint32 = 0x0000_0003
)

// Constants for the ntp package header version field.
const (
	NTP_VN_MASK uint32 = 0x3800_0000
	NTP_VN_V3   uint32 = 0x0000_0003
	NTP_VN_V4   uint32 = 0x0000_0004
)

// Constants for the ntp package header mode field.
const (
	NTP_MODE_MASK        uint32 = 0x0700_0000
	NTP_MODE_RESERVED    uint32 = 0x0000_0000
	NTP_MODE_SYM_ACTIVE  uint32 = 0x0000_0001
	NTP_MODE_SYM_PASSIVE uint32 = 0x0000_0002
	NTP_MODE_CLIENT      uint32 = 0x0000_0003
	NTP_MODE_SERVER      uint32 = 0x0000_0004
	NTP_MODE_BROADCAST   uint32 = 0x0000_0005
	NTP_MODE_CONTROL     uint32 = 0x0000_0006
	NTP_MODE_PRIVATE     uint32 = 0x0000_0007
)

func timestampToNtpSeconds(t time.Time) (secs, fracs uint32) {
	secs = uint32(t.Unix() + int64(NtpDelta))
	fracs = uint32(t.Nanosecond())
	return
}

func ntpSecondsToTimestamp(secs, fracs uint32) time.Time {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint32(buf[0:], secs)
	binary.BigEndian.PutUint32(buf[4:], fracs)

	t := time.Time{}
	return t
}

// This is the ntp package representation. Its received from
// clients and sent to clients as server response.
type NtpPackage struct {
	header             uint32
	rootDelay          uint32
	rootDispersion     uint32
	referenceClockId   uint32
	referenceTimestamp time.Time
	originateTimestamp time.Time
	receiveTimestamp   time.Time
	transmitTimestamp  time.Time
}

// Get the ntp leap indicator.
func (pkg *NtpPackage) GetLeap() uint32 {
	return (pkg.header & NTP_LI_MASK) >> 30
}

// Set the ntp leap indicator.
func (pkg *NtpPackage) SetLeap(value uint32) {
	pkg.header &= ^NTP_LI_MASK
	pkg.header |= (NTP_LI_MASK & (value << 30))
}

// Get the ntp version number.
func (pkg *NtpPackage) GetVersion() uint32 {
	return ((pkg.header & NTP_VN_MASK) >> 27)
}

// Set the ntp version number.
func (pkg *NtpPackage) SetVersion(value uint32) {
	pkg.header &= ^NTP_VN_MASK
	pkg.header |= (NTP_VN_MASK & (value << 27))
}

// Get the ntp mode.
func (pkg *NtpPackage) GetMode() uint32 {
	return ((pkg.header & NTP_MODE_MASK) >> 24)
}

// Set the ntp mode.
func (pkg *NtpPackage) SetMode(value uint32) {
	pkg.header &= ^NTP_MODE_MASK
	pkg.header |= (NTP_MODE_MASK & (value << 24))
}

// Get the ntp stratum value.
func (pkg *NtpPackage) GetStratum() uint32 {
	return ((pkg.header & NTP_STRATUM_MASK) >> 16)
}

// Set the ntp stratum value.
func (pkg *NtpPackage) SetStratum(value uint32) {
	pkg.header &= ^NTP_STRATUM_MASK
	pkg.header |= (NTP_STRATUM_MASK & (value << 16))
}

// Get the ntp poll interval.
func (pkg *NtpPackage) GetPoll() uint32 {
	return ((pkg.header & NTP_POLL_MASK) >> 8)
}

// Set the ntp poll interval.
func (pkg *NtpPackage) SetPoll(value uint32) {
	pkg.header &= ^NTP_POLL_MASK
	pkg.header |= (NTP_POLL_MASK & (value << 8))
}

// Get the ntp precision value.
func (pkg *NtpPackage) GetPrecision() uint32 {
	return ((pkg.header & NTP_PRECISION_MASK) >> 0)
}

// Set the ntp precision value.
func (pkg *NtpPackage) SetPrecision(value uint32) {
	pkg.header &= ^NTP_PRECISION_MASK
	pkg.header |= (NTP_PRECISION_MASK & (value << 0))
}

// Get the ntp root delay.
func (pkg *NtpPackage) GetRootDelay() uint32 {
	return pkg.rootDelay
}

// Set the ntp root delay.
func (pkg *NtpPackage) SetRootDelay(value uint32) {
	pkg.rootDelay = value
}

// Get the ntp root dispersion.
func (pkg *NtpPackage) GetRootDispersion() uint32 {
	return pkg.rootDispersion
}

// Set the ntp root dispersion.
func (pkg *NtpPackage) SetRootDispersion(value uint32) {
	pkg.rootDispersion = value
}

// Get the ntp reference clock identifier.
func (pkg *NtpPackage) GetReferenceClockId() []byte {
	buf := make([]byte, 4)
	return binary.BigEndian.AppendUint32(
		buf, pkg.referenceClockId)
}

// Set the ntp reference clock identifier.
func (pkg *NtpPackage) SetReferenceClockId(value []byte) {
	pkg.referenceClockId = binary.BigEndian.Uint32(value)
}

// Get the ntp reference timestamp.
func (pkg *NtpPackage) GetReferenceTimestamp() time.Time {
	return pkg.referenceTimestamp
}

// Set the ntp reference timestamp.
func (pkg *NtpPackage) SetReferenceTimestamp(value time.Time) {
	pkg.referenceTimestamp = value
}

// Get the ntp originate timestamp.
func (pkg *NtpPackage) GetOriginateTimestamp() time.Time {
	return pkg.originateTimestamp
}

// Set the ntp originate timestamp.
func (pkg *NtpPackage) SetOriginateTimestamp(value time.Time) {
	pkg.originateTimestamp = value
}

// Get the ntp receive timestamp.
func (pkg *NtpPackage) GetReceiveTimestamp() time.Time {
	return pkg.receiveTimestamp
}

// Set the ntp receive timestamp.
func (pkg *NtpPackage) SetReceiveTimestamp(value time.Time) {
	pkg.receiveTimestamp = value
}

// Get the ntp receive timestamp.
func (pkg *NtpPackage) GetTransmitTimestamp() time.Time {
	return pkg.transmitTimestamp
}

// Set the ntp receive timestamp.
func (pkg *NtpPackage) SetTransmitTimestamp(value time.Time) {
	pkg.transmitTimestamp = value
}

// Return ntp package bytes.
func (pkg *NtpPackage) ToBytes() ([]byte, error) {
	return pkg.MarshalBinary()
}

// Parse ntp package from bytes.
func PackageFromBytes(data []byte) (*NtpPackage, error) {
	pkg := NtpPackage{}
	err := pkg.UnmarshalBinary(data)
	if err != nil {
		return nil, err
	}
	return &pkg, nil
}

// String implements the fmt.Stringer interface.
func (pkg *NtpPackage) String() string {
	return fmt.Sprintf("<NtpPackage(mode=%d, version=%d, stratum=%d)>",
		pkg.GetMode(), pkg.GetVersion(), pkg.GetStratum())
}

// MarshalBinary implements encoding.BinaryMarshaler interface.
func (pkg *NtpPackage) MarshalBinary() ([]byte, error) {
	// Create encoder with network byte order
	encoder := binary.BigEndian
	// Create ntp package buffer
	enc := make([]byte, 0, NTP_PACKAGE_SIZE)

	// Encode package data
	enc = encoder.AppendUint32(enc, pkg.header)
	enc = encoder.AppendUint32(enc, pkg.rootDelay)
	enc = encoder.AppendUint32(enc, pkg.rootDispersion)
	enc = encoder.AppendUint32(enc, pkg.referenceClockId)

	// Encode package data timestamps
	secs, fracs := timestampToNtpSeconds(pkg.referenceTimestamp)
	enc = encoder.AppendUint32(enc, uint32(secs))
	enc = encoder.AppendUint32(enc, uint32(fracs))

	secs, fracs = timestampToNtpSeconds(pkg.originateTimestamp)
	enc = encoder.AppendUint32(enc, uint32(secs))
	enc = encoder.AppendUint32(enc, uint32(fracs))

	secs, fracs = timestampToNtpSeconds(pkg.receiveTimestamp)
	enc = encoder.AppendUint32(enc, uint32(secs))
	enc = encoder.AppendUint32(enc, uint32(fracs))

	secs, fracs = timestampToNtpSeconds(pkg.transmitTimestamp)
	enc = encoder.AppendUint32(enc, uint32(secs))
	enc = encoder.AppendUint32(enc, uint32(fracs))

	return enc, nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler interface.
func (pkg *NtpPackage) UnmarshalBinary(data []byte) error {
	// Validate package size
	if len(data) < NTP_PACKAGE_SIZE {
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
	secs, fracs := dec.Uint32(buf[16:]), dec.Uint32(buf[20:])
	pkg.referenceTimestamp = ntpSecondsToTimestamp(secs, fracs)

	secs, fracs = dec.Uint32(buf[24:]), dec.Uint32(buf[28:])
	pkg.originateTimestamp = ntpSecondsToTimestamp(secs, fracs)

	secs, fracs = dec.Uint32(buf[32:]), dec.Uint32(buf[36:])
	pkg.receiveTimestamp = ntpSecondsToTimestamp(secs, fracs)

	secs, fracs = dec.Uint32(buf[40:]), dec.Uint32(buf[44:])
	pkg.transmitTimestamp = ntpSecondsToTimestamp(secs, fracs)

	return nil
}
