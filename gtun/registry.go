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

type RegistryConfig struct {
	Addr  string `json:"addr"`
	Token string `json:"token"`
	Must  bool   `json:"must"`
}

type Registry struct {
	addr  string
	token string
	must  bool
}

func NewRegistry(cfg *RegistryConfig) *Registry {
	return &Registry{
		addr:  cfg.Addr,
		token: cfg.Token,
		must:  cfg.Must,
	}
}

func (r *Registry) Access() (string, error) {
	url := fmt.Sprintf("%s/gtun/access", r.addr)
	body := &common.C2GRegister{
		IsWindows: runtime.GOOS == "windows",
		AuthToken: r.token,
	}

	s, err := PostJson(url, body, nil)
	if err != nil {
		return "", err
	}

	respbody := &common.ResponseBody{}
	err = json.Unmarshal([]byte(s), &respbody)
	if err != nil {
		return "", err
	}

	if respbody.Code != common.CODE_SUCCESS {
		return "", errors.New(respbody.Message)
	}

	bytes, err := json.Marshal(respbody.Data)
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
