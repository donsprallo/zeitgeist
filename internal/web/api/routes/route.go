// Copyright 2024 The Zeitgeist Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package routes

import (
	"encoding/json"
	"github.com/donsprallo/zeitgeist/internal/server"
	"github.com/donsprallo/zeitgeist/internal/web/api"
	"github.com/gorilla/mux"
	"net"
	"net/http"
	"strconv"
	"time"
)

type RouteResponse struct {
	Id     int           `json:"id"`
	Subnet string        `json:"subnet"`
	Timer  TimerResponse `json:"timer"`
}

type RouteAllResponse struct {
	Length int             `json:"length"`
	Routes []RouteResponse `json:"routes"`
}

type RouteEndpoint struct {
	handler http.Handler
	timers  *server.TimerCollection // The registered timers
	routes  *server.RoutingTable    // The registered routes
}

func NewRouteEndpoint(
	timers *server.TimerCollection,
	routes *server.RoutingTable,
) api.Endpoint {
	return &RouteEndpoint{
		timers: timers,
		routes: routes,
	}
}

func (e *RouteEndpoint) RegisterRoutes(router *mux.Router) {
	e.handler = router

	// RouteResponse collection management.
	router.HandleFunc("/",
		e.getAllRoutes).Methods(http.MethodGet)
	router.HandleFunc("/",
		e.newRoute).Methods(http.MethodPut)

	// Specific route management.
	router.HandleFunc("/{id:[0-9]+}",
		e.deleteRoute).Methods(http.MethodDelete)
	router.HandleFunc("/{id:[0-9]+}",
		e.getRoute).Methods(http.MethodGet)
	router.HandleFunc("/{id:[0-9]+}",
		e.updateRoute).Methods(http.MethodPost)

	// Default route management
	router.HandleFunc("/default",
		e.getDefaultRoute).Methods(http.MethodGet)
	router.HandleFunc("/default",
		e.updateDefaultRoute).Methods(http.MethodPost)
}

// Return true if net.IPNet is a default route.
func isDefaultRoute(IPNet net.IPNet) bool {
	if IPNet.IP.IsLoopback() ||
		IPNet.IP.IsUnspecified() ||
		IPNet.IP.IsLinkLocalUnicast() {
		return true
	}
	return false
}

// Get the mode and time info from default route.
func (e *RouteEndpoint) getDefaultRoute(
	w http.ResponseWriter, _ *http.Request,
) {
	// Create response, that contains all default routes.
	response := RouteAllResponse{
		Length: 0,
		Routes: make([]RouteResponse, 0, 3),
	}

	// Iterate through all routes and search for alle default routes.
	// Each default route is added to response on demand.
	for _, entry := range e.routes.All() {
		if isDefaultRoute(entry.IPNet) {
			response.Length++
			response.Routes = append(
				// Build RouteResponse from entry and append to
				// response.
				response.Routes, RouteResponse{
					Id:     entry.Id,
					Subnet: entry.IPNet.String(),
					Timer: TimerResponse{
						Id:    entry.TimerId,
						Type:  server.TimerName(entry.Timer),
						Value: entry.Timer.Get().Format(time.RFC3339),
					},
				},
			)
		}
	}

	// Return as JSON response.
	api.MustJsonResponse(
		w, response, http.StatusOK)
}

// Set the mode to default handler. On specific mode, it's possible
// to update settings.
func (e *RouteEndpoint) updateDefaultRoute(
	w http.ResponseWriter, r *http.Request,
) {
	// Decode body data.
	var request UpdateRouteRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		api.MustJsonResponse(
			w, BodyDecodeError, http.StatusBadRequest)
		return
	}

	// Find timer by id.
	timer := e.timers.Get(request.TimerId)
	if timer.Timer == nil {
		api.MustJsonResponse(
			w, NotFoundError, http.StatusBadRequest)
		return
	}

	// Find route by id and update its timer.
	for _, entry := range e.routes.All() {
		if isDefaultRoute(entry.IPNet) {
			err := e.routes.Set(entry.Id, timer.Timer, timer.Id)
			if err != nil {
				api.MustJsonResponse(
					w, NotFoundError, http.StatusBadRequest)
				return
			}
		}
	}

	// Send success response.
	api.MustJsonResponse(w, MessageResponse{
		Message: "default route update success",
	}, http.StatusOK)
}

// Get all registered routes.
func (e *RouteEndpoint) getAllRoutes(
	w http.ResponseWriter, _ *http.Request,
) {
	routes := e.routes.All()
	lenRoutes := len(routes)
	// Build response from routing table entries. We know the size
	// of routing entries here. So we can allocate the size.
	response := RouteAllResponse{
		Length: lenRoutes,
		Routes: make([]RouteResponse, lenRoutes),
	}
	// Iterate through routing entries and add each entry to response.
	// As subnet, we return the CIDR string representation of the ip net.
	// For timer mode, an extra function is converting the timer to its
	// string representation.
	for idx, entry := range routes {
		response.Routes[idx] = RouteResponse{
			Id:     entry.Id,
			Subnet: entry.IPNet.String(),
			Timer: TimerResponse{
				Id:    entry.TimerId,
				Type:  server.TimerName(entry.Timer),
				Value: entry.Timer.Get().Format(time.RFC3339),
			},
		}
	}
	// Return as JSON response.
	api.MustJsonResponse(
		w, response, http.StatusOK)
}

