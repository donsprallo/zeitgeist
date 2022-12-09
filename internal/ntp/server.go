package ntp

import (
	"net"
	"time"

	log "github.com/sirupsen/logrus"
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
	pkg, err := PackageFromBytes(data)
	if err != nil {
		log.Error(err)
		return
	}
	log.Infof("read ntp request %s", pkg)

	// Create response from requested package
	pkg, err = buildNtpResponse(
		pkg, rx_timestamp)
	if err != nil {
		log.Error(err)
		return
	}

	// Convert package data to bytes array
	res_bytes, err := pkg.ToBytes()
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
	pkg *NtpPackage,
	rx_timestamp time.Time,
) (*NtpPackage, error) {
	// Set header
	pkg.SetLeap(NTP_LI_ADD_SEC)
	pkg.SetVersion(NTP_VN_V3)
	pkg.SetMode(NTP_MODE_SERVER)
	pkg.SetStratum(3)
	pkg.SetPoll(2)
	pkg.SetPrecision(2)

	// Set root delay
	pkg.SetRootDelay(1)
	pkg.SetRootDispersion(2)
	pkg.SetReferenceClockId([]byte("ABCD"))
	pkg.SetReferenceTimestamp(time.Now())
	pkg.SetOriginateTimestamp(time.Now())
	pkg.SetReceiveTimestamp(rx_timestamp)
	// Set transmit timestamp at least before sent
	pkg.SetTransmitTimestamp(time.Now())

	return pkg, nil
}
