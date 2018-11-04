package kcpserver

import (
	"crypto/sha1"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	_ "net/http/pprof"
	"os"
	"time"

	"golang.org/x/crypto/pbkdf2"

	"path/filepath"

	"github.com/golang/snappy"
	kcp "github.com/xtaci/kcp-go"
	"github.com/xtaci/smux"
)

var (
	// VERSION is injected by buildflags
	VERSION = "SELFBUILD"
	// SALT is use for pbkdf2 key expansion
	SALT = "kcp-go"
)

type compStream struct {
	conn net.Conn
	w    *snappy.Writer
	r    *snappy.Reader
}

func (c *compStream) Read(p []byte) (n int, err error) {
	return c.r.Read(p)
}

func (c *compStream) Write(p []byte) (n int, err error) {
	n, err = c.w.Write(p)
	err = c.w.Flush()
	return n, err
}

func (c *compStream) Close() error {
	return c.conn.Close()
}

func newCompStream(conn net.Conn) *compStream {
	c := new(compStream)
	c.conn = conn
	c.w = snappy.NewBufferedWriter(conn)
	c.r = snappy.NewReader(conn)
	return c
}

// handle multiplex-ed connection
func handleMux(conn io.ReadWriteCloser, config *Config) {
	// stream multiplex
	smuxConfig := smux.DefaultConfig()
	smuxConfig.MaxReceiveBuffer = config.SockBuf
	smuxConfig.KeepAliveInterval = time.Duration(config.KeepAlive) * time.Second

	mux, err := smux.Server(conn, smuxConfig)
	if err != nil {
		log.Println(err)
		return
	}
	defer mux.Close()
	for {
		p1, err := mux.AcceptStream()
		if err != nil {
			log.Println(err)
			return
		}
		p2, err := net.DialTimeout("tcp", config.Target, 5*time.Second)
		if err != nil {
			p1.Close()
			log.Println(err)
			continue
		}
		go handleClient(p1, p2, config.Quiet)
	}
}

func handleClient(p1, p2 io.ReadWriteCloser, quiet bool) {
	if !quiet {
		log.Println("stream opened")
		defer log.Println("stream closed")
	}
	defer p1.Close()
	defer p2.Close()

	// start tunnel
	p1die := make(chan struct{})
	go func() { io.Copy(p1, p2); close(p1die) }()

	p2die := make(chan struct{})
	go func() { io.Copy(p2, p1); close(p2die) }()

	// wait for tunnel termination
	select {
	case <-p1die:
	case <-p2die:
	}
}

func checkError(err error) {
	if err != nil {
		log.Printf("%+v\n", err)
		os.Exit(-1)
	}
}

