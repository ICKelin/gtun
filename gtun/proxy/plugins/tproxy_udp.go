package plugins

//
//import (
//	"bytes"
//	"encoding/binary"
//	"fmt"
//	"github.com/ICKelin/gtun/gtun"
//	"github.com/ICKelin/gtun/gtun/route"
//	"github.com/ICKelin/gtun/internal/logs"
//	"github.com/ICKelin/gtun/internal/proto"
//	"github.com/ICKelin/gtun/internal/utils"
//	"github.com/ICKelin/optw/transport"
//	"io"
//	"net"
//	"sync"
//	"syscall"
//	"time"
//	"unsafe"
//)
//
//var (
//	// default udp timeout(read, write)(seconds)
//	defaultUDPTimeout = 10
//
//	// default udp session timeout(seconds)
//	defaultUDPSessionTimeout = 30
//)
//
//type udpSession struct {
//	stream     transport.Stream
//	lastActive time.Time
//}
//
//type UDPForward struct {
//	region         string
//	listenAddr     string
//	sessionTimeout int
//	readTimeout    time.Duration
//	writeTimeout   time.Duration
//	rawfd          int
//
//	// the session manager is the global session manager
//	// it stores opennotr_client to opennotr_server connection
//	routeManager *route.Manager
//
//	// udpSessions stores each client forward stream
//	// the purpose of udpSession is to reuse stream
//	udpSessions map[string]*udpSession
//	udpsessLock sync.Mutex
//
//	ratelimit *utils.RateLimit
//}
//
//func NewUDPForward(region string, cfg UDPForwardConfig, ratelimit *utils.RateLimit) *UDPForward {
//	readTimeout := cfg.ReadTimeout
//	if readTimeout <= 0 {
//		readTimeout = defaultUDPTimeout
//	}
//
//	writeTimeout := cfg.WriteTimeout
//	if writeTimeout <= 0 {
//		writeTimeout = defaultUDPTimeout
//	}
//
//	sessionTimeout := cfg.SessionTimeout
//	if sessionTimeout <= 0 {
//		sessionTimeout = defaultUDPSessionTimeout
//	}
//
//	return &UDPForward{
//		region:         region,
//		listenAddr:     cfg.ListenAddr,
//		readTimeout:    time.Duration(readTimeout) * time.Second,
//		writeTimeout:   time.Duration(writeTimeout) * time.Second,
//		sessionTimeout: sessionTimeout,
//		routeManager:   route.GetRouteManager(),
//		udpSessions:    make(map[string]*udpSession),
//		ratelimit:      ratelimit,
//	}
//}
//
//// Listen listens an udp port, since that we use tproxy to
//// redirect traffic to this listened udp port
//// so the socket should set to ip transparent option
//func (f *UDPForward) Listen() (*net.UDPConn, error) {
//	laddr, err := net.ResolveUDPAddr("udp", f.listenAddr)
//	if err != nil {
//		logs.Error("resolve udp fail: %v", err)
//		return nil, err
//	}
//
//	lconn, err := net.ListenUDP("udp", laddr)
//	if err != nil {
//		return nil, err
//	}
//
//	// set socket with ip transparent option
//	file, err := lconn.File()
//	if err != nil {
//		lconn.Close()
//		return nil, err
//	}
//	defer file.Close()
//
//	err = syscall.SetsockoptInt(int(file.Fd()), syscall.SOL_IP, syscall.IP_TRANSPARENT, 1)
//	if err != nil {
//		lconn.Close()
//		return nil, err
//	}
//
//	// set socket with recv origin dst option
//	err = syscall.SetsockoptInt(int(file.Fd()), syscall.SOL_IP, syscall.IP_RECVORIGDSTADDR, 1)
//	if err != nil {
//		return nil, err
//	}
//
//	// create raw socket fd
//	// we use rawsocket to send udp packet back to client.
//	rawfd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_RAW)
//	if err != nil || rawfd < 0 {
//		logs.Error("call socket fail: %v", err)
//		return nil, err
//	}
//
//	err = syscall.SetsockoptInt(rawfd, syscall.IPPROTO_IP, syscall.IP_HDRINCL, 1)
//	if err != nil {
//		return nil, err
//	}
//
//	f.rawfd = rawfd
//	return lconn, nil
//}
//
//func (f *UDPForward) Serve(lconn *net.UDPConn) error {
//	go f.recycleSession()
//	buf := make([]byte, 64*1024)
//	oob := make([]byte, 1024)
//	for {
//		// udp is not connect oriented, it should use read message
//		// and read the origin dst ip and port from msghdr
//		nr, oobn, _, raddr, err := lconn.ReadMsgUDP(buf, oob)
//		if err != nil {
//			logs.Error("read from udp fail: %v", err)
//			break
//		}
//
//		origindst, err := f.getOriginDst(oob[:oobn])
//		if err != nil {
//			logs.Error("get origin dst fail: %v", err)
//			continue
//		}
//
//		dip, dport, _ := net.SplitHostPort(origindst.String())
//		sip, sport, _ := net.SplitHostPort(raddr.String())
//
//		key := fmt.Sprintf("%s:%s:%s:%s", sip, sport, dip, dport)
//
//		f.udpsessLock.Lock()
//		udpsess := f.udpSessions[key]
//		if udpsess != nil {
//			udpsess.lastActive = time.Now()
//			f.udpsessLock.Unlock()
//		} else {
//			f.udpsessLock.Unlock()
//			sess := f.routeManager.Route(f.region, dip)
//			if sess == nil {
//				logs.Error("no route to host: %s", dip)
//				continue
//			}
//
//			stream, err := sess.OpenStream()
//			if err != nil {
//				logs.Error("open stream fail: %v", err)
//				continue
//			}
//
//			logs.Debug("open new stream for %s", key)
//			udpsess = &udpSession{stream, time.Now()}
//			f.udpsessLock.Lock()
//			f.udpSessions[key] = udpsess
//			f.udpsessLock.Unlock()
//
//			bytes := proto.EncodeProxyProtocol("udp", sip, sport, dip, dport)
//			stream.SetWriteDeadline(time.Now().Add(f.writeTimeout))
//			_, err = stream.Write(bytes)
//			stream.SetWriteDeadline(time.Time{})
//			if err != nil {
//				logs.Error("stream write fail: %v", err)
//				continue
//			}
//
//			go f.forwardUDP(stream, key, origindst, raddr)
//		}
//
//		stream := udpsess.stream
//
//		bytes := proto.EncodeData(buf[:nr])
//		stream.SetWriteDeadline(time.Now().Add(f.writeTimeout))
//		_, err = stream.Write(bytes)
//		stream.SetWriteDeadline(time.Time{})
//		if err != nil {
//			logs.Error("stream write fail: %v", err)
//		}
//	}
//	return nil
//}
//
//// forwardUDP reads from stream and write to tofd via rawsocket
//func (f *UDPForward) forwardUDP(stream transport.Stream, sessionKey string, fromaddr, toaddr *net.UDPAddr) {
//	defer stream.Close()
//	defer func() {
//		f.udpsessLock.Lock()
//		delete(f.udpSessions, sessionKey)
//		f.udpsessLock.Unlock()
//	}()
//
//	hdr := make([]byte, 2)
//	for {
//		nr, err := stream.Read(hdr)
//		if err != nil {
//			if err != io.EOF {
//				logs.Error("read stream fail %v", err)
//			}
//			break
//		}
//		if nr != 2 {
//			logs.Error("invalid bodylen: %d", nr)
//			continue
//		}
//
//		nlen := binary.BigEndian.Uint16(hdr)
//		buf := make([]byte, nlen)
//		stream.SetReadDeadline(time.Now().Add(f.readTimeout))
//		_, err = io.ReadFull(stream, buf)
//		stream.SetReadDeadline(time.Time{})
//		if err != nil {
//			logs.Error("read stream body fail: %v", err)
//			break
//		}
//
//		err = gtun.sendUDPViaRaw(f.rawfd, fromaddr, toaddr, buf)
//		if err != nil {
//			logs.Error("send via raw socket fail: %v", err)
//		}
//
//		f.udpsessLock.Lock()
//		udpsess := f.udpSessions[sessionKey]
//		if udpsess != nil {
//			udpsess.lastActive = time.Now()
//		}
//		f.udpsessLock.Unlock()
//	}
//}
//
//func (f *UDPForward) recycleSession() {
//	tick := time.NewTicker(time.Second * 5)
//	for range tick.C {
//		f.udpsessLock.Lock()
//		for k, s := range f.udpSessions {
//			if time.Now().Sub(s.lastActive).Seconds() > float64(f.sessionTimeout) {
//				logs.Warn("remove udp session")
//				s.stream.Close()
//				delete(f.udpSessions, k)
//			}
//		}
//		f.udpsessLock.Unlock()
//	}
//}
//
//func (f *UDPForward) getOriginDst(hdr []byte) (*net.UDPAddr, error) {
//	msgs, err := syscall.ParseSocketControlMessage(hdr)
//	if err != nil {
//		return nil, err
//	}
//
//	var origindst *net.UDPAddr
//	for _, msg := range msgs {
//		if msg.Header.Level == syscall.SOL_IP &&
//			msg.Header.Type == syscall.IP_RECVORIGDSTADDR {
//			originDstRaw := &syscall.RawSockaddrInet4{}
//			err := binary.Read(bytes.NewReader(msg.Data), binary.LittleEndian, originDstRaw)
//			if err != nil {
//				logs.Error("read origin dst fail: %v", err)
//				continue
//			}
//
//			// only support for ipv4
//			if originDstRaw.Family == syscall.AF_INET {
//				pp := (*syscall.RawSockaddrInet4)(unsafe.Pointer(originDstRaw))
//				p := (*[2]byte)(unsafe.Pointer(&pp.Port))
//				origindst = &net.UDPAddr{
//					IP:   net.IPv4(pp.Addr[0], pp.Addr[1], pp.Addr[2], pp.Addr[3]),
//					Port: int(p[0])<<8 + int(p[1]),
//				}
//			}
//		}
//	}
//
//	if origindst == nil {
//		return nil, fmt.Errorf("get origin dst fail")
//	}
//
//	return origindst, nil
//}
