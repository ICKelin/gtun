package proxy

import (
	"fmt"
	"github.com/ICKelin/gtun/internal/logs"
	"github.com/ICKelin/gtun/internal/utils"
	"strings"
	"sync/atomic"
)

var (
	markID       = int32(1)
	routeTableID = int32(101)
)

func allocateMarkID() int32 {
	return atomic.AddInt32(&markID, 1)
}

func allocateRouteTableID() int32 {
	return atomic.AddInt32(&routeTableID, 1)
}

func initRedirect(proto, region, redirectPort string) {
	setName := ipsetNamePrefix + region

	out, err := utils.ExecCmd("ipset", []string{"create", setName, "hash:net"})
	if err != nil {
		logs.Warn("create ipset fail: %v %s", err, out)
	}

	markID := allocateMarkID()
	routeTable := allocateRouteTableID()

	args := fmt.Sprintf("-t mangle -I PREROUTING -p %s -m set --match-set %s dst -j TPROXY --tproxy-mark %d/%d --on-port %s", proto, setName, markID, markID, redirectPort)
	out, err = utils.ExecCmd("iptables", strings.Split(args, " "))
	if err != nil {
		logs.Warn("%s %s %s", args, err, out)
	}

	args = fmt.Sprintf("-t mangle -I OUTPUT -p %s -m set --match-set %s dst -j MARK --set-mark %d", proto, setName, markID)
	out, err = utils.ExecCmd("iptables", strings.Split(args, " "))
	if err != nil {
		logs.Warn("%s %s %s", args, err, out)
	}

	args = fmt.Sprintf("rule add fwmark %d lookup %d", markID, routeTable)
	out, err = utils.ExecCmd("ip", strings.Split(args, " "))
	if err != nil {
		logs.Warn("%s %s %s", args, err, out)
	}

	args = fmt.Sprintf("ro add local default dev lo table %d", routeTable)
	out, err = utils.ExecCmd("ip", strings.Split(args, " "))
	if err != nil {
		logs.Warn("%s %s %s", args, err, out)
	}
}
