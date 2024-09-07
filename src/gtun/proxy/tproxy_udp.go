package proxy

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/ICKelin/gtun/src/gtun/route"
	"github.com/ICKelin/gtun/src/internal/logs"
	"github.com/ICKelin/gtun/src/internal/proto"
	"github.com/ICKelin/gtun/src/internal/utils"
	transport "github.com/ICKelin/optw"
	"io"
	"net"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

func init() {
	_ = Register("tproxy_udp", NewTProxyUDP)
}

var (
	// default udp timeout(read, write)(seconds)
	defaultUDPTimeout = 10

	// default udp session timeout(seconds)
	defaultUDPSessionTimeout = 30
)

type udpSession struct {
	stream     transport.Stream
	lastActive time.Time
}

type TProxyUDPConfig struct {
	ReadTimeout    int    `json:"read_timeout"`
	WriteTimeout   int    `json:"write_timeout"`
	SessionTimeout int    `json:"session_timeout"`
	ListenAddr     string `json:"listen_addr"`
}

type TProxyUDP struct {
	region         string
	listenAddr     string
	sessionTimeout time.Duration
	readTimeout    time.Duration
	writeTimeout   time.Duration
	rawfd          int

	routeManager *route.Manager

	// udpSessions stores each client forward stream
	// the purpose of udpSession is to reuse stream
	udpSessions map[string]*udpSession
	udpsessLock sync.Mutex
}

func NewTProxyUDP() Proxy {
	return &TProxyUDP{}
}

func (p *TProxyUDP) Name() string {
	return "tproxy_udp"
}

func (p *TProxyUDP) Setup(region string, cfgContent json.RawMessage) error {
	var cfg = TProxyUDPConfig{}
	err := json.Unmarshal(cfgContent, &cfg)
	if err != nil {
		return nil
	}

	err = p.initConfig(region, cfg)
	if err != nil {
		return err
	}

	return nil
}

func (p *TProxyUDP) initConfig(region string, cfg TProxyUDPConfig) error {
	readTimeout := cfg.ReadTimeout
	if readTimeout <= 0 {
		readTimeout = defaultUDPTimeout
	}

	writeTimeout := cfg.WriteTimeout
	if writeTimeout <= 0 {
		writeTimeout = defaultUDPTimeout
	}

	sessionTimeout := cfg.SessionTimeout
	if sessionTimeout <= 0 {
		sessionTimeout = defaultUDPSessionTimeout
	}

	p.region = region
	p.listenAddr = cfg.ListenAddr
	p.writeTimeout = time.Duration(writeTimeout) * time.Second
	p.readTimeout = time.Duration(readTimeout) * time.Second
	p.sessionTimeout = time.Duration(sessionTimeout) * time.Second
	p.udpSessions = make(map[string]*udpSession)
	p.routeManager = route.GetRouteManager()
	return nil
}

// ListenAndServe listens an udp port, since that we use tproxy to
// redirect traffic to this listened udp port
// so the socket should set to ip transparent option
func (p *TProxyUDP) ListenAndServe() error {
	laddr, err := net.ResolveUDPAddr("udp", p.listenAddr)
	if err != nil {
		return err
	}

	lconn, err := net.ListenUDP("udp", laddr)
	if err != nil {
		return err
	}

	// set socket with ip transparent option
	file, err := lconn.File()
	if err != nil {
		lconn.Close()
		return err
	}
	defer file.Close()

	err = syscall.SetsockoptInt(int(file.Fd()), syscall.SOL_IP, syscall.IP_TRANSPARENT, 1)
	if err != nil {
		lconn.Close()
		return err
	}

	// set socket with recv origin dst option
	err = syscall.SetsockoptInt(int(file.Fd()), syscall.SOL_IP, syscall.IP_RECVORIGDSTADDR, 1)
	if err != nil {
		return err
	}

	// create raw socket fd
	// we use rawsocket to send udp packet back to client.
	rawfd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_RAW)
	if err != nil || rawfd < 0 {
		return err
	}

	err = syscall.SetsockoptInt(rawfd, syscall.IPPROTO_IP, syscall.IP_HDRINCL, 1)
	if err != nil {
		return err
	}

	p.rawfd = rawfd
	return p.serve(lconn)
}

