package gtun

import (
	"flag"
	"fmt"
	"os"
)

var usage = `Usage: %s [OPTIONS]
OPTIONS:
`

var usage1 = `
Examples:
	./gtun -s 12.13.14.15:443 -key "auth key" -debug true
	./gtun -s 12.13.14.15:443 -key "auth key" -debug true -tap true
`

type Options struct {
	srv     string
	authkey string
	debug   bool
	tap     bool
}

func ParseArgs() (*Options, error) {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, usage, os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, usage1)
	}

	psrv := flag.String(
		"s",
		"",
		"gtun server address")

	pkey := flag.String(
		"key",
		"",
		"auth key with gtun server")

	pdebug := flag.Bool(
		"debug",
		false,
		"debug mode")

	ptap := flag.Bool(
		"tap",
		false,
		"tap mode, tap mode for layer2 tunnel, default is false")

	flag.Parse()

	opt := &Options{
		srv:     *psrv,
		authkey: *pkey,
		debug:   *pdebug,
		tap:     *ptap,
	}

	return opt, nil
}
