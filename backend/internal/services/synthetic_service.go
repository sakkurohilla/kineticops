package services

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/sakkurohilla/kineticops/backend/internal/logging"
	"github.com/sakkurohilla/kineticops/backend/internal/models"
	"github.com/sakkurohilla/kineticops/backend/internal/repository/postgres"
	"gorm.io/gorm"
)

type SyntheticService struct {
	db *gorm.DB
}

func NewSyntheticService() *SyntheticService {
	return &SyntheticService{
		db: postgres.DB,
	}
}

// Monitor Management
func (s *SyntheticService) CreateMonitor(monitor *models.SyntheticMonitor) error {
	return postgres.DB.Create(monitor).Error
}

func (s *SyntheticService) GetMonitors(tenantID int64) ([]models.SyntheticMonitor, error) {
	var monitors []models.SyntheticMonitor
	err := postgres.DB.Where("tenant_id = ?", tenantID).Find(&monitors).Error
	return monitors, err
}

func (s *SyntheticService) GetMonitor(id int64, tenantID int64) (*models.SyntheticMonitor, error) {
	var monitor models.SyntheticMonitor
	err := postgres.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&monitor).Error
	return &monitor, err
}

func (s *SyntheticService) UpdateMonitor(id int64, tenantID int64, updates map[string]interface{}) error {
	return postgres.DB.Model(&models.SyntheticMonitor{}).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		Updates(updates).Error
}

func (s *SyntheticService) DeleteMonitor(id int64, tenantID int64) error {
	return postgres.DB.Where("id = ? AND tenant_id = ?", id, tenantID).
		Delete(&models.SyntheticMonitor{}).Error
}

// Monitor Execution
func (s *SyntheticService) ExecuteMonitor(monitor *models.SyntheticMonitor) (*models.SyntheticResult, error) {
	result := &models.SyntheticResult{
		TenantID:  monitor.TenantID,
		MonitorID: monitor.ID,
		Location:  "default", // In production, this would be dynamic
		Timestamp: time.Now(),
	}

	switch monitor.Type {
	case "ping":
		return s.executePingMonitor(monitor, result)
	case "simple_browser":
		return s.executeSimpleBrowserMonitor(monitor, result)
	case "api_test":
		return s.executeAPITestMonitor(monitor, result)
	default:
		result.Success = false
		result.Error = "Unsupported monitor type"
		return result, nil
	}
}

