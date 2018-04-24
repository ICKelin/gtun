// +build linux darwin

package main

import (
	"github.com/songgao/water"
)

func NewIfce() (*water.Interface, error) {
	cfg := water.Config{
		DeviceType: water.TUN,
	}

	ifce, err := water.New(cfg)
	if err != nil {
		return nil, err
	}

	return ifce, err
}
