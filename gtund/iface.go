package gtund

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/songgao/water"
)

type Interface struct {
	*water.Interface
	ip   string
	cidr string
}

func NewInterface(istap bool, ip, cidr string) (*Interface, error) {
	iface := &Interface{
		ip:   ip,
		cidr: cidr,
	}

	ifconfig := water.Config{}

	if istap {
		ifconfig.DeviceType = water.TAP
	} else {
		ifconfig.DeviceType = water.TUN
	}

	ifce, err := water.New(ifconfig)
	if err != nil {
		return nil, err
	}

	err = setupDevice(ifce.Name(), ip, cidr)
	if err != nil {
		return nil, err
	}

	iface.Interface = ifce
	return iface, nil
}

func setupDevice(dev, ip, cidr string) (err error) {
	switch runtime.GOOS {
	case "linux":
		out, err := execCmd("ifconfig", []string{dev, "up"})
		if err != nil {
			return fmt.Errorf("ifconfig fail: %s %v", out, err)
		}

		out, err = execCmd("ip", []string{"addr", "add", cidr, "dev", dev})
		if err != nil {
			return fmt.Errorf("ip add addr fail: %s %v", out, err)
		}

	default:
		return fmt.Errorf("unsupported: %s %s", runtime.GOOS, runtime.GOARCH)

	}

	return nil
}

func execCmd(cmd string, args []string) (string, error) {
	b, err := exec.Command(cmd, args...).CombinedOutput()
	return string(b), err
}
