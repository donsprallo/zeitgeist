package main

import (
	"context"
	"flag"
	"github.com/donsprallo/gots/internal/web/api/routes"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/donsprallo/gots/internal/ntp"
	"github.com/donsprallo/gots/internal/server"
	"github.com/donsprallo/gots/internal/web"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

// Variables for command line arguments.
var (
	ntpHost *string
	ntpPort *int
	webHost *string
	webPort *int
)

// Load a string value from environment key. If environment key
// does not exist, a fallback value is returned.
func getEnvStr(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		log.Debugf("get parsed env[%s]: %s", key, value)
		return value
	}
	log.Debugf("get fallback env[%s]: %s", key, fallback)
	return fallback
}

// Load a integer value from environment key. Of environment key
// does not exist, a fallback value is returned.
func getEnvInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		if parsed, err := strconv.Atoi(value); err == nil {
			log.Debugf("get parsed env[%s]: %d", key, parsed)
			return parsed
		}
	}
	log.Debugf("get fallback env[%s]: %d", key, fallback)
	return fallback
}

func init() {
	// Load dotenv when .env file available. When this file
	// does not exist, this is not an error.
	err := godotenv.Load()
	if err != nil {
		log.Warn("no .env file to load")
	}
}

func init() {
	// Setup application logger.
	log.SetLevel(log.DebugLevel)
}

func init() {
	// Setup command line arguments.
	ntpHost = flag.String(
		"host", getEnvStr("NTP_HOST", "localhost"),
		"ntp daemon hostname")
	ntpPort = flag.Int("port", getEnvInt("NTP_PORT", 123),
		"ntp daemon port")
	webHost = flag.String(
		"web-host", getEnvStr("WEB_HOST", "localhost"),
		"web hostname")
	webPort = flag.Int(
		"web-port", getEnvInt("WEB_PORT", 80),
		"web port")
	// Parse command line arguments.
	flag.Parse()
}

func main() {
	// First we create a default ntp package. This is used for set up
	// the default timers in next step. The settings here means, that
	// the ntp server response override incoming requests with this data.
	defaultTimerPackage := ntp.Package{}
	defaultTimerPackage.SetVersion(ntp.VersionV3)
	defaultTimerPackage.SetMode(ntp.ModeServer)
	defaultTimerPackage.SetStratum(1)
	defaultTimerPackage.SetReferenceClockId([]byte("ABCD"))

	// Next we create the default timers. These timers are used for the
	// default route we build in next step. This means that this timer
	// is used for all requests, where no other route match ip address
	// from requested client.
	defaultTimer := &server.SystemTimer{
		NTPPackage: defaultTimerPackage,
	}

	// Create routing protocol for handle requests. For this, we need to create
	// a routing table. The table contains all ip address's and the
	// corresponding timer instances.
	routingTable := server.NewRoutingTable(10)

	// The RoutingStrategy is used to specify, how a request and its ip
	// address is matching a timer. The default timer is used to handle all
	// requests matching the default route.
	routingStrategy := server.NewStaticRouting(
		routingTable, defaultTimer)

	// Create timer collection to collect timers. We need to manage all timers
	// and do this with this collection.
	timers := server.NewTimerCollection(10)
	timers.Add(defaultTimer)

	// Create ntp server and start application. The ntp server handle all
	// ntp requests with a RoutingStrategy.
	ntpServer := server.NewServer(
		*ntpHost, *ntpPort, routingStrategy)
	go ntpServer.Serve()

	// Now we create a web server. First we need a router that handle http
	// requests. The strict slash option is needed here. This means, that
	// a trailing slash in "/route/" is automatically redirect to "/route".
	// This is useful for path naming convention on endpoint registration.
	router := mux.NewRouter()
	router.StrictSlash(true)

	// For the web api we need to create endpoints. An endpoint is a collection
	// of logically related functions for a web API.
	apiHealth := routes.NewHealthcheckEndpoint()
	apiTimer := routes.NewTimerEndpoint(timers)
	apiRoute := routes.NewRouteEndpoint(timers, routingTable)

	// We still need a web server so that we can deliver our routes.
	webServer := web.NewServer(
		*webHost, *webPort, router)

	// The API endpoints must be registered with the web server. Here we define
	// a prefix under which address the endpoint can be reached.
	webServer.RegisterEndpoint("/api/v1/healthcheck", apiHealth)
	webServer.RegisterEndpoint("/api/v1/timer", apiTimer)
	webServer.RegisterEndpoint("/api/v1/route", apiRoute)

	// Now we can start our webserver in background.
	go webServer.Serve()

	// Create ticker to update all timers every second.
	timerTicker := time.NewTicker(1 * time.Second)

	// Gracefully shutdown.
	idleConnectionsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)

		// Block until SIGINT received.
		<-sigint

		// Create a deadline to wait for shutdown.
		wait := 10 * time.Second
		ctx, cancel := context.WithTimeout(
			context.Background(), wait)
		defer cancel()

		// Does not block if no connections, but will otherwise wait
		// until the timeout deadline.
		err := webServer.Shutdown(ctx)
		if err != nil {
			log.Error(err)
		}

		close(idleConnectionsClosed)
	}()

	// Loop infinity until gracefully shutdown.
	for {
		select {
		// On ticker ticks, update all timers.
		case <-timerTicker.C:
			timers.AllUpdate()
		// On gracefully shutdown.
		case <-idleConnectionsClosed:
			break
		}
	}
}
