package api

import (
	"context"
	"fmt"
	"github.com/donsprallo/gots/internal/server"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type Server struct {
	host   string               // The server hostname
	port   int                  // The server port
	router *mux.Router          // The http handler
	rTable *server.RoutingTable // The routing table as database
	server *http.Server         // The http server instance
}

func NewApiServer(
	host string,
	port int,
	router *mux.Router,
	rTable *server.RoutingTable,
) *Server {
	// Create api server
	return &Server{
		host:   host,
		port:   port,
		router: router,
		rTable: rTable,
	}
}

// RegisterRoutes register all possible api routers to handle REST requests.
func (s *Server) RegisterRoutes(
	prefix string,
) {
	// Create api version 1 router from main router.
	apiV1Router := s.router.
		PathPrefix(prefix).Subrouter()
	s.registerApiV1Handlers(apiV1Router)
}

// Serve start listening the Server.
func (s *Server) Serve() {
	// Create http server for REST api.
	s.server = &http.Server{
		Addr:         s.getAddrStr(),
		Handler:      s.router,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	// Start the server by listening.
	log.Infof("api listening on %s", s.getAddrStr())
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

func (s *Server) registerApiV1Handlers(router *mux.Router) {
	// Healthcheck
	router.HandleFunc("/healthcheck",
		s.healthcheckHandler).Methods("GET")

	// Default route management
	router.HandleFunc("/route/default",
		s.getDefaultRouteHandler).Methods("GET")
	router.HandleFunc("/route/default",
		s.updateDefaultRouteHandler).Methods("POST")

	// Route collection management
	router.HandleFunc("/route",
		s.getAllRoutesHandler).Methods("GET")
	router.HandleFunc("/route",
		s.newRouteHandler).Methods("PUT")

	// Specific route management
	router.HandleFunc("/route/{id}",
		s.deleteRouteHandler).Methods("DELETE")
	router.HandleFunc("/route/{id}",
		s.getRouteHandler).Methods("GET")
	router.HandleFunc("/route/{id}",
		s.updateRouteHandler).Methods("POST")
}

// Get the healthcheck response. The healthcheck always return the same
// result. This enables an easy way to automatically check the result.
func (s *Server) healthcheckHandler(
	res http.ResponseWriter,
	_ *http.Request,
) {
	// Always return the same result.
	mustJsonResponse(res, map[string]bool{"ok": true})
}

// Get the mode and time info from default route.
func (s *Server) getDefaultRouteHandler(
	res http.ResponseWriter,
	req *http.Request,
) {
	// Write not implemented status code
	res.WriteHeader(http.StatusNotImplemented)
}

// Set the mode to default router. On specific mode, its possible
// to update settings.
func (s *Server) updateDefaultRouteHandler(
	res http.ResponseWriter,
	req *http.Request,
) {
	// Write not implemented status code
	res.WriteHeader(http.StatusNotImplemented)
}

// Get all registered routes.
func (s *Server) getAllRoutesHandler(
	res http.ResponseWriter,
	_ *http.Request,
) {
	routes := s.rTable.All()
	lenRoutes := len(routes)
	// Build response from routing table entries. We know the size
	// of routing entries here. So we can allocate the size.
	response := Routes{
		Length: lenRoutes,
		Routes: make([]Route, lenRoutes),
	}
	// Iterate through routing entries and add each entry to response.
	// As subnet, we return the CIDR string representation of the ip net.
	// For timer mode, an extra function is converting the timer to its
	// string representation.
	for idx, entry := range routes {
		response.Routes[idx] = Route{
			Id:     idx,
			Subnet: entry.IPNet.String(),
			Timer:  server.TimerName(entry.Timer),
		}
	}
	// Return as JSON response.
	mustJsonResponse(res, response)
}

// Create a new route.
func (s *Server) newRouteHandler(
	res http.ResponseWriter,
	req *http.Request,
) {
	// Write not implemented status code
	res.WriteHeader(http.StatusNotImplemented)
}

// Delete an existing route.
func (s *Server) deleteRouteHandler(
	res http.ResponseWriter,
	req *http.Request,
) {
	// Write not implemented status code
	res.WriteHeader(http.StatusNotImplemented)
}

// Get a specific route.
func (s *Server) getRouteHandler(
	res http.ResponseWriter,
	req *http.Request,
) {
	// Write not implemented status code
	res.WriteHeader(http.StatusNotImplemented)
}

// Update settings of specific route.
func (s *Server) updateRouteHandler(
	res http.ResponseWriter,
	req *http.Request,
) {
	// Write not implemented status code
	res.WriteHeader(http.StatusNotImplemented)
}
