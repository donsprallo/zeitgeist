package server

import (
	"errors"
	"net"

	log "github.com/sirupsen/logrus"
)

// RoutingTableEntry is an entry in a RoutingTable. Each entry contains
// information for a RoutingStrategy to decide, which Timer instance
// can be found.
type RoutingTableEntry struct {
	Id      int       // The unique identifier of the entry.
	IPNet   net.IPNet // IPNet is the net.IP and net.IPMask to match by RoutingStrategy.
	Timer   Timer     // Timer is a Timer instance returned by RoutingStrategy.
	TimerId int
}

func (e *RoutingTableEntry) SetTimer(timer Timer, timerId int) {
	e.Timer = timer
	e.TimerId = timerId
}

func (e *RoutingTableEntry) SetIPNet(ipNet net.IPNet) {
	e.IPNet = ipNet
}

// RoutingTable is a collection of RoutingTableEntry.
type RoutingTable struct {
	nextId  int
	entries []RoutingTableEntry
}

// NewRoutingTable create a new RoutingTable instance with size.
func NewRoutingTable(size int) *RoutingTable {
	return &RoutingTable{
		nextId:  0,
		entries: make([]RoutingTableEntry, 0, size),
	}
}

// All return all RoutingTableEntry objects from RoutingTable.
func (t *RoutingTable) All() []RoutingTableEntry {
	return t.entries
}

// Add adds a net.IP address and Timer to the Table. This address maps
// a net.IP address to a specific Timer.
func (t *RoutingTable) Add(
	ipNet net.IPNet,
	timer Timer,
	timerId int,
) error {
	// IP address must be unique in routing Table.
	if t.Contains(ipNet) {
		return errors.New(
			"key exist in routing Table")
	}
	// Add entry to routing Table.
	t.entries = append(t.entries, RoutingTableEntry{
		Id:      t.nextId,
		IPNet:   ipNet,
		Timer:   timer,
		TimerId: timerId,
	})
	t.nextId++
	return nil
}

func (t *RoutingTable) Get(id int) *RoutingTableEntry {
	for _, entry := range t.entries {
		if entry.Id == id {
			return &entry
		}
	}
	return nil
}

func (t *RoutingTable) Set(id int, timer Timer, timerId int) error {
	for idx, entry := range t.entries {
		if entry.Id == id {
			t.entries[idx].Timer = timer
			t.entries[idx].TimerId = timerId
			return nil
		}
	}
	return errors.New("no route found by id")
}

// MustAdd works how RoutingTable.Add but on an error a panic is used.
// The method adds a net.IP address and Timer to the Table. This address
// maps a net.IP address to a specific Timer.
func (t *RoutingTable) MustAdd(
	ipNet net.IPNet,
	timer Timer,
	timerId int,
) {
	err := t.Add(ipNet, timer, timerId)
	if err != nil {
		log.Panic(err)
	}
}

// Contains checks if a net.IPNet value exists in the collection. Returns true
// if net.IPNet value exists in RoutingTable, otherwise return false.
func (t *RoutingTable) Contains(value net.IPNet) bool {
	for _, entry := range t.entries {
		if entry.IPNet.IP.Equal(value.IP) {
			return true
		}
	}
	return false
}

// RoutingStrategy is an interface to define a strategy for routing net.IP
// addresses to a Timer instance. Each request can get a specified response,
// depends on the response from RoutingStrategy. A net.IP address is mapped
// to a Timer.
type RoutingStrategy interface {

	// FindTimer find a Timer by a net.IP address.
	FindTimer(ip net.IP) (Timer, error)
}

// StaticRouting is a specific RoutingStrategy for simple static routing. This
// means that each net.IP address is managed in a list. To this list net.IP
// addresses and timers are attached. The list is traversed in reverse order
// and checked for a match. If a match is found, then the corresponding timer
// is returned. When no timer is found, a default timer is returned.
type StaticRouting struct {
	Table *RoutingTable
}

// FindTimer search for a Timer by a net.IP address. When no address matches
// one of the timers network mask, an error is returned. But this should
// never have reached in normal system.
func (r *StaticRouting) FindTimer(
	ip net.IP,
) (Timer, error) {
	// First search for a match by equal; We must reverse the
	// static routing Table entries.
	for i := len(r.Table.entries) - 1; i >= 0; i-- {
		entry := r.Table.entries[i]
		if ip.Mask(entry.IPNet.Mask).Equal(entry.IPNet.IP) {
			log.Debugf("host with ip[%s] equal mask[%s] match",
				ip, entry.IPNet.String())
			return entry.Timer, nil
		}
	}
	// Next search for a match by contain; We must reverse the
	// static routing Table entries.
	for i := len(r.Table.entries) - 1; i >= 0; i-- {
		entry := r.Table.entries[i]
		if entry.IPNet.Contains(ip) {
			log.Debugf("host with ip[%s] contains mask[%s] match",
				ip, entry.IPNet.String())
			return entry.Timer, nil
		}
	}
	// No match found. Should never have reached.
	return nil, errors.New(
		"no handler found in routing Table")
}

var (
	defaultRoute = net.IPNet{
		Mask: net.CIDRMask(0, 32),
		IP:   net.ParseIP("0.0.0.0"),
	}
	ipv4Route = net.IPNet{
		Mask: net.CIDRMask(24, 32),
		IP:   net.ParseIP("127.0.0.0"),
	}
	ipv6Route = net.IPNet{
		Mask: net.CIDRMask(120, 128),
		IP:   net.ParseIP("::"),
	}
)

// NewStaticRouting create a new StaticRouting instance. A default Timer
// must be added to be sure that we have a default ntp timer. The default
// Timer is added to the RoutingTable as default route, that handle all
// net.IP addresses without need to add other routes.
func NewStaticRouting(
	table *RoutingTable,
	defaultTimer Timer,
	timerId int,
) *StaticRouting {
	// Create basic structure
	routing := StaticRouting{
		Table: table,
	}
	// Add the default response timer to router.
	routing.Table.MustAdd(defaultRoute, defaultTimer, timerId)
	// Add IPv4 loop back address.
	routing.Table.MustAdd(ipv4Route, defaultTimer, timerId)
	// Add IPv6 loop back address.
	routing.Table.MustAdd(ipv6Route, defaultTimer, timerId)
	return &routing
}
