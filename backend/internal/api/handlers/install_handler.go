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
set -euo pipefail

KINETICOPS_HOST="%s"
INSTALL_DIR="/opt/kineticops-agent"
CONFIG_DIR="/etc/kineticops-agent"
SERVICE_NAME="kineticops-agent"
INSTALLATION_TOKEN="%s"

PROMTAIL_USER=${PROMTAIL_USER:-promtail}
PROMTAIL_BIN=${PROMTAIL_BIN:-/usr/local/bin/promtail}
PROMTAIL_POS_DIR=/var/lib/promtail
PROMTAIL_CONFIG_DST="/etc/promtail/promtail.yaml"
PROMTAIL_SYSTEMD="/etc/systemd/system/promtail.service"

echo "🚀 Installing KineticOps agents (kineticops-agent + promtail)..."

# Ensure script runs as root
if [[ $EUID -ne 0 ]]; then
	echo "❌ Please run as root (use sudo)"
	exit 1
fi

# Detect OS/ARCH
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case $ARCH in
	x86_64) ARCH="amd64" ;;
	aarch64|arm64) ARCH="arm64" ;;
	*) echo "❌ Unsupported arch: $ARCH"; exit 1 ;;
esac

echo "✅ Detected OS: $OS, ARCH: $ARCH"

# Create directories for agent
mkdir -p "$INSTALL_DIR"
mkdir -p "$CONFIG_DIR"
mkdir -p "/var/log/kineticops-agent"

echo "📥 Downloading kineticops-agent binary..."
if command -v curl >/dev/null 2>&1; then
	curl -sSL "$KINETICOPS_HOST/api/v1/install/agent-$OS-$ARCH" -o "$INSTALL_DIR/kineticops-agent"
elif command -v wget >/dev/null 2>&1; then
	wget -q "$KINETICOPS_HOST/api/v1/install/agent-$OS-$ARCH" -O "$INSTALL_DIR/kineticops-agent"
else
	echo "❌ curl or wget required to download agent binary"
	exit 1
fi
chmod +x "$INSTALL_DIR/kineticops-agent"

# Bootstrap request to backend to get loki_url and agent token
echo "🔐 Requesting agent bootstrap from backend to obtain Loki URL and agent token..."
BOOT_RESP=$(mktemp)
BOOT_PAYLOAD="{\"hostname\": \"$(hostname)\"}"
if [[ -n "${CREATE_HOST:-}" && -n "${REG_SECRET:-}" ]]; then
	BOOT_PAYLOAD="{\"hostname\": \"$(hostname)\", \"create_if_missing\": true, \"reg_secret\": \"${REG_SECRET}\"}"
fi
if command -v curl >/dev/null 2>&1; then
	if curl -sS -X POST -H "Content-Type: application/json" -d "$BOOT_PAYLOAD" "$KINETICOPS_HOST/api/v1/agents/bootstrap" -o "$BOOT_RESP"; then
		# parse JSON (jq preferred, python fallback)
		parse_field(){
			field=$1; file=$2
			if command -v jq >/dev/null 2>&1; then
				jq -r ".${field} // empty" "$file" || true
			elif command -v python3 >/dev/null 2>&1; then
				python3 - <<PY
import json,sys
try:
	obj=json.load(open('$file'))
	print(obj.get('$field',''))
except:
	sys.exit(0)
PY
			else
				grep -oP '"'"${field}""'\s*:\s*"\K[^"]+' "$file" 2>/dev/null || true
			fi
		}
		LOKI_URL=$(parse_field loki_url "$BOOT_RESP")
		AGENT_TOKEN=$(parse_field token "$BOOT_RESP")
	else
		echo "⚠️  Bootstrap request failed; proceeding with install but Promtail will need manual config"
	fi
else
	echo "⚠️  curl not available; cannot call bootstrap endpoint. Promtail will need manual LOKI_URL/TOKEN"
fi
rm -f "$BOOT_RESP"

# Create kineticops-agent config using agent token if provided, otherwise fall back to installation token
AGENT_AUTH_TOKEN=${AGENT_TOKEN:-$INSTALLATION_TOKEN}
echo "⚙️  Writing agent config to $CONFIG_DIR/config.yaml"
cat > "$CONFIG_DIR/config.yaml" <<EOF
agent:
	name: "kineticops-agent"
	hostname: "$(hostname)"
	period: 30s

output:
	kineticops:
		hosts: ["$KINETICOPS_HOST"]
		token: "$AGENT_AUTH_TOKEN"
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

# Create systemd service for kineticops-agent
cat > "/etc/systemd/system/$SERVICE_NAME.service" <<EOF
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

echo "� Installing Promtail (for Loki collection)..."
# Create promtail dirs
mkdir -p "$PROMTAIL_POS_DIR"
mkdir -p "$(dirname $PROMTAIL_CONFIG_DST)"

# Try to extract promtail binary from Docker image if not present
if [[ ! -x "$PROMTAIL_BIN" ]]; then
	if command -v docker >/dev/null 2>&1; then
		CONTAINER_ID=$(docker create grafana/promtail:latest /bin/true)
		if docker cp "$CONTAINER_ID":/usr/bin/promtail "$PROMTAIL_BIN" 2>/dev/null || docker cp "$CONTAINER_ID":/bin/promtail "$PROMTAIL_BIN" 2>/dev/null; then
			chmod +x "$PROMTAIL_BIN"
		else
			echo "⚠️  Could not extract promtail binary from image; please install promtail manually"
		fi
		docker rm "$CONTAINER_ID" >/dev/null || true
	else
		echo "⚠️  Docker not available to extract promtail. Please install promtail binary manually at $PROMTAIL_BIN"
	fi
fi

# Write a minimal Promtail config; if LOKI_URL present from bootstrap, embed it
echo "⚙️  Writing Promtail config to $PROMTAIL_CONFIG_DST"
cat > "$PROMTAIL_CONFIG_DST" <<EOF
server:
	http_listen_port: 9080
	grpc_listen_port: 0

positions:
	filename: "$PROMTAIL_POS_DIR/positions.yaml"

clients:
	- url: "${LOKI_URL:-http://loki:3100/loki/api/v1/push}"
EOF

if [[ -n "${AGENT_TOKEN:-}" ]]; then
	cat >> "$PROMTAIL_CONFIG_DST" <<EOF
		headers:
			Authorization: "Bearer ${AGENT_TOKEN}"
EOF
fi

# Create simple promtail systemd unit
cat > "$PROMTAIL_SYSTEMD" <<EOF
[Unit]
Description=Promtail log collector
After=network.target

[Service]
User=$PROMTAIL_USER
ExecStart=$PROMTAIL_BIN -config.file=$PROMTAIL_CONFIG_DST
Restart=always

[Install]
WantedBy=multi-user.target
EOF

# Set ownership and start services
chown -R root:root "$INSTALL_DIR" || true
chmod +x "$INSTALL_DIR/kineticops-agent" || true
systemctl daemon-reload
systemctl enable --now "$SERVICE_NAME" || systemctl restart "$SERVICE_NAME" || true
systemctl enable --now promtail || systemctl restart promtail || true

echo ""
echo "✅ Installation finished. Services status:"
systemctl is-active --quiet "$SERVICE_NAME" && echo "- $SERVICE_NAME: running" || echo "- $SERVICE_NAME: not running"
systemctl is-active --quiet promtail && echo "- promtail: running" || echo "- promtail: not running"

echo "Config locations: $CONFIG_DIR/config.yaml and $PROMTAIL_CONFIG_DST"
echo "If Promtail did not start, install promtail binary at $PROMTAIL_BIN or ensure docker is available to extract it."
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
