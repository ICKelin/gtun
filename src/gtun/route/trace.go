package route

import (
	"encoding/binary"
	"github.com/ICKelin/gtun/src/internal/logs"
	"math"
	"net"
	"sync"
	"time"
)

var tm = &traceManager{
	regionTrace: make(map[string]*trace),
}

type traceManager struct {
	regionTrace map[string]*trace
}

func (m *traceManager) addTraces(traces map[string]*trace) {
	for k, _ := range traces {
		m.regionTrace[k] = traces[k]
	}
}

func (m *traceManager) startTrace() {
	for _, trace := range m.regionTrace {
		logs.Debug("region[%s] running trace", trace.region)
		go trace.runTraceJob()
	}
}

func (m *traceManager) getRegionBestTarget(region string) (traceTarget, bool) {
	regionTrace := m.regionTrace[region]
	if regionTrace == nil {
		logs.Warn("trace for region[%s] not exist", region)
		return traceTarget{}, false
	}

	return regionTrace.getBestNode()
}

type traceTarget struct {
	traceAddr  string
	serverAddr string
	scheme     string
}

type trace struct {
	region        string
	targets       []traceTarget
	targetScoreMu sync.Mutex
	targetScore   map[traceTarget]float64
	totalRtt      int32
}

func newTrace(region string) *trace {
	return &trace{
		region:        region,
		targetScoreMu: sync.Mutex{},
		targetScore:   make(map[traceTarget]float64),
	}
}

func (t *trace) addTarget(target traceTarget) {
	t.targets = append(t.targets, target)
}

func (t *trace) runTraceJob() {
	t.trace()
	tick := time.NewTicker(time.Second * 120)
	for range tick.C {
		t.trace()
	}
}

func (t *trace) trace() {
	for i, target := range t.targets {
		raddr, err := net.ResolveUDPAddr("udp", target.traceAddr)
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
		rconn.Close()

		if rtt < 0 {
			rtt = math.MaxInt
		}

		lossRank := t.calcLossScore(loss)
		delayRank := t.calcRttScore(rtt)
		score := lossRank + delayRank
		logs.Debug("region[%s] %s loss %d rtt %d lossRank %.4f delayRank %.4f score %.4f",
			t.region, target, loss, rtt, lossRank, delayRank, score)
		t.targetScoreMu.Lock()
		t.targetScore[t.targets[i]] = score
		t.targetScoreMu.Unlock()
	}
}

// f(p)  = 50                              p = 0,
// f(p) = 40+(0.75-p)x13                   0%  < p <= 0.75%,
// f(p) = 35+(1.25-p)x10                   0.75% < p <= 1.25%,
// f(p) = 30+(2.25-p)x5                    1.25% < p <= 2.25%,
// f(p) = 30+(p-2.25)x5x-1                 p > 2.25%
func (t *trace) calcLossScore(loss int) float64 {
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

func (t *trace) calcRttScore(rtt int) float64 {
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

func (t *trace) getBestNode() (traceTarget, bool) {
	t.targetScoreMu.Lock()
	defer t.targetScoreMu.Unlock()
	bestScore := float64(-1)
	var node traceTarget
	var ok bool = false
	for target, score := range t.targetScore {
		if bestScore*1000 < score*1000 {
			bestScore = score
			node = target
			ok = true
		}
	}
	return node, ok
}
