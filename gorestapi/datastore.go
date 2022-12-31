package gorestapi

import (
	"github.com/jqrd/gorestapi-mongo/model/db"
	"github.com/jqrd/gorestapi-mongo/store/mongodb"
)

type DataStore interface {
	Widgets() mongodb.MongoCollection[*db.Widget]
	Things() mongodb.MongoCollection[*db.Thing]
}

type dataStore struct {
	widgets mongodb.MongoCollection[*db.Widget]
	things  mongodb.MongoCollection[*db.Thing]
}

func (s *dataStore) Widgets() mongodb.MongoCollection[*db.Widget] {
	return s.widgets
}

func (s *dataStore) Things() mongodb.MongoCollection[*db.Thing] {
	return s.things
}

func NewDataStore(
	widgets mongodb.MongoCollection[*db.Widget],
	things mongodb.MongoCollection[*db.Thing],
) DataStore {
	return &dataStore{
		widgets: widgets,
		things:  things,
	}
}
