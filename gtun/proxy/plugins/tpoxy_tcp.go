package plugins

import (
	"encoding/json"
	"github.com/ICKelin/gtun/gtun/proxy"
	"github.com/ICKelin/gtun/gtun/route"
	"github.com/ICKelin/gtun/internal/logs"
	"github.com/ICKelin/gtun/internal/proto"
	"github.com/ICKelin/gtun/internal/utils"
	"io"
	"net"
	"sync"
	"syscall"
	"time"
)

func init() {
	_ = proxy.Register("tcp", NewTProxyTCP)
}

var (
	// default tcp timeout(read, write), 10 seconds
	defaultTCPTimeout = 10
)

type TProxyTCPConfig struct {
	ReadTimeout  int    `json:"readTimeout"`
	WriteTimeout int    `json:"writeTimeout"`
	ListenAddr   string `json:"listen_addr"`
	RateLimit    int    `json:"rateLimit"`
	Region       string `json:"region"`
}

type TProxyTCP struct {
	region     string
	listenAddr string

	// writeTimeout defines the tcp connection write timeout in second
	// default value set to 10 seconds
	writeTimeout time.Duration

	// readTimeout defines the tcp connection write timeout in second
	// default value set to 10 seconds
	readTimeout time.Duration

	mempool sync.Pool

	ratelimit *utils.RateLimit

	routeManager *route.Manager
}

func NewTProxyTCP() proxy.Proxy {
	return &TProxyTCP{}
}

func (p *TProxyTCP) Name() string {
	return "tproxy_tcp"
}

func (p *TProxyTCP) Setup(cfgContent json.RawMessage) error {
	var cfg = TProxyTCPConfig{}
	err := json.Unmarshal(cfgContent, &cfg)
	if err != nil {
		return nil
	}

	// TODO: verify configuration
	return p.initConfig(cfg)
}

func (p *TProxyTCP) initConfig(cfg TProxyTCPConfig) error {
	tcpReadTimeout := cfg.ReadTimeout
	if tcpReadTimeout <= 0 {
		tcpReadTimeout = defaultTCPTimeout
	}

	tcpWriteTimeout := cfg.WriteTimeout
	if tcpWriteTimeout <= 0 {
		tcpWriteTimeout = defaultTCPTimeout
	}

	rateLimit := utils.NewRateLimit()
	rateLimit.SetRateLimit(int64(cfg.RateLimit * 1024 * 1024))

	p.region = cfg.Region
	p.listenAddr = cfg.ListenAddr
	p.writeTimeout = time.Duration(tcpWriteTimeout) * time.Second
	p.readTimeout = time.Duration(tcpReadTimeout) * time.Second
	p.mempool = sync.Pool{
		New: func() interface{} {
			return make([]byte, 32*1024)
		},
	}
	p.ratelimit = rateLimit
	p.routeManager = route.GetRouteManager()
	return nil
}

func (p *TProxyTCP) ListenAndServe() error {
	listener, err := net.Listen("tcp", p.listenAddr)
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
		return err
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			logs.Error("accept fail: %v", err)
			break
		}

		go p.doProxy(conn)
	}

	return nil
}

func (p *TProxyTCP) doProxy(conn net.Conn) {
	defer conn.Close()

	dip, dport, _ := net.SplitHostPort(conn.LocalAddr().String())
	sip, sport, _ := net.SplitHostPort(conn.RemoteAddr().String())

	sess := p.routeManager.Route(p.region, dip)
	if sess == nil {
		logs.Error("no route to host: %s", dip)
		return
	}
	logs.Debug("%s:%s=>%s:%s", sip, sport, dip, dport)
	stream, err := sess.OpenStream()
	if err != nil {
		logs.Error("open stream fail: %v", err)
		return
	}
	defer stream.Close()

	bytes := proto.EncodeProxyProtocol("tcp", sip, sport, dip, dport)
	_ = stream.SetWriteDeadline(time.Now().Add(p.writeTimeout))
	_, err = stream.Write(bytes)
	_ = stream.SetWriteDeadline(time.Time{})
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
		obj := p.mempool.Get()
		defer p.mempool.Put(obj)
		buf := obj.([]byte)
		io.CopyBuffer(stream, conn, buf)
	}()

	obj := p.mempool.Get()
	defer p.mempool.Put(obj)
	buf := obj.([]byte)
	io.CopyBuffer(conn, stream, buf)
}
