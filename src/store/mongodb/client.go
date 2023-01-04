package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/xid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/jqrd/gorestapi-mongo/pkg/log"
	mongodriver "github.com/jqrd/gorestapi-mongo/store/driver/mongo"
)

const WidgetsCollection string = "widgets"
const ThingsCollection string = "things"

type Config struct {
	mongodriver.Config `conf:",squash"`
}

type Client struct {
	db    *mongo.Database
	newID func() string
}

func New(cfg *Config) (*Client, error) {

	db, err := mongodriver.New(&cfg.Config)
	if err != nil {
		return nil, fmt.Errorf("could not create mongodb client: %v", err.Error())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"name": cfg.Database}
	dbs, err := db.Client().ListDatabaseNames(ctx, filter)
	if err != nil {
		return nil, err
	}

	if len(dbs) == 0 {
		log.Warn("Database not found, creating ", cfg.Database)
		err := db.CreateCollection(ctx, WidgetsCollection)
		if err != nil {
			return nil, fmt.Errorf("could not create %v collection in the database: %v", WidgetsCollection, err.Error())
		}
	}

	return &Client{
		db: db,
		newID: func() string {
			return xid.New().String()
		},
	}, nil

}
