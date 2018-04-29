// +build windows

package main

import (
	"github.com/songgao/water"
)

func NewIfce(tap bool) (*water.Interface, error) {
	cfg := water.Config{
		DeviceType: water.TAP,
	}

	cfg.ComponentID = "tap0901"

	ifce, err := water.New(cfg)
	if err != nil {
		return nil, err
	}

	return ifce, err
}
