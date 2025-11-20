package services

import (
	"context"
	"math"
	"time"

	"github.com/sakkurohilla/kineticops/backend/internal/logging"
	"github.com/sakkurohilla/kineticops/backend/internal/models"
	"github.com/sakkurohilla/kineticops/backend/internal/repository/postgres"
)

// TrendAnalysisService handles trend detection and anomaly detection
type TrendAnalysisService struct{}

func NewTrendAnalysisService() *TrendAnalysisService {
	return &TrendAnalysisService{}
}

// AnalyzeTrends analyzes trends for all active metrics
func (s *TrendAnalysisService) AnalyzeTrends() error {
	// Get distinct metric names
	var metricNames []string
	postgres.DB.Raw("SELECT DISTINCT name FROM metrics WHERE timestamp > NOW() - INTERVAL '1 hour'").Pluck("name", &metricNames)

	for _, metricName := range metricNames {
		// Get all hosts with this metric
		var hostIDs []int64
		postgres.DB.Raw("SELECT DISTINCT host_id FROM metrics WHERE name = ? AND timestamp > NOW() - INTERVAL '1 hour'", metricName).Pluck("host_id", &hostIDs)

		for _, hostID := range hostIDs {
			if err := s.AnalyzeMetricTrend(hostID, metricName); err != nil {
				logging.Errorf("Failed to analyze trend for host=%d metric=%s: %v", hostID, metricName, err)
			}
		}
	}

	return nil
}

// AnalyzeMetricTrend analyzes trend for a specific metric
func (s *TrendAnalysisService) AnalyzeMetricTrend(hostID int64, metricName string) error {
	// Get recent values (last 100 data points)
	var values []struct {
		Value     float64
		Timestamp time.Time
	}

	err := postgres.DB.Raw(`
		SELECT value, timestamp 
		FROM metrics 
		WHERE host_id = ? AND name = ? 
		ORDER BY timestamp DESC 
		LIMIT 100
	`, hostID, metricName).Scan(&values).Error

	if err != nil || len(values) < 10 {
		return nil // Need at least 10 data points
	}

	// Calculate statistics
	var sum, sumSquares float64
	min := math.MaxFloat64
	max := -math.MaxFloat64

	for _, v := range values {
		sum += v.Value
		sumSquares += v.Value * v.Value
		if v.Value < min {
			min = v.Value
		}
		if v.Value > max {
			max = v.Value
		}
	}

	n := float64(len(values))
	mean := sum / n
	variance := (sumSquares / n) - (mean * mean)
	stdDev := math.Sqrt(variance)

	// Calculate moving average
	movingAvg := mean

	// Calculate slope (linear regression)
	slope := s.calculateSlope(values)

	// Detect trend type
	trendType := "stable"
	if math.Abs(slope) > stdDev*0.1 {
		if slope > 0 {
			trendType = "increasing"
		} else {
			trendType = "decreasing"
		}
	}

	// Detect anomalies (values beyond 3 standard deviations)
	isAnomaly := false
	anomalyScore := 0.0
	latestValue := values[0].Value
	deviation := math.Abs(latestValue - mean)

	if deviation > 3*stdDev && stdDev > 0 {
		isAnomaly = true
		anomalyScore = deviation / stdDev
	}

	// Predict next value using simple linear regression
	predictedValue := latestValue + slope

	// Store trend analysis
	trend := models.MetricTrend{
		HostID:         hostID,
		MetricName:     metricName,
		TrendType:      trendType,
		Confidence:     s.calculateConfidence(n, variance),
		MovingAvg:      movingAvg,
		StdDev:         stdDev,
		Slope:          slope,
		IsAnomaly:      isAnomaly,
		AnomalyScore:   anomalyScore,
		PredictedValue: predictedValue,
		AnalyzedAt:     time.Now(),
	}

	return postgres.DB.Create(&trend).Error
}

// calculateSlope computes linear regression slope
func (s *TrendAnalysisService) calculateSlope(values []struct {
	Value     float64
	Timestamp time.Time
}) float64 {
	if len(values) < 2 {
		return 0
	}

	n := float64(len(values))
	var sumX, sumY, sumXY, sumX2 float64

	for i, v := range values {
		x := float64(i)
		y := v.Value
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	// Slope formula: (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)
	numerator := n*sumXY - sumX*sumY
	denominator := n*sumX2 - sumX*sumX

	if denominator == 0 {
		return 0
	}

	return numerator / denominator
}

// calculateConfidence returns a confidence score for the trend
func (s *TrendAnalysisService) calculateConfidence(n, variance float64) float64 {
	// More data points = higher confidence
	dataConfidence := math.Min(n/100, 1.0)

	// Lower variance = higher confidence
	varianceConfidence := 1.0 / (1.0 + variance)

	return (dataConfidence + varianceConfidence) / 2.0
}

// StartTrendAnalysisWorker runs periodic trend analysis
func (s *TrendAnalysisService) StartTrendAnalysisWorker(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.AnalyzeTrends(); err != nil {
				logging.Errorf("Trend analysis failed: %v", err)
			} else {
				logging.Infof("Trend analysis completed")
			}
		}
	}
}

// GetTrends retrieves trend data for a metric
func (s *TrendAnalysisService) GetTrends(hostID int64, metricName string, limit int) ([]models.MetricTrend, error) {
	var trends []models.MetricTrend
	err := postgres.DB.Where("host_id = ? AND metric_name = ?", hostID, metricName).
		Order("analyzed_at DESC").
		Limit(limit).
		Find(&trends).Error
	return trends, err
}

// Global trend analysis service
var TrendAnalysisSvc = NewTrendAnalysisService()
