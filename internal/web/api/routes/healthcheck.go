package routes

import (
	"github.com/donsprallo/gots/internal/web/api"
	"github.com/gorilla/mux"
	"net/http"
)

// HealthcheckEndpoint is used to check anything that may interrupt
// the API from serving incoming requests. The default route "/" is
// designed to check the service as canary. This means that a status
// value other than the expected one indicates a serious error. Other
// routes should be used for further status checks.
type HealthcheckEndpoint struct {
	handler http.Handler
}

// NewHealthcheckEndpoint creates a new api.Endpoint for healthcheck
// capabilities. The endpoint must be registered with a http.server.
func NewHealthcheckEndpoint() api.Endpoint {
	return &HealthcheckEndpoint{}
}

// RegisterRoutes implements api.Endpoint interface.
func (e *HealthcheckEndpoint) RegisterRoutes(router *mux.Router) {
	e.handler = router

	// The only healthcheck routes
	router.HandleFunc("/", e.healthcheck).
		Methods(http.MethodGet)
	router.HandleFunc("/ping", e.ping).
		Methods(http.MethodGet)
}

// HealthcheckResponse is the response type for the HealthcheckEndpoint
// default route. The response contains only a boolean to display the
// API status.
type HealthcheckResponse struct {
	Status bool `json:"status"`
}

// PingResponse is the response type for the HealthcheckEndpoint ping
// route. The response contains only a boolean to display that the API
// is available.
type PingResponse struct {
	Status bool `json:"status"`
}

// The healthcheck route of the HealthcheckEndpoint verifies multiple items
// and responds with the status of the API and its dependencies. The route
// responds with the HealthcheckResponse and is superior to the ping route.
func (e *HealthcheckEndpoint) healthcheck(
	w http.ResponseWriter, _ *http.Request,
) {
	// Disable cache to prevent http caching from serving the request.
	w.Header().Add("Cache-Control", "no-cache")
	api.MustJsonResponse(w, HealthcheckResponse{
		Status: true,
	}, http.StatusOK)
}

// The ping route of the HealthcheckEndpoint barely checks that the API is
// running and the service is accessible. For this the endpoint always return
// the same result where status is true. Any other response indicates a
// critical system failure.
func (e *HealthcheckEndpoint) ping(
	w http.ResponseWriter, _ *http.Request,
) {
	// Disable cache to prevent http caching from serving the request.
	w.Header().Add("Cache-Control", "no-cache")
	api.MustJsonResponse(w, PingResponse{
		Status: true,
	}, http.StatusOK)
}
