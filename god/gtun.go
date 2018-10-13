package god

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/ICKelin/glog"
	"github.com/ICKelin/gtun/common"
)

type gtunConfig struct {
	Listener string   `json:"gtun_listener"`
	Tokens   []string `json:"tokens"` // 用户授权码
}

type gtun struct {
	listener string
	tokens   []string
}

func NewGtun(cfg *gtunConfig) *gtun {
	return &gtun{
		listener: cfg.Listener,
		tokens:   cfg.Tokens,
	}
}

func (g *gtun) Run() error {
	http.HandleFunc("/gtun/register", g.onRegister)
	return http.ListenAndServe(g.listener, nil)
}

func (g *gtun) onRegister(w http.ResponseWriter, r *http.Request) {
	content, err := ioutil.ReadAll(r.Body)
	if err != nil {
		glog.ERROR("read body fail: ", err)
		return
	}
	defer r.Body.Close()

	regInfo := &common.C2GRegister{}
	err = json.Unmarshal(content, &regInfo)
	if err != nil {
		common.Response(nil, err)
		return
	}

	if g.checkAuth(regInfo) == false {
		common.Response(nil, errors.New("auth fail"))
		return
	}

	gtundInfo, err := GetDB().GetAvailableGtund(regInfo.IsWindows)
	if err != nil {
		common.Response(nil, err)
		return
	}

	common.Response(gtundInfo, nil)
	glog.INFO("register from ", r.RemoteAddr)
}

func (g *gtun) checkAuth(regInfo *common.C2GRegister) bool {
	// just write for...
	for _, token := range g.tokens {
		if token == regInfo.AuthToken {
			return true
		}
	}
	return false
}
