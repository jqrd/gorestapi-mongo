package db

import "github.com/jqrd/gorestapi-mongo/store/mongodb"

// TODO is there a better way to do this with generics? at least to share impl.

func (obj *Widget) ID() string {
	return obj.Id
}

func (obj *Widget) SetID(id string) mongodb.MongoDocument {
	if obj == nil {
		obj = &Widget{}
	}
	obj.Id = id
	return obj
}

func (obj *Thing) ID() string {
	return obj.Id
}

func (obj *Thing) SetID(id string) mongodb.MongoDocument {
	if obj == nil {
		obj = &Thing{}
	}
	obj.Id = id
	return obj
}
