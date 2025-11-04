package workers

import (
	"log"
	"time"

	"github.com/sakkurohilla/kineticops/backend/internal/repository/postgres"
)

// StartHeartbeatMonitor monitors host heartbeats and marks offline hosts
func StartHeartbeatMonitor() {
	log.Println("[HEARTBEAT] Starting heartbeat monitor...")
	
	ticker := time.NewTicker(30 * time.Second) // Check every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			checkHostHeartbeats()
		}
	}
}

func checkHostHeartbeats() {
	// Mark hosts offline if no heartbeat for 2 minutes
	timeout := time.Now().Add(-2 * time.Minute)
	
	result := postgres.DB.Exec(`
		UPDATE hosts 
		SET agent_status = 'offline' 
		WHERE agent_status = 'online' 
		AND last_seen < ?
	`, timeout)
	
	if result.Error != nil {
		log.Printf("[HEARTBEAT] Error updating offline hosts: %v", result.Error)
		return
	}
	
	if result.RowsAffected > 0 {
		log.Printf("[HEARTBEAT] Marked %d hosts as offline (no heartbeat > 2min)", result.RowsAffected)
	}
}