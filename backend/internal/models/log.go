package models

import (
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

var LogCollection *mongo.Collection

type Log struct {
	ID        string            `bson:"_id,omitempty" json:"id"`
	TenantID  int64             `bson:"tenant_id" json:"tenant_id"`
	HostID    int64             `bson:"host_id" json:"host_id"`
	Timestamp time.Time         `bson:"timestamp" json:"timestamp"`
	Level     string            `bson:"level" json:"level"`
	Message   string            `bson:"message" json:"message"`
	Meta      map[string]string `bson:"meta" json:"meta"`
	FullText  string            `bson:"full_text,omitempty" json:"full_text,omitempty"`
	CorrelID  string            `bson:"correlation_id,omitempty" json:"correlation_id,omitempty"`
}
