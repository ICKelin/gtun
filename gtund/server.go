package gtund

import (
	"encoding/json"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"github.com/ICKelin/glog"
	"github.com/ICKelin/gtun/common"
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
	stop       chan struct{}

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
		stop:        make(chan struct{}),
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

		// whether we should exit
		if GetConfig().GodCfg.Must {
			os.Exit(-1)
		}
	}()

	return server, nil
}

func (server *Server) Run() {
	go server.readIface()
	go server.snd()

	for {
		conn, err := server.listener.Accept()
		if err != nil {
			glog.ERROR(err)
			break
		}

		go server.onConn(conn)
	}
}

func (server *Server) Stop() {
	server.listener.Close()
	close(server.stop)
}

func (server *Server) onConn(conn net.Conn) {
	defer conn.Close()

	cmd, payload, err := common.Decode(conn)
	if err != nil {
		glog.ERROR("decode auth msg fail:", err)
		return
	}

	if cmd != common.C2S_AUTHORIZE {
		glog.ERROR("invalid authhorize cmd", cmd)
		return
	}

	authMsg := &common.C2SAuthorize{}
	err = json.Unmarshal(payload, &authMsg)
	if err != nil {
		glog.ERROR("invalid auth msg", err)
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

	glog.INFO("accept cloud client from", conn.RemoteAddr().String(), "assign ip", s2c.AccessIP)

	server.god.UpdateClientCount(1)
	defer server.god.UpdateClientCount(-1)

	server.rcv(conn)
}

func (server *Server) rcv(conn net.Conn) {
	defer conn.Close()
	for {
		select {
		case <-server.stop:
			return
		default:

		}
		cmd, pkt, err := common.Decode(conn)
		if err != nil {
			if err != io.EOF {
				glog.ERROR(err)
			}
			break
		}

		switch cmd {
		case common.C2S_HEARTBEAT:
			bytes, err := common.Encode(common.S2C_HEARTBEAT, nil)
			if err != nil {
				glog.ERROR(err)
				continue
			}

			server.sndqueue <- &GtunClientContext{conn: conn, payload: bytes}

		case common.C2C_DATA:
			_, err = server.iface.Write(pkt)
			if err != nil {
				glog.ERROR(err)
			}

		default:
			glog.INFO("unimplement cmd", cmd, len(pkt))
		}
	}
}

func (server *Server) snd() {
	for {
		select {
		case <-server.stop:
			return
		default:
		}

		ctx := <-server.sndqueue
		ctx.conn.SetWriteDeadline(time.Now().Add(time.Second * 30))
		nw, err := ctx.conn.Write(ctx.payload)
		ctx.conn.SetWriteDeadline(time.Time{})
		if err != nil {
			glog.ERROR(err)
		}

		if nw != len(ctx.payload) {
			glog.ERROR("write not full", nw, len(ctx.payload))
		}
	}
}

func (server *Server) readIface() {
	buff := make([]byte, 65536)
	for {
		select {
		case <-server.stop:
			return
		default:
		}

		nr, err := server.iface.Read(buff)
		if err != nil {
			if err != io.EOF {
				glog.ERROR(err)
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
				return
			}

			ethOffset = 14
		}

		p := Packet(buff[ethOffset:nr])

		if p.Invalid() {
			continue
		}

		if p.Version() != 4 {
			continue
		}

		peer := p.Dst()
		err = server.forward.Peer(server.sndqueue, peer, buff[:nr])
		if err != nil {
			glog.ERROR("send to ", peer, err)
			continue
		}
	}
}

func (server *Server) authResp(conn net.Conn, s2c *common.S2CAuthorize) {
	resp, err := json.Marshal(s2c)
	if err != nil {
		glog.ERROR("marshal aut reply fail:", err)
		return
	}

	buff, _ := common.Encode(common.S2C_AUTHORIZE, resp)
	_, err = conn.Write(buff)
	if err != nil {
		glog.ERROR("send auth reply fail:", err)
		return
	}
}

func (server *Server) checkAuth(authMsg *common.C2SAuthorize) bool {
	return authMsg.Key == server.authKey
}

func (server *Server) isNewConnect(authMsg *common.C2SAuthorize) bool {
	return authMsg.AccessIP == ""
}
