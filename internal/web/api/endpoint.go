package api

import (
	"encoding/json"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type Endpoint interface {
	RegisterRoutes(router *mux.Router)
}

// MustJsonResponse encode a value into json string and write the result
// to response. This must always be made. An error will log a status code
// 400 http.StatusInternalServerError is sent.
func MustJsonResponse(w http.ResponseWriter, v any, status int) {
	// Set response header.
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)

	// Encode value into json string and write to response. This
	// must always be made. On error, we sent server error code.
	err := json.NewEncoder(w).Encode(v)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
