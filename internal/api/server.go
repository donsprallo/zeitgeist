package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/donsprallo/gots/internal/ntp"
	"github.com/donsprallo/gots/internal/server"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type Server struct {
	host   string                  // The server hostname
	port   int                     // The server port
	router *mux.Router             // The http handler
	routes *server.RoutingTable    // The registered routes
	timers *server.TimerCollection // The registered timers
	server *http.Server            // The http server instance
}

func NewApiServer(
	host string,
	port int,
	router *mux.Router,
	routes *server.RoutingTable,
	timers *server.TimerCollection,
) *Server {
	// Create api server
	return &Server{
		host:   host,
		port:   port,
		router: router,
		routes: routes,
		timers: timers,
	}
}

// RegisterRoutes register all possible api routers to handle REST requests.
func (s *Server) RegisterRoutes(
	prefix string,
) {
	// Create api version 1 router from main router.
	apiV1Router := s.router.
		PathPrefix(prefix).Subrouter()
	s.registerApiV1Handlers(apiV1Router)
}

// Serve start listening the Server.
func (s *Server) Serve() {
	// Create http server for REST api.
	s.server = &http.Server{
		Addr:         s.getAddrStr(),
		Handler:      s.router,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	// Start the server by listening.
	log.Infof("api listening on %s", s.getAddrStr())
	if err := s.server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

// Shutdown handle gracefully shutdown without interrupt active connections.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// Get the server address string from host and port.
func (s *Server) getAddrStr() string {
	return fmt.Sprintf("%s:%d", s.host, s.port)
}

// Get the server address from host and port.
func (s *Server) getAddr() *net.TCPAddr {
	addr, err := net.ResolveTCPAddr("tcp", s.getAddrStr())
	if err != nil {
		log.Panic(err)
	}
	return addr
}

func (s *Server) registerApiV1Handlers(router *mux.Router) {
	// Healthcheck
	router.HandleFunc("/healthcheck",
		s.healthcheckHandler).Methods("GET")

	// TimerResponse collection management
	router.HandleFunc("/timer",
		s.getAllTimersHandler).Methods("GET")
	router.HandleFunc("/timer/ntp",
		s.newNtpTimerHandler).Methods("PUT")
	router.HandleFunc("/timer/system",
		s.newSystemTimerHandler).Methods("PUT")
	router.HandleFunc("/timer/modify",
		s.newModifyTimerHandler).Methods("PUT")

	// Specific timer management
	router.HandleFunc("/timer/{id}",
		s.deleteTimerHandler).Methods("DELETE")
	router.HandleFunc("/timer/{id}",
		s.getTimerHandler).Methods("GET")
	router.HandleFunc("/timer/{id}",
		s.updateTimerHandler).Methods("POST")

	// RouteResponse collection management
	router.HandleFunc("/route",
		s.getAllRoutesHandler).Methods("GET")
	router.HandleFunc("/route",
		s.newRouteHandler).Methods("PUT")

	// Specific route management
	router.HandleFunc("/route/{id}",
		s.deleteRouteHandler).Methods("DELETE")
	router.HandleFunc("/route/{id}",
		s.getRouteHandler).Methods("GET")
	router.HandleFunc("/route/{id}",
		s.updateRouteHandler).Methods("POST")

	// Default route management
	router.HandleFunc("/route/default",
		s.getDefaultRouteHandler).Methods("GET")
	router.HandleFunc("/route/default",
		s.updateDefaultRouteHandler).Methods("POST")
}

// Get the healthcheck response. The healthcheck always return the same
// result. This enables an easy way to automatically check the result.
func (s *Server) healthcheckHandler(
	res http.ResponseWriter,
	_ *http.Request,
) {
	// Always return the same result.
	mustJsonResponse(res, map[string]bool{"ok": true})
}

// Create a ntp.Package from request data.
func packageFromReq(_ *http.Request) *ntp.Package {
	return &ntp.Package{}
}

// Get all registered timers.
func (s *Server) getAllTimersHandler(
	res http.ResponseWriter,
	_ *http.Request,
) {
	timers := s.timers.All()
	// Build response from timers collection. We know the size
	// of timer collection here. So we can allocate the size.
	response := TimersResponse{
		Length: s.timers.Length(),
		Timers: make([]TimerResponse, s.timers.Length()),
	}
	// Iterate through timers and add each entry to response.
	for idx, entry := range timers {
		response.Timers[idx] = TimerResponse{
			Id:   idx,
			Type: server.TimerName(entry.Timer),
		}
	}
	// Return as JSON response.
	mustJsonResponse(res, response)
}

// Create a new NtpTimer.
func (s *Server) newNtpTimerHandler(
	res http.ResponseWriter,
	req *http.Request,
) {
	// Create new timer from request data.
	ntpPackage := packageFromReq(req)
	timer := &server.NtpTimer{
		NTPPackage: *ntpPackage,
	}
	// Add timer to collection.
	idx := s.timers.Add(timer)
	mustJsonTimerResponse(res, timer, idx)
}

// Create a new SystemTimer.
func (s *Server) newSystemTimerHandler(
	res http.ResponseWriter,
	req *http.Request,
) {
	// Create new timer from request data.
	ntpPackage := packageFromReq(req)
	timer := &server.SystemTimer{
		NTPPackage: *ntpPackage,
	}
	// Add timer to collection.
	idx := s.timers.Add(timer)
	mustJsonTimerResponse(res, timer, idx)
}

// Create a new ModifyTimer.
func (s *Server) newModifyTimerHandler(
	res http.ResponseWriter,
	req *http.Request,
) {
	// Create new timer from request data.
	ntpPackage := packageFromReq(req)
	timer := &server.ModifyTimer{
		NTPPackage: *ntpPackage,
		Time:       time.Now(),
	}
	// Add timer to collection.
	idx := s.timers.Add(timer)
	mustJsonTimerResponse(res, timer, idx)
}

// Delete an existing server.Timer instance from collection.
func (s *Server) deleteTimerHandler(
	res http.ResponseWriter,
	req *http.Request,
) {
	// Parse query parameters.
	vars := mux.Vars(req)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		mustJsonResponse(res, ErrorResponse{
			Message: "invalid query id",
		})
		return
	}
	// Delete timer by id.
	err = s.timers.Delete(id)
	if err != nil {
		mustJsonResponse(res, ErrorResponse{
			Message: err.Error(),
		})
		return
	}
	// Timer successful deleted.
	mustJsonResponse(res, MessageResponse{
		Message: "delete timer success",
	})
}

// Get a specific route.
func (s *Server) getTimerHandler(
	res http.ResponseWriter,
	req *http.Request,
) {
	// Parse query parameters.
	vars := mux.Vars(req)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		mustJsonResponse(res, ErrorResponse{
			Message: "invalid query id",
		})
		return
	}
	// Get timer by id.
	timer := s.timers.Get(id)
	if timer.Timer == nil {
		mustJsonResponse(res, ErrorResponse{
			Message: "can not find timer by id",
		})
		return
	}
	// Make response with timer.
	mustJsonTimerResponse(
		res, timer.Timer, id)
}

