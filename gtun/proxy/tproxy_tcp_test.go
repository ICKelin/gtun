package proxy

import (
	"bytes"
	"net"
	"testing"
	"time"
)

type dummyConn struct {
	readBuf  bytes.Buffer
	writeBuf bytes.Buffer
}

func (d *dummyConn) Read(b []byte) (n int, err error) {
	n = copy(b, d.readBuf.Bytes())
	return n, nil
}

func (d *dummyConn) Write(b []byte) (n int, err error) {
	return d.writeBuf.Write(b)
}

func (d *dummyConn) Close() error {
	d.readBuf.Reset()
	d.writeBuf.Reset()
	return nil
}

func (d *dummyConn) LocalAddr() net.Addr {
	return nil
}

func (d *dummyConn) RemoteAddr() net.Addr {
	return nil
}

func (d *dummyConn) SetDeadline(t time.Time) error {
	return nil
}

func (d *dummyConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (d *dummyConn) SetWriteDeadline(t time.Time) error {
	return nil
}

func TestTProxyTCPDoProxy(t *testing.T) {
	p := NewTProxyTCP()
	cfg := `{}`
	err := p.Setup([]byte(cfg))
	if err != nil {
		t.Error(err)
		return
	}

	conn := &dummyConn{}
	p.(*TProxyTCP).doProxy(conn)
}
