// Copyright 2024 The Zeitgeist Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package server

import (
	"fmt"
	"net"
	"time"

	"github.com/donsprallo/zeitgeist/internal/ntp"
	log "github.com/sirupsen/logrus"
)

// NewServer creates a new ntp server instance. A ntp server is serving
// on an udp port to the host interface. Each connection's ip address is
// passed to the routing to find a specific Timer by a ruleset.
func NewServer(
	host string,
	port int,
	routing RoutingStrategy,
) *Server {
	return &Server{
		host:    host,
		port:    port,
		routing: routing,
	}
}

// Server is the ntp server structure.
type Server struct {
	host    string          // host name of ntp server to listen.
	port    int             // port of ntp server to listen.
	routing RoutingStrategy // routing strategy to find Timer.
}

// Serve start serving of the ntp server. The function is not returning until
// the server received an unhandled error. All known errors are write to log
// and skip the current connection,
func (s *Server) Serve() {
	// Setup socket server address.
	addr := s.getAddr()

	// Listen to address with udp socket.
	conn, err := net.ListenUDP(addr.Network(), addr)
	if err != nil {
		log.Panic(err)
	}

	// Ready for listening, make secure socket closing.
	defer func(conn *net.UDPConn) {
		err := conn.Close()
		if err != nil {
			log.Error(err)
		}
	}(conn)
	log.Infof("server listening on %s", s.getAddrStr())

	for {
		// Read received data from remote udp socket.
		data := make([]byte, 48)
		rLen, rAddr, err := conn.ReadFromUDP(data)
		if err != nil {
			// It is possible that the connection is closed. On this
			// case a panic must be logged, because it is not expected
			// and handled by the current server implementation.
			log.Panic(err)
		}

		// Get receive timestamp so fast as possible.
		rxTimestamp := time.Now()

		// Be sure that remote address is set.
		if rAddr == nil {
			log.Warn("request has missing remote address")
			continue
		}
		log.Infof("read %d bytes of data from %s", rLen, rAddr)

		// Handle connections in background.
		go s.handleRequest(conn, rAddr, data, rxTimestamp)
	}

	// TODO: Need to gracefully shutdown
	// log.Info("shutting down")
}

// Get the server address string from host and port.
func (s *Server) getAddrStr() string {
	return fmt.Sprintf("%s:%d", s.host, s.port)
}

// Get the server address from host and port.
func (s *Server) getAddr() *net.UDPAddr {
	addr, err := net.ResolveUDPAddr("udp", s.getAddrStr())
	if err != nil {
		log.Panic(err)
	}
	return addr
}

// Handle a ntp request from conn and remote addr. The connection must not
// be closed after request is handled, because the server must wait for a
// new connection.
func (s *Server) handleRequest(
	conn *net.UDPConn,
	addr *net.UDPAddr,
	data []byte,
	rxTimestamp time.Time,
) {
	// Parse request data to a ntp package.
	pkg, err := ntp.PackageFromBytes(data)
	if err != nil {
		log.Error(err)
		return
	}

	pkg.SetReceiveTimestamp(rxTimestamp)
	log.Infof("read ntp request %s", pkg)

	// Find response timer by client addr.
	timer, err := s.routing.FindTimer(addr.IP)
	if err != nil {
		log.Error(err)
		return
	}

	// Create response from requested package.
	pkg, err = PackageFromTimer(
		pkg, timer.Package(), timer)
	if err != nil {
		log.Error(err)
		return
	}

	// Convert package data to bytes array.
	resBytes, err := pkg.ToBytes()
	if err != nil {
		log.Error(err)
		return
	}

	// Send response to client.
	log.Infof("write ntp response to %s", addr)
	_, err = conn.WriteToUDP(resBytes, addr)
	if err != nil {
		log.Error(err)
		return
	}
}
