/*

MIT License

Copyright (c) 2018 ICKelin

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

*/

/*
	DESCRIPTION:
				This program is a gtun client for game/ip accelator.

	Author: ICKelin
*/

package main

import (
	"fmt"
	"os"

	"github.com/ICKelin/glog"
)

func main() {
	opts, err := ParseArgs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse args fail: %v\n", err)
		return
	}

	if opts.debug {
		glog.Init("gtun", glog.PRIORITY_DEBUG, "./", glog.OPT_DATE, 1024*10)
	} else {
		glog.Init("gtun", glog.PRIORITY_INFO, "./", glog.OPT_DATE, 1024*10)
	}

	cliConfig := &ClientConfig{
		serverAddr: opts.srv,
		authKey:    opts.authkey,
	}
	NewClient(cliConfig).Run(opts)
}
