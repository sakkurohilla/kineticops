package handlers

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sakkurohilla/kineticops/backend/internal/models"
	"github.com/sakkurohilla/kineticops/backend/internal/repository/postgres"
)

// ServeInstallScript returns a minimal bootstrapper that decodes a base64
// encoded inner installer and executes it. Keeping the inner payload small
// avoids heredoc/quoting issues when piping to sh on target hosts.
func ServeInstallScript(c *fiber.Ctx) error {
	// token may be empty; installer will prompt interactively if not provided.
	token := c.Query("token")

	scheme := "http"
	if c.Get("X-Forwarded-Proto") == "https" || c.Protocol() == "https" {
		scheme = "https"
	}
	host := fmt.Sprintf("%s://%s", scheme, c.Get("Host"))

	// The inner installer is the actual bash payload we want executed on target
	// hosts. We embed it as a Go raw string (no backticks inside) then encode
	// it to base64 before emitting a small POSIX wrapper. This avoids Go parse
	// problems and makes the served wrapper safe to pipe to /bin/sh.
	inner := fmt.Sprintf(`#!/usr/bin/env bash
set -euo pipefail
trap 'rm -f /tmp/kineticops_install.sh' EXIT
KINETICOPS_HOST="%s"
INSTALLATION_TOKEN="%s"
TARGET_OS="%s"
INSTALL_DIR="/opt/kineticops-agent"
mkdir -p "$INSTALL_DIR" /var/log/kineticops-agent
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
if [[ "$OS" != "linux" ]]; then
	echo "Non-Linux OS ($OS) unsupported"
	exit 1
fi
ARCH=$(uname -m)
case "$ARCH" in
x86_64) ARCH=amd64 ;;
aarch64|arm64) ARCH=arm64 ;;
arm*) ARCH=arm64 ;;  # Fallback for armv7 etc.
*) echo "Unsupported arch: $ARCH"; exit 1 ;;
esac
# If no token was provided in the URL, prompt the user interactively (silent input).
if [ -z "$INSTALLATION_TOKEN" ]; then
	echo "No installation token provided. Please enter your installation token."
	printf "Token: "
	read -s -r INSTALLATION_TOKEN
	echo
	if [ -z "$INSTALLATION_TOKEN" ]; then
		echo "No token provided; aborting."
		exit 1
	fi
fi
# Prompt for target OS type when not supplied.
if [ -z "$TARGET_OS" ]; then
	echo "Which OS are you installing for?"
	echo "1) ubuntu"
	echo "2) centos"
	echo "3) other (uses generic linux)"
	printf "Choose [1-3]: "
	read -r choice
	case "$choice" in
		1) TARGET_OS=ubuntu ;;
		2) TARGET_OS=centos ;;
		*) TARGET_OS=other ;;
	esac
fi
# Use TARGET_OS if set, else fall back to OS (e.g., "linux")
OS_PART="${TARGET_OS:-$OS}"
if [[ "$TARGET_OS" == "other" ]]; then
	OS_PART="$OS"
fi
try_names() {
	os_part="$1"
	arch_part="$2"
	# Prefer the short 'agent-<os>-<arch>' artifact names that we publish
	echo "agent-$os_part-$arch_part"
	echo "agent_${os_part}_${arch_part}"
	# Backwards-compatible long name variants
	echo "kineticops-agent-$os_part-$arch_part"
	echo "kineticops-agent_${os_part}_${arch_part}"
}
DL=""
for name in $(try_names "$OS_PART" "$ARCH"); do
	url="$KINETICOPS_HOST/api/v1/install/$name"
	echo "Trying artifact: $name at $url"
	if command -v curl >/dev/null 2>&1; then
		if curl -sfS --head "$url" >/dev/null 2>&1; then
			DL="$name"
			break
		fi
	elif command -v wget >/dev/null 2>&1; then
		if wget -q --spider "$url" >/dev/null 2>&1; then
			DL="$name"
			break
		fi
	fi
done
if [ -z "$DL" ]; then
	echo "No suitable agent binary available for $OS_PART/$ARCH"
	exit 1
fi
echo "Selected artifact: $DL"
BIN_PATH="$INSTALL_DIR/kineticops-agent"
# Also create a stable symlink name for systemd units and other installers
LINK_PATH="$INSTALL_DIR/agent"
download() {
	url="$1"
	echo "Downloading from $url"
	if command -v curl >/dev/null 2>&1; then
		if ! curl -fsSL "$url" -o "$BIN_PATH"; then
			echo "Download failed for $url"
			return 1
		fi
	else
		if ! wget -q "$url" -O "$BIN_PATH"; then
			echo "Download failed for $url"
			return 1
		fi
	fi
	echo "Downloaded to $BIN_PATH"
}
verify_with_gpg() {
	# Fetch public key and signature, import key to temporary GNUPGHOME and verify
	pub_url="$KINETICOPS_HOST/api/v1/install/file/public.key"
	sig_url="$KINETICOPS_HOST/api/v1/install/file/$1.asc"
	tmpdir=$(mktemp -d)
	export GNUPGHOME="$tmpdir"
	if command -v curl >/dev/null 2>&1; then
		curl -sSL "$pub_url" -o "$tmpdir/public.key" || true
		curl -sSL "$sig_url" -o "$BIN_PATH.asc" || true
	else
		wget -q "$pub_url" -O "$tmpdir/public.key" || true
		wget -q "$sig_url" -O "$BIN_PATH.asc" || true
	fi
	if [ -f "$tmpdir/public.key" ]; then
		gpg --import "$tmpdir/public.key" >/dev/null 2>&1 || true
	else
		rm -rf "$tmpdir"
		return 1
	fi
	if [ -f "$BIN_PATH.asc" ]; then
		if gpg --verify "$BIN_PATH.asc" "$BIN_PATH" >/dev/null 2>&1; then
			rm -rf "$tmpdir" "$BIN_PATH.asc"
			return 0
		fi
		rm -f "$BIN_PATH.asc"
	fi
	rm -rf "$tmpdir"
	return 1
}
checksum_ok() {
	sum_url="$KINETICOPS_HOST/api/v1/install/file/$1.sha256"
	expected=$(curl -sSL "$sum_url" | awk '{print $1}' || true)
	if [ -z "$expected" ]; then
		# No checksum published -> fail safe
		echo "Checksum not available for $1"
		return 1
	fi
	actual=$(sha256sum "$BIN_PATH" 2>/dev/null | awk '{print $1}' || true)
	if [ "$actual" = "$expected" ]; then
		return 0
	else
		echo "Checksum mismatch: expected $expected, got $actual"
		return 1
	fi
}
if ! download "$KINETICOPS_HOST/api/v1/install/$DL"; then
	exit 1
fi
# Prefer GPG verification if gpg is available and signatures are published
if command -v gpg >/dev/null 2>&1; then
	if verify_with_gpg "$DL"; then
		echo "GPG signature verified for $DL"
	elif checksum_ok "$DL"; then
		echo "GPG verification failed or not available; checksum OK for $DL"
	else
		echo "Verification failed for $DL"
		exit 1
	fi
else
	if checksum_ok "$DL"; then
		echo "Checksum OK for $DL (GPG unavailable)"
	else
		echo "Checksum not available or mismatch for $DL"
		exit 1
	fi
fi
chmod +x "$BIN_PATH"
# Create a stable symlink so other installers or systemd units that expect
# /opt/kineticops-agent/agent will work regardless of the exact binary name.
ln -sf "$BIN_PATH" "$LINK_PATH" || { echo "Symlink failed"; exit 1; }
echo "Installed agent to $BIN_PATH (symlinked to $LINK_PATH)"
# Write a minimal agent configuration so the installed agent knows how to
# reach this backend. Use an unquoted heredoc so shell variables are
# expanded (KINETICOPS_HOST and INSTALLATION_TOKEN set earlier in the
# script). Use spaces for indentation (YAML disallows tabs).
cat > "$INSTALL_DIR/config.yaml" <<YAML
output:
	kineticops:
		hosts:
			- "${KINETICOPS_HOST}"
		token: "${INSTALLATION_TOKEN}"
		timeout: 30s
		max_retry: 4
logging:
	level: info
# Pipeline controls how frequently the agent batches and sends metrics.
pipeline:
	# send at least every 15s (lower for test/dev, higher for production if desired)
	batch_time: 15s
	# smaller batch size helps ensure regular sends even on low volume hosts
	batch_size: 100
# Metrics inputs — enable common system metrics by default so agents
# continuously collect CPU, memory, disk and network stats.
metrics:
	enabled: true
	inputs:
		- system
		- cpu
		- memory
		- disk
		- network
		- load
YAML
chmod 0644 "$INSTALL_DIR/config.yaml"
# Also write a copy to the agent's default config location so agents
# that read /etc/kineticops-agent/config.yaml (agent default) will
# pick up the same configuration. This is best-effort and won't
# overwrite an existing /etc file unless the installer runs as root.
if [ "$(id -u)" -eq 0 ]; then
	mkdir -p /etc/kineticops-agent || true
	cp -f "$INSTALL_DIR/config.yaml" /etc/kineticops-agent/config.yaml || true
	chmod 0644 /etc/kineticops-agent/config.yaml || true
fi
## Create a systemd service unit that points at the installed binary so the
## agent can be managed as a service. Prefer writing the unit directly when
## running as root; fall back to sudo+tee when the script is executed under a
## non-root user piping into sudo. This avoids permission issues when the
## installer is invoked in different ways (curl | sudo bash vs curl | bash).
if [ "$(id -u)" -eq 0 ]; then
	cat > /etc/systemd/system/kineticops-agent.service << 'EOF'
[Unit]
Description=KineticOps Agent
After=network.target
[Service]
Type=simple
User=root
WorkingDirectory=/opt/kineticops-agent
ExecStart=/opt/kineticops-agent/agent -c /opt/kineticops-agent/config.yaml
Restart=always
RestartSec=30
StandardOutput=journal
StandardError=journal
[Install]
WantedBy=multi-user.target
EOF
else
	sudo tee /etc/systemd/system/kineticops-agent.service > /dev/null << 'EOF'
[Unit]
Description=KineticOps Agent
After=network.target
[Service]
Type=simple
User=root
WorkingDirectory=/opt/kineticops-agent
ExecStart=/opt/kineticops-agent/agent -c /opt/kineticops-agent/config.yaml
Restart=always
RestartSec=30
StandardOutput=journal
StandardError=journal
[Install]
WantedBy=multi-user.target
EOF
fi
# Reload systemd and try to enable/start the service if systemctl exists.
if command -v systemctl >/dev/null 2>&1; then
	if [ "$(id -u)" -eq 0 ]; then
		systemctl daemon-reload || true
		systemctl enable kineticops-agent || true
		systemctl restart kineticops-agent || true
	else
		sudo systemctl daemon-reload || true
		sudo systemctl enable kineticops-agent || true
		sudo systemctl restart kineticops-agent || true
	fi
	echo "kineticops-agent service enabled and restarted (if systemd available)"
else
	echo "systemd not detected; service file created but not started"
fi
`, host, token, c.Query("target_os"))
	encoded := base64.StdEncoding.EncodeToString([]byte(inner))
	// Emit a POSIX-compatible wrapper that writes the inner installer to /tmp
	// and prefers to run it under bash with pipefail when bash exists. This
	// keeps strict behavior on systems with bash while remaining compatible
	// with /bin/sh (dash) when bash is unavailable.
	wrapper := fmt.Sprintf(`#!/bin/sh
set -eu
# Write inner installer (base64) to temporary file
cat <<'BASE64' | base64 -d > /tmp/kineticops_install.sh
%s
BASE64
chmod +x /tmp/kineticops_install.sh
# If bash is available, run the inner installer under bash with pipefail for
# stricter failure handling. Otherwise execute the installer directly.
if command -v bash >/dev/null 2>&1; then
	exec bash -c 'set -euo pipefail; exec /tmp/kineticops_install.sh "$@"' -- "$@"
else
	exec /tmp/kineticops_install.sh "$@"
fi
`, encoded)
	c.Set("Content-Type", "text/plain; charset=utf-8")
	return c.SendString(wrapper)
}

