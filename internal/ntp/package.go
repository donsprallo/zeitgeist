package ntp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"unsafe"

	log "github.com/sirupsen/logrus"
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

type NtpHeader struct {
	Fields uint32
}

func (header NtpHeader) GetLeap() int {
	return int((header.Fields & 0x0000_0003))
}

func (header *NtpHeader) SetLeap(val int) {

}

func (header NtpHeader) GetVersion() int {
	return int((header.Fields & 0x0000_001C) >> 2)
}

func (header *NtpHeader) SetVersion(val int) {

}

func (header NtpHeader) GetMode() int {
	return int((header.Fields & 0x0000_E000) >> 5)
}

func (header *NtpHeader) SetMode(val int) {

}

func (header NtpHeader) GetStratum() int {
	return int((header.Fields & 0x0000_FF00) >> 7)
}

func (header *NtpHeader) SetStratum(val int) {

}

func (header NtpHeader) GetPoll() int {
	return int((header.Fields & 0x00FF_0000) >> 16)
}

func (header *NtpHeader) SetPoll(val int) {

}

func (header NtpHeader) GetPrecision() int {
	return int((header.Fields & 0xFF00_0000) >> 24)
}

func (header *NtpHeader) SetPrecision(val int) {

}

type NtpPackage struct {
	Header               NtpHeader
	RootDelay            int32
	ReferenceId          int32
	ReferenceTimestamp   int64
	OriginTimestamp      int64
	ReceiveTimestamp     int64
	TransmitTimestamp    int64
	DestinationTimestamp int64
	KeyId                int64
	Digest               int64
}

// Stringer interface.
func (pkg *NtpPackage) String() string {
	return fmt.Sprintf("<NtpPackage(tx=%d, stratum=%d)>",
		pkg.TransmitTimestamp, pkg.Header.GetStratum())
}

func (pkg *NtpPackage) ToBytes() []byte {
	return make([]byte, 48)
}

// Parse ntp package from bytes.
func PackageFromBytes(data []byte) (*NtpPackage, error) {
	var pkg NtpPackage
	log.Debug(unsafe.Sizeof(pkg))
	if len(data) != 16 {
		return nil, errors.New(
			"invalid ntp package")
	}
	buffer := bytes.NewBuffer(data)
	binary.Read(buffer, binary.BigEndian, pkg)
	return &pkg, nil
}
