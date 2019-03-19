package gtund

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"

	"github.com/ICKelin/gtun/logs"
)

var (
	defaultRule = "reverse.policy"
)

type ReverseConfig struct {
	Rule string `toml:"rule"`
}

type Reverse struct {
	ruleFile string
}

func NewReverse(cfg *ReverseConfig) *Reverse {
	rule := cfg.Rule
	if rule == "" {
		rule = defaultRule
	}

	return &Reverse{
		ruleFile: rule,
	}
}

type ReversePolicy struct {
	Proto string `json:"proto"`
	From  string `json:"from"`
	To    string `to:"json"`
}

func (r *Reverse) Run() error {
	policy, err := r.LoadPolicy(r.ruleFile)
	if err != nil {
		return err
	}

	for _, r := range policy {
		go proxy(r.Proto, r.From, r.To)
	}

	return nil
}

func (r *Reverse) LoadPolicy(path string) ([]*ReversePolicy, error) {
	fp, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	policies := make([]*ReversePolicy, 0)

	reader := bufio.NewReader(fp)
	for {
		bline, _, err := reader.ReadLine()
		if err != nil {
			if err != io.EOF {
				return nil, err
			}
			break
		}

		r, err := r.loadline(string(bline))
		if err != nil {
			logs.Warn("%v", err)
			continue
		}

		policies = append(policies, r)
	}

	return policies, nil
}

func (r *Reverse) loadline(line string) (*ReversePolicy, error) {
	if len(line) > 0 && line[0] == '#' {
		return nil, fmt.Errorf("ignore comment")
	}

	sp := strings.Split(line, " ")
	if len(sp) != 2 {
		return nil, fmt.Errorf("invalid line %s", line)
	}

	sp1 := strings.Split(sp[1], "->")
	if len(sp1) != 2 {
		return nil, fmt.Errorf("invalid line %s", line)
	}

	p := &ReversePolicy{
		Proto: sp[0],
		From:  sp1[0],
		To:    sp1[1],
	}

	return p, nil
}

func proxy(prot string, from, to string) {
	if strings.ToLower(prot) == "tcp" {
		proxyTCP(from, to)
	}

	if strings.ToLower(prot) == "udp" {
		proxyUDP(from, to)
	}
}

func proxyTCP(from, to string) {
	listener, err := net.Listen("tcp", from)
	if err != nil {
		logs.Error("listen fail: %v", err)
		return
	}

	logs.Info("proxy pass tcp %s => %s", from, to)
	for {
		conn, err := listener.Accept()
		if err != nil {
			logs.Error("accept: %v", err)
			break
		}

		go reverse("tcp", conn, to)
	}
}

func proxyUDP(from, to string) {
	laddr, err := net.ResolveUDPAddr("udp", from)
	if err != nil {
		logs.Error("resolve udp fail: %v", err)
		return
	}

	lconn, err := net.ListenUDP("udp", laddr)
	if err != nil {
		logs.Error("listen udp fail: %v", err)
		return
	}

	logs.Info("proxy pass udp %s => %s", from, to)
	go reverse("udp", lconn, to)
}

func reverse(proto string, clientconn net.Conn, to string) {
	defer clientconn.Close()
	rconn, err := net.Dial(proto, to)
	if err != nil {
		logs.Error("dial: %v", err)
		return
	}
	defer rconn.Close()

	wg := &sync.WaitGroup{}
	wg.Add(2)

	// client->rconn
	go func() {
		defer wg.Done()
		Copy(rconn, clientconn)
	}()

	// rconn->client
	go func() {
		defer wg.Done()
		Copy(clientconn, rconn)
	}()

	wg.Wait()
}

func Copy(dst, src net.Conn) {
	buffer := make([]byte, 2048)
	for {
		nr, err := src.Read(buffer)
		if err != nil {
			if err != io.EOF {
				logs.Error("read fail: %v", err)
			}
			break
		}

		_, err = dst.Write(buffer[:nr])
		if err != nil {
			logs.Error("write to peer fail: %v", err)
			break
		}
	}

}
