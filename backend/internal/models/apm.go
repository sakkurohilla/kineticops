package models

import "time"

// APM (Application Performance Monitoring) Models

type Application struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	TenantID  int64     `gorm:"index" json:"tenant_id"`
	HostID    int64     `gorm:"index" json:"host_id"`
	Name      string    `gorm:"size:128;not null" json:"name"`
	Type      string    `gorm:"size:64" json:"type"` // web, api, worker, etc.
	Language  string    `gorm:"size:32" json:"language"`
	Framework string    `gorm:"size:64" json:"framework"`
	Version   string    `gorm:"size:32" json:"version"`
	Status    string    `gorm:"size:32;default:'unknown'" json:"status"`
	LastSeen  time.Time `json:"last_seen"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

type Transaction struct {
	ID            int64     `gorm:"primaryKey" json:"id"`
	TenantID      int64     `gorm:"index" json:"tenant_id"`
	ApplicationID int64     `gorm:"index" json:"application_id"`
	TraceID       string    `gorm:"size:64;index" json:"trace_id"`
	Name          string    `gorm:"size:256" json:"name"`
	Type          string    `gorm:"size:64" json:"type"` // web, background
	Duration      float64   `json:"duration"`            // milliseconds
	ResponseTime  float64   `json:"response_time"`
	Throughput    float64   `json:"throughput"`
	ErrorRate     float64   `json:"error_rate"`
	Apdex         float64   `json:"apdex"`
	StatusCode    int       `json:"status_code"`
	Method        string    `gorm:"size:16" json:"method"`
	URI           string    `gorm:"size:512" json:"uri"`
	UserAgent     string    `gorm:"size:512" json:"user_agent"`
	RemoteIP      string    `gorm:"size:45" json:"remote_ip"`
	Timestamp     time.Time `gorm:"index" json:"timestamp"`
}

type Trace struct {
	ID            int64     `gorm:"primaryKey" json:"id"`
	TenantID      int64     `gorm:"index" json:"tenant_id"`
	TraceID       string    `gorm:"size:64;unique;index" json:"trace_id"`
	ApplicationID int64     `gorm:"index" json:"application_id"`
	RootSpanID    string    `gorm:"size:64" json:"root_span_id"`
	Duration      float64   `json:"duration"`
	SpanCount     int       `json:"span_count"`
	ErrorCount    int       `json:"error_count"`
	Timestamp     time.Time `gorm:"index" json:"timestamp"`
}

type Span struct {
	ID            int64     `gorm:"primaryKey" json:"id"`
	TenantID      int64     `gorm:"index" json:"tenant_id"`
	TraceID       string    `gorm:"size:64;index" json:"trace_id"`
	SpanID        string    `gorm:"size:64;unique;index" json:"span_id"`
	ParentSpanID  string    `gorm:"size:64;index" json:"parent_span_id"`
	ApplicationID int64     `gorm:"index" json:"application_id"`
	OperationName string    `gorm:"size:256" json:"operation_name"`
	ServiceName   string    `gorm:"size:128" json:"service_name"`
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	Duration      float64   `json:"duration"`
	Tags          string    `gorm:"type:text" json:"tags"` // JSON
	Logs          string    `gorm:"type:text" json:"logs"` // JSON
	Status        string    `gorm:"size:32" json:"status"`
	Error         bool      `json:"error"`
	ErrorMessage  string    `gorm:"size:1024" json:"error_message"`
}

type ErrorEvent struct {
	ID            int64     `gorm:"primaryKey" json:"id"`
	TenantID      int64     `gorm:"index" json:"tenant_id"`
	ApplicationID int64     `gorm:"index" json:"application_id"`
	TraceID       string    `gorm:"size:64;index" json:"trace_id"`
	SpanID        string    `gorm:"size:64;index" json:"span_id"`
	ErrorClass    string    `gorm:"size:256" json:"error_class"`
	ErrorMessage  string    `gorm:"size:1024" json:"error_message"`
	StackTrace    string    `gorm:"type:text" json:"stack_trace"`
	Fingerprint   string    `gorm:"size:64;index" json:"fingerprint"`
	Count         int       `gorm:"default:1" json:"count"`
	FirstSeen     time.Time `json:"first_seen"`
	LastSeen      time.Time `json:"last_seen"`
	Resolved      bool      `gorm:"default:false" json:"resolved"`
	Timestamp     time.Time `gorm:"index" json:"timestamp"`
}

type DatabaseQuery struct {
	ID            int64     `gorm:"primaryKey" json:"id"`
	TenantID      int64     `gorm:"index" json:"tenant_id"`
	ApplicationID int64     `gorm:"index" json:"application_id"`
	TraceID       string    `gorm:"size:64;index" json:"trace_id"`
	SpanID        string    `gorm:"size:64;index" json:"span_id"`
	Database      string    `gorm:"size:128" json:"database"`
	Operation     string    `gorm:"size:64" json:"operation"` // SELECT, INSERT, UPDATE, DELETE
	Table         string    `gorm:"size:128" json:"table"`
	Query         string    `gorm:"type:text" json:"query"`
	Duration      float64   `json:"duration"`
	RowsAffected  int64     `json:"rows_affected"`
	Timestamp     time.Time `gorm:"index" json:"timestamp"`
}

type ExternalService struct {
	ID            int64     `gorm:"primaryKey" json:"id"`
	TenantID      int64     `gorm:"index" json:"tenant_id"`
	ApplicationID int64     `gorm:"index" json:"application_id"`
	TraceID       string    `gorm:"size:64;index" json:"trace_id"`
	SpanID        string    `gorm:"size:64;index" json:"span_id"`
	ServiceName   string    `gorm:"size:128" json:"service_name"`
	URL           string    `gorm:"size:512" json:"url"`
	Method        string    `gorm:"size:16" json:"method"`
	StatusCode    int       `json:"status_code"`
	Duration      float64   `json:"duration"`
	ResponseSize  int64     `json:"response_size"`
	Timestamp     time.Time `gorm:"index" json:"timestamp"`
}

// Performance Metrics
type PerformanceMetric struct {
	ID            int64     `gorm:"primaryKey" json:"id"`
	TenantID      int64     `gorm:"index" json:"tenant_id"`
	ApplicationID int64     `gorm:"index" json:"application_id"`
	MetricName    string    `gorm:"size:128;index" json:"metric_name"`
	MetricType    string    `gorm:"size:32" json:"metric_type"` // counter, gauge, histogram
	Value         float64   `json:"value"`
	Tags          string    `gorm:"type:text" json:"tags"` // JSON
	Timestamp     time.Time `gorm:"index" json:"timestamp"`
}

// Custom Events
type CustomEvent struct {
	ID            int64     `gorm:"primaryKey" json:"id"`
	TenantID      int64     `gorm:"index" json:"tenant_id"`
	ApplicationID int64     `gorm:"index" json:"application_id"`
	EventType     string    `gorm:"size:128;index" json:"event_type"`
	EventName     string    `gorm:"size:256" json:"event_name"`
	Attributes    string    `gorm:"type:text" json:"attributes"` // JSON
	Timestamp     time.Time `gorm:"index" json:"timestamp"`
}
