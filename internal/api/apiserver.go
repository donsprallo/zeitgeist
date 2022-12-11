package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type ApiServer struct {
	Host   string       // Server hostname
	Port   int          // Server port
	router *mux.Router  // Server router
	server *http.Server // Http listener
}

func NewApiServer(
	host string,
	port int,
	router *mux.Router,
) *ApiServer {
	// Create ApiServer
	return &ApiServer{
		Host:   host,
		Port:   port,
		router: router,
	}
}

// Register all possible api routers to handle REST requests.
func (s *ApiServer) RegisterRoutes(
	prefix string,
) {
	// Create api version 1 router from main router
	apiV1Router := s.router.
		PathPrefix(prefix).Subrouter()
	reigisterApiV1Handlers(apiV1Router)
}

func (s *ApiServer) Serve() {
	// Create http server for REST api
	s.server = &http.Server{
		Handler:      s.router,
		Addr:         s.getAddrStr(),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	// Start api server
	log.Infof("api listening on %s", s.getAddrStr())
	if err := s.server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func (s *ApiServer) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func (s *ApiServer) getAddrStr() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

func reigisterApiV1Handlers(router *mux.Router) {
	// Healthcheck
	router.HandleFunc("/healthcheck",
		healthcheck).Methods("GET")

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
// result. This enable an easy way to automatically check the result.
func healthcheck(
	res http.ResponseWriter, req *http.Request) {
	res.Header().Add("Content-Type", "application/json")
	// Always return the same result
	json.NewEncoder(res).Encode(map[string]bool{"ok": true})
}

// Get the mode and time info from default route.
func getDefaultRouteHandler(
	res http.ResponseWriter, req *http.Request) {
	// Write not implemented status code
	res.WriteHeader(http.StatusNotImplemented)
}

// Set the mode to default router. On specific mode, its possible
// to update settings.
func updateDefaultRouteHandler(
	res http.ResponseWriter, req *http.Request) {
	// Write not implemented status code
	res.WriteHeader(http.StatusNotImplemented)
}

// Get all registered routes.
func getAllRoutesHandler(
	res http.ResponseWriter, req *http.Request) {
	// Write not implemented status code
	res.WriteHeader(http.StatusNotImplemented)
}

// Create a new route.
func newRouteHandler(
	res http.ResponseWriter, req *http.Request) {
	// Write not implemented status code
	res.WriteHeader(http.StatusNotImplemented)
}

// Celete an existing route.
func deleteRouteHandler(
	res http.ResponseWriter, req *http.Request) {
	// Write not implemented status code
	res.WriteHeader(http.StatusNotImplemented)
}

// Get a specific route.
func getRouteHandler(
	res http.ResponseWriter, req *http.Request) {
	// Write not implemented status code
	res.WriteHeader(http.StatusNotImplemented)
}

// Update setings of specific route.
func updateRouteHandler(
	res http.ResponseWriter, req *http.Request) {
	// Write not implemented status code
	res.WriteHeader(http.StatusNotImplemented)
}
