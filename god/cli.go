package god

import (
	"flag"
)

type Options struct {
	confPath string
}

func ParseArgs() (*Options, error) {
	flag.Usage = func() {
		flag.PrintDefaults()
	}

	pconfig := flag.String("c", "", "config path")
	flag.Parse()

	opts := &Options{
		confPath: *pconfig,
	}

	return opts, nil
}
