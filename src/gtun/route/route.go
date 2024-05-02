package route

import (
	"fmt"
	"github.com/ICKelin/gtun/src/gtun/config"
	"github.com/ICKelin/gtun/src/internal/logs"
	transport "github.com/ICKelin/optw"
	"sync"
)

var routeManager = &Manager{
	tm:         tm,
	cm:         cm,
	routeTable: make(map[string][]*routeItem),
}

type routeItem struct {
	region     string
	scheme     string
	serverAddr string
	transport.Conn
}

type Manager struct {
	tm *traceManager
	cm *connManager

	routeTableMu sync.Mutex
	routeTable   map[string][]*routeItem
}

func GetRouteManager() *Manager {
	return routeManager
}

func (routeManager *Manager) Route(region, dip string) transport.Conn {
	regionRoutes, ok := routeManager.routeTable[region]
	if !ok {
		return nil
	}

	if len(regionRoutes) <= 0 {
		return nil
	}

	bestNode, ok := routeManager.tm.getRegionBestTarget(region)
	if ok {
		bestAddr := bestNode.serverAddr
		for i := 0; i < len(regionRoutes); i++ {
			it := regionRoutes[i]
			if it.IsClosed() {
				logs.Warn("%s %s is closed", region, it.RemoteAddr())
				continue
			}

			if len(bestAddr) != 0 {
				// scheme://ip:port match
				if bestAddr == it.serverAddr &&
					bestNode.scheme == it.scheme {
					logs.Debug("region[%s] best node match %s://%s",
						region, it.scheme, bestAddr)
					return it
				}
			}
		}
	} else {
		logs.Warn("no best node for region[%s]", region)
	}

	logs.Warn("region[%s] use random hop", region)
	hash := 0
	for _, c := range dip {
		hash += int(c)
	}

	hop := regionRoutes[hash%len(regionRoutes)]
	if hop == nil || hop.IsClosed() {
		return nil
	}
	return hop
}

func (routeManager *Manager) addRoute(region string, item *routeItem) {
	routeManager.routeTableMu.Lock()
	defer routeManager.routeTableMu.Unlock()

	regionItems := routeManager.routeTable[region]
	if regionItems == nil {
		regionItems = make([]*routeItem, 0)
	}

	regionItems = append(regionItems, item)
	routeManager.routeTable[region] = regionItems
}

func (routeManager *Manager) deleteRoute(region string, item *routeItem) {
	routeManager.routeTableMu.Lock()
	defer routeManager.routeTableMu.Unlock()
	regionItems := routeManager.routeTable[region]
	if regionItems == nil {
		return
	}

	conns := make([]*routeItem, 0, len(regionItems))
	for _, it := range regionItems {
		if it == item {
			continue
		}

		conns = append(conns, it)
	}

	routeManager.routeTable[region] = conns
}

func Setup(region string, routes []*config.RouteConfig) error {
	for _, cfg := range routes {
		conn, err := newConn(region, cfg.Scheme, cfg.Server, cfg.AuthKey)
		if err != nil {
			fmt.Printf("region[%s] connect to %s://%s fail: %v\n",
				region, cfg.Scheme, cfg.Server, cfg.AuthKey)
			return err
		}

		cm.regionConn[region] = append(cm.regionConn[region], conn)

		t, ok := tm.regionTrace[region]
		if !ok {
			logs.Debug("add region[%s] trace", region)
			t = newTrace(region)
			tm.regionTrace[region] = t
		} else {
			logs.Debug("region[%s] trace exist", region)
		}

		t.addTarget(traceTarget{
			traceAddr:  cfg.Trace,
			serverAddr: cfg.Server,
			scheme:     cfg.Scheme,
		})
	}

	return nil
}

func Run() {
	go tm.startTrace()
	go cm.startConn()
}
