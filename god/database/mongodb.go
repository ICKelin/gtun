package database

import (
	"github.com/ICKelin/glog"
	"gopkg.in/mgo.v2"
)

type MongoDB struct {
	dbURL   string
	dbName  string
	session *mgo.Session
}

type MongoDBClient struct {
	session *mgo.Session
}

func (this MongoDBClient) Close() {
	this.session.Close()
}

//mongodb://[username:password@]host1[:port1][,host2[:port2],...[,hostN[:portN]]][/[database][?options]]
func NewMongoDB(url, dbName string) *MongoDB {
	var mongoDB *MongoDB = &MongoDB{}
	mongoDB.dbURL = url
	mongoDB.dbName = dbName
	return mongoDB
}

func (this *MongoDB) MainSession() *mgo.Session {
	if this.session == nil {
		var session, err = mgo.Dial(this.dbURL)
		if err != nil {
			glog.FATAL("connect db fail: ", err)
			return nil
		} else {
			this.session = session
		}
	}
	return this.session
}

// 需要调用者自行 Session.Close()
func (this *MongoDB) NewSession() *mgo.Session {
	return this.MainSession().Clone()
}

// 从主 Session 获取一个集合
func (this *MongoDB) C(name string) *mgo.Collection {
	return this.MainSession().DB(this.dbName).C(name)
}
