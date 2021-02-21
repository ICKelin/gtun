// +build linux darwin

package gtun

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/ICKelin/gtun/pkg/logs"
	"github.com/songgao/water"
)

func NewIfce() (*water.Interface, error) {
	cfg := water.Config{}
	cfg.DeviceType = water.TUN
	ifce, err := water.New(cfg)
	if err != nil {
		return nil, err
	}

	return ifce, err
}

type CMD struct {
	cmd  string
	args []string
}

func setupIface(ifce *water.Interface, ip string, gw string) (err error) {
	cmdlist := make([]*CMD, 0)

	switch runtime.GOOS {
	case "linux":
		cmdlist = append(cmdlist, &CMD{cmd: "ifconfig", args: []string{ifce.Name(), "up"}})
		args := strings.Split(fmt.Sprintf("addr add %s/24 dev %s", ip, ifce.Name()), " ")
		cmdlist = append(cmdlist, &CMD{cmd: "ip", args: args})

	case "darwin":
		cmdlist = append(cmdlist, &CMD{cmd: "ifconfig", args: []string{ifce.Name(), "up"}})

		args := strings.Split(fmt.Sprintf("%s %s %s", ifce.Name(), ip, ip), " ")
		cmdlist = append(cmdlist, &CMD{cmd: "ifconfig", args: args})

		args = strings.Split(fmt.Sprintf("add -net %s/24 %s", gw, ip), " ")
		cmdlist = append(cmdlist, &CMD{cmd: "route", args: args})

	case "windows":
		logs.Error("do not support %s", runtime.GOOS)

	}

	logs.Info("set up interface")
	for _, c := range cmdlist {
		logs.Info("%s %s", c.cmd, c.args)
		output, err := exec.Command(c.cmd, c.args...).CombinedOutput()
		if err != nil {
			return fmt.Errorf("run %s error %s", c, string(output))
		}
	}

	return nil
}

func setdownIface(ifce *water.Interface, ip string, gw string) (err error) {
	cmdlist := make([]*CMD, 0)

	switch runtime.GOOS {
	case "linux":
		cmdlist = append(cmdlist, &CMD{cmd: "ifconfig", args: []string{ifce.Name(), "down"}})
		args := strings.Split(fmt.Sprintf("addr del %s/24 dev %s", ip, ifce.Name()), " ")
		cmdlist = append(cmdlist, &CMD{cmd: "ip", args: args})

	case "darwin":
		cmdlist = append(cmdlist, &CMD{cmd: "ifconfig", args: []string{ifce.Name(), "down"}})

		args := strings.Split(fmt.Sprintf("%s %s %s", ifce.Name(), ip, ip), " ")
		cmdlist = append(cmdlist, &CMD{cmd: "ifconfig", args: args})

		args = strings.Split(fmt.Sprintf("del -net %s/24 %s", gw, ip), " ")
		cmdlist = append(cmdlist, &CMD{cmd: "route", args: args})

	case "windows":
		logs.Error("do not support %s", runtime.GOOS)
	}

	logs.Info("set down interface")
	for _, c := range cmdlist {
		logs.Info("%s %s", c.cmd, c.args)
		output, err := exec.Command(c.cmd, c.args...).CombinedOutput()
		if err != nil {
			return fmt.Errorf("run %s error %s", c, string(output))
		}
	}

	return nil
}
