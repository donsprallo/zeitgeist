package ntp

import (
	"encoding/binary"
	"net"
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	ntp_time_delta float64
)

type NtpServer struct {
	Host string // Hostname of ntp server
	Port int    // Port of ntp server
}

// Start serving of ntp server. The function is not returning until
// the server received an unhandled error. All known errors are put
// to log and skip the current connection,
func (server *NtpServer) Serve() {
	// Setup socket server address
	addr := &net.UDPAddr{
		IP:   net.ParseIP(server.Host),
		Port: server.Port,
	}

	// Listen to address with udp socket
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Panic(err)
	}

	// Ready for listening, make secure socket closing
	defer conn.Close()
	log.Infof("server listening on %s", conn.LocalAddr())

	for {
		// Read received data from udp socket
		message := make([]byte, 48)
		rlen, raddr, err := conn.ReadFromUDP(message)
		if err != nil {
			// TODO: what is the error reason. We do not need
			// panic here
			log.Panic(err)
		}

		// Be sure that remote address is set
		if raddr == nil {
			log.Warn("request has missing remote address")
			continue
		}
		log.Infof("read %d bytes of data from %s", rlen, raddr)

		go handleRequest(conn, raddr, message)
	}
}

// Create a new ntp server instance.
func NewNtpServer(host string, port int) *NtpServer {
	return &NtpServer{
		Host: host,
		Port: port,
	}
}

func handleRequest(
	conn *net.UDPConn,
	addr *net.UDPAddr,
	data []byte,
) {
	secs, frac := getNtpSeconds(time.Now())

	response := make([]byte, 48)
	binary.BigEndian.PutUint32(response[40:], uint32(secs))
	binary.BigEndian.PutUint32(response[44:], uint32(frac))

	// Send response to client
	log.Infof("write ntp response to %s", addr)
	_, err := conn.WriteToUDP(response, addr)
	if err != nil {
		log.Error(err)
		return
	}
}

// Build a server response.
func buildResponse() (*NtpPackage, error) {
	return &NtpPackage{}, nil
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
