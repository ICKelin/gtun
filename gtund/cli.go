package gtund

import (
	"errors"
	"flag"
)

type Options struct {
	authKey     string
	gateway     string
	listenAddr  string
	routeUrl    string
	nameserver  string
	reverseFile string
	tapMode     bool
	debug       bool
}

func ParseArgs() (*Options, error) {
	flag.Usage = func() {
		flag.PrintDefaults()
	}

	pkey := flag.String(
		"k",
		"",
		"auth key for client connect checking")

	pgateway := flag.String(
		"g",
		"",
		"gateway address, local tun/tap device ip, dhcp pool set to $gateway/24")

	plisten := flag.String(
		"l",
		"",
		"gtun server listen address")

	proute := flag.String(
		"r",
		"",
		"route rules file path, gtun server deploy the file content for gtun client\n"+
			"gtun client insert those ip into route table")

	pnameserver := flag.String(
		"n",
		"",
		"nameserver deploy to gtun client. now it's NOT works")

	preverse := flag.String(
		"p",
		"",
		"reverse proxy policy file path")

	ptap := flag.Bool(
		"t",
		false,
		"use tap device for layer2 forward")

	pdebug := flag.Bool(
		"debug",
		false,
		"debug mode")

	flag.Parse()

	if *pgateway == "" {
		return nil, errors.New("gateway MUST SET")
	}

	opts := &Options{
		authKey:     *pkey,
		gateway:     *pgateway,
		listenAddr:  *plisten,
		routeUrl:    *proute,
		nameserver:  *pnameserver,
		reverseFile: *preverse,
		tapMode:     *ptap,
		debug:       *pdebug,
	}

	return opts, nil
}
