package mongodriver

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Config struct {
	Username string `conf:"username" default:"postgres"`
	Password string `conf:"password" default:"password"`
	Host     string `conf:"host" default:"localhost"`
	Port     string `conf:"port" default:"27017"`
	Database string `conf:"database" default:"mongo"`
	// AutoCreate          bool          `conf:"auto_create" default:"false"`
	// SearchPath          string        `conf:"search_path" default:""`
	// SSLMode             string        `conf:"sslmode" default:"false"`
	// SSLCert             string        `conf:"sslcert" default:""`
	// SSLKey              string        `conf:"sslkey" default:""`
	// SSLRootCert         string        `conf:"sslrootcert" default:""`
	// Retries             int           `conf:"retries" default:"5"`
	// SleepBetweenRetries time.Duration `conf:"sleep_between_retries" default:"7s"`
	// MaxConnections      int           `conf:"max_connections" default:"40"`
	// WipeConfirm         bool          `conf:"wipe_confirm" default:"false"`

	//Logger          Logger
	//MigrationSource source.Driver
}

// New returns a new database client
func New(c *Config) (*mongo.Database, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// TODO proper auth, etc https://pkg.go.dev/go.mongodb.org/mongo-driver/mongo#Connect

	uri := fmt.Sprintf("mongodb://%v:%v", c.Host, c.Port)
	cred := options.Credential{
		AuthMechanism: "SCRAM-SHA-256",
		Username:      c.Username,
		Password:      c.Password,
	}
	options := options.Client().
		ApplyURI(uri).
		SetAuth(cred)
	client, err := mongo.Connect(ctx, options)
	if err != nil {
		return nil, err
	}

	db := client.Database(c.Database)
	return db, nil
}

type KnownError int

const (
	Error_Unknown KnownError = iota
	Error_NoDocumentsFound
)

func GetKnownError(err error) KnownError {
	if strings.Contains(err.Error(), "no documents in result") {
		return Error_NoDocumentsFound
	}

	return Error_Unknown
}
