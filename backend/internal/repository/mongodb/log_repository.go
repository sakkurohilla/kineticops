package mongodb

import (
	"context"
	"time"

	"github.com/sakkurohilla/kineticops/backend/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func InsertLog(ctx context.Context, log *models.Log) error {
	_, err := models.LogCollection.InsertOne(ctx, log)
	return err
}

func SearchLogs(ctx context.Context, tenantID int64, filters map[string]interface{}, text string, limit int, skip int) ([]models.Log, error) {
	q := bson.M{"tenant_id": tenantID}
	for k, v := range filters {
		q[k] = v
	}
	if text != "" {
		q["$text"] = bson.M{"$search": text}
	}
	opts := options.Find().SetSort(bson.M{"timestamp": -1}).SetLimit(int64(limit)).SetSkip(int64(skip))
	cur, err := models.LogCollection.Find(ctx, q, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var logs []models.Log
	for cur.Next(ctx) {
		var log models.Log
		if err = cur.Decode(&log); err == nil {
			logs = append(logs, log)
		}
	}
	return logs, nil
}

func DeleteOldLogs(ctx context.Context, cutoff time.Time) error {
	_, err := models.LogCollection.DeleteMany(ctx, bson.M{"timestamp": bson.M{"$lt": cutoff}})
	return err
}
