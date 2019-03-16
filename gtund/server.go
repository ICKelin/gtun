package gtund

import (
	"encoding/json"
	"io"
	"net"
	"strings"
	"time"

	"github.com/ICKelin/gtun/common"
	"github.com/ICKelin/gtun/logs"
)

var (
	authFailMsg    = "authorize fail"
	authSuccessMsg = "authorize success"
)

type ServerConfig struct {
	listenAddr  string
	authKey     string
	gateway     string
	routeUrl    string
	nameservers string
	reverseFile string
	tapMode     bool
}

type Server struct {
	listenAddr string
	listener   net.Listener
	authKey    string
	gateway    string
	sndqueue   chan *GtunClientContext
	done       chan struct{}

	iface       *Interface
	reverse     *Reverse
	dhcp        *DHCP
	forward     *Forward
	god         *God
	routeUrl    string
	nameservers []string
}

type GtunClientContext struct {
	conn    net.Conn
	payload []byte
}

func NewServer(cfg *ServerConfig) (*Server, error) {
	server := &Server{
		listenAddr:  cfg.listenAddr,
		authKey:     cfg.authKey,
		gateway:     cfg.gateway,
		routeUrl:    cfg.routeUrl,
		nameservers: strings.Split(cfg.nameservers, ","),
		forward:     NewForward(),
		sndqueue:    make(chan *GtunClientContext),
		done:        make(chan struct{}),
	}

	// init server listener
	listener, err := net.Listen("tcp", cfg.listenAddr)
	if err != nil {
		return nil, err
	}
	server.listener = listener

	// init virtual device
	devConfig := &InterfaceConfig{
		ip:     cfg.gateway,
		gw:     cfg.gateway,
		tapDev: cfg.tapMode,
	}
	ifce, err := NewInterface(devConfig)
	if err != nil {
		return nil, err
	}
	server.iface = ifce

	// init dhcp pool
	dhcpCfg := &DHCPConfig{
		gateway: cfg.gateway,
	}
	dhcp, err := NewDHCP(dhcpCfg)
	if err != nil {
		return nil, err
	}
	server.dhcp = dhcp

	// init reverse
	if cfg.reverseFile != "" {
		reverseCfg := &ReverseConfig{
			ruleFile: cfg.reverseFile,
		}
		r, err := NewReverse(reverseCfg)
		if err != nil {
			return nil, err
		}
		server.reverse = r
	}

	// init god module
	g := NewGod(GetConfig().GodCfg)
	server.god = g
	go func() {
		g.Run()

	}()

	return server, nil
}

func (server *Server) Run() {
	go server.readIface()
	go server.snd()

	for {
		select {
		case <-server.done:
			return
		default:
		}

		conn, err := server.listener.Accept()
		if err != nil {
			logs.Error("accept: %v", err)
			break
		}

		go server.onConn(conn)
	}
}

func (server *Server) Close() {
	server.listener.Close()
	close(server.done)
}

func (server *Server) onConn(conn net.Conn) {
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

	if !server.checkAuth(authMsg) {
		s2c.Status = authFailMsg
		server.authResp(conn, s2c)
		return
	}

	s2c.AccessIP, err = server.dhcp.SelectIP(authMsg.AccessIP)
	if err != nil {
		s2c.Status = err.Error()
		server.authResp(conn, s2c)
		return
	}

	defer server.dhcp.RecycleIP(s2c.AccessIP)

	s2c.Status = authSuccessMsg
	s2c.Gateway = server.gateway
	s2c.RouteScriptUrl = server.routeUrl
	s2c.Nameservers = server.nameservers
	server.authResp(conn, s2c)

	server.forward.Add(s2c.AccessIP, conn)
	defer server.forward.Del(s2c.AccessIP)

	logs.Info("accept cloud client from %s assign ip %s", conn.RemoteAddr().String(), s2c.AccessIP)

	go func() {
		server.god.UpdateClientCount(1)
		defer server.god.UpdateClientCount(-1)
	}()

	server.rcv(conn)
}

func (server *Server) rcv(conn net.Conn) {
	defer conn.Close()

	for {
		select {
		case <-server.done:
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

			server.sndqueue <- &GtunClientContext{conn: conn, payload: bytes}

		case common.C2C_DATA:
			_, err = server.iface.Write(pkt)
			if err != nil {
				logs.Error("read from iface: %v", err)
			}

		default:
			logs.Info("unimplement cmd %d, size: %d", cmd, len(pkt))
		}
	}
}

func (server *Server) snd() {
	for {
		select {
		case <-server.done:
			return
		default:
		}

		ctx := <-server.sndqueue
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

func (server *Server) readIface() {
	buff := make([]byte, 65536)
	for {
		select {
		case <-server.done:
			logs.Info("server receive done signal")
			return

		default:
		}

		nr, err := server.iface.Read(buff)
		if err != nil {
			if err != io.EOF {
				logs.Error("read from iface fail: %v", err)
			}
			continue
		}

		ethOffset := 0

		if server.iface.IsTAP() {
			f := Frame(buff[:nr])
			if f.Invalid() {
				continue
			}

			if !f.IsIPV4() {
				// broadcast
				server.forward.Broadcast(server.sndqueue, buff[:nr])
				continue
			}

			ethOffset = 14
		}

		p := Packet(buff[ethOffset:nr])

		if p.Invalid() {
			logs.Warn("invalid packet")
			continue
		}

		logs.Debug("%s %s", p.Src(), p.Dst())

		if p.Version() != 4 {
			logs.Warn("only support for ipv4")
			continue
		}

		peer := p.Dst()
		err = server.forward.Peer(server.sndqueue, peer, buff[:nr])
		if err != nil {
			logs.Error("send to ", peer, err)
			continue
		}

	}
}

func (server *Server) authResp(conn net.Conn, s2c *common.S2CAuthorize) {
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

func (server *Server) checkAuth(authMsg *common.C2SAuthorize) bool {
	return authMsg.Key == server.authKey
}

func (server *Server) isNewConnect(authMsg *common.C2SAuthorize) bool {
	return authMsg.AccessIP == ""
}
