package services

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sakkurohilla/kineticops/backend/internal/logging"
	"github.com/sakkurohilla/kineticops/backend/internal/models"
	"github.com/sakkurohilla/kineticops/backend/internal/repository/postgres"
	ws "github.com/sakkurohilla/kineticops/backend/internal/websocket"
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
	// Test SSH connection FIRST - REQUIRED for production
	err := s.testSSHConnection(req)
	if err != nil {
		return nil, fmt.Errorf("SSH authentication failed: %v", err)
	}

	// Generate session token
	token, err := s.generateSessionToken(req.HostID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate session token: %v", err)
	}

	// Encrypt credentials for storage (only for session duration)
	encryptedPassword := ""
	encryptedSSHKey := ""

	if req.Password != "" {
		encryptedPassword, err = EncryptCredential(req.Password, s.jwtSecret)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt password: %v", err)
		}
	}

	if req.SSHKey != "" {
		encryptedSSHKey, err = EncryptCredential(req.SSHKey, s.jwtSecret)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt SSH key: %v", err)
		}
	}

	// Get agent for this host
	agent, err := s.agentRepo.GetByHostID(req.HostID)
	var agentID *int
	if err == nil {
		agentID = &agent.ID
	}

	// Create session with encrypted credentials
	session := &models.WorkflowSession{
		HostID:       req.HostID,
		UserID:       userID,
		AgentID:      agentID,
		SessionToken: token,
		Status:       "active",
		// SSH workflow sessions expire after 10 minutes for security
		ExpiresAt: time.Now().Add(10 * time.Minute),
		Username:  req.Username,
		Password:  encryptedPassword,
		SSHKey:    encryptedSSHKey,
	}

	err = s.workflowRepo.CreateSession(session)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %v", err)
	}

	logging.Infof("Workflow session created for host=%d user=%d with %s authentication",
		req.HostID, userID, func() string {
			if req.SSHKey != "" {
				return "SSH key"
			}
			return "password"
		}())

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
	if err := s.workflowRepo.UpdateSession(token); err != nil {
		logging.Warnf("failed to update workflow session last-activity for token=%s: %v", token, err)
	}

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

