package mongodb

import (
	"fmt"

	"github.com/rs/xid"
	"go.mongodb.org/mongo-driver/mongo"

	mongodriver "github.com/jqrd/gorestapi-mongo/store/driver/mongo"
)

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
		return nil, fmt.Errorf("could not create mongodb client: %w", err)
	}

	return &Client{
		db: db,
		newID: func() string {
			return xid.New().String()
		},
	}, nil

}
