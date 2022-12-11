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
	ntp_host *string
	ntp_port *int
	api_host *string
	api_port *int
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
	ntp_host = flag.String(
		"host", getEnvStr("NTP_HOST", "localhost"),
		"ntp daemon hostname")
	ntp_port = flag.Int("port", getEnvInt("NTP_PORT", 123),
		"ntp daemon port")
	api_host = flag.String(
		"api-host", getEnvStr("API_HOST", "localhost"),
		"api hostname")
	api_port = flag.Int(
		"api-port", getEnvInt("API_PORT", 80),
		"api port")
	// Parse command line arguments
	flag.Parse()
}

func main() {
	// Create routing protocol for handle requests
	defaultTimer := &server.SystemNtpTimer{
		Version: ntp.NTP_VN_V3,
		Mode:    ntp.NTP_MODE_SERVER,
		Stratum: 1,
		Id:      []byte("ABCD"),
	}
	routing := server.NewStaticRouting(defaultTimer)
	// Create ntp server and start application
	server := server.NewNtpServer(
		*ntp_host, *ntp_port, routing)
	go server.Serve()

	// Create REST api server
	router := mux.NewRouter()
	apiServer := api.NewApiServer(
		*api_host, *api_port, router)
	apiServer.RegisterRoutes("/api/v1")
	go apiServer.Serve()

	// Gracefully shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	// Block until SIGINT received
	<-c

	// Create a deadline to wait for
	wait := 10 * time.Second
	ctx, cancel := context.WithTimeout(
		context.Background(), wait)
	defer cancel()

	// Does not block if no connections, but will otherwise wait
	// unitl the timeout deadline.
	apiServer.Shutdown(ctx)
	log.Info("shutting down")
	os.Exit(0)
}
