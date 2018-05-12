package dns

import (
	"fmt"
	"net"
	"strings"

	"github.com/ICKelin/glog"
	"github.com/miekg/dns"
)

type DNS struct {
	addr   string
	worker *Worker
}

func NewDNS(addr string, worker *Worker) *DNS {
	return &DNS{
		addr:   addr,
		worker: worker,
	}
}

func (this *DNS) String() string {
	return fmt.Sprintf("addr %s worker size %d", this.addr, this.worker.Size())
}

func (this *DNS) Run() {
	laddr, err := net.ResolveUDPAddr("udp", this.addr)
	if err != nil {
		glog.ERROR(err)
		return
	}

	conn, err := net.ListenUDP("udp", laddr)
	if err != nil {
		glog.ERROR(err)
		return
	}

	glog.INFO("DNS Run", this.String())

	for {
		data := make([]byte, 512)
		n, raddr, err := conn.ReadFromUDP(data)
		if err != nil || raddr == nil {
			continue
		}

		if n > 512 {
			glog.WARM("dns pkt too large", n)
		}

		this.worker.Notify(conn, data[:n], raddr)
	}
}

func Decode(data []byte) (*dns.Msg, error) {
	var dnsMsg = &dns.Msg{}
	if err := dnsMsg.Unpack(data); err != nil {
		return nil, fmt.Errorf("worker, unpack msg, err: %s", err.Error())
	}

	return dnsMsg, nil
}

func GetDNSQuestions(msg *dns.Msg) string {
	x := []string{}
	for i := range msg.Question {
		x = append(x, msg.Question[i].Name+dns.TypeToString[msg.Question[i].Qtype])
	}
	return strings.Join(x, ",")
}

func GetDNSAnswers(msg *dns.Msg) string {
	x := []string{}
	for i := range msg.Answer {
		x = append(x, msg.Answer[i].Header().Name+dns.TypeToString[msg.Answer[i].Header().Rrtype])
	}
	return strings.Join(x, ",")
}

func String(msg *dns.Msg) string {
	return fmt.Sprintf("Q:%s, R:%s", GetDNSQuestions(msg), GetDNSAnswers(msg))
}

func FindAnswers(msg *dns.Msg) ([]string, bool) {
	x := []string{}
	for i := range msg.Answer {
		if msg.Answer[i] == nil {
			continue
		}
		if msg.Answer[i].Header().Rrtype == dns.TypeA {
			if an, ok := msg.Answer[i].(*dns.A); ok {
				ip := an.A.To4()
				sip := fmt.Sprintf("%d.%d.%d.%d", ip[0], ip[1], ip[2], ip[3])
				x = append(x, sip)
			}
		}
		if msg.Answer[i].Header().Rrtype == dns.TypeAAAA {
			if an, ok := msg.Answer[i].(*dns.AAAA); ok {
				ip16 := an.AAAA.To16()
				x = append(x, ip16.String())
			}
		}
	}
	return x, len(x) > 0
}

func HasRecord(msg *dns.Msg, recordType uint16) bool {
	for _, n := range msg.Ns {
		if n.Header().Rrtype == recordType {
			return true
		}
	}
	return false
}
func HasNsRecord(msg *dns.Msg) bool {
	return HasRecord(msg, dns.TypeNS)
}

func HasARecord(msg *dns.Msg) bool {
	return HasRecord(msg, dns.TypeA)
}

func HasCNameRecord(msg *dns.Msg) bool {
	return HasRecord(msg, dns.TypeCNAME)
}

func HasAAAARecord(msg *dns.Msg) bool {
	return HasRecord(msg, dns.TypeAAAA)
}

func HasPTRRecord(msg *dns.Msg) bool {
	return HasRecord(msg, dns.TypePTR)
}

func HandleNSResponse(response *dns.Msg, server net.Conn) (*dns.Msg, error) {
	return nil, fmt.Errorf("unimplement ns record")
}

func HandleCNameResponse(response *dns.Msg, server net.Conn) (*dns.Msg, error) {
	return nil, fmt.Errorf("unimplements cname record")
}
