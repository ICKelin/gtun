package gtun

import (
	"fmt"
	"io"
	"net"
	"sync"
	"syscall"
	"time"

	"github.com/ICKelin/gtun/gtun/forward"

	"github.com/ICKelin/gtun/internal/logs"
	"github.com/ICKelin/gtun/internal/proto"
)

func init() {
	forward.RegisterForward("tcp", NewTCPForward)
}

var (
	// default tcp timeout(read, write), 10 seconds
	defaultTCPTimeout = 10
)

type TCPForwardConfig struct {
	ReadTimeout  int    `json:"readTimeout"`
	WriteTimeout int    `json:"writeTimeout"`
	RateLimit    int    `json:"rateLimit"`
	Region       string `json:"region"`
}

type TCPForward struct {
	region     string
	listenAddr string

	// writeTimeout defines the tcp connection write timeout in second
	// default value set to 10 seconds
	writeTimeout time.Duration

	// readTimeout defines the tcp connection write timeout in second
	// default value set to 10 seconds
	readTimeout time.Duration

	mempool sync.Pool

	ratelimit *RateLimit
}

func NewTCPForward() *TCPForward {
	return &TCPForward{}
}

func Setup(cfg TCPForwardConfig) error {
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
		ratelimit: ratelimit,
	}
}

func (f *TCPForward) ListenAndServe(listener net.Listener) error {
	listener, err := net.Listen("tcp", f.listenAddr)
	if err != nil {
		return err
	}

	// set socket with ip transparent option
	file, err := listener.(*net.TCPListener).File()
	if err != nil {
		return err
	}
	defer file.Close()

	err = syscall.SetsockoptInt(int(file.Fd()), syscall.SOL_IP, syscall.IP_TRANSPARENT, 1)
	if err != nil {
		return nil, err
	}
	return listener, nil

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

	sess := f.sessMgr.GetSession(f.region, dip)
	if sess == nil {
		logs.Error("no route to host: %s", dip)
		f.forwardDirect(conn, dip, dport)
		return
	}
	logs.Debug("%s:%s=>%s:%s", sip, sport, dip, dport)
	stream, err := sess.conn.OpenStream()
	if err != nil {
		logs.Error("open stream fail: %v", err)
		f.forwardDirect(conn, dip, dport)
		return
	}
	defer stream.Close()

	bytes := proto.EncodeProxyProtocol("tcp", sip, sport, dip, dport)
	stream.SetWriteDeadline(time.Now().Add(f.writeTimeout))
	_, err = stream.Write(bytes)
	stream.SetWriteDeadline(time.Time{})
	if err != nil {
		logs.Error("stream write fail: %v", err)
		f.forwardDirect(conn, dip, dport)
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
		defer f.mempool.Put(obj)
		buf := obj.([]byte)
		io.CopyBuffer(stream, conn, buf)
	}()

	obj := f.mempool.Get()
	defer f.mempool.Put(obj)
	buf := obj.([]byte)
	io.CopyBuffer(conn, stream, buf)
}

func (f *TCPForward) forwardDirect(conn net.Conn, dip, dport string) {
	rconn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", dip, dport))
	if err != nil {
		logs.Error("forward direct fail: %v", err)
		return
	}
	defer rconn.Close()

	wg := &sync.WaitGroup{}
	wg.Add(1)
	defer wg.Wait()

	go func() {
		defer wg.Done()
		defer rconn.Close()
		defer conn.Close()
		obj := f.mempool.Get()
		defer f.mempool.Put(obj)
		buf := obj.([]byte)
		copyBuffer(rconn, conn, buf, f.ratelimit)
	}()

	obj := f.mempool.Get()
	defer f.mempool.Put(obj)
	buf := obj.([]byte)
	copyBuffer(conn, rconn, buf, f.ratelimit)
}

func copyBuffer(dst io.Writer, src io.Reader, buf []byte, limiter *RateLimit) (written int64, err error) {
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			limiter.LimitRate(int64((nr)))
			nw, ew := dst.Write(buf[0:nr])
			if nw < 0 || nr < nw {
				nw = 0
				if ew == nil {
					ew = fmt.Errorf("invalid write")
				}
			}
			written += int64(nw)
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}
	return written, err
}