// Update settings of specific route.
func (s *Server) updateTimerHandler(
	res http.ResponseWriter,
	req *http.Request,
) {
	// Parse query parameters.
	vars := mux.Vars(req)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		mustJsonResponse(res, ErrorResponse{
			Message: "invalid query id",
		})
		return
	}
	// Get timer by id.
	timer := s.timers.Get(id)
	if timer.Timer == nil {
		mustJsonResponse(res, ErrorResponse{
			Message: "can not find timer by id",
		})
		return
	}

	// Build response from timer type.
	switch timer.Timer.(type) {
	case *server.ModifyTimer:
		// Parse body parameters for ModifyTimer.
		body := make(map[string]string, 0)
		err := json.NewDecoder(req.Body).Decode(&body)
		if err != nil {
			mustJsonResponse(res, ErrorResponse{
				Message: "can not decode body data",
			})
			return
		}
		// Parse time value from body
		timeLayout := time.RFC822
		timeVal, err := time.Parse(
			timeLayout, body["time"])
		if err != nil {
			mustJsonResponse(res, ErrorResponse{
				Message: "can not parse time",
			})
			return
		}
		// Set timer with value.
		timer.Timer.Set(timeVal)
		mustJsonResponse(res, MessageResponse{
			Message: "timer update successful",
		})
		return
	default:
		mustJsonResponse(res, ErrorResponse{
			Message: "timer can not modified",
		})
		return
	}
}

