package gtund

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"syscall"
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
	routeFile   string
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
	routes      []string
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

	// init routes deploy
	if cfg.routeFile != "" {
		routes, err := loadRouteRules(cfg.routeFile)
		if err != nil {
			return nil, err
		}
		server.routes = routes
	}

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
	s2c.RouteRule = server.routes
	s2c.Nameservers = server.nameservers
	server.authResp(conn, s2c)

	server.forward.Add(s2c.AccessIP, conn)
	defer server.forward.Del(s2c.AccessIP)

	glog.INFO("accept cloud client from", conn.RemoteAddr().String(), "assign ip", s2c.AccessIP)

	server.rcv(conn)
}

func (server *Server) rcv(conn net.Conn) {
	defer conn.Close()
	for {
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
			if nr < 14 {
				glog.WARM("too short ethernet frame", nr)
				continue
			}

			// Not eq ip pkt, just broadcast it
			// This handle maybe dangerous
			if whichProtocol(buff) != syscall.IPPROTO_IP {
				server.forward.table.Range(func(key, val interface{}) bool {
					conn, ok := val.(net.Conn)
					if ok {
						bytes, _ := common.Encode(common.C2C_DATA, buff[:nr])
						server.sndqueue <- &GtunClientContext{conn: conn, payload: bytes}
					}
					return true
				})
			}

			ethOffset = 14
		}

		if server.iface.IsTUN() {
			if nr < 20 {
				glog.WARM("too short ippkt", nr)
				continue
			}
		}

		if nr < ethOffset+20 {
			glog.WARM("to short ippkt", nr, ethOffset+20)
			continue
		}

		// TODO ip version
		dst := ""
		if isIPV4(buff[ethOffset]) {
			dst = fmt.Sprintf("%d.%d.%d.%d", buff[ethOffset+16], buff[ethOffset+17], buff[ethOffset+18], buff[ethOffset+19])
		} else {
			glog.WARM("not support ipv6")
		}
		c := server.forward.Get(dst)
		if c != nil {
			bytes, err := common.Encode(common.C2C_DATA, buff[:nr])
			if err != nil {
				glog.ERROR(err)
				continue
			}

			server.sndqueue <- &GtunClientContext{conn: c, payload: bytes}
		} else {
			glog.ERROR(dst, "offline")
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

func whichProtocol(frame []byte) int {
	if len(frame) > 14 {
		return int(frame[12])<<8 + int(frame[13])
	}
	return -1
}

func isIPV4(vhl byte) bool {
	if (vhl >> 4) == 4 {
		return true
	}
	return false
}

func loadRouteRules(rfile string) ([]string, error) {
	fp, err := os.Open(rfile)
	if err != nil {
		return nil, err
	}

	routes := make([]string, 0)

	linecount := 0
	maxbytes := 0xff00
	curbytes := 0
	reader := bufio.NewReader(fp)
	for {
		bline, _, err := reader.ReadLine()
		if err != nil {
			if err != io.EOF {
				return nil, err
			}
			break
		}

		line := string(bline)
		linecount += 1

		// 2018.04.20 rule store max 20 rule record
		// There is no plan to fix this "feature"
		if linecount > 20 {
			return nil, fmt.Errorf("rules set max record set to 20, suggest using url instead of rule file")
		}

		// 2018.04.20 check max bytes
		// since the protocol header set 2bytes for pkt header
		// once overflow, cli json decode fail
		curbytes += len(bline)
		if curbytes > maxbytes {
			return nil, fmt.Errorf("rule set max bytes 0xff00")
		}

		// ignore comment
		if len(line) > 0 && line[0] == '#' {
			continue
		}

		routes = append(routes, line)
	}

	return routes, nil
}
