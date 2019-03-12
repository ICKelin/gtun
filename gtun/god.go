package gtun

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime"
	"time"

	"github.com/ICKelin/gtun/common"
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

	s, err := PostJson(url, body, nil)
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

func PostJson(uri string, jsonbody interface{}, header map[string]string) (data string, err error) {
	tr := &http.Transport{
		MaxIdleConns:    10,
		IdleConnTimeout: 30 * time.Second,
	}
	client := &http.Client{Transport: tr}

	params, err := json.Marshal(jsonbody)
	if err != nil {
		return "", err
	}

	request, err := http.NewRequest("POST", uri, bytes.NewReader(params))
	if err != nil {
		return "", err
	}

	request.Header.Set("Content-Type", "application/json")
	for key, value := range header {
		request.Header.Add(key, value)
	}

	resp, err := client.Do(request)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	bdata, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(bdata), nil
}
