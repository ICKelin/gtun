package gtun

import (
	"encoding/json"
	"errors"
	"fmt"
	"runtime"

	"github.com/ICKelin/gtun/common"

	"github.com/ICKelin/gone/net/ghttp"
)

var (
	defaultGod   = "http://127.0.0.1:2002" // just hard code, rewrite with your god address
	defaultToken = "gtun-cg-token"
)

type GodConfig struct {
	GodAddr  string `json:"god_addr"`
	GodToken string `json:"token"`
	Must     bool   `json:"must"`
}

type God struct {
	serverAddr string
	token      string
	must       bool
}

func NewGod(cfg *GodConfig) *God {
	server := cfg.GodAddr
	if server == "" {
		server = defaultGod
	}

	token := cfg.GodToken
	if token == "" {
		token = defaultToken
	}

	return &God{
		serverAddr: server,
		token:      token,
		must:       cfg.Must,
	}
}

func (g *God) Access() (string, error) {
	url := fmt.Sprintf("%s/gtun/access", g.serverAddr)
	body := &common.C2GRegister{
		IsWindows: runtime.GOOS == "windows",
		AuthToken: g.token,
	}
	s, err := ghttp.PostJson(url, body, nil)
	if err != nil {
		return "", err
	}

	r := &common.ResponseBody{}
	err = json.Unmarshal([]byte(s), &r)
	if err != nil {
		return "", err
	}

	if r.Code != common.CODE_SUCCESS {
		return "", errors.New(r.Message)
	}

	bytes, err := json.Marshal(r.Data)
	if err != nil {
		return "", err
	}

	var gtunInfo common.G2CRegister
	err = json.Unmarshal(bytes, &gtunInfo)
	if err != nil {
		return "", err
	}

	if gtunInfo.ServerAddress == "" {
		return "", fmt.Errorf("empty server address")
	}
	return gtunInfo.ServerAddress, nil
}
