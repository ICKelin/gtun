package gtun

import (
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/ICKelin/gtun/common"

	"github.com/ICKelin/gone/net/ghttp"
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
	return &God{
		serverAddr: cfg.GodAddr,
		token:      cfg.GodToken,
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

	fmt.Println(s)
	r := &common.ResponseBody{}
	err = json.Unmarshal([]byte(s), &r)
	if err != nil {
		return "", err
	}

	gtunInfo, ok := r.Data.(*common.G2CRegister)
	if !ok {
		return "", fmt.Errorf("invalid response")
	}
	if gtunInfo.ServerAddress == "" {
		return "", fmt.Errorf("empty server address")
	}
	return gtunInfo.ServerAddress, nil
}
