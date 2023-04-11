package infrastructure

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

type Promotion struct {
	ID             string  `bson:"id"`
	Price          float64 `bson:"price"`
	ExpirationDate string  `bson:"expiration_date"`
}

func InitDataLayer() *mongo.Client {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://mongo2:30002,mongo3:30003/?replicaSet=my-replica-set"))
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println("Connected to Database")
	}

	return client
}

func InitPrimeDataLayer() *mongo.Client {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://mongo1:30001"))
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println("Connected to Database")
	}

	return client
}

func CloseClient(client *mongo.Client) {
	if client == nil {
		return
	}

	err := client.Disconnect(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connection to MongoDB closed.")
}

func AddPromotions(promotions []Promotion, client *mongo.Client) {

	var interfacePromotions []interface{}
	for _, s := range promotions {
		interfacePromotions = append(interfacePromotions, s)
	}

	promotionCollection := client.Database("vervegroup").Collection("promotions")
	_, err := promotionCollection.InsertMany(context.TODO(), interfacePromotions)
	if err != nil {
		log.Fatal(err)
	}
}

func DeletePromotionsCollection(client *mongo.Client) {

	promotionCollection := client.Database("vervegroup").Collection("promotions")
	promotionCollection.Drop(context.TODO())
}

func FindPromotions(id string, client *mongo.Client) []Promotion {

	var result []Promotion

	promotionCollection := client.Database("vervegroup").Collection("promotions")
	promotionCursor, err := promotionCollection.Find(context.TODO(), bson.D{{"id", id}})
	if err != nil {
		log.Fatal(err)
	}

	if err = promotionCursor.All(context.TODO(), &result); err != nil {
		log.Fatal(err)
	}
	return result
}
