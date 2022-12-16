package routes

import (
	"github.com/donsprallo/gots/internal/web/api"
	"github.com/gorilla/mux"
	"net/http"
)

// Healthy interface is used to check the health status of a system.
// If IsHealthy returns false, then the system is in failure stat. The
// error can be obtained with the builtin error interface.
type Healthy interface {

	// IsHealthy checks if something is OK or faulty. When all is OK
	// true is returned, otherwise false is returned.
	IsHealthy() bool

	error
}

// HealthcheckEndpoint is used to check anything that may interrupt
// the API from serving incoming requests. The default route "/" is
// designed to check the service as canary. This means that a status
// value other than the expected one indicates a serious error. Other
// routes should be used for further status checks.
type HealthcheckEndpoint struct {
	handler  http.Handler       // The http handler
	checkers map[string]Healthy // A map of health checkers
}

// NewHealthcheckEndpoint creates a new api.Endpoint for healthcheck
// capabilities. The endpoint must be registered with a http.server.
func NewHealthcheckEndpoint() api.Endpoint {
	return &HealthcheckEndpoint{
		checkers: make(map[string]Healthy),
	}
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

// AddChecker adds a Healthy checkers with a name to the HealthcheckEndpoint.
// The checkers are used in a healthcheck request to check the state of the
// system.
func (e *HealthcheckEndpoint) AddChecker(
	name string, checker Healthy) {
	e.checkers[name] = checker
}

// RemoveChecker deletes a Healthy checkers from the HealthcheckEndpoint.
func (e *HealthcheckEndpoint) RemoveChecker(name string) {
	delete(e.checkers, name)
}

// HealthcheckResponse is the response type for the HealthcheckEndpoint
// healthcheck route. The response contains a boolean to display the API
// status and a map of errors.
type HealthcheckResponse struct {
	Status bool              `json:"status"`
	Errors map[string]string `json:"errors"`
}

// PingResponse is the response type for the HealthcheckEndpoint ping
// route. The response contains only a string to display that the API
// is available.
type PingResponse struct {
	Status string `json:"status"`
}

// The healthcheck route of the HealthcheckEndpoint verifies multiple items
// and responds with the status of the API and its dependencies. The route
// responds with the HealthcheckResponse and is superior to the ping route.
func (e *HealthcheckEndpoint) healthcheck(
	w http.ResponseWriter, _ *http.Request,
) {
	// Check all dependencies. On error add information to map.
	apiErrors := make(map[string]string)
	for name, checker := range e.checkers {
		if !checker.IsHealthy() {
			// Add info on error detection.
			apiErrors[name] = checker.Error()
		}
	}
	// Set response status indicators.
	hasErrors := len(apiErrors) != 0
	statusCode := http.StatusOK
	if hasErrors {
		statusCode = http.StatusBadRequest
	}
	// Disable cache to prevent http caching from serving the request.
	w.Header().Add("Cache-Control", "no-cache")
	api.MustJsonResponse(w, HealthcheckResponse{
		Status: !hasErrors,
		Errors: apiErrors,
	}, statusCode)
}

// The ping route of the HealthcheckEndpoint barely checks that the API is
// running and the service is accessible. For this the endpoint always return
// the same result where status is "running". Any other response indicates a
// critical system failure.
func (e *HealthcheckEndpoint) ping(
	w http.ResponseWriter, _ *http.Request,
) {
	// Disable cache to prevent http caching from serving the request.
	w.Header().Add("Cache-Control", "no-cache")
	api.MustJsonResponse(w, PingResponse{
		Status: "running",
	}, http.StatusOK)
}
