package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/donsprallo/zeitgeist/internal/web/api/routes"
	"github.com/donsprallo/zeitgeist/pkg/config"
	"os"
	"os/signal"
	"time"

	"github.com/donsprallo/zeitgeist/internal/ntp"
	"github.com/donsprallo/zeitgeist/internal/server"
	"github.com/donsprallo/zeitgeist/internal/web"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

// Variables add by linker flags.
var (
	version string // Application version
)

// Variables for command line arguments.
var (
	ntpHost     *string
	ntpPort     *int
	webHost     *string
	webPort     *int
	showVersion *bool
	logLevel    *string
)

// Default command line argument values.
var (
	defaultNtpHost  string
	defaultNtpPort  int
	defaultWebHost  string
	defaultWebPort  int
	defaultLogLevel string
)

// Load dotenv when .env file available. When this file
// does not exist, this is not an error.
func init() {
	err := godotenv.Load()
	if err != nil {
		log.Debug("no .env file to load")
	}
}

// Initialize default command line argument values. This values can be
// overwritten from environment variables. When no environment variable
// is set, a fallback value is used.
func init() {
	defaultNtpHost = config.GetEnvStr("NTP_HOST", "localhost")
	defaultNtpPort = config.GetEnvInt("NTP_PORT", 123)
	defaultWebHost = config.GetEnvStr("WEB_HOST", "localhost")
	defaultWebPort = config.GetEnvInt("WEB_PORT", 80)
	defaultLogLevel = config.GetEnvStr("LOGLEVEL", "debug")
}

// Setup command line arguments.
func init() {
	// Ntp server arguments.
	ntpHost = flag.String(
		"host", defaultNtpHost,
		"ntp daemon host interface name")
	ntpPort = flag.Int("port", defaultNtpPort,
		"ntp daemon host interface port")
	// Web server arguments.
	webHost = flag.String(
		"web-host", defaultWebHost,
		"web host interface name")
	webPort = flag.Int(
		"web-port", defaultWebPort,
		"web host interface port")
	showVersion = flag.Bool(
		"version", false,
		"show version information and exit")
	logLevel = flag.String(
		"loglevel", defaultLogLevel,
		"set application logger level")
	// Parse command line arguments.
	flag.Parse()
}

// Setup application logger.
func init() {
	level := log.DebugLevel
	switch *logLevel {
	case "debug":
		level = log.DebugLevel
	case "info":
		level = log.InfoLevel
	case "warn":
		level = log.WarnLevel
	case "error":
		level = log.ErrorLevel
	default:
		log.Warn("no valid log level set")
	}
	log.SetLevel(level)
}

func main() {
	// When version flag is set, just display version information and exit.
	if *showVersion == true {
		fmt.Printf("time server version %s\n", version)
		os.Exit(0)
	}

	// First we create a default ntp package. This is used for set up
	// the default timers in next step. The settings here means, that
	// the ntp server response override incoming requests with this data.
	defaultTimerPackage := ntp.Package{}
	defaultTimerPackage.SetVersion(ntp.VersionV3)
	defaultTimerPackage.SetMode(ntp.ModeServer)
	defaultTimerPackage.SetStratum(1)
	defaultTimerPackage.SetReferenceClockId([]byte("NICO"))

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

	// Create timer collection to collect timers. We need to manage all timers
	// and do this with this collection. The timer id is a unique identifier
	// for the timer.
	timers := server.NewTimerCollection(10)
	timerId := timers.Add(defaultTimer)

	// The RoutingStrategy is used to specify, how a request and its ip
	// address is matching a timer. The default timer is used to handle all
	// requests matching the default route.
	routingStrategy := server.NewStaticRouting(
		routingTable, defaultTimer, timerId)

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
	apiHealth := routes.NewHealthEndpoint()
	apiTimer := routes.NewTimerEndpoint(timers)
	apiRoute := routes.NewRouteEndpoint(timers, routingTable)

	// We still need a web server so that we can deliver our routes.
	webServer := web.NewServer(
		*webHost, *webPort, router)

	// The API endpoints must be registered with the web server. Here we define
	// a prefix under which address the endpoint can be reached.
	webServer.RegisterEndpoint("/api/v1/health", apiHealth)
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
