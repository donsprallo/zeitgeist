package server

import (
	"net"
	"testing"
)

func TestFindTimer(t *testing.T) {
	// Create test Table; The message is an identifier, to check which
	// response timer is returned from routing strategy.
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
	// response timers here. One default and one for each network.
	defaultTimer := DummyTimer{Message: "default"}
	net1Timer := DummyTimer{Message: "net1"}
	net2Timer := DummyTimer{Message: "net2"}
	routing := NewStaticRouting(defaultTimer)
	// Add timer that matches 192.168.1.0 network
	routing.Table.MustAdd(net.IPNet{
		Mask: net.CIDRMask(24, 32),
		IP:   net.ParseIP("192.168.1.0"),
	}, net1Timer)
	// Add timer that matches 192.168.2.11 host but
	// not the 192.168.2.0 network.
	routing.Table.MustAdd(net.IPNet{
		Mask: net.CIDRMask(32, 32),
		IP:   net.ParseIP("192.168.2.11"),
	}, net2Timer)

	// Test all values
	for _, table := range tables {
		// Try to find response timer; this should always return
		// a timer.
		timer, err := routing.FindTimer(table.IP)
		if err != nil {
			t.Errorf("ip[%s] err: %s",
				table.IP, err)
		}
		// Check timer; the ip must resolve by a specific timer
		dummy := timer.(DummyTimer)
		if dummy.Message != table.Message {
			t.Errorf("ip[%s] found incorrect timer: want '%s' get '%s'",
				table.IP, table.Message, dummy.Message)
		}
	}
}
