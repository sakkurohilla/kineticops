package postgres

import (
	"encoding/json"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/sakkurohilla/kineticops/backend/internal/models"
)

type AgentRepository struct {
	db *sqlx.DB
}

func NewAgentRepository(db *sqlx.DB) *AgentRepository {
	return &AgentRepository{db: db}
}

func (r *AgentRepository) Create(agent *models.Agent) error {
	query := `
		INSERT INTO agents (host_id, token, status, version, os_info, system_info, installation_log)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at`
	
	return r.db.QueryRow(query, agent.HostID, agent.Token, agent.Status, agent.Version,
		agent.OSInfo, agent.SystemInfo, agent.InstallationLog).Scan(
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
	systemInfoJSON, _ := json.Marshal(heartbeat.SystemInfo)
	_, err = tx.Exec(`
		UPDATE agents 
		SET last_heartbeat = $1, system_info = $2, status = 'online', updated_at = CURRENT_TIMESTAMP
		WHERE token = $3`,
		time.Now(), systemInfoJSON, token)
	if err != nil {
		return err
	}

	// Get agent to find host_id
	var hostID int
	err = tx.Get(&hostID, "SELECT host_id FROM agents WHERE token = $1", token)
	if err != nil {
		return err
	}

	// Update host metrics
	_, err = tx.Exec(`
		INSERT INTO host_metrics (host_id, cpu_usage, memory_usage, disk_usage, timestamp)
		VALUES ($1, $2, $3, $4, $5)`,
		hostID, heartbeat.CPUUsage, heartbeat.MemoryUsage, heartbeat.DiskUsage, time.Now())
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

	// Update services
	for _, service := range heartbeat.Services {
		_, err = tx.Exec(`
			INSERT INTO host_services (host_id, name, status, pid, memory_usage, cpu_usage, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP)
			ON CONFLICT (host_id, name) 
			DO UPDATE SET status = $3, pid = $4, memory_usage = $5, cpu_usage = $6, updated_at = CURRENT_TIMESTAMP`,
			hostID, service.Name, service.Status, service.PID, service.MemoryUsage, service.CPUUsage)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *AgentRepository) UpdateStatus(id int, status string, log string) error {
	query := `
		UPDATE agents 
		SET status = $1, installation_log = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $3`
	_, err := r.db.Exec(query, status, log, id)
	return err
}

func (r *AgentRepository) GetServices(hostID int) ([]models.HostService, error) {
	var services []models.HostService
	query := `SELECT * FROM host_services WHERE host_id = $1 ORDER BY name`
	err := r.db.Select(&services, query, hostID)
	return services, err
}