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
	// Setup application logger
	log.SetLevel(log.DebugLevel)
}

func init() {
	// Setup command line arguments
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
	// Parse command line arguments
	flag.Parse()
}

func main() {
	// Create routing protocol for handle requests
	systemTimerPackage := ntp.Package{}
	systemTimerPackage.SetVersion(ntp.VersionV3)
	systemTimerPackage.SetMode(ntp.ModeServer)
	systemTimerPackage.SetStratum(1)
	systemTimerPackage.SetReferenceClockId([]byte("ABCD"))

	defaultTimer := &server.SystemTimer{
		NTPPackage: systemTimerPackage,
	}

	routing := server.NewStaticRouting(defaultTimer)
	// Create ntp server and start application
	ntpServer := server.NewNtpServer(
		*ntpHost, *ntpPort, routing)
	go ntpServer.Serve()

	// Create REST api server
	router := mux.NewRouter()
	apiServer := api.NewApiServer(
		*apiHost, *apiPort, router)
	apiServer.RegisterRoutes("/api/v1")
	go apiServer.Serve()

	// Gracefully shutdown
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

	<-idleConnectionsClosed
	log.Info("shutting down")
}
