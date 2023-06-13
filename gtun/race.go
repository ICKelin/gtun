package gtun

import (
	"encoding/binary"
	"math"
	"net"
	"sync"
	"time"

	"github.com/ICKelin/gtun/internal/logs"
)

var gRaceManager *RaceManager
var singleton sync.Once

// RaceManager manage region race
type RaceManager struct {
	regionRaceMu sync.Mutex
	regionRace   map[string]*Race
}

// NewRaceManager is a singleton of create a race manager
func NewRaceManager() *RaceManager {
	singleton.Do(func() {
		gRaceManager = &RaceManager{
			regionRace: make(map[string]*Race),
		}
	})
	return gRaceManager
}

// AddRegionRace adds a race instance for region
func (m *RaceManager) AddRegionRace(region string, race *Race) {
	m.regionRaceMu.Lock()
	defer m.regionRaceMu.Unlock()
	m.regionRace[region] = race
}

//
//func (m *RaceManager) DeleteRegionNode(region, node string) {
//	m.regionRaceMu.Lock()
//	defer m.regionRaceMu.Unlock()
//	race := m.RegionRace[region]
//	if race == nil {
//		return
//	}
//}

// GetBestNode returns the hightest score of region target region
func (m *RaceManager) GetBestNode(region string) string {
	regionRace := m.regionRace[region]
	if regionRace == nil {
		return ""
	}

	return regionRace.GetBestNode()
}

// Race is a region race instance
type Race struct {
	targets       []string
	targetScoreMu sync.Mutex
	targetScore   map[string]float64
	totalRtt      int32
}

// NewRace return race instance
func NewRace(targets []string) *Race {
	return &Race{
		targets:       targets,
		targetScoreMu: sync.Mutex{},
		targetScore:   make(map[string]float64),
	}
}

// Run race job
func (r *Race) Run() {
	r.race()
	tick := time.NewTicker(time.Second * 120)
	for range tick.C {
		r.race()
	}
}

func (r *Race) race() {
	for _, target := range r.targets {
		raddr, err := net.ResolveUDPAddr("udp", target)
		if err != nil {
			logs.Error("resolve udp addr: %v", err)
			continue
		}

		rconn, err := net.DialUDP("udp", nil, raddr)
		if err != nil {
			logs.Error("dial udp fail: %v", err)
			continue
		}

		rtt := -1
		loss := 60
		seq := uint64(0)
		buf := make([]byte, 8)
		for i := 0; i < 60; i++ {
			seq++
			binary.BigEndian.PutUint64(buf, seq)

			beg := time.Now()
			rconn.SetWriteDeadline(time.Now().Add(time.Second * 2))
			_, err := rconn.Write(buf)
			rconn.SetWriteDeadline(time.Time{})
			if err != nil {
				logs.Error("write to udp fail: %v", err)
				continue
			}

			rconn.SetReadDeadline(time.Now().Add(time.Second * 2))
			_, _, err = rconn.ReadFromUDP(buf)
			rconn.SetReadDeadline(time.Time{})
			if err != nil {
				logs.Error("read from udp fail: %v", err)
				continue
			}
			loss--
			diff := time.Now().Sub(beg).Milliseconds()
			rtt += int(diff)
		}
		remoteAddr := rconn.RemoteAddr().String()
		rconn.Close()

		if rtt < 0 {
			rtt = math.MaxInt
		}

		lossRank := r.calcLossScore(loss)
		delayRank := r.calcRttScore(rtt)
		score := lossRank + delayRank
		logs.Debug("%s loss %d rtt %d lossRank %.4f delayRank %.4f score %.4f", target, loss, rtt, lossRank, delayRank, score)
		r.targetScoreMu.Lock()
		r.targetScore[remoteAddr] = score
		r.targetScoreMu.Unlock()
	}
}

// f(p)  = 50                              p = 0,
// f(p) = 40+(0.75-p)x13                   0%  < p <= 0.75%,
// f(p) = 35+(1.25-p)x10                   0.75% < p <= 1.25%,
// f(p) = 30+(2.25-p)x5                    1.25% < p <= 2.25%,
// f(p) = 30+(p-2.25)x5x-1                 p > 2.25%
func (r *Race) calcLossScore(loss int) float64 {
	lossRate := float64(loss) / 60
	if 0 < lossRate && lossRate <= 0.75 {
		return 40 + (0.75-lossRate)*13
	} else if 0.75 < lossRate && lossRate <= 1.25 {
		return 35 + (1.25-lossRate)*10
	} else if 1.25 < lossRate && lossRate <= 2.25 {
		return 30 + (2.25-lossRate)*5
	} else if lossRate > 2.25 {
		return 30 + (lossRate-2.25)*5*(-1)
	}
	return 50
}

func (r *Race) calcRttScore(rtt int) float64 {
	avgRtt := float64(rtt) / 60
	if 0 < avgRtt && avgRtt < 45.0 {
		return 50
	} else if 45.0 < avgRtt && avgRtt <= 90.0 {
		return 40 + (90-avgRtt)*0.2
	} else if 90.0 < avgRtt && avgRtt <= 120.0 {
		return 35 + (120-avgRtt)*0.17
	} else if 120.0 < avgRtt && avgRtt <= 180.0 {
		return 30 + (180-avgRtt)*0.08
	} else if avgRtt > 180 {
		return 30 + (avgRtt-180)*0.08*(-1)
	}

	return 0
}

// GetBestNode of all the targets of race
func (r *Race) GetBestNode() string {
	r.targetScoreMu.Lock()
	defer r.targetScoreMu.Unlock()
	bestScore := float64(-1)
	node := ""
	for target, score := range r.targetScore {
		if bestScore < score {
			bestScore = score
			node = target
		}
	}
	return node
}