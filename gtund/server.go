package gtund

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/ICKelin/gtun/pkg/logs"
	"github.com/ICKelin/gtun/pkg/proto"
	"github.com/hashicorp/yamux"
)

var (
	authFailMsg    = "authorize fail"
	authSuccessMsg = "authorize success"
)

type ServerConfig struct {
	Listen  string `toml:"listen"`
	AuthKey string `toml:"authKey"`
}

type Server struct {
	listenAddr string
	authKey    string
}

func NewServer(cfg ServerConfig) (*Server, error) {
	s := &Server{
		listenAddr: cfg.Listen,
		authKey:    cfg.AuthKey,
	}

	return s, nil
}

func (s *Server) Run() error {
	listener, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		return err
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok {
				if ne.Temporary() {
					continue
				}
			}
			return err
		}

		go s.onConn(conn)
	}
}

func (s *Server) onConn(conn net.Conn) {
	defer conn.Close()
	sess, err := yamux.Server(conn, nil)
	if err != nil {
		logs.Error("create yamux server fail: %v", err)
		return
	}

	for {
		stream, err := sess.AcceptStream()
		if err != nil {
			logs.Error("accept stream fail: %v", err)
			break
		}
		go s.handleStream(stream)
	}
}

func (s *Server) handleStream(stream *yamux.Stream) {
	lenbuf := make([]byte, 2)
	_, err := stream.Read(lenbuf)
	if err != nil {
		log.Println(err)
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

func (s *Server) tcpProxy(stream *yamux.Stream, p *proto.ProxyProtocol) {
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
		io.Copy(remoteConn, stream)
	}()

	go func() {
		defer remoteConn.Close()
		defer stream.Close()
		io.Copy(stream, remoteConn)
	}()
}

func (s *Server) udpProxy(stream *yamux.Stream, p *proto.ProxyProtocol) {
	addr := fmt.Sprintf("%s:%s", p.DstIP, p.DstPort)
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

	go func() {
		defer remoteConn.Close()
		defer stream.Close()
		buf := make([]byte, 64*1024)
		for {
			nr, err := remoteConn.Read(buf)
			if err != nil {
				logs.Error("read from remote fail: %v", err)
				break
			}

			bytes := encode(buf[:nr])
			_, err = stream.Write(bytes)
			if err != nil {
				logs.Error("stream write fail: %v", err)
				break
			}
		}
	}()
}

func encode(raw []byte) []byte {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, uint16(len(raw)))
	buf = append(buf, raw...)
	return buf
}