// Get the mode and time info from default route.
func (s *Server) getDefaultRouteHandler(
	res http.ResponseWriter,
	req *http.Request,
) {
	// Write not implemented status code
	res.WriteHeader(http.StatusNotImplemented)
}

// Set the mode to default router. On specific mode, it's possible
// to update settings.
func (s *Server) updateDefaultRouteHandler(
	res http.ResponseWriter,
	req *http.Request,
) {
	// Write not implemented status code
	res.WriteHeader(http.StatusNotImplemented)
}

// Get all registered routes.
func (s *Server) getAllRoutesHandler(
	res http.ResponseWriter,
	_ *http.Request,
) {
	routes := s.routes.All()
	lenRoutes := len(routes)
	// Build response from routing table entries. We know the size
	// of routing entries here. So we can allocate the size.
	response := RoutesResponse{
		Length: lenRoutes,
		Routes: make([]RouteResponse, lenRoutes),
	}
	// Iterate through routing entries and add each entry to response.
	// As subnet, we return the CIDR string representation of the ip net.
	// For timer mode, an extra function is converting the timer to its
	// string representation.
	for idx, entry := range routes {
		response.Routes[idx] = RouteResponse{
			Id:     idx,
			Subnet: entry.IPNet.String(),
			// TODO: We can not get timer id here
			Timer: TimerResponse{
				Id:   -1,
				Type: server.TimerName(entry.Timer),
			},
		}
	}
	// Return as JSON response.
	mustJsonResponse(res, response)
}

type NewRouteRequest struct {
	TimerId int    `json:"timerId"`
	Subnet  string `json:"subnet"`
}

// Create a new route.
func (s *Server) newRouteHandler(
	res http.ResponseWriter,
	req *http.Request,
) {
	// Parse body data.
	var routeRequest NewRouteRequest
	err := json.NewDecoder(req.Body).Decode(&routeRequest)
	if err != nil {
		mustJsonResponse(res, ErrorResponse{
			Message: "can not decode body data",
		})
		return
	}

	// Find timer by id.
	timer := s.timers.Get(routeRequest.TimerId)
	if timer.Timer == nil {
		mustJsonResponse(res, ErrorResponse{
			Message: "can not find timer",
		})
		return
	}

	// Parse subnet to net.IPNet.
	_, ipNet, err := net.ParseCIDR(routeRequest.Subnet)
	if err != nil {
		mustJsonResponse(res, ErrorResponse{
			Message: "can not parse subnet",
		})
		return
	}

	// Add net.IPNet to routing and map to timer instance.
	err = s.routes.Add(*ipNet, timer.Timer)
	if err != nil {
		mustJsonResponse(res, ErrorResponse{
			Message: "route with subnet exist",
		})
		return
	}

	// Build success response.
	mustJsonResponse(res, MessageResponse{
		Message: "create new route success",
	})
}

// Delete an existing route.
func (s *Server) deleteRouteHandler(
	res http.ResponseWriter,
	req *http.Request,
) {
	// Write not implemented status code
	res.WriteHeader(http.StatusNotImplemented)
}

// Get a specific route.
func (s *Server) getRouteHandler(
	res http.ResponseWriter,
	req *http.Request,
) {
	// Write not implemented status code
	res.WriteHeader(http.StatusNotImplemented)
}

// Update settings of specific route.
func (s *Server) updateRouteHandler(
	res http.ResponseWriter,
	req *http.Request,
) {
	// Write not implemented status code
	res.WriteHeader(http.StatusNotImplemented)
}
