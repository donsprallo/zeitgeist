package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/donsprallo/gots/internal/api"
	"github.com/donsprallo/gots/internal/ntp"
	"github.com/donsprallo/gots/internal/server"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

// Variables for command line arguments.
var (
	ntpHost *string
	ntpPort *int
	apiHost *string
	apiPort *int
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
	apiHost = flag.String(
		"api-host", getEnvStr("API_HOST", "localhost"),
		"api hostname")
	apiPort = flag.Int(
		"api-port", getEnvInt("API_PORT", 80),
		"api port")
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

	// Now we create an api server. Here we can edit ntp server settings with
	// a universal rest client.
	router := mux.NewRouter()
	apiServer := api.NewApiServer(
		*apiHost, *apiPort, router, routingTable)
	apiServer.RegisterRoutes("/api/v1")
	go apiServer.Serve()

	// Create ticker to update all timers every second.
	timerTicker := time.NewTicker(1 * time.Second)

	// Gracefully shutdown.
	idleConnectionsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)

		// Block until SIGINT received
		<-sigint

		// Create a deadline to wait for shutdown.
		wait := 10 * time.Second
		ctx, cancel := context.WithTimeout(
			context.Background(), wait)
		defer cancel()

		// Does not block if no connections, but will otherwise wait
		// until the timeout deadline.
		err := apiServer.Shutdown(ctx)
		if err != nil {
			log.Error(err)
		}

		close(idleConnectionsClosed)
	}()

	// Loop infinity.
	for {
		select {
		// On ticker ticks, we update all timers.
		case <-timerTicker.C:
			timers.AllUpdate()
		// On gracefully shutdown.
		case <-idleConnectionsClosed:
			break
		}
	}
}
