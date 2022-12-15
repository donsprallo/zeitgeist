package routes

import (
	"encoding/json"
	"github.com/donsprallo/gots/internal/server"
	"github.com/donsprallo/gots/internal/web/api"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"time"
)

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

type TimerEndpoint struct {
	handler http.Handler
	timers  *server.TimerCollection // The registered timers
}

func NewTimerEndpoint(
	timers *server.TimerCollection,
) api.Endpoint {
	return &TimerEndpoint{
		timers: timers,
	}
}

func (e *TimerEndpoint) RegisterRoutes(router *mux.Router) {
	e.handler = router

	// TimerResponse collection management.
	router.HandleFunc("/",
		e.getAllTimers).Methods(http.MethodGet)
	router.HandleFunc("/ntp",
		e.newNtpTimer).Methods(http.MethodPut)
	router.HandleFunc("/system",
		e.newSystemTimer).Methods(http.MethodPut)
	router.HandleFunc("/modify",
		e.newModifyTimer).Methods(http.MethodPut)

	// Specific timer management.
	router.HandleFunc("/{id}",
		e.deleteTimer).Methods(http.MethodDelete)
	router.HandleFunc("/{id}",
		e.getTimer).Methods(http.MethodGet)
	router.HandleFunc("/{id}",
		e.updateTimer).Methods(http.MethodPost)
}

// Get all registered timers.
func (e *TimerEndpoint) getAllTimers(
	w http.ResponseWriter, _ *http.Request,
) {
	timers := e.timers.All()
	// Build response from timers collection. We know the size
	// of timer collection here. So we can allocate the size.
	response := TimersResponse{
		Length: e.timers.Length(),
		Timers: make([]TimerResponse, e.timers.Length()),
	}
	// Iterate through timers and add each entry to response.
	for idx, entry := range timers {
		response.Timers[idx] = TimerResponse{
			Id:   idx,
			Type: server.TimerName(entry.Timer),
		}
	}
	// Return as JSON response.
	api.MustJsonResponse(
		w, response, http.StatusOK)
}

// Create a new NtpTimer.
func (e *TimerEndpoint) newNtpTimer(
	w http.ResponseWriter, r *http.Request,
) {
	// Create new timer from request data.
	ntpPackage := packageFromReq(r)
	timer := &server.NtpTimer{
		NTPPackage: *ntpPackage,
	}
	// Add timer to collection.
	idx := e.timers.Add(timer)
	mustJsonTimerResponse(
		w, timer, idx, http.StatusCreated)
}

// Create a new SystemTimer.
func (e *TimerEndpoint) newSystemTimer(
	w http.ResponseWriter, r *http.Request,
) {
	// Create new timer from request data.
	ntpPackage := packageFromReq(r)
	timer := &server.SystemTimer{
		NTPPackage: *ntpPackage,
	}
	// Add timer to collection.
	idx := e.timers.Add(timer)
	mustJsonTimerResponse(
		w, timer, idx, http.StatusCreated)
}

// Create a new ModifyTimer.
func (e *TimerEndpoint) newModifyTimer(
	w http.ResponseWriter, r *http.Request,
) {
	// Create new timer from request data.
	ntpPackage := packageFromReq(r)
	timer := &server.ModifyTimer{
		NTPPackage: *ntpPackage,
		Time:       time.Now(),
	}
	// Add timer to collection.
	idx := e.timers.Add(timer)
	mustJsonTimerResponse(
		w, timer, idx, http.StatusCreated)
}

// Delete an existing server.Timer instance from collection.
func (e *TimerEndpoint) deleteTimer(
	w http.ResponseWriter, r *http.Request,
) {
	// Parse query parameters.
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		api.MustJsonResponse(w, ErrorResponse{
			Message: "invalid query id",
		}, http.StatusOK)
		return
	}
	// Delete timer by id.
	err = e.timers.Delete(id)
	if err != nil {
		api.MustJsonResponse(w, ErrorResponse{
			Message: err.Error(),
		}, http.StatusOK)
		return
	}
	// Timer successful deleted.
	api.MustJsonResponse(w, MessageResponse{
		Message: "delete timer success",
	}, http.StatusAccepted)
}

// Get a specific route.
func (e *TimerEndpoint) getTimer(
	w http.ResponseWriter, r *http.Request,
) {
	// Parse query parameters.
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		api.MustJsonResponse(w, ErrorResponse{
			Message: "invalid query id",
		}, http.StatusOK)
		return
	}
	// Get timer by id.
	timer := e.timers.Get(id)
	if timer.Timer == nil {
		api.MustJsonResponse(w, ErrorResponse{
			Message: "can not find timer by id",
		}, http.StatusOK)
		return
	}
	// Make response with timer.
	mustJsonTimerResponse(
		w, timer.Timer, id, http.StatusOK)
}

// Update settings of specific route.
func (e *TimerEndpoint) updateTimer(
	w http.ResponseWriter, r *http.Request,
) {
	// Parse query parameters.
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		api.MustJsonResponse(w, ErrorResponse{
			Message: "invalid query id",
		}, http.StatusBadRequest)
		return
	}
	// Get timer by id.
	timer := e.timers.Get(id)
	if timer.Timer == nil {
		api.MustJsonResponse(w, ErrorResponse{
			Message: "can not find timer by id",
		}, http.StatusNoContent)
		return
	}

	// Build response from timer type.
	switch timer.Timer.(type) {
	case *server.ModifyTimer:
		// Parse body parameters for ModifyTimer.
		body := make(map[string]string, 0)
		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			api.MustJsonResponse(w, ErrorResponse{
				Message: "can not decode body data",
			}, http.StatusBadRequest)
			return
		}
		// Parse time value from body
		timeLayout := time.RFC822
		timeVal, err := time.Parse(
			timeLayout, body["time"])
		if err != nil {
			api.MustJsonResponse(w, ErrorResponse{
				Message: "can not parse time",
			}, http.StatusBadRequest)
			return
		}
		// Set timer with value.
		timer.Timer.Set(timeVal)
		api.MustJsonResponse(w, MessageResponse{
			Message: "timer update successful",
		}, http.StatusOK)
		return
	default:
		api.MustJsonResponse(w, ErrorResponse{
			Message: "timer can not modified",
		}, http.StatusConflict)
		return
	}
}
