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

// Create a new ntp server instance.
func NewNtpServer(host string, port int) *NtpServer {
	return &NtpServer{
		Host: host,
		Port: port,
	}
}

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
		data := make([]byte, 48)
		rlen, raddr, err := conn.ReadFromUDP(data)
		if err != nil {
			// Its possible that the connection is closed. This case is a panic
			// because it is not expected and handled by the server.
			log.Panic(err)
		}

		// Get receive timestamp so fast as possible.
		rx_timestamp := time.Now()

		// Be sure that remote address is set
		if raddr == nil {
			log.Warn("request has missing remote address")
			continue
		}
		log.Infof("read %d bytes of data from %s", rlen, raddr)

		// Handle connection in background. We can wait here for
		// new packages
		go handleNtpRequest(conn, raddr, data, rx_timestamp)
	}
}

// Handle an ntp request from conn and remote addr. The connection is not
// closed after request is handled, because the server must wait for a new
// connection.
func handleNtpRequest(
	conn *net.UDPConn,
	addr *net.UDPAddr,
	data []byte,
	rx_timestamp time.Time,
) {
	// Parse request data to a ntp package
	req_ntp_pkg, err := PackageFromBytes(data)
	if err != nil {
		log.Error(err)
		return
	}
	log.Infof("read ntp request %s", req_ntp_pkg)

	// Create response from requested package
	res, err := buildNtpResponse(req_ntp_pkg)
	if err != nil {
		log.Error(err)
		return
	}

	// Convert package data to bytes array
	res_bytes, err := res.ToBytes()
	if err != nil {
		log.Error(err)
		return
	}

	// Send response to client
	log.Infof("write ntp response to %s", addr)
	_, err = conn.WriteToUDP(res_bytes, addr)
	if err != nil {
		log.Error(err)
		return
	}
}

// Build a ntp package from response and current time.
func buildNtpResponse(
	reqPkg *NtpPackage,
) (*NtpPackage, error) {
	secs, frac := getNtpSeconds(time.Now())

	response := make([]byte, 48)
	binary.BigEndian.PutUint32(response[40:], uint32(secs))
	binary.BigEndian.PutUint32(response[44:], uint32(frac))

	return PackageFromBytes(response)
}
