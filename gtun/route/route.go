package route

import (
	"github.com/ICKelin/gtun/internal/logs"
	"strings"
	"sync"

	"github.com/ICKelin/optw/transport"
)

var routeManager = &Manager{
	regionHops:  make(map[string][]*HopInfo),
	raceManager: GetRaceManager(),
}

type Manager struct {
	raceManager  *RaceManager
	regionHopsMu sync.RWMutex
	regionHops   map[string][]*HopInfo
}

type HopInfo struct {
	transport.Conn
}

func GetRouteManager() *Manager {
	return routeManager
}

func (routeManager *Manager) AddRoute(region string, hop *HopInfo) {
	routeManager.regionHopsMu.Lock()
	defer routeManager.regionHopsMu.Unlock()
	regionHops := routeManager.regionHops[region]
	if regionHops == nil {
		regionHops = make([]*HopInfo, 0)
	}
	regionHops = append(regionHops, hop)
	routeManager.regionHops[region] = regionHops

}

func (routeManager *Manager) Route(region, dip string) *HopInfo {
	routeManager.regionHopsMu.RLock()
	defer routeManager.regionHopsMu.RUnlock()

	regionHops, ok := routeManager.regionHops[region]
	if !ok {
		return nil
	}

	if len(regionHops) <= 0 {
		return nil
	}

	bestNode := routeManager.raceManager.GetBestNode(region)
	bestIP := strings.Split(bestNode, ":")[0]
	for i := 0; i < len(regionHops); i++ {
		hop := regionHops[i]
		if hop.IsClosed() {
			logs.Warn("%s %s is closed", region, hop.RemoteAddr())
			continue
		}

		if len(bestIP) != 0 {
			hopIP := strings.Split(hop.RemoteAddr().String(), ":")[0]
			if bestIP == hopIP {
				logs.Debug("best ip match %s", bestIP)
				return hop
			}
		}
	}

	logs.Warn("use random hop")
	hash := 0
	for _, c := range dip {
		hash += int(c)
	}
	return regionHops[hash%len(regionHops)]
}

func (routeManager *Manager) DeleteRoute(region string, hop *HopInfo) {
	routeManager.regionHopsMu.Lock()
	defer routeManager.regionHopsMu.Unlock()
	regionHops := routeManager.regionHops[region]
	if regionHops == nil {
		return
	}

	hops := make([]*HopInfo, 0, len(regionHops))
	for _, s := range regionHops {
		if s.RemoteAddr().String() == hop.RemoteAddr().String() {
			continue
		}

		hops = append(hops, s)
	}

	routeManager.regionHops[region] = hops
}
