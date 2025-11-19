package services

//nolint:staticcheck // intentional usage of deprecated package for compatibility; see repository/postgres/connection.go
import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/sakkurohilla/kineticops/backend/internal/logging"
	"github.com/sakkurohilla/kineticops/backend/internal/models"
	"github.com/sakkurohilla/kineticops/backend/internal/repository/postgres"
)

type AgentService struct {
	agentRepo  *postgres.AgentRepository
	hostRepo   *postgres.HostRepository
	sshService *SSHService
}

func NewAgentService(agentRepo *postgres.AgentRepository, hostRepo *postgres.HostRepository, sshService *SSHService) *AgentService {
	return &AgentService{
		agentRepo:  agentRepo,
		hostRepo:   hostRepo,
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
					if derr := postgres.DeleteHost(postgres.DB, h.ID); derr != nil {
						logging.Warnf("failed to cleanup failed host %s (id=%d): %v", h.Hostname, h.ID, derr)
					} else {
						logging.Infof("Cleaned up failed host creation attempt: %s (%s)", h.Hostname, h.IP)
					}
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
		Hostname:    req.Hostname,
		IP:          req.IP,
		SSHUser:     req.Username,
		SSHPassword: req.Password,
		SSHKey:      req.SSHKey,
		SSHPort:     int64(req.Port),
		OS:          "linux",
		Group:       "default",
		AgentStatus: "installing",
		TenantID:    1, // Default tenant
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
	if err := s.agentRepo.CreateInstallLog(installLog); err != nil {
		logging.Warnf("failed to create install log for agent %d: %v", installLog.AgentID, err)
	}

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
	logging.Infof("Starting automatic setup for agent %d", agent.ID)

	// Update status to installing
	if err := s.agentRepo.UpdateStatus(agent.ID, "installing", "Starting automatic installation..."); err != nil {
		logging.Warnf("failed to update agent %d status to installing: %v", agent.ID, err)
	}

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
		logging.Errorf("Automatic setup failed for agent %d: %v", agent.ID, err)
		if uerr := s.agentRepo.UpdateStatus(agent.ID, "failed", fmt.Sprintf("Installation failed: %v", err)); uerr != nil {
			logging.Errorf("failed to mark agent %d failed after SSH error: %v", agent.ID, uerr)
		}
		return
	}

	if merr := s.agentRepo.MarkInstalled(agent.ID); merr != nil {
		logging.Errorf("failed to mark agent %d installed: %v", agent.ID, merr)
	}
	if uerr := s.agentRepo.UpdateStatus(agent.ID, "installed", "Agent installed successfully via SSH"); uerr != nil {
		logging.Errorf("failed to update agent %d status to installed: %v", agent.ID, uerr)
	}
	logging.Infof("Automatic setup completed for agent %d", agent.ID)
}

