package gtund

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"time"

	"github.com/ICKelin/gtun/common"
	"github.com/ICKelin/gtun/pkg/logs"
)

var (
	authFailMsg    = "authorize fail"
	authSuccessMsg = "authorize success"
)

type ServerConfig struct {
	Listen      string   `toml:"listen"`
	AuthKey     string   `toml:"auth_key"`
	RouteUrl    string   `toml:"route_url"`
	Nameservers []string `toml:"nameservers"`
}

type Server struct {
	listenAddr  string
	authKey     string
	gateway     string
	routeUrl    string
	nameservers []string

	iface   *Interface
	dhcp    *DHCP
	forward *Forward
}

type GtunClientContext struct {
	conn    net.Conn
	payload []byte
}

func NewServer(cfg ServerConfig, dhcp *DHCP, iface *Interface) (*Server, error) {
	s := &Server{
		listenAddr: cfg.Listen,
		authKey:    cfg.AuthKey,
		routeUrl:   cfg.RouteUrl,
		forward:    NewForward(),
		iface:      iface,
		dhcp:       dhcp,
	}

	return s, nil
}

func (s *Server) Run() error {
	go s.readIface()
	// go s.snd()

	listener, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		return err
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}

		go s.onConn(conn)
	}
}

func (s *Server) onConn(conn net.Conn) {
	defer conn.Close()

	cmd, payload, err := common.Decode(conn)
	if err != nil {
		logs.Error("decode auth msg fail: %v", err)
		return
	}

	if cmd != common.C2S_AUTHORIZE {
		logs.Error("invalid authhorize cmd: %d", cmd)
		return
	}

	authMsg := &common.C2SAuthorize{}
	err = json.Unmarshal(payload, &authMsg)
	if err != nil {
		logs.Error("invalid auth msg: %v", err)
		return
	}

	s2c := &common.S2CAuthorize{}

	if !s.checkAuth(authMsg) {
		s2c.Status = authFailMsg
		s.authResp(conn, s2c)
		return
	}

	s2c.AccessIP, err = s.dhcp.SelectIP()
	if err != nil {
		s2c.Status = err.Error()
		s.authResp(conn, s2c)
		return
	}

	defer s.dhcp.RecycleIP(s2c.AccessIP)

	s2c.Status = authSuccessMsg
	s2c.Gateway = s.dhcp.gateway
	s2c.RouteScriptUrl = s.routeUrl
	s2c.Nameservers = s.nameservers
	s.authResp(conn, s2c)

	sndbuf := make(chan []byte)
	s.forward.Add(s2c.AccessIP, sndbuf)
	defer s.forward.Del(s2c.AccessIP)

	logs.Info("accept cloud client from %s assign ip %s", conn.RemoteAddr().String(), s2c.AccessIP)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go s.snd(ctx, conn, sndbuf)
	s.recv(conn, sndbuf)
}

func (s *Server) recv(conn net.Conn, sndbuf chan []byte) {
	defer conn.Close()

	for {
		cmd, pkt, err := common.Decode(conn)
		if err != nil {
			if err != io.EOF {
				logs.Error("decode fail: ", err)
			}
			break
		}

		switch cmd {
		case common.C2S_HEARTBEAT:
			logs.Debug("on C2S_HEARTBEAT: %s", conn.RemoteAddr().String())
			bytes, err := common.Encode(common.S2C_HEARTBEAT, nil)
			if err != nil {
				logs.Error("encode fail: %v", err)
				continue
			}

			sndbuf <- bytes

		case common.C2C_DATA:
			_, err = s.iface.Write(pkt)
			if err != nil {
				logs.Error("read from iface: %v", err)
			}

		default:
			logs.Info("unimplement cmd %d, size: %d", cmd, len(pkt))
		}
	}
}

func (s *Server) snd(ctx context.Context, conn net.Conn, sndbuf chan []byte) {
	for {
		select {
		case <-ctx.Done():
			logs.Warn("close send")
			return

		case payload := <-sndbuf:
			conn.SetWriteDeadline(time.Now().Add(time.Second * 30))
			nw, err := conn.Write(payload)
			conn.SetWriteDeadline(time.Time{})
			if err != nil {
				logs.Error("write to peer fail: %v", err)
			}

			if nw != len(payload) {
				logs.Error("write not full %d %d", nw, len(payload))
			}
		}
	}
}

func (s *Server) readIface() {
	buff := make([]byte, 65536)
	for {
		nr, err := s.iface.Read(buff)
		if err != nil {
			if err != io.EOF {
				logs.Error("read from iface fail: %v", err)
			}
			continue
		}

		ethOffset := 0

		if s.iface.IsTAP() {
			f := Frame(buff[:nr])
			if f.Invalid() {
				continue
			}

			if !f.IsIPV4() {
				// broadcast
				s.forward.Broadcast(buff[:nr])
				continue
			}

			ethOffset = 14
		}

		p := Packet(buff[ethOffset:nr])

		if p.Invalid() {
			logs.Warn("invalid packet")
			continue
		}

		if p.Version() != 4 {
			logs.Warn("only support for ipv4")
			continue
		}

		peer := p.Dst()
		err = s.forward.Peer(peer, buff[:nr])
		if err != nil {
			logs.Error("send to ", peer, err)
			continue
		}

	}
}

func (s *Server) authResp(conn net.Conn, s2c *common.S2CAuthorize) {
	resp, err := json.Marshal(s2c)
	if err != nil {
		logs.Error("marshal aut reply fail:", err)
		return
	}

	buff, _ := common.Encode(common.S2C_AUTHORIZE, resp)
	_, err = conn.Write(buff)
	if err != nil {
		logs.Error("send auth reply fail:", err)
		return
	}
}

func (s *Server) checkAuth(authMsg *common.C2SAuthorize) bool {
	return authMsg.Key == s.authKey
}
