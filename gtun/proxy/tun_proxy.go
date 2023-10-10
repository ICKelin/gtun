package proxy

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/ICKelin/gtun/gtun/route"
	"github.com/ICKelin/gtun/internal/logs"
	"github.com/ICKelin/gtun/internal/proto"
	"github.com/ICKelin/gtun/internal/utils"
	"github.com/ICKelin/optw/transport"
	"io"
	"time"
)

func init() {
	Register("tun_proxy", NewTunProxy)
}

type TunProxyConfig struct {
	Region       string `json:"region"`
	MTU          int    `json:"mtu"`
	WriteTimeout int    `json:"write_timeout"`
	ReadTimeout  int    `json:"read_timeout"`
}

type TunProxy struct {
	config TunProxyConfig
	dev    *utils.Interface
}

func NewTunProxy() Proxy {
	return &TunProxy{}
}

func (p *TunProxy) Name() string {
	return "tun_proxy"
}

func (p *TunProxy) Setup(cfg json.RawMessage) error {
	var config = TunProxyConfig{}
	err := json.Unmarshal(cfg, &config)
	if err != nil {
		return err
	}

	if config.MTU <= 0 {
		return fmt.Errorf("%s invalid mtu", p.Name())
	}

	dev, err := utils.NewInterface()
	if err != nil {
		return err
	}
	err = dev.SetMTU(config.MTU)
	if err != nil {
		return err
	}

	p.config = config
	p.dev = dev
	return nil
}

func (p *TunProxy) ListenAndServe() error {
	// tun proxy use only one stream for long live connection
	var nextHopStream transport.Stream
	var nextHopConn *route.HopInfo
	for {
		buf, err := p.dev.Read()
		if err != nil {
			return err
		}

		if nextHopConn == nil || nextHopConn.IsClosed() {
			nextHopConn := route.GetRouteManager().Route(p.config.Region, "")
			if nextHopConn == nil {
				logs.Warn("route to next hop fail")
				continue
			}

			nextHopStream, err = nextHopConn.OpenStream()
			if err != nil {
				logs.Warn("open stream fail: %v", err)
				continue
			}

			// encode proxy protocol
			bytes := proto.EncodeProxyProtocol("tun_proxy", "", "0", "", "0")
			_ = nextHopStream.SetWriteDeadline(time.Now().Add(time.Duration(p.config.WriteTimeout)))
			_, err = nextHopStream.Write(bytes)
			_ = nextHopStream.SetWriteDeadline(time.Time{})

			go p.readFromRemote(nextHopStream)
		}

		bytes := proto.EncodeData(buf)
		nextHopStream.SetWriteDeadline(time.Now().Add(time.Duration(p.config.WriteTimeout)))
		_, err = nextHopStream.Write(bytes)
		nextHopStream.SetWriteDeadline(time.Time{})
		if err != nil {
			nextHopStream.Close()
			nextHopConn.Close()
			logs.Error("stream write fail: %v", err)
		}
	}
}

func (p *TunProxy) readFromRemote(stream transport.Stream) {
	defer stream.Close()
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
		stream.SetReadDeadline(time.Now().Add(time.Duration(p.config.ReadTimeout)))
		_, err = io.ReadFull(stream, buf)
		stream.SetReadDeadline(time.Time{})
		if err != nil {
			logs.Error("read stream body fail: %v", err)
			break
		}

		_, err = p.dev.Write(buf)
		if err != nil {
			logs.Warn("write to dev fail: %v", err)
			return
		}
	}
}
