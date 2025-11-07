package services

import (
	"context"
	"strings"
	"time"

	"github.com/sakkurohilla/kineticops/backend/internal/models"
	"github.com/sakkurohilla/kineticops/backend/internal/repository/mongodb"
)

func ParseAndEnrichLog(log *models.Log) {
	if log.Message != "" {
		if strings.Contains(log.Message, "correl_id:") {
			parts := strings.Split(log.Message, "correl_id:")
			if len(parts) > 1 {
				log.CorrelID = strings.TrimSpace(parts[1])
			}
		}
	}
	log.FullText = log.Message + " " + formatMeta(log.Meta)
}

func formatMeta(meta map[string]string) string {
	var out []string
	for k, v := range meta {
		out = append(out, k+"="+v)
	}
	return strings.Join(out, " ")
}

func CollectLog(ctx context.Context, log *models.Log) error {
	ParseAndEnrichLog(log)
	// Store logs in MongoDB by default (ES integration removed)
	return mongodb.InsertLog(ctx, log)
}

func SearchLogs(ctx context.Context, tenantID int64, filters map[string]interface{}, text string, limit int, skip int) ([]models.Log, error) {
	// Use MongoDB for searches
	return mongodb.SearchLogs(ctx, tenantID, filters, text, limit, skip)
}

// CountLogs returns the total number of logs matching the query
func CountLogs(ctx context.Context, tenantID int64, filters map[string]interface{}, text string) (int64, error) {
	return mongodb.CountLogs(ctx, tenantID, filters, text)
}

func EnforceLogRetention(days int) error {
	cutoff := time.Now().AddDate(0, 0, -days)
	// Enforce retention in MongoDB
	return mongodb.DeleteOldLogs(context.Background(), cutoff)
}
