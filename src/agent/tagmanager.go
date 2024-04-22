package main

import (
	"github.com/gogo/protobuf/proto"
	"io/ioutil"
	"v2ray.com/core/app/router"
)

const (
	tagAll = "all"
)

type TagManager struct {
	tagIPlist   map[string][]string
	tagSitelist map[string][]string
}

func NewTagManager(geoIPFile, siteFile string) (*TagManager, error) {
	t := &TagManager{
		tagSitelist: make(map[string][]string),
		tagIPlist:   make(map[string][]string),
	}
	content, err := ioutil.ReadFile(siteFile)
	if err != nil {
		return nil, err
	}

	protoList := router.GeoSiteList{}
	err = proto.Unmarshal(content, &protoList)
	if err != nil {
		return nil, err
	}
	for _, list := range protoList.Entry {
		for _, domain := range list.Domain {
			t.tagSitelist[list.CountryCode] = append(t.tagSitelist[list.CountryCode], domain.String())
		}
	}

	content, err = ioutil.ReadFile(geoIPFile)
	if err != nil {
		return nil, err
	}

	iplist := router.GeoIPList{}
	err = proto.Unmarshal(content, &iplist)
	if err != nil {
		return nil, err
	}

	for _, entry := range iplist.Entry {
		for _, cidr := range entry.Cidr {
			t.tagIPlist[entry.CountryCode] = append(t.tagIPlist[entry.CountryCode], cidr.String())
		}
	}

	return t, nil
}

func (tm *TagManager) GetTagIPList(tag string) []string {
	return tm.tagIPlist[tag]
}

func (tm *TagManager) GetTagSiteList(tag string) []string {
	return tm.tagSitelist[tag]
}
