package models

import "time"

type AlertRule struct {
	ID                  int64     `json:"id"`
	TenantID            int64     `json:"tenant_id"`
	MetricName          string    `json:"metric_name"`
	Operator            string    `json:"operator"` // ">", "<", etc.
	Threshold           float64   `json:"threshold"`
	Window              int       `json:"window"`    // minutes
	Frequency           int       `json:"frequency"` // # of breaches to trigger
	NotificationWebhook string    `json:"notification_webhook"`
	EscalationPolicy    string    `json:"escalation_policy"` // store as JSON string
	CreatedAt           time.Time `json:"created_at"`
}
