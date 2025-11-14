package mongodb

import (
	"context"
	"time"

	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/sakkurohilla/kineticops/backend/internal/logging"
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
		logging.Errorf("MongoDB connection failed: %v", err)
		return err
	}
	if err = Client.Ping(ctx, nil); err != nil {
		logging.Errorf("MongoDB ping error: %v", err)
		return err
	}
	return nil
}
