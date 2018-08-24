package gtun

import (
	"fmt"
	"os"

	"github.com/ICKelin/glog"
)

func Main() {
	opts, err := ParseArgs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse args fail: %v\n", err)
		return
	}

	if opts.debug {
		glog.Init("gtun", glog.PRIORITY_DEBUG, "./", glog.OPT_DATE, 1024*10)
	} else {
		glog.Init("gtun", glog.PRIORITY_WARN, "./", glog.OPT_DATE, 1024*10)
	}

	cliConfig := &ClientConfig{
		serverAddr: opts.srv,
		authKey:    opts.authkey,
	}
	NewClient(cliConfig).Run(opts)
}
