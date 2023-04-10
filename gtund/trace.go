package gtund

import (
	"github.com/ICKelin/gtun/internal/logs"
	"net"
)

type TraceServer struct {
	listenAddr string
}

func NewTraceServer(listenAddr string) *TraceServer {
	return &TraceServer{listenAddr: listenAddr}
}

func (s *TraceServer) ListenAndServe() {
	laddr, err := net.ResolveUDPAddr("udp", s.listenAddr)
	if err != nil {
		logs.Error("invalid udp address: %v", err)
		return
	}

	lconn, err := net.ListenUDP("udp", laddr)
	if err != nil {
		logs.Error("listen udp fail: %v", err)
		return
	}
	defer lconn.Close()

	buf := make([]byte, 100)
	for {
		nr, raddr, err := lconn.ReadFromUDP(buf)
		if err != nil {
			logs.Error("read trace fail: %v", err)
		}

		_, err = lconn.WriteToUDP(buf[:nr], raddr)
		if err != nil {
			logs.Error("write trace fail: %v", err)
		}
	}
}
