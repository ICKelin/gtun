package database

import (
	"github.com/ICKelin/gtun/registry/config"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var dbMongoDB *MongoDB

type ModelManager struct {
}

func (this *ModelManager) GetMongoDB() *MongoDB {
	if dbMongoDB == nil {
		dbMongoDB = NewMongoDB(config.GetConfig().DBConfig.Url, config.GetConfig().DBConfig.DBName)
	}
	return dbMongoDB
}

func (this *ModelManager) Session(callback func(s *mgo.Session)) {
	var s = this.GetMongoDB().NewSession()
	defer s.Close()
	callback(s)
}

func (this *ModelManager) Insert(cName string, model interface{}) (err error) {
	this.C(cName).Insert(model)
	return err
}

func (this *ModelManager) Update(cName string, selector bson.M, update interface{}) (err error) {
	this.C(cName).Update(selector, update)
	return err
}

func (this *ModelManager) UpdateAll(cName string, selector bson.M, update interface{}) (info *mgo.ChangeInfo, err error) {
	this.C(cName).UpdateAll(selector, update)
	return info, err
}

func (this *ModelManager) UpdateWithId(cName string, id bson.ObjectId, update interface{}) (err error) {
	this.C(cName).UpdateId(id, update)
	return err
}

func (this *ModelManager) FindOne(cName string, query bson.M, result interface{}) (err error) {
	this.C(cName).Find(query).One(result)
	return err
}

func (this *ModelManager) FindAll(cName string, query bson.M, results interface{}) (err error) {
	this.C(cName).Find(query).All(results)
	return err
}

func (this *ModelManager) Remove(cName string, selector bson.M) (err error) {
	this.C(cName).Remove(selector)
	return err
}

func (this *ModelManager) RemoveWithId(cName string, id bson.ObjectId) (err error) {
	this.C(cName).RemoveId(id)
	return err
}

func (this *ModelManager) Count(cName string, selector bson.M) (count int, err error) {
	this.C(cName).Find(selector).Count()
	return count, err
}

func (this *ModelManager) DropCollection(cName string) (err error) {
	err = this.C(cName).DropCollection()
	return err
}

func (this *ModelManager) C(name string) *mgo.Collection {
	return this.GetMongoDB().C(name)
}
