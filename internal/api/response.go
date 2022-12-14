package api

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
)

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

type TimerResponse struct {
	Id   int    `json:"id"`
	Type string `json:"type"`
}

type TimerValueResponse struct {
	TimerResponse
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
