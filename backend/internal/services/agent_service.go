package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"strings"

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
	// Clean up any failed previous attempts and check for duplicates
	existingHosts, err := postgres.ListHosts(postgres.DB, 0, 100, 0)
	if err == nil {
		for _, h := range existingHosts {
			// Check for duplicate hostname OR IP
			if h.Hostname == req.Hostname || h.IP == req.IP {
				if h.AgentStatus == "installing" || h.AgentStatus == "failed" || h.AgentStatus == "" {
					// Remove failed attempt
					postgres.DeleteHost(postgres.DB, h.ID)
					log.Printf("Cleaned up failed host creation attempt: %s (%s)", h.Hostname, h.IP)
				} else {
					if h.Hostname == req.Hostname {
						return nil, fmt.Errorf("host with hostname '%s' already exists", req.Hostname)
					}
					if h.IP == req.IP {
						return nil, fmt.Errorf("host with IP '%s' already exists (hostname: %s)", req.IP, h.Hostname)
					}
				}
			}
		}
	}

	// Test SSH connection BEFORE creating host
	if req.SetupMethod == "automatic" {
		var sshErr error
		if req.SSHKey != "" {
			sshErr = TestSSHConnectionWithKey(req.IP, req.Port, req.Username, "", req.SSHKey)
		} else {
			sshErr = TestSSHConnection(req.IP, req.Port, req.Username, req.Password)
		}
		if sshErr != nil {
			return nil, fmt.Errorf("SSH connection failed: %v", sshErr)
		}
	}

	// Create host only after SSH test passes
	host := &models.Host{
		Hostname:     req.Hostname,
		IP:           req.IP,
		SSHUser:      req.Username,
		SSHPassword:  req.Password,
		SSHKey:       req.SSHKey,
		SSHPort:      int64(req.Port),
		OS:           "linux",
		Group:        "default",
		AgentStatus:  "installing",
		TenantID:     1, // Default tenant
	}
	
	err = s.hostRepo.Create(host)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			if strings.Contains(err.Error(), "hostname") {
				return nil, fmt.Errorf("host with hostname '%s' already exists", req.Hostname)
			}
			if strings.Contains(err.Error(), "ip") {
				return nil, fmt.Errorf("host with IP '%s' already exists", req.IP)
			}
		}
		return nil, fmt.Errorf("failed to create host: %v", err)
	}

	// Generate unique token
	token, err := s.GenerateAgentToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %v", err)
	}

	// Create agent record
	agent := &models.Agent{
		HostID:      int(host.ID),
		AgentToken:  token,
		Status:      "pending",
		SetupMethod: req.SetupMethod,
	}
	
	err = s.agentRepo.Create(agent)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %v", err)
	}

	// Create installation log
	installLog := &models.AgentInstallationLog{
		AgentID:     agent.ID,
		SetupMethod: req.SetupMethod,
		Status:      "started",
		Logs:        "Installation initiated",
	}
	s.agentRepo.CreateInstallLog(installLog)

	response := &models.AgentSetupResponse{
		HostID:      int(host.ID),
		AgentID:     agent.ID,
		Token:       token,
		SetupMethod: req.SetupMethod,
		Status:      "pending",
	}

	if req.SetupMethod == "automatic" {
		go s.SetupAgentAutomatic(agent, req)
		response.Message = "Automatic installation started"
	} else {
		script := s.GenerateInstallScript(token)
		response.InstallScript = script
		response.Message = "Manual installation script generated"
	}

	return response, nil
}

func (s *AgentService) SetupAgentAutomatic(agent *models.Agent, req *models.AgentSetupRequest) {
	log.Printf("Starting automatic setup for agent %d", agent.ID)
	
	// Update status to installing
	s.agentRepo.UpdateStatus(agent.ID, "installing", "Starting automatic installation...")

	// Generate install script
	script := s.GenerateInstallScript(agent.AgentToken)
	
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

	s.agentRepo.MarkInstalled(agent.ID)
	s.agentRepo.UpdateStatus(agent.ID, "installed", "Agent installed successfully via SSH")
	log.Printf("Automatic setup completed for agent %d", agent.ID)
}