func (p *TProxyUDP) serve(lconn *net.UDPConn) error {
	logs.Info("region[%s] %s listen %s", p.region, p.Name(), p.listenAddr)
	go p.recycleSession()
	buf := make([]byte, 64*1024)
	oob := make([]byte, 1024)
	for {
		// udp is not connect oriented, it should use read message
		// and read the origin dst ip and port from msghdr
		nr, oobn, _, raddr, err := lconn.ReadMsgUDP(buf, oob)
		if err != nil {
			logs.Error("read from udp fail: %v", err)
			break
		}

		origindst, err := p.getOriginDst(oob[:oobn])
		if err != nil {
			logs.Error("get origin dst fail: %v", err)
			continue
		}

		dip, dport, _ := net.SplitHostPort(origindst.String())
		sip, sport, _ := net.SplitHostPort(raddr.String())

		key := fmt.Sprintf("%s:%s:%s:%s", sip, sport, dip, dport)

		p.udpsessLock.Lock()
		udpsess := p.udpSessions[key]
		if udpsess != nil {
			udpsess.lastActive = time.Now()
			p.udpsessLock.Unlock()
		} else {
			p.udpsessLock.Unlock()
			sess := p.routeManager.Route(p.region, dip)
			if sess == nil {
				logs.Error("no route to host: %s", dip)
				continue
			}

			stream, err := sess.OpenStream()
			if err != nil {
				// force close to trigger reconnect
				// quic CAN'T get close state by sess.IsClose()
				// Close to trigger quic reconnect
				sess.Close()
				logs.Error("open stream fail: %v", err)
				continue
			}

			logs.Debug("open new stream for %s", key)
			udpsess = &udpSession{stream, time.Now()}
			p.udpsessLock.Lock()
			p.udpSessions[key] = udpsess
			p.udpsessLock.Unlock()

			bytes := proto.EncodeProxyProtocol("udp", sip, sport, dip, dport)
			stream.SetWriteDeadline(time.Now().Add(p.writeTimeout))
			_, err = stream.Write(bytes)
			stream.SetWriteDeadline(time.Time{})
			if err != nil {
				logs.Error("stream write fail: %v", err)
				continue
			}

			go p.doProxy(stream, key, origindst, raddr)
		}

		stream := udpsess.stream

		bytes := proto.EncodeData(buf[:nr])
		stream.SetWriteDeadline(time.Now().Add(p.writeTimeout))
		_, err = stream.Write(bytes)
		stream.SetWriteDeadline(time.Time{})
		if err != nil {
			logs.Error("stream write fail: %v", err)
		}
	}
	return nil
}

// doProxy reads from stream and write to tofd via rawsocket
func (p *TProxyUDP) doProxy(stream transport.Stream, sessionKey string, fromaddr, toaddr *net.UDPAddr) {
	defer stream.Close()
	defer func() {
		p.udpsessLock.Lock()
		delete(p.udpSessions, sessionKey)
		p.udpsessLock.Unlock()
	}()

	hdr := make([]byte, 2)
	for {
		nr, err := stream.Read(hdr)
		if err != nil {
			if err != io.EOF {
				logs.Error("read stream fail %v", err)
			}
			break
		}
		if nr != 2 {
			logs.Error("invalid bodylen: %d", nr)
			continue
		}

		nlen := binary.BigEndian.Uint16(hdr)
		buf := make([]byte, nlen)
		stream.SetReadDeadline(time.Now().Add(p.readTimeout))
		_, err = io.ReadFull(stream, buf)
		stream.SetReadDeadline(time.Time{})
		if err != nil {
			logs.Error("read stream body fail: %v", err)
			break
		}

		err = utils.SendUDPViaRaw(p.rawfd, fromaddr, toaddr, buf)
		if err != nil {
			logs.Error("send via raw socket fail: %v", err)
		}

		p.udpsessLock.Lock()
		udpsess := p.udpSessions[sessionKey]
		if udpsess != nil {
			udpsess.lastActive = time.Now()
		}
		p.udpsessLock.Unlock()
	}
}

func (p *TProxyUDP) recycleSession() {
	tick := time.NewTicker(time.Second * 5)
	for range tick.C {
		p.udpsessLock.Lock()
		for k, s := range p.udpSessions {
			if time.Now().Sub(s.lastActive).Seconds() > float64(p.sessionTimeout) {
				logs.Warn("remove udp session")
				s.stream.Close()
				delete(p.udpSessions, k)
			}
		}
		p.udpsessLock.Unlock()
	}
}

func (p *TProxyUDP) getOriginDst(hdr []byte) (*net.UDPAddr, error) {
	msgs, err := syscall.ParseSocketControlMessage(hdr)
	if err != nil {
		return nil, err
	}

	var origindst *net.UDPAddr
	for _, msg := range msgs {
		if msg.Header.Level == syscall.SOL_IP &&
			msg.Header.Type == syscall.IP_RECVORIGDSTADDR {
			originDstRaw := &syscall.RawSockaddrInet4{}
			err := binary.Read(bytes.NewReader(msg.Data), binary.LittleEndian, originDstRaw)
			if err != nil {
				logs.Error("read origin dst fail: %v", err)
				continue
			}

			// only support for ipv4
			if originDstRaw.Family == syscall.AF_INET {
				pp := (*syscall.RawSockaddrInet4)(unsafe.Pointer(originDstRaw))
				p := (*[2]byte)(unsafe.Pointer(&pp.Port))
				origindst = &net.UDPAddr{
					IP:   net.IPv4(pp.Addr[0], pp.Addr[1], pp.Addr[2], pp.Addr[3]),
					Port: int(p[0])<<8 + int(p[1]),
				}
			}
		}
	}

	if origindst == nil {
		return nil, fmt.Errorf("get origin dst fail")
	}

	return origindst, nil
}
