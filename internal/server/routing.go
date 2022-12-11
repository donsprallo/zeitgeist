package server

import (
	"errors"
	"net"

	log "github.com/sirupsen/logrus"
)

type RoutingTableEntry struct {
	IPNet net.IPNet
	Timer Timer
}

type RoutingTable []RoutingTableEntry

// Check that routing table contains value. Return true if
// value is in table, otherwise return false.
func (t *RoutingTable) Contains(value net.IPNet) bool {
	for _, entry := range *t {
		if entry.IPNet.IP.Equal(value.IP) {
			return true
		}
	}
	return false
}

// Each client can get a generic or special case ntp packageresponse. To
// identify the client, his network address is used. The ntp.Routing maps
// the clients address to a corresponding ntp.ResponseBuilder. This is called
// routing in our ntp context.
type RoutingStrategy interface {

	// Find a ntp.ResponseBuilder by a network address. The ntp.ResponseBuilder
	// is used to build a ntp package response.
	FindTimer(ip net.IP) (Timer, error)
}

// The ntp.StaticRouting is using a simple routing algorithm. Each client
// can map his network address to a single ntp.NtpTimer. When no
// timer is found, an default timer is returned.
type StaticRouting struct {
	table RoutingTable
}

// Add a network address to the router. This address is mapping a clients
// network address to a specific response timer. The first addedd must be
// the last one, that is checked by a find.
func (r *StaticRouting) AddTimer(
	ipnet net.IPNet,
	timer Timer,
) error {
	// IPNet must be unique in routing table
	if r.table.Contains(ipnet) {
		return errors.New(
			"key exist in routing table")
	}
	// Add entry to routing table
	r.table = append(r.table, RoutingTableEntry{
		IPNet: ipnet,
		Timer: timer,
	})
	return nil
}

// Search for a response timer by a net.Addr. When no address matches
// one of the timers network mask, an error is returned. But this should
// never be the case.
func (r *StaticRouting) FindTimer(
	ip net.IP,
) (Timer, error) {
	// First search for a match by equal; We must reverse the
	// static routing table entries.
	for i := len(r.table) - 1; i >= 0; i-- {
		entry := r.table[i]
		if ip.Mask(entry.IPNet.Mask).Equal(entry.IPNet.IP) {
			log.Debugf("host[%s] euqals mask[%s] ip[%s]",
				ip, entry.IPNet.Mask, entry.IPNet.IP)
			return entry.Timer, nil
		}
	}
	// Next search for a match by contain; We must reverse the
	// static routing table entries.
	for i := len(r.table) - 1; i >= 0; i-- {
		entry := r.table[i]
		if entry.IPNet.Contains(ip) {
			log.Debugf("host[%s] contains mask[%s] ip[%s]",
				ip, entry.IPNet.Mask, entry.IPNet.IP)
			return entry.Timer, nil
		}
	}
	// No match found
	return nil, errors.New(
		"no handler found in routing table")
}

var (
	defaultRoute = net.IPNet{
		Mask: net.CIDRMask(0, 32),
		IP:   net.ParseIP("0.0.0.0"),
	}
	ipv4Route = net.IPNet{
		Mask: net.CIDRMask(24, 32),
		IP:   net.ParseIP("127.0.0.1"),
	}
	ipv6Route = net.IPNet{
		Mask: net.CIDRMask(120, 128),
		IP:   net.ParseIP("::1"),
	}
)

// Create a new ntp.StaticRouting instance. A default ntp.NtpTimer
// must be added to be sure that we have a default ntp timer.
func NewStaticRouting(defaultTimer Timer) *StaticRouting {
	// Create basic structure
	routing := StaticRouting{
		table: make(RoutingTable, 0, 10),
	}
	// Add the default response timer to router
	routing.AddTimer(
		defaultRoute,
		defaultTimer,
	)
	// Add IPv4 loopback address
	routing.AddTimer(
		ipv4Route,
		defaultTimer,
	)
	// Add IPv6 loopback address
	routing.AddTimer(
		ipv6Route,
		defaultTimer,
	)
	return &routing
}
