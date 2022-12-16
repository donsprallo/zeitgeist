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

	// The only healthcheck route.
	router.HandleFunc("/", e.healthcheck).
		Methods(http.MethodGet)
}

// HealthcheckResponse is the response type for the HealthcheckEndpoint
// default route. The response contains only a boolean to display the
// API status.
type HealthcheckResponse struct {
	Status bool `json:"status"`
}

// This is the default route for the HealthcheckEndpoint. The healthcheck
// route always response with the HealthcheckResponse and status true. Any
// other response indicates a critical system failure.
func (e *HealthcheckEndpoint) healthcheck(
	w http.ResponseWriter, _ *http.Request,
) {
	// Disable cache to prevent http caching from serving the
	// request. As a result, every request to the endpoint returns
	// the most up-to-date status of the service.
	w.Header().Add("Cache-Control", "no-cache")
	api.MustJsonResponse(w, HealthcheckResponse{
		Status: true,
	}, http.StatusOK)
}