// DiscoverServicesRealtime gets real-time services from websocket cache (same as Services page)
func (s *WorkflowService) DiscoverServicesRealtime(hostID int, sessionToken string) (interface{}, error) {
	// Validate session
	_, err := s.ValidateSession(sessionToken)
	if err != nil {
		return nil, err
	}

	// Get last cached services from websocket hub
	servicesData, err := ws.GetLastServicesForHost(int64(hostID))
	if err != nil {
		// Fallback to database if no websocket data available yet
		logging.Warnf("No real-time services for host %d, using database fallback: %v", hostID, err)
		return s.DiscoverServices(hostID, sessionToken)
	}

	// Extract services from the websocket message
	services, ok := servicesData["services"]
	if !ok {
		return nil, fmt.Errorf("services field not found in cached data")
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

	if err := s.workflowRepo.CreateControlLog(controlLog); err != nil {
		logging.Warnf("failed to write service control log for service=%d host=%d: %v", serviceID, session.HostID, err)
	}

	return &models.ServiceControlResponse{
		Success: err == nil,
		Output:  output,
		Error: func() string {
			if err != nil {
				return err.Error()
			}
			return ""
		}(),
	}, nil
}

func (s *WorkflowService) ExecuteRemoteCommand(hostID int, command string, sessionToken string) (string, error) {
	// Get session with credentials
	session, err := s.workflowRepo.GetSession(sessionToken)
	if err != nil {
		return "", fmt.Errorf("invalid session: %v", err)
	}

	// Get host details
	host, err := s.hostRepo.GetByID(hostID)
	if err != nil {
		return "", fmt.Errorf("host not found: %v", err)
	}

	// Decrypt credentials
	password := ""
	sshKey := ""

	if session.Password != "" {
		password, err = DecryptCredential(session.Password, s.jwtSecret)
		if err != nil {
			return "", fmt.Errorf("failed to decrypt password: %v", err)
		}
	}

	if session.SSHKey != "" {
		sshKey, err = DecryptCredential(session.SSHKey, s.jwtSecret)
		if err != nil {
			return "", fmt.Errorf("failed to decrypt SSH key: %v", err)
		}
	}

	// Create SSH client with decrypted credentials
	var sshClient *SSHClient
	if sshKey != "" {
		sshClient, err = NewSSHClientWithKey(host.IP, 22, session.Username, "", sshKey)
	} else {
		sshClient, err = NewSSHClient(host.IP, 22, session.Username, password)
	}

	if err != nil {
		return "", fmt.Errorf("SSH connection failed: %v", err)
	}
	defer sshClient.Close()

	// Execute command
	output, err := sshClient.ExecuteCommand(command)
	if err != nil {
		return "", fmt.Errorf("command execution failed: %v", err)
	}

	logging.Infof("Executed command on host %d: %s", hostID, command)
	return output, nil
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

func (s *WorkflowService) GetHostWorkflowRealtime(hostID int, sessionToken string) (map[string]interface{}, error) {
	// Validate session
	_, err := s.ValidateSession(sessionToken)
	if err != nil {
		return nil, err
	}

	// Get services from websocket cache (real-time)
	services, err := s.DiscoverServicesRealtime(hostID, sessionToken)
	if err != nil {
		// Fallback to database
		services, err = s.DiscoverServices(hostID, sessionToken)
		if err != nil {
			return nil, err
		}
	}

	// Get latest host metrics from database
	var metrics struct {
		CPUUsage       float64 `db:"cpu_usage"`
		MemoryUsage    float64 `db:"memory_usage"`
		MemoryFree     float64 `db:"memory_free"`
		DiskUsage      float64 `db:"disk_usage"`
		DiskReadSpeed  float64 `db:"disk_read_speed"`
		DiskWriteSpeed float64 `db:"disk_write_speed"`
		NetworkIn      float64 `db:"network_in"`
		NetworkOut     float64 `db:"network_out"`
		LoadAverage    float64 `db:"load_average"`
	}

	query := `
		SELECT cpu_usage, memory_usage, memory_free, disk_usage, 
		       disk_read_speed, disk_write_speed, network_in, network_out, load_average
		FROM host_metrics
		WHERE host_id = $1
		ORDER BY timestamp DESC
		LIMIT 1
	`

	// Use GetDB() method to access the database
	err = s.workflowRepo.GetDB().Get(&metrics, query, hostID)
	if err != nil {
		logging.Warnf("Failed to fetch metrics for host %d: %v", hostID, err)
		// Return services without metrics
		return map[string]interface{}{
			"services": services,
			"status":   "active",
		}, nil
	}

	// Get recent logs
	logs, err := s.workflowRepo.GetControlLogs(hostID, 10)
	if err != nil {
		logs = []models.ServiceControlLog{} // Empty if error
	}

	return map[string]interface{}{
		"services": services,
		"metrics": map[string]interface{}{
			"cpu_usage":        metrics.CPUUsage,
			"memory_usage":     metrics.MemoryUsage,
			"memory_free":      metrics.MemoryFree,
			"disk_usage":       metrics.DiskUsage,
			"disk_read_speed":  metrics.DiskReadSpeed,
			"disk_write_speed": metrics.DiskWriteSpeed,
			"network_in":       metrics.NetworkIn,
			"network_out":      metrics.NetworkOut,
			"load_average":     metrics.LoadAverage,
		},
		"logs":   logs,
		"status": "active",
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
		return fmt.Errorf("host not found: %v", err)
	}

	// Use host IP for SSH connection
	hostIP := host.IP
	if hostIP == "" {
		return fmt.Errorf("host IP not configured")
	}

	// Validate credentials are provided
	if req.Username == "" {
		return fmt.Errorf("SSH username is required")
	}

	if req.SSHKey == "" && req.Password == "" {
		return fmt.Errorf("SSH password or key is required")
	}

	// Test the actual SSH connection
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
