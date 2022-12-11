package server

import (
	"fmt"
	"net"
	"time"

	"github.com/donsprallo/gots/internal/ntp"
	log "github.com/sirupsen/logrus"
)

// Create a new ntp server instance.
func NewNtpServer(
	host string,
	port int,
	routing RoutingStrategy,
) *NtpServer {
	return &NtpServer{
		Host:    host,
		Port:    port,
		Routing: routing,
	}
}

type NtpServer struct {
	Host    string          // Hostname of ntp server
	Port    int             // Port of ntp server
	Routing RoutingStrategy // Routing table
}

// Get the server address string.
func (s *NtpServer) getAddrStr() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

// Start serving of ntp server. The function is not returning until
// the server received an unhandled error. All known errors are put
// to log and skip the current connection,
func (s *NtpServer) Serve() {
	// Setup socket server address
	addr, err := net.ResolveUDPAddr("udp", s.getAddrStr())
	if err != nil {
		log.Panic(err)
	}

	// Listen to address with udp socket
	conn, err := net.ListenUDP(addr.Network(), addr)
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
		go s.handleRequest(conn, raddr, data, rx_timestamp)
	}
}

// Handle an ntp request from conn and remote addr. The connection is not
// closed after request is handled, because the server must wait for a new
// connection.
func (s *NtpServer) handleRequest(
	conn *net.UDPConn,
	addr *net.UDPAddr,
	data []byte,
	rx_timestamp time.Time,
) {
	// Parse request data to a ntp package
	pkg, err := ntp.PackageFromBytes(data)
	if err != nil {
		log.Error(err)
		return
	}

	pkg.SetReceiveTimestamp(rx_timestamp)
	log.Infof("read ntp request %s", pkg)

	// Find response builder by client addr
	handler, err := s.Routing.
		FindResponseBuilder(addr.IP)
	if err != nil {
		log.Error(err)
		return
	}

	// Create response from requested package
	pkg, err = handler.BuildResponse(pkg)
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
