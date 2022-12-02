package main

import (
	"flag"
	"os"
	"strconv"

	"github.com/donsprallo/gots/internal/ntp"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

// Variables for command line arguments.
var (
	host *string
	port *int
)

// Load a string value from environment key. If environment key
// does not exist, a fallback value is returned.
func getEnvStr(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// Load a integer value from environment key. Of environment key
// does not exist, a fallback value is returned.
func getEnvInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		if parsed, err := strconv.Atoi(value); err != nil {
			return parsed
		}
	}
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
	// Setup command line arguments
	host = flag.String(
		"host", getEnvStr("NTP_HOST", "localhost"),
		"ntp daemon hostname")
	port = flag.Int("port", getEnvInt("NTP_PORT", 123),
		"ntp daemon port")
	// Parse command line arguments
	flag.Parse()
}

func main() {
	// Create ntp server and start application
	server := ntp.NewNtpServer(*host, *port)
	server.Serve()
}
