// Copyright 2024 The Zeitgeist Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package web

import (
	"context"
	"fmt"
	"github.com/donsprallo/zeitgeist/internal/web/api"
	"github.com/gorilla/mux"
	"net"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

type Server struct {
	host    string       // The server hostname
	port    int          // The server port
	handler *mux.Router  // The http handler
	server  *http.Server // The http server instance
}

// NewServer creates a new web server instance. The server is listening on
// host interface and port. A handler handles incoming requests.
func NewServer(
	host string,
	port int,
	handler *mux.Router,
) *Server {
	// Create web server
	return &Server{
		host:    host,
		port:    port,
		handler: handler,
	}
}

// Serve start listening the Server.
func (s *Server) Serve() {
	// Create http server for REST web.
	s.server = &http.Server{
		Addr:         s.getAddrStr(),
		Handler:      s.handler,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	// Start the server by listening.
	log.Infof("web server listening on %s", s.getAddrStr())
	if err := s.server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

// Shutdown handle gracefully shutdown without interrupt active connections.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// Get the server address string from host and port.
func (s *Server) getAddrStr() string {
	return fmt.Sprintf("%s:%d", s.host, s.port)
}

// Get the server address from host and port.
func (s *Server) getAddr() *net.TCPAddr {
	addr, err := net.ResolveTCPAddr("tcp", s.getAddrStr())
	if err != nil {
		log.Panic(err)
	}
	return addr
}

// RegisterEndpoint add an endpoint to the server. A prefix can be used to
// specify a sub route that is handled by the endpoint.
func (s *Server) RegisterEndpoint(
	prefix string,
	endpoint api.Endpoint,
) {
	// Create sub router for an endpoint. The endpoint can register
	// its routes to this router.
	router := s.handler.
		PathPrefix(prefix).
		Subrouter()
	// The endpoint must register its routes to the sub router.
	endpoint.RegisterRoutes(router)
}
