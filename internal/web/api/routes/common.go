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

// Response errors.
var (
	QueryParameterError = ErrorResponse{
		Message: "invalid query parameter"}
	BodyDecodeError = ErrorResponse{
		Message: "can not decode body data"}
	NotFoundError = ErrorResponse{
		Message: "entity not found"}
)

// Create a ntp.Package from request data.
func packageFromReq(_ *http.Request) *ntp.Package {
	// Create default ntp package.
	var pkg ntp.Package
	pkg.SetVersion(ntp.VersionV3)
	pkg.SetMode(ntp.ModeServer)
	pkg.SetStratum(1)
	pkg.SetReferenceClockId([]byte("NICO"))
	return &pkg
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
