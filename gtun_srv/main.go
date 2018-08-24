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
				This program is a gtun server for game/ip accelator.

	Author: ICKelin
*/

package main

import (
	"fmt"
	"strings"

	"github.com/ICKelin/glog"
	"github.com/ICKelin/gtun/reverse"
)

func main() {
	opts, err := ParseArgs()
	if err != nil {
		fmt.Printf("parse args fail: %v", err)
	}

	if opts.debug {
		glog.Init("gtund", glog.PRIORITY_DEBUG, "./", glog.OPT_DATE, 1024*10)
	} else {
		glog.Init("gtund", glog.PRIORITY_WARN, "./", glog.OPT_DATE, 1024*10)
	}

	if opts.routeFile != "" {
		err := LoadRules(opts.routeFile)
		if err != nil {
			glog.WARM("load rules fail: ", err)
		}
	}

	if opts.nameserver != "" {
		gNameserver = strings.Split(opts.nameserver, ",")
	}

	if opts.gateway == "" {
		glog.ERROR("gateway MUST NOT be empty")
		return
	}

	sp := strings.Split(opts.gateway, ".")
	if len(sp) != 4 {
		glog.ERROR("ip address format fail", opts.gateway)
		return
	}

	prefix := fmt.Sprintf("%s.%s.%s", sp[0], sp[1], sp[2])
	dhcppool = NewDHCPPool(prefix)

	// 2018.05.03
	// Support for reverse proxy
	if opts.reverseFile != "" {
		err := LoadReversePolicy(opts.reverseFile)
		if err != nil {
			glog.WARM("load reverse policy fail:", err)
		} else {
			for _, r := range gReversePolicy {
				go reverse.Proxy(r.Proto, r.From, r.To)
			}
		}
	}

	GtunServe(opts)
}