// ServeAgentBinary serves agent binaries by full artifact name (e.g., /install/agent-ubuntu-amd64).
// Updated to match script URL pattern.
func ServeAgentBinary(c *fiber.Ctx) error {
	name := c.Params("name") // Now :name instead of :os/:arch
	if name == "" {
		return c.Status(400).SendString("agent name required")
	}
	// Build a list of candidate paths to look for the artifact. Prefer
	// locations relative to the running executable (useful in container
	// deployments), then the process working directory, then a few common
	// absolute locations.
	candidates := []string{}

	// Add executable-relative build directory if possible
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(exeDir, "build", name),
			filepath.Join(exeDir, "build", strings.ReplaceAll(name, "_", "-")),
		)
	}

	// Add paths relative to current working directory
	candidates = append(candidates,
		filepath.Join("build", name),
		filepath.Join("build", strings.ReplaceAll(name, "_", "-")),
	)

	// Common absolute locations inside containers / typical deployments
	candidates = append(candidates,
		filepath.Join("/app", "build", name),
		filepath.Join("/build", name),
		filepath.Join("/usr/local/bin", name),
	)

	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return c.SendFile(p)
		}
	}
	return c.Status(404).SendString(fmt.Sprintf("Agent binary %s not available", name))
}

// ServeArtifact serves any file located in the build/ directory by name.
// This is used for serving checksum files like agent-... .sha256 alongside
// binaries.
func ServeArtifact(c *fiber.Ctx) error {
	name := c.Params("name")
	if name == "" {
		return c.Status(400).SendString("artifact name required")
	}
	// Look in executable-relative build dir, working dir build/, and
	// common absolute locations. This mirrors ServeAgentBinary's behavior.
	candidates := []string{}
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(exeDir, "build", name),
		)
	}
	candidates = append(candidates,
		filepath.Join("build", name),
		filepath.Join("/app", "build", name),
		filepath.Join("/build", name),
		filepath.Join("/usr/local/bin", "build", name),
	)
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return c.SendFile(p)
		}
	}
	return c.Status(404).SendString("artifact not found")
}

