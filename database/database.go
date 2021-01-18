package database

import (
	"context"
	"log"
	"time"

	"github.com/xngln/photo-server/graph/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DB struct {
	client *mongo.Client
}

func Connect() *DB {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	return &DB{
		client: client,
	}
}

func (db *DB) Disconnect() {
	if db.client == nil {
		return
	}

	err := db.client.Disconnect(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
}

func (db *DB) InsertImage(input *model.NewImage) *model.ImageDB {
	collection := db.client.Database("store").Collection("images")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	res, err := collection.InsertOne(ctx, input)
	if err != nil {
		log.Fatal(err)
	}
	id := res.InsertedID.(primitive.ObjectID).Hex()
	return &model.ImageDB{
		ID:    id,
		Name:  input.Name,
		Price: input.Price,
	}
}

//func (db *DB) FindImageByID(ID string) *model.ImageDB {

//}

//func (db *DB) AllImages(ID string) []*model.ImageDB {
//}

//func (db *DB) DeleteImageByID(ID string) bool {

//}
