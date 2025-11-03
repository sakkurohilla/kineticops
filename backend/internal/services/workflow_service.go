package services

import (
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sakkurohilla/kineticops/backend/internal/models"
	"github.com/sakkurohilla/kineticops/backend/internal/repository/postgres"
)

type WorkflowService struct {
	workflowRepo *postgres.WorkflowRepository
	agentRepo    *postgres.AgentRepository
	hostRepo     *postgres.HostRepository
	sshService   *SSHService
	jwtSecret    string
}

func NewWorkflowService(workflowRepo *postgres.WorkflowRepository, agentRepo *postgres.AgentRepository, hostRepo *postgres.HostRepository, sshService *SSHService, jwtSecret string) *WorkflowService {
	return &WorkflowService{
		workflowRepo: workflowRepo,
		agentRepo:    agentRepo,
		hostRepo:     hostRepo,
		sshService:   sshService,
		jwtSecret:    jwtSecret,
	}
}

func (s *WorkflowService) CreateWorkflowSession(req *models.WorkflowSessionRequest, userID int) (*models.WorkflowSessionResponse, error) {
	// Generate session token
	token, err := s.generateSessionToken(req.HostID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate session token: %v", err)
	}

	// Get agent for this host
	agent, err := s.agentRepo.GetByHostID(req.HostID)
	var agentID *int
	if err == nil {
		agentID = &agent.ID
	}

	// Create session
	session := &models.WorkflowSession{
		HostID:       req.HostID,
		UserID:       userID,
		AgentID:      agentID,
		SessionToken: token,
		Status:       "active",
		ExpiresAt:    time.Now().Add(1 * time.Hour),
	}

	err = s.workflowRepo.CreateSession(session)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %v", err)
	}

	// Test SSH connection - REQUIRED for production
	err = s.testSSHConnection(req)
	if err != nil {
		return nil, fmt.Errorf("SSH connection failed: %v", err)
	}

	return &models.WorkflowSessionResponse{
		SessionToken: token,
		ExpiresAt:    session.ExpiresAt,
		HostID:       req.HostID,
		Status:       "active",
	}, nil
}

func (s *WorkflowService) ValidateSession(token string) (*models.WorkflowSession, error) {
	session, err := s.workflowRepo.GetSession(token)
	if err != nil {
		return nil, fmt.Errorf("invalid session: %v", err)
	}

	// Update last activity
	s.workflowRepo.UpdateSession(token)

	return session, nil
}

func (s *WorkflowService) DiscoverServices(hostID int, sessionToken string) ([]models.AgentService, error) {
	// Validate session
	_, err := s.ValidateSession(sessionToken)
	if err != nil {
		return nil, err
	}

	// Get agent for this host
	agent, err := s.agentRepo.GetByHostID(hostID)
	if err != nil {
		return nil, fmt.Errorf("agent not found for host: %v", err)
	}

	// Get services from agent
	services, err := s.agentRepo.GetServices(agent.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get services: %v", err)
	}

	return services, nil
}

func (s *WorkflowService) ControlService(serviceID int, action models.ControlAction, sessionToken string, userID int) (*models.ServiceControlResponse, error) {
	// Validate session
	session, err := s.ValidateSession(sessionToken)
	if err != nil {
		return nil, err
	}

	// Get service details
	services, err := s.agentRepo.GetServices(*session.AgentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get services: %v", err)
	}

	var targetService *models.AgentService
	for _, svc := range services {
		if svc.ID == serviceID {
			targetService = &svc
			break
		}
	}

	if targetService == nil {
		return nil, fmt.Errorf("service not found")
	}

	// Execute command via SSH
	command := s.buildServiceCommand(targetService.ServiceName, action)
	output, err := s.ExecuteRemoteCommand(session.HostID, command, sessionToken)

	// Log the action
	controlLog := &models.ServiceControlLog{
		ServiceID:   &serviceID,
		ServiceName: targetService.ServiceName,
		HostID:      session.HostID,
		Action:      string(action),
		Status:      "success",
		Output:      output,
		ExecutedBy:  userID,
	}

	if err != nil {
		controlLog.Status = "failed"
		controlLog.ErrorMessage = err.Error()
	}

	s.workflowRepo.CreateControlLog(controlLog)

	return &models.ServiceControlResponse{
		Success: err == nil,
		Output:  output,
		Error:   func() string { if err != nil { return err.Error() }; return "" }(),
	}, nil
}

func (s *WorkflowService) ExecuteRemoteCommand(hostID int, command string, sessionToken string) (string, error) {
	// This would use cached SSH credentials from session
	// Execute actual SSH command
	log.Printf("Executing command on host %d: %s", hostID, command)
	// TODO: Implement actual SSH command execution
	return "", fmt.Errorf("SSH command execution not implemented")
}

func (s *WorkflowService) GetHostWorkflow(hostID int, sessionToken string) (map[string]interface{}, error) {
	// Validate session
	_, err := s.ValidateSession(sessionToken)
	if err != nil {
		return nil, err
	}

	// Get services
	services, err := s.DiscoverServices(hostID, sessionToken)
	if err != nil {
		return nil, err
	}

	// Get recent logs
	logs, err := s.workflowRepo.GetControlLogs(hostID, 10)
	if err != nil {
		logs = []models.ServiceControlLog{} // Empty if error
	}

	return map[string]interface{}{
		"services": services,
		"logs":     logs,
		"status":   "active",
	}, nil
}

func (s *WorkflowService) CloseSession(sessionToken string) error {
	return s.workflowRepo.ExpireSession(sessionToken)
}

// Helper methods

func (s *WorkflowService) generateSessionToken(hostID, userID int) (string, error) {
	claims := jwt.MapClaims{
		"host_id": hostID,
		"user_id": userID,
		"exp":     time.Now().Add(1 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *WorkflowService) testSSHConnection(req *models.WorkflowSessionRequest) error {
	// Get host details to get the actual IP address
	host, err := s.hostRepo.GetByID(req.HostID)
	if err != nil {
		// For development: allow workflow without existing host
		log.Printf("[WARN] Host %d not found, skipping SSH test: %v", req.HostID, err)
		return nil
	}
	
	// Use host IP for SSH connection
	hostIP := host.IP
	if hostIP == "" {
		return fmt.Errorf("host IP not configured")
	}
	
	if req.SSHKey != "" {
		return TestSSHConnectionWithKey(hostIP, 22, req.Username, "", req.SSHKey)
	}
	return TestSSHConnection(hostIP, 22, req.Username, req.Password)
}

func (s *WorkflowService) buildServiceCommand(serviceName string, action models.ControlAction) string {
	switch action {
	case models.ActionStart:
		return fmt.Sprintf("sudo systemctl start %s", serviceName)
	case models.ActionStop:
		return fmt.Sprintf("sudo systemctl stop %s", serviceName)
	case models.ActionRestart:
		return fmt.Sprintf("sudo systemctl restart %s", serviceName)
	case models.ActionEnable:
		return fmt.Sprintf("sudo systemctl enable %s", serviceName)
	case models.ActionDisable:
		return fmt.Sprintf("sudo systemctl disable %s", serviceName)
	default:
		return fmt.Sprintf("sudo systemctl status %s", serviceName)
	}
}