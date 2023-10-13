package utils

import (
	"fmt"
	"github.com/ICKelin/gtun/internal/logs"
	"os/exec"
	"runtime"
	"time"

	"github.com/songgao/water"
)

type Interface struct {
	tun *water.Interface
}

func NewInterface() (*Interface, error) {
	iface := &Interface{}

	ifconfig := water.Config{
		DeviceType: water.TUN,
	}

	for i := 0; i < 10; i++ {
		ifconfig.Name = fmt.Sprintf("gtun.%d", i)

		ifce, err := water.New(ifconfig)
		if err != nil {
			logs.Error("new interface %s fail: %v", ifconfig.Name, err)
			time.Sleep(time.Second * 1)
			continue
		}

		iface.tun = ifce
		return iface, nil
	}
	return nil, fmt.Errorf("new interface %s fail", ifconfig.Name)
}

func (iface *Interface) SetMTU(mtu int) error {
	out, err := ExecCmd("ifconfig", []string{iface.tun.Name(), "mtu", fmt.Sprintf("%d", mtu)})
	if err != nil {
		return fmt.Errorf("set mtu fail: %s %v", out, err)
	}
	return nil
}

func (iface *Interface) Up() error {
	switch runtime.GOOS {
	case "linux":
		out, err := ExecCmd("ifconfig", []string{iface.tun.Name(), "up"})
		if err != nil {
			return fmt.Errorf("ifconfig fail: %s %v", out, err)
		}

	default:
		return fmt.Errorf("unsupported: %s %s", runtime.GOOS, runtime.GOARCH)

	}

	return nil
}

func (iface *Interface) Read() ([]byte, error) {
	buf := make([]byte, 2048)
	n, err := iface.tun.Read(buf)
	if err != nil {
		return nil, err
	}

	return buf[:n], nil
}

func (iface *Interface) Write(buf []byte) (int, error) {
	return iface.tun.Write(buf)
}

func (iface *Interface) Close() {
	iface.tun.Close()
}

func ExecCmd(cmd string, args []string) (string, error) {
	b, err := exec.Command(cmd, args...).CombinedOutput()
	return string(b), err
}
