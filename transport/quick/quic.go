package quick

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"math/big"

	"github.com/ICKelin/gtun/transport"
	"github.com/lucas-clemente/quic-go"
)

var _ transport.Listener = &Listener{}
var _ transport.Dialer = &Dialer{}
var _ transport.Conn = &Conn{}

type Dialer struct{}

type Listener struct {
	quic.Listener
}

type Conn struct {
	sess quic.Session
}

func (c *Conn) OpenStream() (transport.Stream, error) {
	stream, err := c.sess.OpenStreamSync(context.Background())
	if err != nil {
		return nil, err
	}
	return stream, nil
}

func (c *Conn) AcceptStream() (transport.Stream, error) {
	return c.sess.AcceptStream(context.Background())
}

func (c *Conn) Close() {
	c.sess.CloseWithError(0, "close normal")
}

func (c *Conn) IsClosed() bool {
	return false
}

func (d *Dialer) Dial(remote string) (transport.Conn, error) {
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"quic-echo-example"},
	}

	sess, err := quic.DialAddr(remote, tlsConf, nil)
	if err != nil {
		return nil, err
	}

	return &Conn{sess: sess}, nil
}

func (l *Listener) Accept() (transport.Conn, error) {
	conn, err := l.Listener.Accept(context.Background())
	if err != nil {
		return nil, err
	}

	return &Conn{conn}, nil
}

func (l *Listener) Close() error {
	return l.Listener.Close()
}

func Listen(laddr string) (transport.Listener, error) {
	listener, err := quic.ListenAddr(laddr, generateTLSConfig(), nil)
	if err != nil {
		return nil, err
	}

	return &Listener{Listener: listener}, nil
}

// Setup a bare-bones TLS config for the server
func generateTLSConfig() *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"quic-echo-example"},
	}
}
