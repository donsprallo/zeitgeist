package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/donsprallo/gots/internal/server"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type Server struct {
	host    string               // The server hostname
	port    int                  // The server port
	router  *mux.Router          // The http handler
	routing *server.RoutingTable // The routing as database
	server  *http.Server         // The http server instance
}

func NewApiServer(
	host string,
	port int,
	router *mux.Router,
	routing *server.RoutingTable,
) *Server {
	// Create api server
	return &Server{
		host:    host,
		port:    port,
		router:  router,
		routing: routing,
	}
}

// RegisterRoutes register all possible api routers to handle REST requests.
func (s *Server) RegisterRoutes(
	prefix string,
) {
	// Create api version 1 router from main router.
	apiV1Router := s.router.
		PathPrefix(prefix).Subrouter()
	registerApiV1Handlers(apiV1Router)
}

// Serve start listening the Server.
func (s *Server) Serve() {
	// Create http server for REST api.
	s.server = &http.Server{
		Handler:      s.router,
		Addr:         s.getAddrStr(),
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

func (s *Server) getAddrStr() string {
	return fmt.Sprintf("%s:%d", s.host, s.port)
}

func registerApiV1Handlers(router *mux.Router) {
	// Healthcheck
	router.HandleFunc("/healthcheckHandler",
		healthcheckHandler).Methods("GET")

	// Default route management
	router.HandleFunc("/route/default",
		getDefaultRouteHandler).Methods("GET")
	router.HandleFunc("/route/default",
		updateDefaultRouteHandler).Methods("POST")

	// Route collection management
	router.HandleFunc("/route",
		getAllRoutesHandler).Methods("GET")
	router.HandleFunc("/route",
		newRouteHandler).Methods("PUT")

	// Specific route management
	router.HandleFunc("/route/{id}",
		deleteRouteHandler).Methods("DELETE")
	router.HandleFunc("/route/{id}",
		getRouteHandler).Methods("GET")
	router.HandleFunc("/route/{id}",
		updateRouteHandler).Methods("POST")
}

// Get the healthcheck response. The healthcheck always return the same
// result. This enables an easy way to automatically check the result.
func healthcheckHandler(
	res http.ResponseWriter,
	req *http.Request,
) {
	// Set response header.
	res.Header().Add("Content-Type", "application/json")

	// Always return the same result.
	err := json.NewEncoder(res).Encode(map[string]bool{"ok": true})
	if err != nil {
		log.Panic(err)
	}
}

// Get the mode and time info from default route.
func getDefaultRouteHandler(
	res http.ResponseWriter,
	req *http.Request,
) {
	// Write not implemented status code
	res.WriteHeader(http.StatusNotImplemented)
}

// Set the mode to default router. On specific mode, its possible
// to update settings.
func updateDefaultRouteHandler(
	res http.ResponseWriter,
	req *http.Request,
) {
	// Write not implemented status code
	res.WriteHeader(http.StatusNotImplemented)
}

// Get all registered routes.
func getAllRoutesHandler(
	res http.ResponseWriter,
	req *http.Request,
) {
	// Write not implemented status code
	res.WriteHeader(http.StatusNotImplemented)
}

// Create a new route.
func newRouteHandler(
	res http.ResponseWriter,
	req *http.Request,
) {
	// Write not implemented status code
	res.WriteHeader(http.StatusNotImplemented)
}

// Delete an existing route.
func deleteRouteHandler(
	res http.ResponseWriter,
	req *http.Request,
) {
	// Write not implemented status code
	res.WriteHeader(http.StatusNotImplemented)
}

// Get a specific route.
func getRouteHandler(
	res http.ResponseWriter,
	req *http.Request,
) {
	// Write not implemented status code
	res.WriteHeader(http.StatusNotImplemented)
}

// Update settings of specific route.
func updateRouteHandler(
	res http.ResponseWriter,
	req *http.Request,
) {
	// Write not implemented status code
	res.WriteHeader(http.StatusNotImplemented)
}
