package database

import (
	"gopkg.in/mgo.v2/bson"
)

type Model struct {
	Id        bson.ObjectId `json:"id" bson:"_id,omitempty"`
	CreatedAt int64         `json:"created_at" bson:"created_at"`
	UpdatedAt int64         `json:"updated_at" bson:"updated_at"`
	Invalid   bool          `json:"invalid" bson:"invalid"`
	InvalidAt int64         `json:"invalid_at" bson:"invalid_at"`
}
