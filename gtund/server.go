package gtund

import (
	"bufio"
	"encoding/json"
	"errors"
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
	clients     *ClientPool
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
		clients:     NewClientPool(),
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
		routes, err := LoadRules(cfg.routeFile)
		if err != nil {
			return nil, err
		}
		server.routes = routes
	}

	return server, nil
}

func (server *Server) Run() {
	go server.readDevice()
	go server.pickAndSend()

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

	accessip, err := server.auth(conn)
	if err != nil {
		glog.ERROR(err)
		return
	}

	server.clients.Add(accessip, conn)
	defer server.clients.Del(accessip)
	defer server.dhcp.RecycleIP(accessip)
	glog.INFO("accept cloud client from", conn.RemoteAddr().String(), "assign ip", accessip)

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

// Purpose:
//			Read pkt/frame from tun/tap device and send back to gtun_cli
//			For tap device, I do not record MAC address table, STILL USE
//			IP ADDRESS AS SESSION KEY.
// Parameters:
//			ifce => device to read
//			sndqueue => send back queue
//
func (server *Server) readDevice() {
	buff := make([]byte, 65536)
	for {
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
			if WhichProtocol(buff) != syscall.IPPROTO_IP {
				server.clients.Lock()
				for _, c := range server.clients.client {
					bytes, _ := common.Encode(common.C2C_DATA, buff[:nr])

					server.sndqueue <- &GtunClientContext{conn: c, payload: bytes}
				}
				server.clients.Unlock()
				continue
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
		c := server.clients.Get(dst)
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

func (server *Server) pickAndSend() {
	for {
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

func (server *Server) auth(conn net.Conn) (accessip string, err error) {
	cmd, payload, err := common.Decode(conn)
	if err != nil {
		return "", err
	}

	if cmd != common.C2S_AUTHORIZE {
		return "", fmt.Errorf("invalid authhorize cmd %d", cmd)
	}

	auth := &common.C2SAuthorize{}
	err = json.Unmarshal(payload, &auth)
	if err != nil {
		return "", err
	}

	accessip = auth.AccessIP

	s2cauthorize := &common.S2CAuthorize{
		AccessIP:    accessip,
		Status:      "authorize fail",
		RouteRule:   make([]string, 0),
		Nameservers: make([]string, 0),
		Gateway:     server.gateway,
	}

	if auth.Key == server.authKey {
		s2cauthorize.Status = "authorize success"
		if accessip == "" {
			accessip = server.dhcp.SelectIP()
			if accessip == "" {
				return "", errors.New("not enough ip in dhcp pool")
			}
			s2cauthorize.AccessIP = accessip
		}
		s2cauthorize.RouteRule = server.routes
		s2cauthorize.Nameservers = server.nameservers
	}

	resp, err := json.Marshal(s2cauthorize)
	if err != nil {
		return "", err
	}

	buff, _ := common.Encode(common.S2C_AUTHORIZE, resp)
	_, err = conn.Write(buff)
	if err != nil {
		return "", err
	}

	return accessip, nil
}

func WhichProtocol(frame []byte) int {
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

// Purpose:
//			Loading ip/cidr from file and deploy to gtun_cli
//			It seems like deploy router table to client, tell
//			the client to route these ips/cidrs
//			THERE IS NOT IP VALIDATE
func LoadRules(rfile string) ([]string, error) {
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
