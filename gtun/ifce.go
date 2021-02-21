// +build linux darwin

package gtun

import (
	"fmt"
	"os/exec"
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

	output, err := exec.Command("ifconfig", []string{ifce.Name(), "up"}...).CombinedOutput()
	if err != nil {
		logs.Error("up inface %s fail: %s %v", ifce.Name(), string(output), err)
		return nil, err
	}

	return ifce, err
}

func SetIfaceIP(ifce *water.Interface, ip string) error {
	args := strings.Split(fmt.Sprintf("addr add %s/24 dev %s", ip, ifce.Name()), " ")
	output, err := exec.Command("ip", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("set ip fail: %s %v", string(output), err)
	}

	return nil
}

func RemoveIfaceIP(ifce *water.Interface, ip string) (err error) {
	args := strings.Split(fmt.Sprintf("addr del %s/24 dev %s", ip, ifce.Name()), " ")
	output, err := exec.Command("ip", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("remove ip fail: %s %v", string(output), err)
	}

	return nil
}
