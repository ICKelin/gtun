package dns

import (
	"bufio"
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"

	"github.com/pelletier/go-toml"
)

type Config struct {
	Upper       []string `json:"upper"`           // 上游dns
	Custom      bool     `json:"custom"`          // 自定义模式
	Cache       bool     `json:"cache"`           // 缓存开关
	Workers     int      `json:"worker"`          // 工作线程数
	BufferSize  int      `json:"buffer_size"`     // 缓冲队列大小
	Ipset       bool     `json:"ipset"`           // ipset开关
	RulesDir    string   `json:"rulesDir"`        // 配置规则路径
	ResolveFile string   `json:"resolve_file"`    // 上游dns配置文件，兼容dnsmasq
	BuildIn     bool     `json:"build_in_domain"` // 开启内置域名映射功能
	AccessToken string   `json:"access_token"`    // 访问授权码
}

var (
	gConfig *Config
)

func LoadConfig(fpath string) (*Config, error) {
	cnt, err := ioutil.ReadFile(fpath)
	if err != nil {
		return nil, err
	}

	cnt = removeComment(cnt)

	var conf = &Config{}
	json.Unmarshal(cnt, conf)

	upper, err := LoadResolveFile(conf.ResolveFile)
	if err == nil {
		conf.Upper = upper
	}

	gConfig = conf

	return conf, err
}

// 兼容dnsmasq，从指定路径加载上游dns地址
func LoadResolveFile(resolve string) ([]string, error) {
	file, err := os.Open(resolve)
	if err != nil {
		return nil, err
	}

	upper := make([]string, 0)
	br := bufio.NewReader(file)
	for {
		line, _, err := br.ReadLine()
		if err != nil {
			break
		}

		str := string(line)
		// comment
		if strings.HasPrefix(str, "#") {
			continue
		}

		if strings.HasPrefix(str, "nameserver") {
			sp := strings.Split(str, " ")
			if len(sp) != 2 {
				continue
			}
			ip := sp[1]
			upper = append(upper, ip)
		}
	}
	return upper, err
}

func removeComment(content []byte) []byte {
	bc := make([]byte, 0)
	status := 0
	for _, v := range content {
		if v == '#' {
			status = 1
			continue
		}

		if v == '\n' {
			status = 0
		}

		if status == 1 {
			continue
		}

		bc = append(bc, v)
	}
	return bc
}

func GetConfig() *Config {
	return gConfig
}

func SetConfig(conf *Config) {
	gConfig = conf
}

func (this *Config) IsCacheOn() bool {
	return this.Cache
}

func (this *Config) IsCustomOn() bool {
	return this.Custom
}

func (this *Config) String() string {
	d, err := toml.Marshal(*this)
	if err != nil {
		return err.Error()
	}
	return string(d)
}
