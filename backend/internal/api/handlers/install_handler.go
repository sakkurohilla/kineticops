package handlers

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sakkurohilla/kineticops/backend/internal/models"
	"github.com/sakkurohilla/kineticops/backend/internal/repository/postgres"
)

func ServeInstallScript(c *fiber.Ctx) error {
	token := c.Query("token")
	if token == "" {
		return c.Status(400).SendString("Token required: add ?token=your_token")
	}

	// Get the backend host from request - use actual server IP/hostname
	scheme := "http"
	if c.Get("X-Forwarded-Proto") == "https" || c.Protocol() == "https" {
		scheme = "https"
	}

	// Get the actual host from the request header, but replace localhost with actual IP
	requestHost := c.Get("Host")
	if requestHost == "localhost:8080" || requestHost == "127.0.0.1:8080" {
		// Try to get the actual network IP
		requestHost = "192.168.2.54:8080" // Actual server IP
	}
	host := fmt.Sprintf("%s://%s", scheme, requestHost)

	script := fmt.Sprintf(`#!/bin/bash
set -e

KINETICOPS_HOST="%s"
INSTALL_DIR="/opt/kineticops-agent"
CONFIG_DIR="/etc/kineticops-agent"
SERVICE_NAME="kineticops-agent"
INSTALLATION_TOKEN="%s"

echo "🚀 Installing KineticOps Agent..."

# Check if running as root
if [[ $EUID -ne 0 ]]; then
   echo "❌ This script must be run as root (use sudo)"
   exit 1
fi

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) echo "❌ Unsupported architecture: $ARCH"; exit 1 ;;
esac

echo "✅ Detected OS: $OS, Architecture: $ARCH"

# Create directories
mkdir -p "$INSTALL_DIR"
mkdir -p "$CONFIG_DIR"
mkdir -p "/var/log/kineticops-agent"

# Download agent binary from your backend
echo "📥 Downloading agent binary..."
if command -v curl >/dev/null 2>&1; then
    curl -sSL "$KINETICOPS_HOST/api/v1/install/agent-$OS-$ARCH" -o "$INSTALL_DIR/kineticops-agent"
elif command -v wget >/dev/null 2>&1; then
    wget -q "$KINETICOPS_HOST/api/v1/install/agent-$OS-$ARCH" -O "$INSTALL_DIR/kineticops-agent"
else
    echo "❌ Neither curl nor wget found. Please install one of them."
    exit 1
fi

chmod +x "$INSTALL_DIR/kineticops-agent"

# Create configuration file
echo "⚙️  Creating configuration..."
cat > "$CONFIG_DIR/config.yaml" << EOF
agent:
  name: "kineticops-agent"
  hostname: "$(hostname)"
  period: 30s

output:
  kineticops:
    hosts: ["$KINETICOPS_HOST"]
    token: "$INSTALLATION_TOKEN"
    timeout: 30s
    max_retry: 3

modules:
  system:
    enabled: true
    period: 30s
    cpu:
      enabled: true
    memory:
      enabled: true
    network:
      enabled: true
    filesystem:
      enabled: true

logging:
  level: "info"
  to_file: true
  file: "/var/log/kineticops-agent/agent.log"
EOF

# Create systemd service
echo "🔧 Creating systemd service..."
cat > "/etc/systemd/system/$SERVICE_NAME.service" << EOF
[Unit]
Description=KineticOps Monitoring Agent
After=network.target

[Service]
Type=simple
User=root
ExecStart=$INSTALL_DIR/kineticops-agent -c $CONFIG_DIR/config.yaml
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

# Enable and start service
echo "🚀 Starting service..."
systemctl daemon-reload
systemctl enable "$SERVICE_NAME"
systemctl start "$SERVICE_NAME"

# Check status
sleep 2
if systemctl is-active --quiet "$SERVICE_NAME"; then
    echo "✅ KineticOps Agent installed and started successfully!"
    echo "📊 Service status: $(systemctl is-active $SERVICE_NAME)"
    echo "📝 View logs: journalctl -u $SERVICE_NAME -f"
    echo "⚙️  Config file: $CONFIG_DIR/config.yaml"
else
    echo "❌ Service failed to start. Check logs: journalctl -u $SERVICE_NAME"
    exit 1
fi

echo ""
echo "🎉 Installation completed successfully!"
echo "📈 Your host should appear in the KineticOps dashboard within 30 seconds."
echo "🌐 Dashboard: $KINETICOPS_HOST"
`, host, token)

	c.Set("Content-Type", "text/plain")
	return c.SendString(script)
}

func ServeAgentBinary(c *fiber.Ctx) error {
	osType := c.Params("os")
	arch := c.Params("arch")

	// Look for pre-built agent binary
	binaryPath := filepath.Join("/opt/kineticops/agent", fmt.Sprintf("kineticops-agent-%s-%s", osType, arch))

	// If not found, try to serve the local built binary
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		// Fallback to local agent binary (for development)
		localBinary := "/home/akash/kineticops/agent/kineticops-agent"
		if _, err := os.Stat(localBinary); err == nil {
			return c.SendFile(localBinary)
		}
		return c.Status(404).SendString("Agent binary not found. Please build the agent first.")
	}

	return c.SendFile(binaryPath)
}

func GenerateInstallationToken(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized - no user_id"})
	}

	// Convert userID to int64 safely
	var uid int64
	switch v := userID.(type) {
	case int64:
		uid = v
	case int:
		uid = int64(v)
	case float64:
		uid = int64(v)
	default:
		return c.Status(401).JSON(fiber.Map{"error": fmt.Sprintf("Invalid user_id type: %T", userID)})
	}

	// Generate a simple token for now
	token := fmt.Sprintf("install_%d_%d", uid, c.Context().Time().Unix())

	// Store the token in database
	installToken := models.InstallationToken{
		Token:     token,
		UserID:    uint(uid),
		TenantID:  uint(uid),
		ExpiresAt: time.Now().Add(24 * time.Hour),
		Used:      false,
	}
	if err := postgres.DB.Create(&installToken).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create installation token"})
	}

	// Get the backend host - use actual server IP/hostname
	scheme := "http"
	if c.Get("X-Forwarded-Proto") == "https" || c.Protocol() == "https" {
		scheme = "https"
	}

	// Get the actual host from the request header, but replace localhost with actual IP
	requestHost := c.Get("Host")
	if requestHost == "localhost:8080" || requestHost == "127.0.0.1:8080" {
		// Try to get the actual network IP
		requestHost = "192.168.2.54:8080" // Actual server IP
	}
	host := fmt.Sprintf("%s://%s", scheme, requestHost)

	command := fmt.Sprintf("curl -sSL %s/api/v1/install/agent.sh?token=%s | sudo bash", host, token)

	return c.JSON(fiber.Map{
		"token":        token,
		"command":      command,
		"expires_in":   86400, // 24 hours
		"instructions": "Run this command on your target server as root",
	})
}
