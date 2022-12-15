package routes

import (
	"github.com/donsprallo/gots/internal/ntp"
	"github.com/donsprallo/gots/internal/server"
	"github.com/donsprallo/gots/internal/web/api"
	"net/http"
	"time"
)

type MessageResponse struct {
	Message string `json:"message"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

// Create a ntp.Package from request data.
func packageFromReq(_ *http.Request) *ntp.Package {
	return &ntp.Package{}
}

// mustJsonTimerResponse encode a Timer instance to json string and write the
// result to response. This must always be made. An error will log with panic.
func mustJsonTimerResponse(
	w http.ResponseWriter,
	timer server.Timer,
	id int,
	status int,
) {
	// Build response with timer data.
	response := TimerValueResponse{
		Id:    id,
		Type:  server.TimerName(timer),
		Value: timer.Get().Format(time.RFC3339),
	}
	api.MustJsonResponse(w, response, status)
}
