package models

import (
	"time"

	"github.com/ICKelin/gtun/common"
	"github.com/ICKelin/gtun/registry/database"
	"gopkg.in/mgo.v2/bson"
)

type Gtund struct {
	database.Model `bson:",inline"`
	PublicIP       string `json:"public_ip" bson:"public_ip"`
	Port           int    `json:"listen_port" bson:"listen_port"`
	CIDR           string `json:"cidr" bson:"cidr"`
	Token          string `json:"token" bson:"token"`
	Count          int    `json:"count" bson:"count"`
	MaxClientCount int    `json:"max_client_count" bson:"max_client_count"`
	IsWindows      bool   `json:"is_windows" bson:"is_windows"`
	Name           string `json:"name" bson:"name"`
}

type GtundManager struct {
	CollectionName string
	database.ModelManager
}

func GetGtundManager() *GtundManager {
	return &GtundManager{
		CollectionName: "register_gtund",
	}
}

func (m *GtundManager) NewGtund(regInfo *common.S2GRegister) (*Gtund, error) {
	gtund := &Gtund{}
	gtund.Id = bson.NewObjectId()
	gtund.CreatedAt = time.Now().Unix()
	gtund.UpdatedAt = gtund.CreatedAt
	gtund.PublicIP = regInfo.PublicIP
	gtund.Port = regInfo.Port
	gtund.CIDR = regInfo.CIDR
	gtund.Token = regInfo.Token
	gtund.Count = regInfo.Count
	gtund.MaxClientCount = regInfo.MaxClientCount
	gtund.IsWindows = regInfo.IsWindows
	gtund.Name = regInfo.Name
	err := m.Insert(m.CollectionName, gtund)
	if err != nil {
		return nil, err
	}
	return gtund, nil
}

func (m *GtundManager) RemoveGtund(id bson.ObjectId) error {
	return m.RemoveWithId(m.CollectionName, id)
}

func (m *GtundManager) IncReferenceCount(count int) (gtundInfo *Gtund, err error) {
	return nil, nil
}

func (m *GtundManager) GetAvailableGtund(isWindows bool) (*Gtund, error) {
	return nil, nil
}

func (m *GtundManager) GtundList() ([]*Gtund, error) {
	return nil, nil
}
