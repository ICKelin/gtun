// +build linux darwin

package main

import (
	"github.com/songgao/water"
)

func NewIfce(tap bool) (*water.Interface, error) {
	cfg := water.Config{}
	if tap {
		cfg.DeviceType = water.TAP
	} else {
		cfg.DeviceType = water.TUN
	}
	ifce, err := water.New(cfg)
	if err != nil {
		return nil, err
	}

	return ifce, err
}
