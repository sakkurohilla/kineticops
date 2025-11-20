package models

import "time"

type UserSettings struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	UserID    int64     `json:"user_id" gorm:"column:user_id;uniqueIndex"`
	
	// Account settings
	CompanyName  string `json:"company_name" gorm:"column:company_name"`
	Timezone     string `json:"timezone" gorm:"column:timezone"`
	DateFormat   string `json:"date_format" gorm:"column:date_format"`
	
	// Notification settings
	EmailNotifications  bool   `json:"email_notifications" gorm:"column:email_notifications;default:true"`
	SlackNotifications  bool   `json:"slack_notifications" gorm:"column:slack_notifications;default:false"`
	WebhookNotifications bool  `json:"webhook_notifications" gorm:"column:webhook_notifications;default:false"`
	AlertEmail          string `json:"alert_email" gorm:"column:alert_email"`
	SlackWebhook        string `json:"slack_webhook" gorm:"column:slack_webhook"`
	CustomWebhook       string `json:"custom_webhook" gorm:"column:custom_webhook"`
	
	// Security settings
	RequireMFA      bool `json:"require_mfa" gorm:"column:require_mfa;default:false"`
	SessionTimeout  int  `json:"session_timeout" gorm:"column:session_timeout;default:30"`
	PasswordExpiry  int  `json:"password_expiry" gorm:"column:password_expiry;default:90"`
	
	// Data retention settings
	MetricsRetention int `json:"metrics_retention" gorm:"column:metrics_retention;default:30"`
	LogsRetention    int `json:"logs_retention" gorm:"column:logs_retention;default:7"`
	TracesRetention  int `json:"traces_retention" gorm:"column:traces_retention;default:7"`
	
	// Performance settings
	AutoRefresh        bool `json:"auto_refresh" gorm:"column:auto_refresh;default:true"`
	RefreshInterval    int  `json:"refresh_interval" gorm:"column:refresh_interval;default:30"`
	MaxDashboardWidgets int `json:"max_dashboard_widgets" gorm:"column:max_dashboard_widgets;default:20"`
	
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
