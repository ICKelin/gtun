package forward

import (
	"fmt"
	"github.com/ICKelin/gtun/transport"
	"github.com/ICKelin/gtun/transport/transport_api"
	"math"
	"sync"
)

var (
	errNoRoute = fmt.Errorf("no route to host")
	maxRtt     = math.MinInt32
)

type RouteEntry struct {
	scheme, addr string
	rtt          int32
	conn         transport.Conn
}

type RouteTable struct {
	// key: scheme://addr
	tableMu   sync.RWMutex
	table     map[string]*RouteEntry
	minRttKey string
}

func NewRouteTable() *RouteTable {
	return &RouteTable{
		table: make(map[string]*RouteEntry),
	}
}

func (r *RouteTable) Add(scheme, addr, cfg string) error {
	dialer, err := transport_api.NewDialer(scheme, addr, cfg)
	if err != nil {
		return err
	}

	conn, err := dialer.Dial()
	if err != nil {
		return err
	}

	entry := &RouteEntry{
		scheme: scheme,
		addr:   addr,
		conn:   conn,
	}

	entryKey := fmt.Sprintf("%s://%s", scheme, addr)
	r.tableMu.Lock()
	defer r.tableMu.Unlock()
	r.table[entryKey] = entry
	return nil
}

func (r *RouteTable) Del(scheme, addr string) {
	r.tableMu.Lock()
	defer r.tableMu.Unlock()
	for key, entry := range r.table {
		if entry.scheme == scheme &&
			entry.addr == addr {
			delete(r.table, key)
			break
		}
	}
}

func (r *RouteTable) Route() (*RouteEntry, error) {
	r.tableMu.RLock()
	defer r.tableMu.RUnlock()
	if len(r.table) <= 0 {
		return nil, errNoRoute
	}

	entry, ok := r.table[r.minRttKey]
	if !ok {
		return nil, errNoRoute
	}

	return entry, nil
}
