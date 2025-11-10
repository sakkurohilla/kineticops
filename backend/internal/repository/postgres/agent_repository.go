package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/sakkurohilla/kineticops/backend/internal/models"
	"github.com/sakkurohilla/kineticops/backend/internal/telemetry"
	ws "github.com/sakkurohilla/kineticops/backend/internal/websocket"
)

type AgentRepository struct {
	db *sqlx.DB
}

func NewAgentRepository(db *sqlx.DB) *AgentRepository {
	return &AgentRepository{db: db}
}

func (r *AgentRepository) Create(agent *models.Agent) error {
	query := `
		INSERT INTO agents (host_id, token, status, version)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`

	return r.db.QueryRow(query, agent.HostID, agent.AgentToken, agent.Status,
		agent.Version).Scan(
		&agent.ID, &agent.CreatedAt, &agent.UpdatedAt)
}

func (r *AgentRepository) GetByToken(token string) (*models.Agent, error) {
	var agent models.Agent
	query := `SELECT * FROM agents WHERE token = $1`
	err := r.db.Get(&agent, query, token)
	if err != nil {
		return nil, err
	}
	return &agent, nil
}

// GetAgentByToken is a convenience wrapper that uses the package-level SqlxDB
// to resolve an agent by its token. This is handy for middleware or callers
// that don't have an AgentRepository instance.
func GetAgentByToken(token string) (*models.Agent, error) {
	var agent models.Agent
	query := `SELECT * FROM agents WHERE token = $1`
	err := SqlxDB.Get(&agent, query, token)
	if err != nil {
		return nil, err
	}
	return &agent, nil
}

func (r *AgentRepository) GetByHostID(hostID int) (*models.Agent, error) {
	var agent models.Agent
	query := `SELECT * FROM agents WHERE host_id = $1`
	err := r.db.Get(&agent, query, hostID)
	if err != nil {
		return nil, err
	}
	return &agent, nil
}

