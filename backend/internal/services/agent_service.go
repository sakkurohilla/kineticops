package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/sakkurohilla/kineticops/backend/internal/models"
	"github.com/sakkurohilla/kineticops/backend/internal/repository/postgres"
)

type AgentService struct {
	agentRepo *postgres.AgentRepository
	hostRepo  *postgres.HostRepository
	sshService *SSHService
}

func NewAgentService(agentRepo *postgres.AgentRepository, hostRepo *postgres.HostRepository, sshService *SSHService) *AgentService {
	return &AgentService{
		agentRepo: agentRepo,
		hostRepo: hostRepo,
		sshService: sshService,
	}
}

func (s *AgentService) SetupAgent(req *models.AgentSetupRequest) (*models.AgentSetupResponse, error) {
	// Create host first
	host := &models.Host{
		Hostname:     req.Hostname,
		IP:           req.IP,
		AgentStatus:  "installing",
	}
	
	err := s.hostRepo.Create(host)
	if err != nil {
		return nil, fmt.Errorf("failed to create host: %v", err)
	}

	// Generate unique token
	token, err := s.generateToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %v", err)
	}

	// Create agent record
	agent := &models.Agent{
		HostID: int(host.ID),
		Token:  token,
		Status: "pending",
	}
	
	err = s.agentRepo.Create(agent)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %v", err)
	}

	response := &models.AgentSetupResponse{
		HostID:      int(host.ID),
		AgentID:     agent.ID,
		Token:       token,
		SetupMethod: req.SetupMethod,
		Status:      "pending",
	}

	if req.SetupMethod == "automatic" {
		go s.performAutomaticSetup(agent, req)
		response.Message = "Automatic installation started"
	} else {
		script := s.generateInstallScript(token)
		response.InstallScript = script
		response.Message = "Manual installation script generated"
	}

	return response, nil
}

func (s *AgentService) performAutomaticSetup(agent *models.Agent, req *models.AgentSetupRequest) {
	log.Printf("Starting automatic setup for agent %d", agent.ID)
	
	// Update status to installing
	s.agentRepo.UpdateStatus(agent.ID, "installing", "Starting automatic installation...")

	// Generate install script
	script := s.generateInstallScript(agent.Token)
	
	// Connect via SSH and execute
	var err error
	if req.SSHKey != "" {
		err = s.sshService.ExecuteScriptWithKey(req.IP, req.Username, req.SSHKey, script, req.Port)
	} else {
		err = s.sshService.ExecuteScript(req.IP, req.Username, req.Password, script, req.Port)
	}

	if err != nil {
		log.Printf("Automatic setup failed for agent %d: %v", agent.ID, err)
		s.agentRepo.UpdateStatus(agent.ID, "failed", fmt.Sprintf("Installation failed: %v", err))
		return
	}

	s.agentRepo.UpdateStatus(agent.ID, "installed", "Agent installed successfully via SSH")
	log.Printf("Automatic setup completed for agent %d", agent.ID)
}

func (s *AgentService) generateInstallScript(token string) string {
	return fmt.Sprintf(`#!/bin/bash
set -e

echo "Installing KineticOps Agent..."

# Create agent directory
mkdir -p ~/.kineticops
cd ~/.kineticops

# Download agent binary (placeholder - in production this would download from your server)
cat > agent << 'EOF'
#!/bin/bash
# KineticOps Agent v1.0
TOKEN="%s"
SERVER_URL="http://localhost:8080"

while true; do
    # Collect system metrics
    CPU_USAGE=$(top -bn1 | grep "Cpu(s)" | awk '{print $2}' | sed 's/%%us,//')
    MEMORY_USAGE=$(free | grep Mem | awk '{printf("%.1f", $3/$2 * 100.0)}')
    DISK_USAGE=$(df -h / | awk 'NR==2{printf("%.1f", $5)}' | sed 's/%%//')
    
    # Get running services
    SERVICES=$(ps aux | grep -E "(nginx|apache|mysql|postgres|redis|docker)" | grep -v grep | awk '{print $11}' | sort | uniq | head -10 | tr '\n' ',' | sed 's/,$//')
    
    # Send heartbeat
    curl -s -X POST "$SERVER_URL/api/v1/agents/heartbeat" \
        -H "Content-Type: application/json" \
        -d "{
            \"token\": \"$TOKEN\",
            \"cpu_usage\": ${CPU_USAGE:-0},
            \"memory_usage\": ${MEMORY_USAGE:-0},
            \"disk_usage\": ${DISK_USAGE:-0},
            \"services\": [],
            \"system_info\": {
                \"hostname\": \"$(hostname)\",
                \"os\": \"$(uname -s)\",
                \"kernel\": \"$(uname -r)\",
                \"uptime\": \"$(uptime -p)\"
            }
        }" || true
    
    sleep 30
done
EOF

chmod +x agent

# Create systemd service
sudo tee /etc/systemd/system/kineticops-agent.service > /dev/null << EOF
[Unit]
Description=KineticOps Monitoring Agent
After=network.target

[Service]
Type=simple
User=$USER
WorkingDirectory=$HOME/.kineticops
ExecStart=$HOME/.kineticops/agent
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable kineticops-agent
sudo systemctl start kineticops-agent

echo "KineticOps Agent installed and started successfully!"
echo "Service status:"
sudo systemctl status kineticops-agent --no-pager
`, token)
}

func (s *AgentService) ProcessHeartbeat(heartbeat *models.AgentHeartbeat) error {
	return s.agentRepo.UpdateHeartbeat(heartbeat.Token, heartbeat)
}

func (s *AgentService) GetAgentStatus(id int) (*models.Agent, error) {
	return s.agentRepo.GetByHostID(id)
}

func (s *AgentService) GetHostServices(hostID int) ([]models.HostService, error) {
	return s.agentRepo.GetServices(hostID)
}

func (s *AgentService) generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}