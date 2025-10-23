package mongodb

import (
	"context"
	"log"
	"time"

	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client

func Init() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOpts := options.Client().
		ApplyURI(viper.GetString("MONGO_URI")).
		SetMaxPoolSize(20)

	var err error
	Client, err = mongo.Connect(ctx, clientOpts)
	if err != nil {
		log.Printf("MongoDB connection failed: %v", err)
		return err
	}
	if err = Client.Ping(ctx, nil); err != nil {
		log.Printf("MongoDB ping error: %v", err)
		return err
	}
	return nil
}
