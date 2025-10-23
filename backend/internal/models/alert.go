package models

import "time"

type Alert struct {
	ID               int64      `json:"id"`
	TenantID         int64      `json:"tenant_id"`
	RuleID           int64      `json:"rule_id"`
	MetricName       string     `json:"metric_name"`
	HostID           int64      `json:"host_id"`
	Value            float64    `json:"value"`
	Status           string     `json:"status"`
	DedupHash        string     `json:"dedup_hash"`
	EscalatedLevel   int        `json:"escalated_level"`
	TriggeredAt      time.Time  `json:"triggered_at"`
	ClosedAt         *time.Time `json:"closed_at,omitempty"`
	NotificationSent bool       `json:"notification_sent"`
}
