package server

import (
	"errors"
	"fmt"
	"net"
	"testing"

	"github.com/donsprallo/gots/internal/ntp"
)

// Just a dummy to mock response builder.
type DummyResponseBuilder struct {
	Message string
}

// Implement ntp.ResponseBuilder interface.
func (rb DummyResponseBuilder) BuildResponse(
	pkg *ntp.NtpPackage) (*ntp.NtpPackage, error) {
	// Just return an error
	return nil, errors.New(
		"not implemented")
}

// Implement stringer interface.
func (rb DummyResponseBuilder) String() string {
	return fmt.Sprintf(rb.Message)
}

func TestFindResponseBuilder(t *testing.T) {
	// Create test table; The message is an identifier, to check which
	// response builder is returned from routing strategy.
	tables := []struct {
		Message string
		IP      net.IP
	}{
		{"default", net.ParseIP("0.0.0.0")},
		{"default", net.ParseIP("127.0.0.1")},
		{"default", net.ParseIP("::1")},
		{"net1", net.ParseIP("192.168.1.10")},
		{"net1", net.ParseIP("192.168.1.11")},
		{"net2", net.ParseIP("192.168.2.11")},
		{"default", net.ParseIP("192.168.2.10")},
	}

	// Create test routing strategy; we are using three different
	// response builders here. One default and one for each network.
	defaultBuilder := DummyResponseBuilder{Message: "default"}
	net1Builder := DummyResponseBuilder{Message: "net1"}
	net2Builder := DummyResponseBuilder{Message: "net2"}
	routing := NewStaticRouting(defaultBuilder)
	// Add builder that matches 192.168.1.0 network
	routing.AddResponseBuilder(net.IPNet{
		Mask: net.CIDRMask(24, 32),
		IP:   net.ParseIP("192.168.1.0"),
	}, net1Builder)
	// Add builder that matches 192.168.2.11 host but
	// not the 192.168.2.0 network.
	routing.AddResponseBuilder(net.IPNet{
		Mask: net.CIDRMask(32, 32),
		IP:   net.ParseIP("192.168.2.11"),
	}, net2Builder)

	// Test all values
	for _, table := range tables {
		// Try to find response builder; this should always return
		// a builder.
		builder, err := routing.FindResponseBuilder(table.IP)
		if err != nil {
			t.Errorf("ip[%s] err: %s",
				table.IP, err)
		}
		// Check builder; the ip must resolved by a specific builder
		dummy := builder.(DummyResponseBuilder)
		if dummy.Message != table.Message {
			t.Errorf("ip[%s] found incorrect builder: want '%s' get '%s'",
				table.IP, table.Message, dummy.Message)
		}
	}
}
