package postgres

import (
	"github.com/jmoiron/sqlx"
	"github.com/sakkurohilla/kineticops/backend/internal/models"
)

type WorkflowRepository struct {
	db *sqlx.DB
}

func NewWorkflowRepository(db *sqlx.DB) *WorkflowRepository {
	return &WorkflowRepository{db: db}
}

func (r *WorkflowRepository) GetDB() *sqlx.DB {
	return r.db
}

func (r *WorkflowRepository) CreateSession(session *models.WorkflowSession) error {
	query := `
		INSERT INTO workflow_sessions (host_id, user_id, agent_id, session_token, expires_at, username, password_encrypted, ssh_key_encrypted)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, authenticated_at, last_activity`

	return r.db.QueryRow(query, session.HostID, session.UserID, session.AgentID,
		session.SessionToken, session.ExpiresAt, session.Username, session.Password, session.SSHKey).Scan(
		&session.ID, &session.AuthenticatedAt, &session.LastActivity)
}

func (r *WorkflowRepository) GetSession(token string) (*models.WorkflowSession, error) {
	var session models.WorkflowSession
	// Check both expires_at AND last_activity (5 min inactivity timeout)
	query := `SELECT * FROM workflow_sessions 
		WHERE session_token = $1 
		AND expires_at > NOW() 
		AND last_activity > NOW() - INTERVAL '5 minutes'
		AND status = 'active'`
	err := r.db.Get(&session, query, token)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *WorkflowRepository) UpdateSession(token string) error {
	query := `
		UPDATE workflow_sessions 
		SET last_activity = CURRENT_TIMESTAMP
		WHERE session_token = $1`
	_, err := r.db.Exec(query, token)
	return err
}

func (r *WorkflowRepository) ExpireSession(token string) error {
	query := `
		UPDATE workflow_sessions 
		SET status = 'expired', expires_at = CURRENT_TIMESTAMP
		WHERE session_token = $1`
	_, err := r.db.Exec(query, token)
	return err
}

func (r *WorkflowRepository) GetSessionByHostAndUser(hostID, userID int) (*models.WorkflowSession, error) {
	var session models.WorkflowSession
	query := `
		SELECT * FROM workflow_sessions 
		WHERE host_id = $1 AND user_id = $2 AND expires_at > NOW() AND status = 'active'
		ORDER BY authenticated_at DESC LIMIT 1`
	err := r.db.Get(&session, query, hostID, userID)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *WorkflowRepository) CreateControlLog(log *models.ServiceControlLog) error {
	query := `
		INSERT INTO service_control_logs (service_id, service_name, host_id, action, status, output, error_message, executed_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, executed_at`

	return r.db.QueryRow(query, log.ServiceID, log.ServiceName, log.HostID,
		log.Action, log.Status, log.Output, log.ErrorMessage, log.ExecutedBy).Scan(
		&log.ID, &log.ExecutedAt)
}

func (r *WorkflowRepository) GetControlLogs(hostID int, limit int) ([]models.ServiceControlLog, error) {
	var logs []models.ServiceControlLog
	query := `
		SELECT * FROM service_control_logs 
		WHERE host_id = $1 
		ORDER BY executed_at DESC 
		LIMIT $2`
	err := r.db.Select(&logs, query, hostID, limit)
	return logs, err
}

func (r *WorkflowRepository) CleanupExpiredSessions() error {
	// Expire sessions based on expires_at OR inactivity (5 min)
	query := `
		UPDATE workflow_sessions 
		SET status = 'expired' 
		WHERE status = 'active' 
		AND (expires_at < NOW() OR last_activity < NOW() - INTERVAL '5 minutes')`
	_, err := r.db.Exec(query)
	return err
}
