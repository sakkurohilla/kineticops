package services

import (
	"context"
	"time"

	"github.com/sakkurohilla/kineticops/backend/internal/logging"
	"github.com/sakkurohilla/kineticops/backend/internal/models"
	"github.com/sakkurohilla/kineticops/backend/internal/repository/postgres"
)

// AgentHealthService manages agent health monitoring
type AgentHealthService struct{}

func NewAgentHealthService() *AgentHealthService {
	return &AgentHealthService{}
}

// UpdateHealth updates agent health status based on heartbeat
func (s *AgentHealthService) UpdateHealth(agentID int64, latency float64, hasError bool) error {
	var health models.AgentHealth
	err := postgres.DB.Where("agent_id = ?", agentID).First(&health).Error

	// Create new health record if not exists
	if err != nil {
		health = models.AgentHealth{
			AgentID:       agentID,
			HealthScore:   100,
			Status:        "healthy",
			LastHeartbeat: time.Now(),
		}
		return postgres.DB.Create(&health).Error
	}

	// Update heartbeat
	health.LastHeartbeat = time.Now()
	health.HeartbeatMissed = 0

	// Update latency (moving average)
	if health.AvgLatency == 0 {
		health.AvgLatency = latency
	} else {
		health.AvgLatency = (health.AvgLatency*0.7 + latency*0.3) // Weighted average
	}

	// Update error rate
	if hasError {
		health.ErrorRate = (health.ErrorRate*0.9 + 100*0.1) // Increase error rate
	} else {
		health.ErrorRate = health.ErrorRate * 0.95 // Decay error rate
	}

	// Calculate health score (0-100)
	health.HealthScore = s.calculateHealthScore(health.AvgLatency, health.ErrorRate, health.HeartbeatMissed)
	health.Status = s.getStatusFromScore(health.HealthScore)

	return postgres.DB.Save(&health).Error
}

// MarkOffline marks agent as offline after missed heartbeats
func (s *AgentHealthService) MarkOffline(agentID int64) error {
	var health models.AgentHealth
	err := postgres.DB.Where("agent_id = ?", agentID).First(&health).Error
	if err != nil {
		return err
	}

	health.HeartbeatMissed++
	health.HealthScore = s.calculateHealthScore(health.AvgLatency, health.ErrorRate, health.HeartbeatMissed)
	health.Status = s.getStatusFromScore(health.HealthScore)

	if health.HeartbeatMissed >= 3 {
		health.Status = "offline"
		health.HealthScore = 0
	}

	return postgres.DB.Save(&health).Error
}

// CheckOfflineAgents checks for agents that haven't sent heartbeats
func (s *AgentHealthService) CheckOfflineAgents(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Find agents with last heartbeat > 2 minutes ago
			cutoff := time.Now().Add(-2 * time.Minute)

			var healthRecords []models.AgentHealth
			postgres.DB.Where("last_heartbeat < ? AND status != ?", cutoff, "offline").Find(&healthRecords)

			for _, health := range healthRecords {
				if err := s.MarkOffline(health.AgentID); err != nil {
					logging.Errorf("Failed to mark agent %d offline: %v", health.AgentID, err)
				}
			}
		}
	}
}

// GetHealth retrieves agent health status
func (s *AgentHealthService) GetHealth(agentID int64) (*models.AgentHealth, error) {
	var health models.AgentHealth
	err := postgres.DB.Where("agent_id = ?", agentID).First(&health).Error
	return &health, err
}

// calculateHealthScore computes a 0-100 health score
func (s *AgentHealthService) calculateHealthScore(latency, errorRate float64, missedHeartbeats int) int {
	score := 100

	// Deduct for high latency (>500ms is bad)
	if latency > 500 {
		score -= int((latency - 500) / 50) // -2 per 100ms over 500ms
	}

	// Deduct for error rate
	score -= int(errorRate / 2) // -50 for 100% error rate

	// Deduct for missed heartbeats
	score -= missedHeartbeats * 20

	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return score
}

// getStatusFromScore converts score to status
func (s *AgentHealthService) getStatusFromScore(score int) string {
	if score >= 80 {
		return "healthy"
	} else if score >= 50 {
		return "degraded"
	} else if score > 0 {
		return "unhealthy"
	}
	return "offline"
}

// Global health service instance
var AgentHealthSvc = NewAgentHealthService()
