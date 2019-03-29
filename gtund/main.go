package gtund

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/ICKelin/gtun/logs"
)

var version = ""

func Main() {
	flgVersion := flag.Bool("v", false, "print version")
	flgConf := flag.String("c", "", "config file")
	flag.Parse()

	if *flgVersion {
		fmt.Println(version)
		return
	}

	conf, err := ParseConfig(*flgConf)
	if err != nil {
		logs.Error("parse config file fail: %s %v", *flgConf, err)
		return
	}

	dhcp, err := NewDHCP(conf.DHCPConfig)
	if err != nil {
		logs.Error("init dhcp fail: %v", err)
		return
	}

	iface, err := NewInterface(conf.InterfaceConfig, conf.DHCPConfig.Gateway, conf.DHCPConfig.CIDR)
	if err != nil {
		logs.Error("init interface fail: %v", err)
		return
	}

	var registry *Registry
	if conf.RegistryConfig != nil {
		sp := strings.Split(conf.ServerConfig.Listen, ":")
		if len(sp) != 2 {
			logs.Error("invalid listen address")
			return
		}

		port, _ := strconv.Atoi(sp[1])

		registry = NewRegistry(conf.RegistryConfig, &Service{
			Name:        conf.Name,
			PublicIP:    GetPublicIP(),
			Port:        port,
			CIDR:        conf.DHCPConfig.CIDR,
			ClientLimit: conf.DHCPConfig.ClientCount,
			IsTap:       conf.InterfaceConfig.IsTap,
		})

		go registry.Run()
	}

	if conf.ReverseConfig != nil {
		r := NewReverse(conf.ReverseConfig)
		err := r.Run()
		if err != nil {
			logs.Warn("run reverse fail: %v", err)
		}
	}

	server, err := NewServer(conf.ServerConfig, dhcp, iface, registry)
	if err != nil {
		logs.Error("new server: %v", err)
		return
	}

	server.Run()
	server.Close()
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
