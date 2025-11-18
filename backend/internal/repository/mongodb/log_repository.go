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

// CountLogs returns the total number of log documents matching the filters.
func CountLogs(ctx context.Context, tenantID int64, filters map[string]interface{}, text string) (int64, error) {
	q := bson.M{"tenant_id": tenantID}
	for k, v := range filters {
		q[k] = v
	}
	if text != "" {
		q["$text"] = bson.M{"$search": text}
	}
	return models.LogCollection.CountDocuments(ctx, q)
}

func DeleteOldLogs(ctx context.Context, cutoff time.Time) error {
	_, err := models.LogCollection.DeleteMany(ctx, bson.M{"timestamp": bson.M{"$lt": cutoff}})
	return err
}

// GetLogSources returns distinct sources (from meta.source) and levels for a tenant.
func GetLogSources(ctx context.Context, tenantID int64) ([]string, []string, error) {
	filter := bson.M{"tenant_id": tenantID}

	// Distinct sources stored in meta.source (best-effort)
	sourcesIface, err := models.LogCollection.Distinct(ctx, "meta.source", filter)
	if err != nil {
		return nil, nil, err
	}
	levelsIface, err := models.LogCollection.Distinct(ctx, "level", filter)
	if err != nil {
		return nil, nil, err
	}

	sources := make([]string, 0, len(sourcesIface))
	for _, s := range sourcesIface {
		if ss, ok := s.(string); ok && ss != "" {
			sources = append(sources, ss)
		}
	}
	levels := make([]string, 0, len(levelsIface))
	for _, l := range levelsIface {
		if ll, ok := l.(string); ok && ll != "" {
			levels = append(levels, ll)
		}
	}
	return sources, levels, nil
}

// DistinctValues returns distinct values for the given field restricted by tenant_id
func DistinctValues(ctx context.Context, tenantID int64, field string) ([]string, error) {
	q := bson.M{"tenant_id": tenantID}
	vals, err := models.LogCollection.Distinct(ctx, field, q)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(vals))
	for _, v := range vals {
		if s, ok := v.(string); ok {
			out = append(out, s)
		}
	}
	return out, nil
}
