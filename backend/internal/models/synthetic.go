package models

import "time"

// Synthetic Monitoring Models (like New Relic Synthetics)

type SyntheticMonitor struct {
	ID          int64     `gorm:"primaryKey" json:"id"`
	TenantID    int64     `gorm:"index" json:"tenant_id"`
	Name        string    `gorm:"size:256;not null" json:"name"`
	Type        string    `gorm:"size:32;not null" json:"type"` // ping, simple_browser, scripted_browser, api_test
	Status      string    `gorm:"size:32;default:'enabled'" json:"status"`
	Frequency   int       `gorm:"default:300" json:"frequency"` // seconds
	Locations   string    `gorm:"type:text" json:"locations"` // JSON array of locations
	Config      string    `gorm:"type:text" json:"config"` // JSON configuration
	Script      string    `gorm:"type:text" json:"script"` // For scripted monitors
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type SyntheticResult struct {
	ID          int64     `gorm:"primaryKey" json:"id"`
	TenantID    int64     `gorm:"index" json:"tenant_id"`
	MonitorID   int64     `gorm:"index" json:"monitor_id"`
	Location    string    `gorm:"size:64" json:"location"`
	Success     bool      `json:"success"`
	Duration    float64   `json:"duration"` // milliseconds
	StatusCode  int       `json:"status_code"`
	Error       string    `gorm:"size:1024" json:"error"`
	Screenshot  string    `gorm:"size:512" json:"screenshot"` // URL to screenshot
	Metrics     string    `gorm:"type:text" json:"metrics"` // JSON metrics
	Timestamp   time.Time `gorm:"index" json:"timestamp"`
}

type SyntheticAlert struct {
	ID          int64     `gorm:"primaryKey" json:"id"`
	TenantID    int64     `gorm:"index" json:"tenant_id"`
	MonitorID   int64     `gorm:"index" json:"monitor_id"`
	Type        string    `gorm:"size:64" json:"type"` // failure, slow_response, error_rate
	Threshold   float64   `json:"threshold"`
	Duration    int       `json:"duration"` // minutes
	Enabled     bool      `gorm:"default:true" json:"enabled"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// Browser Monitoring Models

type BrowserSession struct {
	ID            int64     `gorm:"primaryKey" json:"id"`
	TenantID      int64     `gorm:"index" json:"tenant_id"`
	ApplicationID int64     `gorm:"index" json:"application_id"`
	SessionID     string    `gorm:"size:64;unique;index" json:"session_id"`
	UserID        string    `gorm:"size:128;index" json:"user_id"`
	UserAgent     string    `gorm:"size:512" json:"user_agent"`
	Browser       string    `gorm:"size:64" json:"browser"`
	BrowserVersion string   `gorm:"size:32" json:"browser_version"`
	OS            string    `gorm:"size:64" json:"os"`
	Device        string    `gorm:"size:64" json:"device"`
	Country       string    `gorm:"size:64" json:"country"`
	Region        string    `gorm:"size:64" json:"region"`
	City          string    `gorm:"size:64" json:"city"`
	StartTime     time.Time `json:"start_time"`
	EndTime       *time.Time `json:"end_time"`
	Duration      float64   `json:"duration"`
	PageViews     int       `json:"page_views"`
	Bounced       bool      `json:"bounced"`
}

type PageView struct {
	ID            int64     `gorm:"primaryKey" json:"id"`
	TenantID      int64     `gorm:"index" json:"tenant_id"`
	ApplicationID int64     `gorm:"index" json:"application_id"`
	SessionID     string    `gorm:"size:64;index" json:"session_id"`
	PageID        string    `gorm:"size:64" json:"page_id"`
	URL           string    `gorm:"size:1024" json:"url"`
	Title         string    `gorm:"size:512" json:"title"`
	Referrer      string    `gorm:"size:1024" json:"referrer"`
	LoadTime      float64   `json:"load_time"`
	DOMReady      float64   `json:"dom_ready"`
	FirstPaint    float64   `json:"first_paint"`
	FirstContentfulPaint float64 `json:"first_contentful_paint"`
	LargestContentfulPaint float64 `json:"largest_contentful_paint"`
	CumulativeLayoutShift float64 `json:"cumulative_layout_shift"`
	FirstInputDelay float64 `json:"first_input_delay"`
	TimeToInteractive float64 `json:"time_to_interactive"`
	Timestamp     time.Time `gorm:"index" json:"timestamp"`
}

type JavaScriptError struct {
	ID            int64     `gorm:"primaryKey" json:"id"`
	TenantID      int64     `gorm:"index" json:"tenant_id"`
	ApplicationID int64     `gorm:"index" json:"application_id"`
	SessionID     string    `gorm:"size:64;index" json:"session_id"`
	PageID        string    `gorm:"size:64" json:"page_id"`
	ErrorMessage  string    `gorm:"size:1024" json:"error_message"`
	ErrorType     string    `gorm:"size:128" json:"error_type"`
	StackTrace    string    `gorm:"type:text" json:"stack_trace"`
	FileName      string    `gorm:"size:512" json:"file_name"`
	LineNumber    int       `json:"line_number"`
	ColumnNumber  int       `json:"column_number"`
	UserAgent     string    `gorm:"size:512" json:"user_agent"`
	URL           string    `gorm:"size:1024" json:"url"`
	Timestamp     time.Time `gorm:"index" json:"timestamp"`
}

type AjaxRequest struct {
	ID            int64     `gorm:"primaryKey" json:"id"`
	TenantID      int64     `gorm:"index" json:"tenant_id"`
	ApplicationID int64     `gorm:"index" json:"application_id"`
	SessionID     string    `gorm:"size:64;index" json:"session_id"`
	PageID        string    `gorm:"size:64" json:"page_id"`
	URL           string    `gorm:"size:1024" json:"url"`
	Method        string    `gorm:"size:16" json:"method"`
	StatusCode    int       `json:"status_code"`
	Duration      float64   `json:"duration"`
	RequestSize   int64     `json:"request_size"`
	ResponseSize  int64     `json:"response_size"`
	Timestamp     time.Time `gorm:"index" json:"timestamp"`
}