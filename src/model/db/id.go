package db

import "github.com/jqrd/gorestapi-mongo/store/mongodb"

// TODO is there a better way to do this with generics? at least to share impl.
// TODO probably will scrap protobuf and go with hand rolled go structs...

func (obj *Widget) SetId(id string) mongodb.MongoDocument {
	if obj == nil {
		obj = &Widget{}
	}
	obj.Id = id
	return obj
}

func (obj *Thing) SetId(id string) mongodb.MongoDocument {
	if obj == nil {
		obj = &Thing{}
	}
	obj.Id = id
	return obj
}