func KCPServer(config *Config) error {
	rand.Seed(int64(time.Now().Nanosecond()))

	switch config.Mode {
	case "normal":
		config.NoDelay, config.Interval, config.Resend, config.NoCongestion = 0, 40, 2, 1
	case "fast":
		config.NoDelay, config.Interval, config.Resend, config.NoCongestion = 0, 30, 2, 1
	case "fast2":
		config.NoDelay, config.Interval, config.Resend, config.NoCongestion = 1, 20, 2, 1
	case "fast3":
		config.NoDelay, config.Interval, config.Resend, config.NoCongestion = 1, 10, 2, 1
	}

	log.Println("version:", VERSION)
	pass := pbkdf2.Key([]byte(config.Key), []byte(SALT), 4096, 32, sha1.New)
	var block kcp.BlockCrypt
	switch config.Crypt {
	case "sm4":
		block, _ = kcp.NewSM4BlockCrypt(pass[:16])
	case "tea":
		block, _ = kcp.NewTEABlockCrypt(pass[:16])
	case "xor":
		block, _ = kcp.NewSimpleXORBlockCrypt(pass)
	case "none":
		block, _ = kcp.NewNoneBlockCrypt(pass)
	case "aes-128":
		block, _ = kcp.NewAESBlockCrypt(pass[:16])
	case "aes-192":
		block, _ = kcp.NewAESBlockCrypt(pass[:24])
	case "blowfish":
		block, _ = kcp.NewBlowfishBlockCrypt(pass)
	case "twofish":
		block, _ = kcp.NewTwofishBlockCrypt(pass)
	case "cast5":
		block, _ = kcp.NewCast5BlockCrypt(pass[:16])
	case "3des":
		block, _ = kcp.NewTripleDESBlockCrypt(pass[:24])
	case "xtea":
		block, _ = kcp.NewXTEABlockCrypt(pass[:16])
	case "salsa20":
		block, _ = kcp.NewSalsa20BlockCrypt(pass)
	default:
		config.Crypt = "aes"
		block, _ = kcp.NewAESBlockCrypt(pass)
	}

	lis, err := kcp.ListenWithOptions(config.Listen, block, config.DataShard, config.ParityShard)
	if err != nil {
		return err
	}

	log.Println("listening on:", lis.Addr())
	log.Println("target:", config.Target)
	log.Println("encryption:", config.Crypt)
	log.Println("nodelay parameters:", config.NoDelay, config.Interval, config.Resend, config.NoCongestion)
	log.Println("sndwnd:", config.SndWnd, "rcvwnd:", config.RcvWnd)
	log.Println("compression:", !config.NoComp)
	log.Println("mtu:", config.MTU)
	log.Println("datashard:", config.DataShard, "parityshard:", config.ParityShard)
	log.Println("acknodelay:", config.AckNodelay)
	log.Println("dscp:", config.DSCP)
	log.Println("sockbuf:", config.SockBuf)
	log.Println("keepalive:", config.KeepAlive)
	log.Println("snmplog:", config.SnmpLog)
	log.Println("snmpperiod:", config.SnmpPeriod)
	log.Println("pprof:", config.Pprof)
	log.Println("quiet:", config.Quiet)

	if err := lis.SetDSCP(config.DSCP); err != nil {
		log.Println("SetDSCP:", err)
	}
	if err := lis.SetReadBuffer(config.SockBuf); err != nil {
		log.Println("SetReadBuffer:", err)
	}
	if err := lis.SetWriteBuffer(config.SockBuf); err != nil {
		log.Println("SetWriteBuffer:", err)
	}

	go snmpLogger(config.SnmpLog, config.SnmpPeriod)

	for {
		if conn, err := lis.AcceptKCP(); err == nil {
			conn.SetStreamMode(true)
			conn.SetWriteDelay(true)
			conn.SetNoDelay(config.NoDelay, config.Interval, config.Resend, config.NoCongestion)
			conn.SetMtu(config.MTU)
			conn.SetWindowSize(config.SndWnd, config.RcvWnd)
			conn.SetACKNoDelay(config.AckNodelay)

			log.Println("connection:", conn.LocalAddr(), "->", conn.RemoteAddr())
			if config.NoComp {
				go handleMux(conn, config)
			} else {
				go handleMux(newCompStream(conn), config)
			}
		} else {
			log.Printf("%+v", err)
		}
	}
}

func snmpLogger(path string, interval int) {
	if path == "" || interval == 0 {
		return
	}
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			// split path into dirname and filename
			logdir, logfile := filepath.Split(path)
			// only format logfile
			f, err := os.OpenFile(logdir+time.Now().Format(logfile), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
			if err != nil {
				log.Println(err)
				return
			}
			w := csv.NewWriter(f)
			// write header in empty file
			if stat, err := f.Stat(); err == nil && stat.Size() == 0 {
				if err := w.Write(append([]string{"Unix"}, kcp.DefaultSnmp.Header()...)); err != nil {
					log.Println(err)
				}
			}
			if err := w.Write(append([]string{fmt.Sprint(time.Now().Unix())}, kcp.DefaultSnmp.ToSlice()...)); err != nil {
				log.Println(err)
			}
			kcp.DefaultSnmp.Reset()
			w.Flush()
			f.Close()
		}
	}
}

func Main() {
	path := flag.String("c", "", "config file path")
	flag.Parse()

	config, err := ParseConfig(*path)
	if err != nil {
		log.Fatal(err)
	}

	KCPServer(config)
}
