package gtun

import (
	"io"
	"net"
	"sync"
	"syscall"
	"time"

	"github.com/ICKelin/gtun/internal/logs"
	"github.com/ICKelin/gtun/internal/proto"
)

var (
	// default tcp timeout(read, write), 10 seconds
	defaultTCPTimeout = 10
)

type TCPForward struct {
	region     string
	listenAddr string

	// writeTimeout defines the tcp connection write timeout in second
	// default value set to 10 seconds
	writeTimeout time.Duration

	// readTimeout defines the tcp connection write timeout in second
	// default value set to 10 seconds
	readTimeout time.Duration

	// the session manager is the global session manager
	// it stores opennotr_client to opennotr_server connection
	sessMgr *SessionManager

	mempool sync.Pool
}

func NewTCPForward(region string, cfg TCPForwardConfig) *TCPForward {
	tcpReadTimeout := cfg.ReadTimeout
	if tcpReadTimeout <= 0 {
		tcpReadTimeout = defaultTCPTimeout
	}

	tcpWriteTimeout := cfg.WriteTimeout
	if tcpWriteTimeout <= 0 {
		tcpWriteTimeout = int(defaultTCPTimeout)
	}

	return &TCPForward{
		region:       region,
		listenAddr:   cfg.ListenAddr,
		writeTimeout: time.Duration(tcpWriteTimeout) * time.Second,
		readTimeout:  time.Duration(tcpReadTimeout) * time.Second,
		sessMgr:      GetSessionManager(),
		mempool: sync.Pool{
			New: func() interface{} {
				return make([]byte, 32*1024)
			},
		},
	}
}

func (f *TCPForward) Listen() (net.Listener, error) {
	listener, err := net.Listen("tcp", f.listenAddr)
	if err != nil {
		return nil, err
	}

	// set socket with ip transparent option
	file, err := listener.(*net.TCPListener).File()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	err = syscall.SetsockoptInt(int(file.Fd()), syscall.SOL_IP, syscall.IP_TRANSPARENT, 1)
	if err != nil {
		return nil, err
	}
	return listener, nil
}

func (f *TCPForward) Serve(listener net.Listener) error {
	for {
		conn, err := listener.Accept()
		if err != nil {
			logs.Error("accept fail: %v", err)
			break
		}

		go f.forwardTCP(conn)
	}

	return nil
}

func (f *TCPForward) forwardTCP(conn net.Conn) {
	defer conn.Close()

	dip, dport, _ := net.SplitHostPort(conn.LocalAddr().String())
	sip, sport, _ := net.SplitHostPort(conn.RemoteAddr().String())

	sess := f.sessMgr.GetSession(f.region)
	if sess == nil {
		logs.Error("no route to host: %s", dip)
		return
	}
	logs.Debug("%s:%s=>%s:%s", sip, sport, dip, dport)
	stream, err := sess.conn.OpenStream()
	if err != nil {
		logs.Error("open stream fail: %v", err)
		return
	}
	defer stream.Close()

	bytes := proto.EncodeProxyProtocol("tcp", sip, sport, dip, dport)
	stream.SetWriteDeadline(time.Now().Add(f.writeTimeout))
	_, err = stream.Write(bytes)
	stream.SetWriteDeadline(time.Time{})
	if err != nil {
		logs.Error("stream write fail: %v", err)
		return
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	defer wg.Wait()

	go func() {
		defer wg.Done()
		defer stream.Close()
		defer conn.Close()
		obj := f.mempool.Get()
		buf := obj.([]byte)
		io.CopyBuffer(stream, conn, buf)
	}()

	obj := f.mempool.Get()
	buf := obj.([]byte)
	io.CopyBuffer(conn, stream, buf)
}
