package routes

import (
	"github.com/donsprallo/gots/internal/web/api"
	"github.com/gorilla/mux"
	"net/http"
)

type HealthcheckEndpoint struct {
	handler http.Handler
}

func NewHealthcheckEndpoint() api.Endpoint {
	return &HealthcheckEndpoint{}
}

func (e *HealthcheckEndpoint) RegisterRoutes(router *mux.Router) {
	e.handler = router

	// The only healthcheck route.
	router.HandleFunc("/", e.healthcheck).
		Methods(http.MethodGet)
}

type HealthcheckResponse struct {
	Status bool `json:"status"`
}

// Get the healthcheck response. The healthcheck always return the same
// result. This enables an easy way to automatically check the result.
func (e *HealthcheckEndpoint) healthcheck(
	w http.ResponseWriter, _ *http.Request,
) {
	api.MustJsonResponse(w, HealthcheckResponse{
		Status: true,
	}, http.StatusOK)
}
