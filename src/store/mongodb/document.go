package mongodb

type MongoDocument interface {
	GetId() string
	SetID(id string) MongoDocument
}
