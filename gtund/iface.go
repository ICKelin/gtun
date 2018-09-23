package gtund

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/ICKelin/glog"
	"github.com/songgao/water"
)

var (
	defaultMask = "255.255.255.0"
)

type InterfaceConfig struct {
	name   string
	ip     string
	mask   string
	gw     string
	tapDev bool
}

type Interface struct {
	*water.Interface
	name string
	ip   string
	mask string
	gw   string
}

func NewInterface(cfg *InterfaceConfig) (*Interface, error) {
	iface := &Interface{
		name: cfg.name,
		ip:   cfg.ip,
		gw:   cfg.gw,
	}

	ifconfig := water.Config{}

	if cfg.tapDev {
		ifconfig.DeviceType = water.TAP
	} else {
		ifconfig.DeviceType = water.TUN
	}

	ifce, err := water.New(ifconfig)
	if err != nil {
		return nil, err
	}

	err = setupDevice(ifce.Name(), cfg.gw)
	if err != nil {
		return nil, err
	}

	iface.Interface = ifce
	iface.name = ifce.Name()
	iface.mask = defaultMask
	return iface, nil
}

func setupDevice(dev, ip string) (err error) {
	type CMD struct {
		cmd  string
		args []string
	}

	cmdlist := make([]*CMD, 0)

	switch runtime.GOOS {
	case "linux":
		cmdlist = append(cmdlist, &CMD{cmd: "ifconfig", args: []string{dev, "up"}})

		args := strings.Split(fmt.Sprintf("addr add %s/24 dev %s", ip, dev), " ")
		cmdlist = append(cmdlist, &CMD{cmd: "ip", args: args})

	case "darwin":
		cmdlist = append(cmdlist, &CMD{cmd: "ifconfig", args: []string{dev, "up"}})

		args := strings.Split(fmt.Sprintf("%s %s %s", dev, ip, ip), " ")
		cmdlist = append(cmdlist, &CMD{cmd: "ifconfig", args: args})

		args = strings.Split(fmt.Sprintf("add -net %s/24 %s", ip, ip), " ")
		cmdlist = append(cmdlist, &CMD{cmd: "route", args: args})

	default:
		glog.FATAL("unsupported: ", runtime.GOOS, runtime.GOARCH)

	}
	for _, c := range cmdlist {
		output, err := exec.Command(c.cmd, c.args...).CombinedOutput()
		if err != nil {
			return fmt.Errorf("run %s error %s", c, string(output))
		}
	}

	return nil
}
