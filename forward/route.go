package forward

import (
	"fmt"
	"github.com/ICKelin/gtun/transport"
)

var (
	errNoRoute = fmt.Errorf("no route to host")
)

type RouteEntry struct {
	rtt        int32
	lastActive int64
	conn       transport.Conn
}

type RouteTable struct {
	Table []*RouteEntry
}

func NewRouteTable() *RouteTable {
	return &RouteTable{}
}

func (r *RouteTable) Route() (*RouteEntry, error) {
	if len(r.Table) <= 0 {
		return nil, errNoRoute
	}

	return r.Table[0], nil
}
