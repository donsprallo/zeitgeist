package main

import (
	"flag"
	"fmt"
	"github.com/donsprallo/zeitgeist/internal/ntp"
)

// Variables for command line arguments.
var (
	ntpHost *string
	ntpPort *int
)

// Setup command line arguments.
func init() {
	ntpHost = flag.String(
		"host", "localhost", "request host address")
	ntpPort = flag.Int(
		"port", 123, "request port")
	// Parse command line arguments.
	flag.Parse()
}

func main() {
	// Request a ntp package from remote server.
	pkg, err := ntp.Request(*ntpHost, *ntpPort)
	if err != nil {
		fmt.Printf(err.Error())
		return
	}

	// Print request result to user.
	fmt.Println("header:")
	fmt.Printf("leap: %d\n", pkg.GetLeap())
	fmt.Printf("version: %d\n", pkg.GetVersion())
	fmt.Printf("mode: %d\n", pkg.GetMode())
	fmt.Printf("stratum: %d\n", pkg.GetStratum())
	fmt.Printf("poll: %d\n", pkg.GetPoll())
	fmt.Printf("precision: %d\n", pkg.GetPrecision())

	fmt.Println("\npackage:")
	fmt.Printf("root delay: %d\n", pkg.GetRootDelay())
	fmt.Printf("root dispersion: %d\n", pkg.GetRootDispersion())
	fmt.Printf("ref clock id: 0x%X\n", pkg.GetReferenceClockId())
	fmt.Printf("ref timestamp: %v\n", pkg.GetReferenceTimestamp())
	fmt.Printf("originate timestamp: %v\n", pkg.GetOriginateTimestamp())
	fmt.Printf("recv timestamp: %v\n", pkg.GetReceiveTimestamp())
	fmt.Printf("transmit timestamp: %v\n", pkg.GetTransmitTimestamp())
}