func (s *SyntheticService) executePingMonitor(monitor *models.SyntheticMonitor, result *models.SyntheticResult) (*models.SyntheticResult, error) {
	// Parse config
	var config struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal([]byte(monitor.Config), &config); err != nil {
		result.Success = false
		result.Error = "Invalid monitor configuration"
		return result, nil
	}

	// Execute ping test
	start := time.Now()
	resp, err := http.Get(config.URL)
	duration := time.Since(start).Milliseconds()

	result.Duration = float64(duration)

	if err != nil {
		result.Success = false
		result.Error = err.Error()
	} else {
		result.Success = resp.StatusCode >= 200 && resp.StatusCode < 400
		result.StatusCode = resp.StatusCode
		resp.Body.Close()
	}

	// Save result
	if err := postgres.DB.Create(result).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (s *SyntheticService) executeSimpleBrowserMonitor(monitor *models.SyntheticMonitor, result *models.SyntheticResult) (*models.SyntheticResult, error) {
	// Parse config
	var config struct {
		URL     string `json:"url"`
		Timeout int    `json:"timeout"`
	}
	if err := json.Unmarshal([]byte(monitor.Config), &config); err != nil {
		result.Success = false
		result.Error = "Invalid monitor configuration"
		return result, nil
	}

	// Execute browser test (simplified - in production use headless browser)
	start := time.Now()
	client := &http.Client{
		Timeout: time.Duration(config.Timeout) * time.Second,
	}

	resp, err := client.Get(config.URL)
	duration := time.Since(start).Milliseconds()

	result.Duration = float64(duration)

	if err != nil {
		result.Success = false
		result.Error = err.Error()
	} else {
		result.Success = resp.StatusCode >= 200 && resp.StatusCode < 400
		result.StatusCode = resp.StatusCode

		// Collect basic metrics
		metrics := map[string]interface{}{
			"response_time":  duration,
			"status_code":    resp.StatusCode,
			"content_length": resp.ContentLength,
		}
		metricsJSON, _ := json.Marshal(metrics)
		result.Metrics = string(metricsJSON)

		resp.Body.Close()
	}

	// Save result
	if err := postgres.DB.Create(result).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (s *SyntheticService) executeAPITestMonitor(monitor *models.SyntheticMonitor, result *models.SyntheticResult) (*models.SyntheticResult, error) {
	// Parse config
	var config struct {
		URL     string            `json:"url"`
		Method  string            `json:"method"`
		Headers map[string]string `json:"headers"`
		Body    string            `json:"body"`
		Timeout int               `json:"timeout"`
	}
	if err := json.Unmarshal([]byte(monitor.Config), &config); err != nil {
		result.Success = false
		result.Error = "Invalid monitor configuration"
		return result, nil
	}

	// Execute API test
	start := time.Now()
	client := &http.Client{
		Timeout: time.Duration(config.Timeout) * time.Second,
	}

	req, err := http.NewRequest(config.Method, config.URL, nil)
	if err != nil {
		result.Success = false
		result.Error = err.Error()
		return result, nil
	}

	// Add headers
	for key, value := range config.Headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	duration := time.Since(start).Milliseconds()

	result.Duration = float64(duration)

	if err != nil {
		result.Success = false
		result.Error = err.Error()
	} else {
		result.Success = resp.StatusCode >= 200 && resp.StatusCode < 400
		result.StatusCode = resp.StatusCode
		resp.Body.Close()
	}

	// Save result
	if err := postgres.DB.Create(result).Error; err != nil {
		return nil, err
	}

	return result, nil
}

// Results Management
func (s *SyntheticService) GetResults(monitorID int64, tenantID int64, start, end time.Time, limit int) ([]models.SyntheticResult, error) {
	var results []models.SyntheticResult
	query := postgres.DB.Where("monitor_id = ? AND tenant_id = ?", monitorID, tenantID)

	if !start.IsZero() && !end.IsZero() {
		query = query.Where("timestamp BETWEEN ? AND ?", start, end)
	}

	err := query.Order("timestamp DESC").Limit(limit).Find(&results).Error
	return results, err
}

func (s *SyntheticService) GetMonitorStats(monitorID int64, tenantID int64, start, end time.Time) (map[string]interface{}, error) {
	var stats struct {
		TotalRuns      int64   `json:"total_runs"`
		SuccessfulRuns int64   `json:"successful_runs"`
		FailedRuns     int64   `json:"failed_runs"`
		AvgDuration    float64 `json:"avg_duration"`
		Uptime         float64 `json:"uptime"`
	}

	query := `
		SELECT 
			COUNT(*) as total_runs,
			SUM(CASE WHEN success THEN 1 ELSE 0 END) as successful_runs,
			SUM(CASE WHEN NOT success THEN 1 ELSE 0 END) as failed_runs,
			AVG(duration) as avg_duration
		FROM synthetic_results 
		WHERE monitor_id = ? AND tenant_id = ? AND timestamp BETWEEN ? AND ?
	`

	err := postgres.DB.Raw(query, monitorID, tenantID, start, end).Scan(&stats).Error
	if err != nil {
		return nil, err
	}

	// Calculate uptime percentage
	if stats.TotalRuns > 0 {
		stats.Uptime = float64(stats.SuccessfulRuns) / float64(stats.TotalRuns) * 100
	}

	result := map[string]interface{}{
		"total_runs":      stats.TotalRuns,
		"successful_runs": stats.SuccessfulRuns,
		"failed_runs":     stats.FailedRuns,
		"avg_duration":    stats.AvgDuration,
		"uptime":          stats.Uptime,
	}

	return result, nil
}

// Alert Management
func (s *SyntheticService) CreateAlert(alert *models.SyntheticAlert) error {
	return postgres.DB.Create(alert).Error
}

func (s *SyntheticService) GetAlerts(monitorID int64, tenantID int64) ([]models.SyntheticAlert, error) {
	var alerts []models.SyntheticAlert
	err := postgres.DB.Where("monitor_id = ? AND tenant_id = ?", monitorID, tenantID).Find(&alerts).Error
	return alerts, err
}

func (s *SyntheticService) CheckAlerts(result *models.SyntheticResult) error {
	// Get alerts for this monitor
	var alerts []models.SyntheticAlert
	err := postgres.DB.Where("monitor_id = ? AND tenant_id = ? AND enabled = ?",
		result.MonitorID, result.TenantID, true).Find(&alerts).Error
	if err != nil {
		return err
	}

	for _, alert := range alerts {
		triggered := false

		switch alert.Type {
		case "failure":
			triggered = !result.Success
		case "slow_response":
			triggered = result.Duration > alert.Threshold
		case "error_rate":
			// Check error rate over time window
			// This would require more complex logic in production
			triggered = !result.Success
		}

		if triggered {
			// In production, send notification
			logging.Infof("ALERT: Monitor %d triggered alert %s", result.MonitorID, alert.Type)
		}
	}

	return nil
}

// Browser Monitoring
func (s *SyntheticService) RecordBrowserSession(session *models.BrowserSession) error {
	return postgres.DB.Create(session).Error
}

func (s *SyntheticService) RecordPageView(pageView *models.PageView) error {
	return postgres.DB.Create(pageView).Error
}

func (s *SyntheticService) RecordJavaScriptError(jsError *models.JavaScriptError) error {
	return postgres.DB.Create(jsError).Error
}

func (s *SyntheticService) RecordAjaxRequest(ajaxReq *models.AjaxRequest) error {
	return postgres.DB.Create(ajaxReq).Error
}

func (s *SyntheticService) GetBrowserSessions(appID int64, tenantID int64, start, end time.Time, limit int) ([]models.BrowserSession, error) {
	var sessions []models.BrowserSession
	query := postgres.DB.Where("application_id = ? AND tenant_id = ?", appID, tenantID)

	if !start.IsZero() && !end.IsZero() {
		query = query.Where("start_time BETWEEN ? AND ?", start, end)
	}

	err := query.Order("start_time DESC").Limit(limit).Find(&sessions).Error
	return sessions, err
}

func (s *SyntheticService) GetPageViews(appID int64, tenantID int64, start, end time.Time, limit int) ([]models.PageView, error) {
	var pageViews []models.PageView
	query := postgres.DB.Where("application_id = ? AND tenant_id = ?", appID, tenantID)

	if !start.IsZero() && !end.IsZero() {
		query = query.Where("timestamp BETWEEN ? AND ?", start, end)
	}

	err := query.Order("timestamp DESC").Limit(limit).Find(&pageViews).Error
	return pageViews, err
}

func (s *SyntheticService) GetJavaScriptErrors(appID int64, tenantID int64, start, end time.Time, limit int) ([]models.JavaScriptError, error) {
	var errors []models.JavaScriptError
	query := postgres.DB.Where("application_id = ? AND tenant_id = ?", appID, tenantID)

	if !start.IsZero() && !end.IsZero() {
		query = query.Where("timestamp BETWEEN ? AND ?", start, end)
	}

	err := query.Order("timestamp DESC").Limit(limit).Find(&errors).Error
	return errors, err
}

// Scheduler for running monitors
func (s *SyntheticService) ScheduleMonitors() {
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		for range ticker.C {
			s.runScheduledMonitors()
		}
	}()
}

func (s *SyntheticService) runScheduledMonitors() {
	var monitors []models.SyntheticMonitor
	err := postgres.DB.Where("status = ?", "enabled").Find(&monitors).Error
	if err != nil {
		return
	}

	for _, monitor := range monitors {
		// Check if it's time to run this monitor
		var lastResult models.SyntheticResult
		err := postgres.DB.Where("monitor_id = ?", monitor.ID).
			Order("timestamp DESC").First(&lastResult).Error

		shouldRun := false
		if err != nil {
			// No previous results, run now
			shouldRun = true
		} else {
			// Check if enough time has passed
			timeSinceLastRun := time.Since(lastResult.Timestamp).Seconds()
			shouldRun = timeSinceLastRun >= float64(monitor.Frequency)
		}

		if shouldRun {
			go func(m models.SyntheticMonitor) {
				result, err := s.ExecuteMonitor(&m)
				if err == nil && result != nil {
					if cerr := s.CheckAlerts(result); cerr != nil {
						logging.Warnf("CheckAlerts failed for monitor=%d: %v", result.MonitorID, cerr)
					}
				}
			}(monitor)
		}
	}
}
