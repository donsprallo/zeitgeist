package api

import (
	"encoding/json"
	"github.com/donsprallo/gots/internal/server"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

type MessageResponse struct {
	Message string `json:"message"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

type TimerResponse struct {
	Id   int    `json:"id"`
	Type string `json:"type"`
}

type TimerValueResponse struct {
	Id    int    `json:"id"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

type TimersResponse struct {
	Length int             `json:"length"`
	Timers []TimerResponse `json:"timers"`
}

type RouteResponse struct {
	Id     int           `json:"id"`
	Subnet string        `json:"subnet"`
	Timer  TimerResponse `json:"timer"`
}

type RoutesResponse struct {
	Length int             `json:"length"`
	Routes []RouteResponse `json:"routes"`
}

// mustJsonResponse encode a value into json string and write the result
// to response. This must always be made. An error will log with panic.
func mustJsonResponse(res http.ResponseWriter, v any) {
	// Set response header.
	res.Header().Add("Content-Type", "application/json")

	// Encode value into json string and write to response. This
	// must always be made. On error, we log with panic.
	err := json.NewEncoder(res).Encode(v)
	if err != nil {
		log.Panic(err)
	}
}

// mustJsonTimerResponse encode a Timer instance to json string and write the
// result to response. This must always be made. An error will log with panic.
func mustJsonTimerResponse(
	res http.ResponseWriter,
	timer server.Timer, id int,
) {
	// Build response with timer data.
	response := TimerValueResponse{
		Id:    id,
		Type:  server.TimerName(timer),
		Value: timer.Get().Format(time.RFC3339),
	}
	mustJsonResponse(res, response)
}
