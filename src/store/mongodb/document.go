package mongodb

type MongoDocument interface {
	GetId() string
	SetId(id string) MongoDocument
}