func (s *AgentService) GenerateAgentToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (s *AgentService) GenerateInstallScript(token string) string {
	return fmt.Sprintf(`#!/bin/bash
set -e

echo "Installing KineticOps Agent..."

# Create agent directory
sudo mkdir -p /opt/kineticops-agent
cd /opt/kineticops-agent

# Create agent script
sudo tee agent > /dev/null << 'EOF'
#!/bin/bash
# KineticOps Agent v1.0
TOKEN="%s"
SERVER_URL="http://localhost:8080"

while true; do
    # Collect system metrics with error handling
    CPU_USAGE=$(awk '/cpu /{u=$2+$4; t=$2+$3+$4+$5; if (NR==1){u1=u; t1=t;} else print (u-u1) * 100 / (t-t1); exit}' <(grep 'cpu ' /proc/stat; sleep 1; grep 'cpu ' /proc/stat) 2>/dev/null || echo "0")
    MEMORY_USAGE=$(free | awk '/^Mem:/{printf "%.1f", $3/$2 * 100.0}' 2>/dev/null || echo "0")
    DISK_USAGE=$(df / | awk 'NR==2{gsub(/%/,""); print $5}' 2>/dev/null || echo "0")
    
    # Get running services
    SERVICES=$(systemctl list-units --type=service --state=running --no-pager --no-legend | head -10 | awk '{print $1}' | sed 's/.service$//' | tr '\n' ',' | sed 's/,$//' 2>/dev/null || echo "")
    
    # Send heartbeat
    curl -s -X POST "$SERVER_URL/api/v1/agents/heartbeat" \
        -H "Content-Type: application/json" \
        -d "{
            \"token\": \"$TOKEN\",
            \"cpu_usage\": ${CPU_USAGE:-0},
            \"memory_usage\": ${MEMORY_USAGE:-0},
            \"disk_usage\": ${DISK_USAGE:-0},
            \"services\": [],
            \"metadata\": {
                \"hostname\": \"$(hostname)\",
                \"os\": \"$(uname -s)\",
                \"kernel\": \"$(uname -r)\",
                \"uptime\": \"$(cat /proc/uptime | awk '{print int($1)}')\"
            }
        }" 2>/dev/null || echo "Heartbeat failed"
    
    sleep 30
done
EOF

sudo chmod +x agent

# Create systemd service
sudo tee /etc/systemd/system/kineticops-agent.service > /dev/null << EOF
[Unit]
Description=KineticOps Monitoring Agent
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/kineticops-agent
ExecStart=/opt/kineticops-agent/agent
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable kineticops-agent
sudo systemctl start kineticops-agent

echo "KineticOps Agent installed and started successfully!"
echo "Service status:"
sudo systemctl status kineticops-agent --no-pager -l
`, token)
}

func (s *AgentService) RegisterAgentHeartbeat(heartbeat *models.AgentHeartbeat) error {
	return s.agentRepo.UpdateHeartbeat(heartbeat.Token, heartbeat)
}

func (s *AgentService) CheckAgentStatus(hostID int) (*models.Agent, error) {
	return s.agentRepo.GetByHostID(hostID)
}

func (s *AgentService) FetchServices(agentID int) ([]models.AgentService, error) {
	return s.agentRepo.GetServices(agentID)
}

func (s *AgentService) GetManualSetupInstructions(token string) string {
	script := s.GenerateInstallScript(token)
	return fmt.Sprintf(`Manual Installation Instructions:

1. Copy the following script to your host:
%s

2. Save it as: ~/.kineticops-install.sh
3. Make executable: chmod +x ~/.kineticops-install.sh
4. Run: bash ~/.kineticops-install.sh
5. Agent will start automatically and connect to server

The agent will appear online in your dashboard within 30 seconds.`, script)
}

func (s *AgentService) GetAgentStatus(id int) (*models.Agent, error) {
	return s.agentRepo.GetByHostID(id)
}

func (s *AgentService) GetHostServices(hostID int) ([]models.AgentService, error) {
	agent, err := s.agentRepo.GetByHostID(hostID)
	if err != nil {
		return nil, err
	}
	return s.agentRepo.GetServices(agent.ID)
}

