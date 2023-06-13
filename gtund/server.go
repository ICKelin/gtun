package gtund

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/ICKelin/gtun/internal/logs"
	"github.com/ICKelin/gtun/internal/proto"
	"github.com/ICKelin/optw/transport"
)

var (
	authFailMsg    = "authorize fail"
	authSuccessMsg = "authorize success"
)

type ServerConfig struct {
	Listen         string `yaml:"listen"`
	AuthKey        string `yaml:"authKey"`
	Scheme         string `yaml:"scheme"`
	ListenerConfig string `yaml:"listenerConfig"`
}

type Server struct {
	listener transport.Listener
	udpPool  sync.Pool
	tcpPool  sync.Pool
}

func NewServer(listener transport.Listener) *Server {
	return &Server{
		listener: listener,
		udpPool: sync.Pool{
			New: func() interface{} {
				return make([]byte, 1024*64)
			},
		},
		tcpPool: sync.Pool{
			New: func() interface{} {
				return make([]byte, 1024*8)
			},
		},
	}
}

func (s *Server) Run() error {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			logs.Error("listener accept fail: %v", err)
			if ne, ok := err.(net.Error); ok {
				if ne.Temporary() {
					time.Sleep(time.Millisecond * 100)
					continue
				}
			}
			return err
		}

		logs.Info("accept connect from remote: %s", conn.RemoteAddr().String())
		go s.onConn(conn)
	}
}

func (s *Server) onConn(conn transport.Conn) {
	defer conn.Close()
	for {
		stream, err := conn.AcceptStream()
		if err != nil {
			logs.Error("accept %s stream fail: %v", conn.RemoteAddr().String(), err)
			break
		}
		go s.handleStream(stream)
	}
}

func (s *Server) handleStream(stream transport.Stream) {
	lenbuf := make([]byte, 2)
	_, err := stream.Read(lenbuf)
	if err != nil {
		logs.Error("read stream len fail: %v", err)
		stream.Close()
		return
	}

	bodylen := binary.BigEndian.Uint16(lenbuf)
	buf := make([]byte, bodylen)
	nr, err := io.ReadFull(stream, buf)
	if err != nil {
		logs.Error("read proxy protocol fail: %v", err)
		stream.Close()
		return
	}

	proxyProtocol := proto.ProxyProtocol{}
	err = json.Unmarshal(buf[:nr], &proxyProtocol)
	if err != nil {
		logs.Error("unmarshal proxy protocol fail: %v", err)
		return
	}

	switch proxyProtocol.Protocol {
	case "tcp":
		s.tcpProxy(stream, &proxyProtocol)
	case "udp":
		s.udpProxy(stream, &proxyProtocol)
	}
}

func (s *Server) tcpProxy(stream transport.Stream, p *proto.ProxyProtocol) {
	addr := fmt.Sprintf("%s:%s", p.DstIP, p.DstPort)
	remoteConn, err := net.DialTimeout("tcp", addr, time.Second*10)
	if err != nil {
		logs.Error("dial server %s fail: %v", addr, err)
		stream.Close()
		return
	}

	go func() {
		defer remoteConn.Close()
		defer stream.Close()
		obj := s.tcpPool.Get()
		defer s.tcpPool.Put(obj)
		buf := obj.([]byte)
		io.CopyBuffer(remoteConn, stream, buf)
	}()

	defer remoteConn.Close()
	defer stream.Close()
	obj := s.tcpPool.Get()
	defer s.tcpPool.Put(obj)
	buf := obj.([]byte)
	io.CopyBuffer(stream, remoteConn, buf)
}

func (s *Server) udpProxy(stream transport.Stream, p *proto.ProxyProtocol) {
	addr := net.JoinHostPort(p.DstPort, p.DstPort)
	raddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		logs.Error("resolve %s fail: %v", addr, err)
		stream.Close()
		return
	}

	remoteConn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		logs.Error("dial %s udp fail: %v", raddr, err)
		return
	}
	defer remoteConn.Close()

	go func() {
		defer remoteConn.Close()
		defer stream.Close()
		hdr := make([]byte, 2)
		for {
			_, err := io.ReadFull(stream, hdr)
			if err != nil {
				logs.Error("read stream fail: %v", err)
				break
			}
			nlen := binary.BigEndian.Uint16(hdr)
			buf := make([]byte, nlen)
			_, err = io.ReadFull(stream, buf)
			if err != nil {
				logs.Error("read stream body fail: %v", err)
				break
			}

			remoteConn.Write(buf)
		}
	}()

	defer remoteConn.Close()
	defer stream.Close()
	obj := s.udpPool.Get()
	defer s.udpPool.Put(obj)
	buf := obj.([]byte)
	for {
		remoteConn.SetReadDeadline(time.Now().Add(time.Minute * 1))
		nr, err := remoteConn.Read(buf)
		remoteConn.SetReadDeadline(time.Time{})
		if err != nil {
			logs.Error("read from remote %s fail: %v", remoteConn.RemoteAddr().String(), err)
			break
		}

		bytes := proto.EncodeData(buf[:nr])
		stream.SetWriteDeadline(time.Now().Add(time.Second * 10))
		nw, err := stream.Write(bytes)
		stream.SetWriteDeadline(time.Time{})
		if err != nil {
			logs.Error("stream write fail: %v", err)
			break
		}

		if nw != len(bytes) {
			logs.Error("bug: stream write %d bytes, expected %d", nw, len(bytes))
		}
	}
}
