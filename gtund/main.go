package gtund

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/ICKelin/glog"
)

func Main() {
	opts, err := ParseArgs()
	if err != nil {
		fmt.Printf("parse args fail: %v", err)
		return
	}

	if opts.debug {
		glog.Init("gtund", glog.PRIORITY_DEBUG, "./", glog.OPT_DATE, 1024*10)
	} else {
		glog.Init("gtund", glog.PRIORITY_WARN, "./", glog.OPT_DATE, 1024*10)
	}

	if opts.confpath != "" {
		_, err = ParseConfig(opts.confpath)
		if err != nil {
			glog.FATAL("parse config file fail:", opts.confpath, err)
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
		glog.ERROR(err)
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
