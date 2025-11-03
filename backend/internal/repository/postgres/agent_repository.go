package postgres

import (
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