type NewRouteRequest struct {
	TimerId int    `json:"timerId"`
	Subnet  string `json:"subnet"`
}

// Create a new route.
func (e *RouteEndpoint) newRoute(
	w http.ResponseWriter, r *http.Request,
) {
	// Parse body data.
	var routeRequest NewRouteRequest
	err := json.NewDecoder(r.Body).Decode(&routeRequest)
	if err != nil {
		api.MustJsonResponse(
			w, BodyDecodeError, http.StatusBadRequest)
		return
	}

	// Find timer by id.
	timer := e.timers.Get(routeRequest.TimerId)
	if timer.Timer == nil {
		api.MustJsonResponse(w, ErrorResponse{
			Message: "can not find timer",
		}, http.StatusBadRequest)
		return
	}

	// Parse subnet to net.IPNet.
	_, ipNet, err := net.ParseCIDR(routeRequest.Subnet)
	if err != nil {
		api.MustJsonResponse(w, ErrorResponse{
			Message: "can not parse subnet",
		}, http.StatusBadRequest)
		return
	}

	// Add net.IPNet to routing and map to timer instance.
	err = e.routes.Add(*ipNet, timer.Timer, timer.Id)
	if err != nil {
		api.MustJsonResponse(w, ErrorResponse{
			Message: "route with subnet exist",
		}, http.StatusConflict)
		return
	}

	// Build success response.
	api.MustJsonResponse(w, MessageResponse{
		Message: "create new route success",
	}, http.StatusCreated)
}

// Delete an existing route.
func (e *RouteEndpoint) deleteRoute(
	w http.ResponseWriter, r *http.Request,
) {
	// Parse query parameters.
	var vars = mux.Vars(r)
	routeId, err := strconv.Atoi(vars["id"])
	if err != nil {
		api.MustJsonResponse(
			w, QueryParameterError, http.StatusBadRequest)
		return
	}

	// Find route by id.
	route := e.routes.Get(routeId)
	if route == nil {
		api.MustJsonResponse(
			w, NotFoundError, http.StatusBadRequest)
		return
	}

	// Protect default route from deletion.
	if isDefaultRoute(route.IPNet) {
		api.MustJsonResponse(
			w, ErrorResponse{
				Message: "can not delete default route",
			}, http.StatusForbidden)
		return
	}

	// Delete route from routing.
	err = e.routes.Remove(routeId)
	if err != nil {
		api.MustJsonResponse(
			w, NotFoundError, http.StatusBadRequest)
		return
	}

	// Deletion success response.
	api.MustJsonResponse(w, MessageResponse{
		Message: "deletion route success",
	}, http.StatusCreated)
}

// Get a specific route.
func (e *RouteEndpoint) getRoute(
	w http.ResponseWriter, r *http.Request,
) {
	// Parse query parameters.
	var vars = mux.Vars(r)
	routeId, err := strconv.Atoi(vars["id"])
	if err != nil {
		api.MustJsonResponse(
			w, QueryParameterError, http.StatusBadRequest)
		return
	}

	// Find route by id and update its timer.
	route := e.routes.Get(routeId)
	if route == nil {
		api.MustJsonResponse(
			w, NotFoundError, http.StatusBadRequest)
		return
	}

	// Send success response.
	api.MustJsonResponse(w, RouteResponse{
		Id:     route.Id,
		Subnet: route.IPNet.String(),
		Timer: TimerResponse{
			Id:    route.TimerId,
			Type:  server.TimerName(route.Timer),
			Value: route.Timer.Get().Format(time.RFC3339),
		},
	}, http.StatusOK)
}

type UpdateRouteRequest struct {
	TimerId int `json:"timerId"`
}

// Update settings of specific route.
func (e *RouteEndpoint) updateRoute(
	w http.ResponseWriter, r *http.Request,
) {
	// Parse query parameters.
	var vars = mux.Vars(r)
	routeId, err := strconv.Atoi(vars["id"])
	if err != nil {
		api.MustJsonResponse(
			w, QueryParameterError, http.StatusBadRequest)
		return
	}

	// Decode body data.
	var request UpdateRouteRequest
	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		api.MustJsonResponse(
			w, BodyDecodeError, http.StatusBadRequest)
		return
	}

	// Find timer by id.
	timer := e.timers.Get(request.TimerId)
	if timer.Timer == nil {
		api.MustJsonResponse(
			w, NotFoundError, http.StatusBadRequest)
		return
	}

	// Find route by id and update its timer.
	err = e.routes.Set(
		routeId, timer.Timer, timer.Id)
	if err != nil {
		api.MustJsonResponse(
			w, NotFoundError, http.StatusBadRequest)
		return
	}

	// Send success response.
	api.MustJsonResponse(w, MessageResponse{
		Message: "route updated successful",
	}, http.StatusOK)
}
