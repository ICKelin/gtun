package gtund

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/ICKelin/gtun/logs"
)

func Main() {
	opts, err := ParseArgs()
	if err != nil {
		fmt.Printf("parse args fail: %v", err)
		return
	}

	if opts.confpath != "" {
		_, err = ParseConfig(opts.confpath)
		if err != nil {
			logs.Error("parse config file fail: %s %v", opts.confpath, err)
			return
		}
	}

	serverCfg := &ServerConfig{
		listenAddr:  opts.listenAddr,
		authKey:     opts.authKey,
		gateway:     opts.gateway,
		routeUrl:    opts.routeUrl,
		nameservers: opts.nameserver,
		reverseFile: opts.reverseFile,
	}

	server, err := NewServer(serverCfg)
	if err != nil {
		logs.Error("new server: %v", err)
		return
	}

	server.Run()
	server.Stop()
}

func GetPublicIP() string {
	resp, err := http.Get("http://ipv4.icanhazip.com")
	if err != nil {
		return ""
	}

	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	str := string(content)
	idx := strings.LastIndex(str, "\n")
	return str[:idx]
}