// GenerateInstallationToken creates and stores a short-lived installation token
// used by installers to bootstrap agents.
func GenerateInstallationToken(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
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
	// Allow client to request a token scoped to a target OS (ubuntu/centos/other)
	var body struct {
		TargetOS string `json:"target_os"`
	}
	// Best-effort parse JSON body; fall back to query param
	if err := c.BodyParser(&body); err != nil {
		// Body may be empty or not JSON; log for diagnostics but continue
		fmt.Printf("install_handler: BodyParser failed: %v\n", err)
	}
	if body.TargetOS == "" {
		body.TargetOS = c.Query("target_os")
	}
	token := fmt.Sprintf("install_%d_%d", uid, c.Context().Time().Unix())
	installToken := models.InstallationToken{
		Token:     token,
		UserID:    uint(uid),
		TenantID:  uint(uid),
		ExpiresAt: time.Now().Add(24 * time.Hour),
		Used:      false,
		TargetOS:  body.TargetOS,
	}
	if err := postgres.DB.Create(&installToken).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create installation token"})
	}
	scheme := "http"
	if c.Get("X-Forwarded-Proto") == "https" || c.Protocol() == "https" {
		scheme = "https"
	}
	// Allow operators to explicitly set the external URL used in generated
	// install commands via the EXTERNAL_URL environment variable (full URL
	// including scheme, e.g. https://example.com). As a fallback, honor
	// PUBLIC_HOST (host[:port]) or the request Host header. This avoids
	// generating curl commands that point at localhost when the backend is
	// being accessed via a public address or reverse proxy.
	var hostURL string
	if ext := os.Getenv("EXTERNAL_URL"); ext != "" {
		hostURL = ext
	} else if pub := os.Getenv("PUBLIC_HOST"); pub != "" {
		if strings.HasPrefix(pub, "http://") || strings.HasPrefix(pub, "https://") {
			hostURL = pub
		} else {
			hostURL = fmt.Sprintf("%s://%s", scheme, pub)
		}
	} else {
		requestHost := c.Get("Host")
		if requestHost == "" {
			requestHost = "localhost:8080"
		}
		hostURL = fmt.Sprintf("%s://%s", scheme, requestHost)
	}
	// Ensure hostURL has no trailing slash to avoid double slashes
	hostURL = strings.TrimRight(hostURL, "/")
	// URL-escape token and target OS values to avoid breaking the generated
	// curl command when special characters are present.
	escToken := url.QueryEscape(token)
	cmd := fmt.Sprintf("curl -sSL %s/api/v1/install/agent.sh?token=%s", hostURL, escToken)
	if installToken.TargetOS != "" {
		cmd = fmt.Sprintf("%s&target_os=%s", cmd, url.QueryEscape(installToken.TargetOS))
	}
	cmd = fmt.Sprintf("%s | sudo bash", cmd)
	return c.JSON(fiber.Map{
		"token":        token,
		"command":      cmd,
		"expires_in":   86400,
		"instructions": "Run this command on your target server as root",
	})
}
