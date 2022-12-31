package mongodb

type MongoDocument interface {
	ID() string
	SetID(id string) MongoDocument
}
