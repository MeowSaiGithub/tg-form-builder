//go:build mongo || all
// +build mongo all

package mongo

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"go-tg-support-ticket/form"
	"go-tg-support-ticket/internal/store"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"time"
)

func init() {
	store.RegisterAdaptor(&adaptor{})
}

type adaptor struct {
	client *mongo.Client
	coll   *mongo.Collection
}

func (a *adaptor) Open(dns string) error {
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(dns))
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	if err = client.Ping(ctx, readpref.Primary()); err != nil {
		return fmt.Errorf("failed to ping to mongo server: %v", err)
	}
	a.client = client
	a.coll = client.Database("test").Collection("supports")

	return nil
}

func (a *adaptor) GetName() string {
	return "mongo"
}

func (a *adaptor) Migrate(_ *form.Form) error {
	return nil
}

func (a *adaptor) InsertUserInputs(_ string, fields []form.Field) error {

	// Generate a UUID v7 for the _id field
	id, _ := uuid.NewV7()

	// Build the MongoDB document
	doc := bson.M{"_id": id}
	for _, field := range fields {
		if field.DBType != "" {
			doc[field.Name] = field.UserValue
		}
	}

	// Insert the document into the collection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := a.coll.InsertOne(ctx, doc)
	if err != nil {
		return fmt.Errorf("failed to add into database: %w", err)
	}

	return nil
}