func (s *AgentService) GenerateAgentToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (s *AgentService) GenerateInstallScript(token string) string {
	// Use a template and replace a token placeholder to avoid fmt interpreting
	// % sequences in the script (awk/printf patterns use % formats).
	tpl := `#!/bin/bash
set -e

echo "Installing KineticOps Agent..."

# Create agent directory
sudo mkdir -p /opt/kineticops-agent
cd /opt/kineticops-agent

# Create agent script
sudo tee agent > /dev/null << 'EOF'
#!/bin/bash
# KineticOps Agent v1.0
TOKEN="{{TOKEN}}"
SERVER_URL="http://localhost:8080"

while true; do
	# Collect system metrics with improved accuracy and robustness
	# CPU usage as fraction (0..1) computed from /proc/stat
	CPU_USAGE=$(bash -lc 'read cpu user1 nice1 system1 idle1 rest < /proc/stat; sleep 1; read cpu user2 nice2 system2 idle2 rest < /proc/stat; prev_total=$((user1+nice1+system1+idle1)); total=$((user2+nice2+system2+idle2)); idle_diff=$((idle2-idle1)); total_diff=$((total-prev_total)); if [ "$total_diff" -gt 0 ]; then awk -v td="$total_diff" -v id="$idle_diff" "BEGIN { printf "%.4f", (td-id)/td }"; else echo "0"; fi' 2>/dev/null || echo "0")

	# Memory: get total, available, and used (in MB)
	read MEMORY_TOTAL_KB MEMORY_AVAILABLE_KB <<< $(awk '/MemTotal/ {total=$2} /MemAvailable/ {avail=$2} END {print total, avail}' /proc/meminfo 2>/dev/null || echo "0 0")
	MEMORY_TOTAL_MB=$(awk -v kb="$MEMORY_TOTAL_KB" 'BEGIN { printf "%.2f", kb/1024 }')
	MEMORY_AVAILABLE_MB=$(awk -v kb="$MEMORY_AVAILABLE_KB" 'BEGIN { printf "%.2f", kb/1024 }')
	MEMORY_USED_MB=$(awk -v tot="$MEMORY_TOTAL_KB" -v avail="$MEMORY_AVAILABLE_KB" 'BEGIN { printf "%.2f", (tot-avail)/1024 }')
	
	# Memory usage as percent (0..100)
	if [ "$MEMORY_TOTAL_KB" -gt 0 ]; then
		MEMORY_USAGE=$(awk -v tot="$MEMORY_TOTAL_KB" -v avail="$MEMORY_AVAILABLE_KB" 'BEGIN { printf "%.1f", (tot-avail)/tot*100 }')
	else
		MEMORY_USAGE=0
	fi

	# Disk usage: report both total and used bytes for the root mount so the
	# backend can compute a canonical percentage and avoid discrepancies.
	DISK_TOTAL_BYTES=$(df --output=size -B1 / | tail -1 | tr -dc '0-9' 2>/dev/null || echo "0")
	DISK_USED_BYTES=$(df --output=used -B1 / | tail -1 | tr -dc '0-9' 2>/dev/null || echo "0")
	# Also compute a best-effort percent for backward compatibility
	if [ "$DISK_TOTAL_BYTES" -gt 0 ]; then
		DISK_USAGE=$(awk -v used="$DISK_USED_BYTES" -v tot="$DISK_TOTAL_BYTES" 'BEGIN{ if(tot>0) printf "%.2f", used/tot*100; else print "0" }')
	else
		DISK_USAGE=0
	fi

	# Get ONLY user-installed services (exclude system defaults)
	# Filter out common system services that come with OS
	SERVICES=$(systemctl list-units --type=service --state=running --no-pager --no-legend | \
		awk '{print $1}' | sed 's/.service$//' | \
		grep -vE '^(systemd|dbus|accounts-daemon|acpid|atd|cron|irqbalance|multipathd|networkd|plymouth|polkit|rsyslog|snapd|ssh|udev|udisks2|wpa_supplicant|ModemManager|NetworkManager|bluetooth|avahi|cups|gdm|lightdm|thermald|unattended-upgrades|packagekit|apparmor|apport|kerneloops)' | \
		head -20 | tr '\n' ',' | sed 's/,$//' 2>/dev/null || echo "")
    
    # Send heartbeat
    curl -s -X POST "$SERVER_URL/api/v1/agents/heartbeat" \
        -H "Content-Type: application/json" \
        -d "{
            \"token\": \"$TOKEN\",
            \"cpu_usage\": ${CPU_USAGE:-0},
            \"memory_usage\": ${MEMORY_USAGE:-0},
            \"memory_total\": ${MEMORY_TOTAL_MB:-0},
            \"memory_used\": ${MEMORY_USED_MB:-0},
            \"disk_usage\": ${DISK_USAGE:-0},
			"services": [],
			"metadata": {
				"hostname": "$(hostname)",
				"os": "$(uname -s)",
				"kernel": "$(uname -r)",
				"uptime": $(awk '{print int($1)}' /proc/uptime)
			},
			"disk_total_bytes": ${DISK_TOTAL_BYTES:-0},
			"disk_used_bytes": ${DISK_USED_BYTES:-0}
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
`

	// Replace our token placeholder safely (script may contain % characters)
	// Note: using strings.ReplaceAll avoids fmt interpreting % sequences in the
	// template which would cause fmt.Sprintf to expect additional args.
	return strings.ReplaceAll(tpl, "{{TOKEN}}", token)

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
