package dns

import (
	"fmt"
	"net"
	"os/exec"
	"strings"

	"github.com/ICKelin/glog"
	"github.com/ICKelin/gone/algo"
	"github.com/miekg/dns"
)

type Job struct {
	dnsData    []byte
	clientConn *net.UDPConn
	clientAddr *net.UDPAddr
}

type Worker struct {
	size       int
	selfDefine *SelfDefine
	resolver   *Resolver
	joblist    *algo.BufferQueue
}

func NewWorker(wSize, jSize int, selfDefine *SelfDefine, resolver *Resolver) *Worker {
	return &Worker{
		size:       wSize,
		selfDefine: selfDefine,
		resolver:   resolver,
		joblist:    algo.NewBufferQueue(jSize),
	}
}

func (this *Worker) Size() int {
	return this.size
}

func (this *Worker) String() string {
	return fmt.Sprintf("worker size %d, jobsize %d", this.size, this.joblist.Size())
}

func (this *Worker) Run() {
	glog.INFO("Worker Run", this.String())
	for x := 0; x < this.size; x++ {
		go this.HandleJob()
	}
}

func (this *Worker) FinishJob(pkt *Job, msg *dns.Msg) error {
	msg.Response = true
	msg.Rcode = dns.RcodeSuccess
	msg.RecursionAvailable = true

	var err error
	if pkt.dnsData, err = msg.Pack(); err != nil {
		return err
	}

	pkt.clientConn.WriteToUDP(pkt.dnsData, pkt.clientAddr)
	return nil
}

func (this *Worker) HandleJob() {
	for {
		ele := this.joblist.Pop()
		dnsQuery, ok := ele.(*Job)
		if !ok {
			glog.ERROR("queue element type assert fail")
			break
		}

		msg, err := Decode(dnsQuery.dnsData)
		if err != nil {
			glog.ERROR(err)
			return
		}

		// 内置域名解析
		if GetConfig().BuildIn {
			response, err := this.selfDefine.GetBuildInDomain(msg)
			if err == nil {
				this.FinishJob(dnsQuery, response)
				return
			}
		}

		servers, set4, set6 := make([]string, 0), make([]string, 0), make([]string, 0)
		if GetConfig().IsCustomOn() {
			servers, set4, set6 = this.selfDefine.GetCustomUpper(msg)
			if len(servers) == 0 {
				servers = GetConfig().Upper
			}
		}

		for _, srv := range servers {
			response, err := this.resolver.Resolve(msg, srv)
			if err != nil {
				glog.ERROR(err)
				continue
			}

			if HasAAAARecord(response) && set6 != nil && len(set6) != 0 {
				AddToSet(set6, response)
			}
			if HasARecord(response) && set4 != nil && len(set4) != 0 {
				AddToSet(set4, response)
			}

			this.FinishJob(dnsQuery, response)

			// 是否开启审计功能
			if GetConfig().Audit {
				audit.Add(GetDNSQuestions(response), strings.Split(dnsQuery.clientAddr.String(), ":")[0])
			}
			break
		}
	}
}

func (this *Worker) Notify(conn *net.UDPConn, data []byte, raddr *net.UDPAddr) {
	this.joblist.Push(&Job{
		clientConn: conn,
		dnsData:    data,
		clientAddr: raddr,
	})
}

func AddToSet(setnames []string, msg *dns.Msg) {
	ips, _ := FindAnswers(msg)
	for _, ip := range ips {
		for _, s := range setnames {
			go addToSet(s, ip)
		}
	}
}

func addToSet(setname string, ip string) {
	runShell(fmt.Sprintf("ipset add %s %s", setname, ip))
}

func runShell(exeStr string) (string, error) {
	cmd := exec.Command("/bin/sh", "-c", exeStr)
	bytes, err := cmd.CombinedOutput()
	return string(bytes), err
}
