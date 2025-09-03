package mongo_helper

import (
	"context"
	"fmt"
	"time"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"crypto_rates_auth/internal/config"
)

var dbName = "authDB"

type Client struct {
	Client        *mongo.Client
	UsersCol      *mongo.Collection
	CodesCol      *mongo.Collection
	SecretsCol    *mongo.Collection
	LogsStreamCol *mongo.Collection
	LogsRestCol   *mongo.Collection
}

func NewClient(cfg *config.Config) (*Client, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

	uri := fmt.Sprintf("mongodb://%s:%s@%s/?authSource=authDB", 
		cfg.Mongo.User, cfg.Mongo.Password, cfg.Mongo.Addr,
	)

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("[mongo] failed to connect: %v", err)
	}
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("[mongo] failed to ping: %v", err)
	}

	db := client.Database(dbName)

	return &Client{
		Client:        client,
		UsersCol:      db.Collection("users"),
		CodesCol:      db.Collection("login_codes"),
		SecretsCol:    db.Collection("active_secrets"),
		LogsStreamCol: db.Collection("logs_stream"),
		LogsRestCol:   db.Collection("logs_rest"),
	}, nil
}

func (c *Client) FindUser(ctx context.Context, email string) (_ *User, err error) {
	var user User
	err = c.UsersCol.FindOne(ctx, map[string]string{"email": email}).Decode(&user)
	if err != nil && errors.Is(err, mongo.ErrNoDocuments) {
		err = nil
	}

	return &user, err
}

func (c *Client) InsertUser(ctx context.Context, user User) error {
	_, err := c.UsersCol.InsertOne(ctx, user)

	return err
}

func (c *Client) InsertLoginCode(ctx context.Context, code LoginCode) error {
	_, err := c.CodesCol.InsertOne(ctx, code)

	return err
}

func (c *Client) ValidateAndConsumeCode(ctx context.Context, email, codeHash string) (*LoginCode, error) {
	var lc LoginCode

	err := c.CodesCol.FindOne(ctx, bson.M{
        "email": email, 
        "code_hash": codeHash, 
        "expires_at": bson.M{"$gt": time.Now()},
    }).Decode(&lc)
	if err != nil {
		return nil, err
	}
	_, _ = c.CodesCol.DeleteOne(ctx, bson.M{"_id": lc.ID})

	return &lc, nil
}

func (c *Client) InsertActiveSecret(ctx context.Context, sec ActiveSecret) error {
	_, err := c.SecretsCol.InsertOne(ctx, sec)

	return err
}

func (c *Client) RemoveActiveSecret(ctx context.Context, sec ActiveSecret) error {
	_, err := c.SecretsCol.DeleteOne(ctx, sec)

	return err
}

func (c *Client) InsertStreamLog(ctx context.Context, logEntry LogEntry) error {
	_, err := c.LogsStreamCol.InsertOne(ctx, logEntry)

	return err
}

func (c *Client) InsertRestLog(ctx context.Context, logEntry LogEntry) error {
	_, err := c.LogsRestCol.InsertOne(ctx, logEntry)

	return err
}