func (r *AgentRepository) UpdateHeartbeat(token string, heartbeat *models.AgentHeartbeat) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Update agent heartbeat
	_, err = tx.Exec(`
		UPDATE agents 
		SET last_heartbeat = $1, status = 'online', updated_at = CURRENT_TIMESTAMP
		WHERE token = $2`,
		time.Now(), token)
	if err != nil {
		return err
	}

	// Get agent to find host_id and agent_id
	var hostID, agentID int
	err = tx.QueryRow("SELECT host_id, id FROM agents WHERE token = $1", token).Scan(&hostID, &agentID)
	if err != nil {
		// If the agent token isn't found, check for a recent tombstone for
		// this hostname (agent may be using old credentials). If the host was
		// recently deleted, refuse to auto-recreate it to avoid unwanted
		// re-registration after manual removal.
		if heartbeat != nil && heartbeat.Metadata.Hostname != "" {
			var tombstones int
			// count deleted_hosts entries in last 7 days for this hostname
			_ = r.db.Get(&tombstones, "SELECT count(*) FROM deleted_hosts WHERE hostname = $1 AND deleted_at > NOW() - INTERVAL '7 days'", heartbeat.Metadata.Hostname)
			if tombstones > 0 {
				fmt.Printf("[AGENT] Refusing heartbeat from token=%s hostname=%s: host recently deleted\n", token, heartbeat.Metadata.Hostname)
				return fmt.Errorf("token not recognized and host %s was recently deleted", heartbeat.Metadata.Hostname)
			}
		}
		// Otherwise return the original error (unknown token)
		return err
	}

	// Compute server-side disk usage percent when disk bytes are provided.
	diskUsagePct := heartbeat.DiskUsage
	if heartbeat != nil && heartbeat.DiskTotalBytes > 0 {
		// use provided byte counts to compute a canonical percent on the server
		diskUsagePct = 0.0
		if heartbeat.DiskTotalBytes > 0 {
			diskUsagePct = (float64(heartbeat.DiskUsedBytes) / float64(heartbeat.DiskTotalBytes)) * 100.0
		}
	}

	// Update host metrics - use upsert to avoid duplicate-key errors when a row already exists
	_, err = tx.Exec(`
		INSERT INTO host_metrics (host_id, cpu_usage, memory_usage, disk_usage, disk_total, disk_used, network_in, network_out, uptime, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (host_id) DO UPDATE SET
			cpu_usage = EXCLUDED.cpu_usage,
			memory_usage = EXCLUDED.memory_usage,
			-- Compute disk_usage server-side: prefer computing from provided bytes (excluded) when available,
			-- otherwise fall back to the provided EXCLUDED.disk_usage value.
			disk_usage = CASE
				WHEN COALESCE(EXCLUDED.disk_total, host_metrics.disk_total) IS NOT NULL AND COALESCE(EXCLUDED.disk_total, host_metrics.disk_total) <> 0
					THEN (COALESCE(EXCLUDED.disk_used, host_metrics.disk_used)::double precision / NULLIF(COALESCE(EXCLUDED.disk_total, host_metrics.disk_total),0)::double precision) * 100.0
				ELSE EXCLUDED.disk_usage
			END,
			disk_total = COALESCE(EXCLUDED.disk_total, host_metrics.disk_total),
			disk_used = COALESCE(EXCLUDED.disk_used, host_metrics.disk_used),
			network_in = COALESCE(EXCLUDED.network_in, host_metrics.network_in),
			network_out = COALESCE(EXCLUDED.network_out, host_metrics.network_out),
			uptime = EXCLUDED.uptime,
			timestamp = EXCLUDED.timestamp`,
		hostID, heartbeat.CPUUsage, heartbeat.MemoryUsage, diskUsagePct,
		heartbeat.DiskTotalBytes, heartbeat.DiskUsedBytes, 0.0, 0.0, heartbeat.Metadata.Uptime, time.Now())
	if err != nil {
		return err
	}

	// Update host status
	_, err = tx.Exec(`
		UPDATE hosts 
		SET agent_status = 'online', last_seen = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2`,
		time.Now(), hostID)
	if err != nil {
		return err
	}

	// Update agent services
	for _, service := range heartbeat.Services {
		_, err = tx.Exec(`
			INSERT INTO agent_services (agent_id, service_name, status, process_id, memory_usage, cpu_usage, last_check)
			VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP)
			ON CONFLICT (agent_id, service_name) 
			DO UPDATE SET status = $3, process_id = $4, memory_usage = $5, cpu_usage = $6, last_check = CURRENT_TIMESTAMP`,
			agentID, service.Name, service.Status, service.PID, service.MemoryUsage, service.CPUUsage)
		if err != nil {
			return err
		}
	}

	// After upsert, compute and persist a canonical disk_usage percent from stored bytes
	// (done inside transaction to avoid races). This ensures an accurate percent even if
	// the EXCLUDED binding did not produce the expected value in the upsert path.
	_, err = tx.Exec(`UPDATE host_metrics SET disk_usage = (disk_used::double precision / NULLIF(disk_total,0)::double precision) * 100.0 WHERE host_id = $1`, hostID)
	if err != nil {
		// If this fails, fall back to commit and return the original error
		_ = tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	// disk_usage is now computed server-side in the upsert SQL (from disk bytes when provided)

	// Broadcast a lightweight metric update for realtime UI updates
	payload := map[string]interface{}{
		"type":         "metric",
		"host_id":      hostID,
		"cpu_usage":    heartbeat.CPUUsage,
		"memory_usage": heartbeat.MemoryUsage,
		// disk_usage is the server-canonical percent (computed from bytes if available)
		"disk_usage": diskUsagePct,
		"disk_total": heartbeat.DiskTotalBytes,
		"disk_used":  heartbeat.DiskUsedBytes,
		"uptime":     0,
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
	}

	// If the agent included uptime in metadata, prefer that value (seconds)
	if heartbeat != nil && heartbeat.Metadata.Uptime > 0 {
		payload["uptime"] = heartbeat.Metadata.Uptime
	}
	if b, jerr := json.Marshal(payload); jerr == nil {
		fmt.Printf("[WS BROADCAST] host=%d heartbeat\n", hostID)
		if gh := ws.GetGlobalHub(); gh != nil {
			gh.RememberMessage(b)
		}
		ws.BroadcastToClients(b)
		telemetry.IncWSBroadcast(context.Background(), 1)
	}

	return nil
}

func (r *AgentRepository) UpdateStatus(id int, status string, log string) error {
	query := `
		UPDATE agents 
		SET status = $1, installation_log = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $3`
	_, err := r.db.Exec(query, status, log, id)
	return err
}

func (r *AgentRepository) GetServices(agentID int) ([]models.AgentService, error) {
	var services []models.AgentService
	query := `SELECT * FROM agent_services WHERE agent_id = $1 ORDER BY service_name`
	err := r.db.Select(&services, query, agentID)
	return services, err
}

func (r *AgentRepository) MarkInstalled(id int) error {
	query := `
		UPDATE agents 
		SET status = 'installed', installed_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *AgentRepository) CreateInstallLog(log *models.AgentInstallationLog) error {
	query := `
		INSERT INTO agent_installation_logs (agent_id, setup_method, status, logs, error_message)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, started_at`

	return r.db.QueryRow(query, log.AgentID, log.SetupMethod, log.Status,
		log.Logs, log.ErrorMessage).Scan(&log.ID, &log.StartedAt)
}

func (r *AgentRepository) UpdateInstallLog(id int, status, logs, errorMsg string) error {
	query := `
		UPDATE agent_installation_logs 
		SET status = $1, logs = $2, error_message = $3, completed_at = CURRENT_TIMESTAMP
		WHERE id = $1`
	_, err := r.db.Exec(query, status, logs, errorMsg, id)
	return err
}

// UpdateRevoked sets the revoked flag for an agent record.
func (r *AgentRepository) UpdateRevoked(agentID int, revoked bool) error {
	query := `UPDATE agents SET revoked = $1, revoked_at = CASE WHEN $1 THEN CURRENT_TIMESTAMP ELSE NULL END, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
	_, err := r.db.Exec(query, revoked, agentID)
	return err
}
