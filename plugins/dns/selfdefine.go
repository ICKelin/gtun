package dns

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"github.com/ICKelin/glog"
	"github.com/ICKelin/gone/cache"
	"github.com/miekg/dns"
)

type SelfDefine struct {
	rulepath     string
	buildinpath  string
	customCache  *cache.Cache
	buildInCache *cache.Cache
}

func NewSelfDefine(rule, buildin string) *SelfDefine {
	return &SelfDefine{
		rulepath:     rule,
		buildinpath:  buildin,
		customCache:  cache.NewCache(cache.ALGO_TRIE),
		buildInCache: cache.NewCache(cache.ALGO_HASH),
	}
}

func (this *SelfDefine) Run() {
	this.LoadDNSRules(this.rulepath)
	this.LoadBuildInDomain(this.buildinpath)
}

// 适配业务的域名
type CustomDomain struct {
	Domain      string   `json:"domain"`
	Set4        []string `json:"set4"`
	Set6        []string `json:"set6"`
	UpperServer string   `json:"upper"`
	IP          []string `json:"ip"`
}

// 自定义解析的域名
type BuildInDomain struct {
	Domain string   `json:"domain"`
	IP     []string `json:"ip"`
}

func (this *SelfDefine) LoadDNSRules(dir string) error {
	files, err := GetDirFiles(dir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if strings.Contains(file, ".conf") {
			this.loadDNSRules(file)
		}
	}

	return err
}

func (this *SelfDefine) loadDNSRules(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	rx := make(map[string]*CustomDomain)

	reader := bufio.NewReader(f)

	for {
		bline, _, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}

		line := string(bline)

		if len(line) < 7 {
			continue
		}

		txt := strings.Split(line, "/")
		if len(txt) != 3 {
			continue
		}

		domain := "*." + txt[1]

		if _, ok := rx[domain]; !ok {
			rx[domain] = &CustomDomain{}
		}

		rx[domain].Domain = domain

		if txt[0] == "server=" {
			rx[domain].UpperServer = strings.Split(txt[2], "#")[0]
		} else if txt[0] == "ipset=" {
			sp := strings.Split(txt[2], ",")
			for _, s := range sp {
				rx[domain].Set4 = append(rx[domain].Set4, s)
			}
		} else if txt[0] == "ipset6=" {
			sp := strings.Split(txt[2], ",")
			for _, s := range sp {
				rx[domain].Set6 = append(rx[domain].Set6, s)
			}
		}

		this.customCache.Add(domain, rx[domain])
		glog.INFO(fmt.Sprintf("%s ==> %s", domain, rx[domain].UpperServer))
	}

	return nil
}

func (this *SelfDefine) LoadBuildInDomain(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	reader := bufio.NewReader(f)
	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			break
		}

		split := strings.Split(string(line), "/")
		if len(split) != 3 {
			continue
		}

		domain := strings.ToLower(split[1])
		ip := []string{split[2]}
		buildInDomain := &BuildInDomain{domain, ip}
		this.buildInCache.Add(domain, buildInDomain)
	}
	return nil
}

func (this *SelfDefine) Save() error {
	buildInDomain := this.buildInCache.String()
	fp, err := os.Open(this.buildinpath)
	if err != nil {
		return err
	}

	defer fp.Close()

	fp.Write([]byte(buildInDomain))
	return nil
}

func (this *SelfDefine) GetBuildInDomain(query *dns.Msg) (*dns.Msg, error) {
	for _, q := range query.Question {
		ele, err := this.buildInCache.Get(strings.ToLower(strings.TrimRight(q.Name, ".")))
		if err != nil {
			continue
		}

		if domainInfo, ok := ele.(*BuildInDomain); ok {
			response := query.Copy()
			response.Answer = append(response.Answer, &dns.A{
				Hdr: dns.RR_Header{
					Name:     q.Name,
					Rrtype:   q.Qtype,
					Class:    q.Qclass,
					Ttl:      500,
					Rdlength: 4,
				},
				A: net.ParseIP(domainInfo.IP[0]).To4(),
			})
			return response, nil
		}

	}
	return nil, fmt.Errorf("not build in domain")
}

func (this *SelfDefine) GetCustomUpper(msg *dns.Msg) ([]string, []string, []string) {
	for _, q := range msg.Question {
		customDomain, err := this.getCustomUpper(strings.TrimRight(q.Name, "."), dns.TypeToString[q.Qtype])
		if err != nil {
			continue
		}

		return []string{customDomain.UpperServer}, customDomain.Set4, customDomain.Set6
	}
	return nil, nil, nil
}

func (this *SelfDefine) getCustomUpper(domain, qtype string) (*CustomDomain, error) {
	record, err := this.customCache.Get(domain)
	if err != nil {
		return nil, err
	}

	if cd, ok := record.(*CustomDomain); ok {
		return cd, nil
	}

	return nil, fmt.Errorf("not custom domain")
}

func GetDirFiles(dir string) ([]string, error) {
	if !dirExist(dir) {
		return nil, fmt.Errorf("build in dns rules not exist")
	}

	fp, _ := os.Open(dir)
	fileInfos, err := fp.Readdir(0)
	if err != nil {
		return nil, err
	}

	defer fp.Close()

	var files = make([]string, 0)
	for _, file := range fileInfos {
		if !file.IsDir() {
			files = append(files, dir+"/"+file.Name())
		}
	}

	return files, nil
}

func dirExist(file string) bool {
	fileInfo, err := os.Stat(file)
	if err != nil {
		return false
	}

	if !fileInfo.IsDir() {
		return false
	}

	return true
}
