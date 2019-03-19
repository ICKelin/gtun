package gtund

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/ICKelin/gtun/common"
	"github.com/ICKelin/gtun/logs"
)

var (
	authFailMsg    = "authorize fail"
	authSuccessMsg = "authorize success"
)

type ServerConfig struct {
	Listen   string `toml:"listen"`
	AuthKey  string `toml:"auth_key"`
	RouteUrl string `toml:"route_url"`
}

type Server struct {
	listenAddr string
	authKey    string
	gateway    string
	sndqueue   chan *GtunClientContext
	done       chan struct{}

	iface       *Interface
	dhcp        *DHCP
	forward     *Forward
	registry    *Registry
	routeUrl    string
	nameservers []string
}

type GtunClientContext struct {
	conn    net.Conn
	payload []byte
}

func NewServer(cfg *ServerConfig, dhcp *DHCP, iface *Interface, registry *Registry) (*Server, error) {
	s := &Server{
		listenAddr: cfg.Listen,
		authKey:    cfg.AuthKey,
		routeUrl:   cfg.RouteUrl,
		forward:    NewForward(),
		sndqueue:   make(chan *GtunClientContext),
		done:       make(chan struct{}),
		iface:      iface,
		dhcp:       dhcp,
		registry:   registry,
	}

	return s, nil
}

func (s *Server) Run() error {
	go s.readIface()
	go s.snd()

	listener, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		return err
	}
	defer listener.Close()

	for {
		select {
		case <-s.done:
			return fmt.Errorf("receive done signal")

		default:
		}

		conn, err := listener.Accept()
		if err != nil {
			return err
		}

		go s.onConn(conn)
	}
}

func (s *Server) Close() {
	close(s.done)
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

	s.forward.Add(s2c.AccessIP, conn)
	defer s.forward.Del(s2c.AccessIP)

	logs.Info("accept cloud client from %s assign ip %s", conn.RemoteAddr().String(), s2c.AccessIP)

	if s.registry != nil {
		s.registry.Sync(1)
		defer s.registry.Sync(-1)
	}

	s.handleClient(conn)
}

func (s *Server) handleClient(conn net.Conn) {
	defer conn.Close()

	for {
		select {
		case <-s.done:
			logs.Info("server receive done signal")
			return
		default:
		}

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

			s.sndqueue <- &GtunClientContext{conn: conn, payload: bytes}

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

func (s *Server) snd() {
	for {
		select {
		case <-s.done:
			logs.Info("snd receive done signal")
			return

		case ctx := <-s.sndqueue:
			ctx.conn.SetWriteDeadline(time.Now().Add(time.Second * 30))
			nw, err := ctx.conn.Write(ctx.payload)
			ctx.conn.SetWriteDeadline(time.Time{})
			if err != nil {
				logs.Error("write to peer fail: %v", err)
			}

			if nw != len(ctx.payload) {
				logs.Error("write not full %d %d", nw, len(ctx.payload))
			}
		}
	}
}

func (s *Server) readIface() {
	buff := make([]byte, 65536)
	for {
		select {
		case <-s.done:
			logs.Info("server receive done signal")
			return

		default:
		}

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
				s.forward.Broadcast(s.sndqueue, buff[:nr])
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
		err = s.forward.Peer(s.sndqueue, peer, buff[:nr])
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
