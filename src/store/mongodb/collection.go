package mongodb

import (
	"context"
	"fmt"
	"log"

	"github.com/jqrd/gorestapi-mongo/store"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoCollection[T MongoDocument] interface {
	InsertOne(ctx context.Context, obj T, options ...*options.InsertOneOptions) error
	FindOne(ctx context.Context, id string, options ...*options.FindOptions) (T, error)
	UpdateOne(ctx context.Context, obj T, options ...*options.UpdateOptions) error
	DeleteOne(ctx context.Context, id string, options ...*options.DeleteOptions) error

	InsertMany(ctx context.Context, obj []T, options ...*options.InsertManyOptions) error
	Find(ctx context.Context, filter bson.M, options ...*options.FindOptions) ([]T, error)
	DeleteMany(ctx context.Context, filter bson.M, options ...*options.DeleteOptions) error
}

type mongoCollection[T MongoDocument] struct {
	col *mongo.Collection
}

func Collection[T MongoDocument](c *Client, collectionName string) MongoCollection[T] {
	return &mongoCollection[T]{c.db.Collection(collectionName)}
}

func (c *mongoCollection[T]) InsertOne(ctx context.Context, obj T, options ...*options.InsertOneOptions) error {
	result, err := c.col.InsertOne(ctx, obj, options...)
	if err != nil {
		return err
	}

	id := result.InsertedID.(primitive.ObjectID)
	obj.SetId(id.Hex())

	return nil
}

func (c *mongoCollection[T]) InsertMany(ctx context.Context, obj []T, options ...*options.InsertManyOptions) error {
	nonGeneric := make([]interface{}, len(obj))
	for i, o := range obj {
		nonGeneric[i] = o
	}
	result, err := c.col.InsertMany(ctx, nonGeneric, options...)
	if err != nil {
		return err
	}

	for i, insertedId := range result.InsertedIDs {
		id := insertedId.(primitive.ObjectID)
		obj[i].SetId(id.Hex())
	}

	return nil
}

func (c *mongoCollection[T]) FindOne(ctx context.Context, id string, options ...*options.FindOptions) (T, error) {
	docId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return *new(T), store.ErrNotFound
	}

	result := c.col.FindOne(ctx, bson.M{"_id": docId})
	obj := new(T)
	err = result.Decode(obj)
	return *obj, store.TryTranslateMongoError(err)
}

func (c *mongoCollection[T]) Find(ctx context.Context, filter bson.M, options ...*options.FindOptions) ([]T, error) {
	if filter == nil {
		filter = bson.M{}
	}
	result, err := c.col.Find(ctx, filter, options...)
	if err != nil {
		return nil, store.TryTranslateMongoError(err)
	}

	found := make([]T, 0)
	err = result.All(ctx, &found)
	return found, err
}

func (c *mongoCollection[T]) UpdateOne(ctx context.Context, obj T, options ...*options.UpdateOptions) error {
	docId, err := primitive.ObjectIDFromHex(obj.GetId())
	if err != nil {
		return store.ErrNotFound
	}

	result, err := c.col.UpdateOne(ctx, bson.M{"_id": docId}, obj, options...)
	if err != nil {
		return err
	}

	if result.UpsertedID == nil {
		return store.ErrNotFound
	}

	id := result.UpsertedID.(primitive.ObjectID)
	if id.Hex() != obj.GetId() {
		return store.ErrNotFound
	}

	return nil
}

func (c *mongoCollection[T]) DeleteOne(ctx context.Context, id string, options ...*options.DeleteOptions) error {
	docId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return store.ErrNotFound
	}

	result, err := c.col.DeleteOne(ctx, bson.M{"_id": docId})
	if err != nil {
		return store.TryTranslateMongoError(err)
	}
	if result.DeletedCount > 1 {
		return fmt.Errorf("expected to delete 0 or 1, deleted %v", result.DeletedCount)
	}
	if result.DeletedCount != 1 {
		return store.ErrNotFound
	}
	return nil
}

func (c *mongoCollection[T]) DeleteMany(ctx context.Context, filter bson.M, options ...*options.DeleteOptions) error {
	// TODO
	log.Panic("not implemented")
	return nil
}
