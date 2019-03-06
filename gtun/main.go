package gtun

import (
	"fmt"
	"os"
)

func Main() {
	opts, err := ParseArgs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse args fail: %v\n", err)
		return
	}

	cliConfig := &ClientConfig{
		serverAddr: opts.srv,
		authKey:    opts.authkey,
	}
	NewClient(cliConfig).Run(opts)
}
