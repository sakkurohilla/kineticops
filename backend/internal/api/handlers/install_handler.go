package handlers

import (
	"encoding/base64"
	"fmt"
	"os"
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

	// Prefer more specific artifact names (kineticops-agent-<os>-<arch>) and
	// perform GPG verification (preferred) then fall back to SHA256.
	inner := fmt.Sprintf(`#!/bin/sh
set -euo pipefail

KINETICOPS_HOST="%s"
INSTALLATION_TOKEN="%s"
TARGET_OS="%s"
INSTALL_DIR="/opt/kineticops-agent"

mkdir -p "$INSTALL_DIR" /var/log/kineticops-agent

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
	x86_64) ARCH=amd64 ;;
	aarch64|arm64) ARCH=arm64 ;;
	*) echo "Unsupported arch: $ARCH"; exit 1 ;;
esac

# If no token was provided in the URL, prompt the user interactively.
if [ -z "$INSTALLATION_TOKEN" ]; then
	echo "No installation token provided. Please enter your installation token."
	printf "Token: "
	read -r INSTALLATION_TOKEN
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
	echo "3) other"
	printf "Choose [1-3]: "
	read -r choice
	case "$choice" in
		1) TARGET_OS=ubuntu ;;
		2) TARGET_OS=centos ;;
		*) TARGET_OS=other ;;
	esac
fi

try_names() {
	os_part="$1"
	arch_part="$2"
	target="$3"
	if [ -n "$target" ]; then
		echo "kineticops-agent-$target-$os_part-$arch_part"
		echo "kineticops-agent_$target_${os_part}_${arch_part}"
	fi
	echo "kineticops-agent-$os_part-$arch_part"
	echo "agent-$os_part-$arch_part"
	echo "agent_${os_part}_${arch_part}"
}

DL=""
for name in $(try_names "$OS" "$ARCH" "$TARGET_OS"); do
	url="$KINETICOPS_HOST/api/v1/install/$name"
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
	echo "No suitable agent binary available for $OS/$ARCH"
	exit 1
fi

BIN_PATH="$INSTALL_DIR/kineticops-agent"

download() {
	url="$1"
	if command -v curl >/dev/null 2>&1; then
		curl -sSL "$url" -o "$BIN_PATH"
	else
		wget -q "$url" -O "$BIN_PATH"
	fi
}

verify_with_gpg() {
	# Fetch public key and signature, import key to temporary GNUPGHOME and verify
	pub_url="$KINETICOPS_HOST/api/v1/install/file/public.key"
	sig_url="$KINETICOPS_HOST/api/v1/install/$1.asc"
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
		return 1
	fi
	if [ -f "$BIN_PATH.asc" ]; then
		if gpg --verify "$BIN_PATH.asc" "$BIN_PATH" >/dev/null 2>&1; then
			rm -rf "$tmpdir" "$BIN_PATH.asc"
			return 0
		fi
	fi
	rm -rf "$tmpdir" "$BIN_PATH.asc" || true
	return 1
}

checksum_ok() {
	sum_url="$KINETICOPS_HOST/api/v1/install/$1.sha256"
	expected=$(curl -sSL "$sum_url" || true)
	if [ -z "$expected" ]; then
		# No checksum published -> fail safe
		echo "Checksum not available for $1"
		return 1
	fi
	actual=$(sha256sum "$BIN_PATH" 2>/dev/null | awk '{print $1}' || true)
	if [ "$actual" = "$expected" ]; then
		return 0
	fi
	return 1
}

download "$KINETICOPS_HOST/api/v1/install/$DL"
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
		echo "Checksum OK for $DL"
	else
		echo "Checksum not available or mismatch for $DL"
		exit 1
	fi
fi

chmod +x "$BIN_PATH"
echo "Installed agent to $BIN_PATH"
`, host, token, c.Query("target_os"))

	encoded := base64.StdEncoding.EncodeToString([]byte(inner))
	wrapper := fmt.Sprintf(`#!/bin/sh
set -euo pipefail
cat <<'BASE64' | base64 -d > /tmp/kineticops_install.sh
%s
BASE64
chmod +x /tmp/kineticops_install.sh
exec /tmp/kineticops_install.sh "$@"
`, encoded)

	c.Set("Content-Type", "text/plain; charset=utf-8")
	return c.SendString(wrapper)
}

// ServeAgentBinary is a placeholder that returns 404 for agent binaries.
// Implement actual binary serving from build artifacts if/when available.
func ServeAgentBinary(c *fiber.Ctx) error {
	osParam := c.Params("os")
	archParam := c.Params("arch")
	// Look for a built artifact in backend/build named agent-<os>-<arch> or agent_<os>_<arch>
	// e.g., backend/build/agent-linux-amd64 or backend/build/agent_linux_amd64
	paths := []string{
		fmt.Sprintf("build/agent-%s-%s", osParam, archParam),
		fmt.Sprintf("build/agent_%s_%s", osParam, archParam),
		fmt.Sprintf("build/kineticops-agent-%s-%s", osParam, archParam),
		fmt.Sprintf("build/kineticops-agent_%s_%s", osParam, archParam),
	}
	// Also check common absolute locations inside containers where the server may run
	absPaths := []string{
		fmt.Sprintf("/build/agent-%s-%s", osParam, archParam),
		fmt.Sprintf("/build/agent_%s_%s", osParam, archParam),
		fmt.Sprintf("/usr/local/bin/build/agent-%s-%s", osParam, archParam),
		fmt.Sprintf("/usr/local/bin/build/agent_%s_%s", osParam, archParam),
		fmt.Sprintf("/usr/local/bin/kineticops-agent-%s-%s", osParam, archParam),
		fmt.Sprintf("/app/build/agent-%s-%s", osParam, archParam),
	}
	for _, p := range paths {
		// serve if file exists
		if _, err := os.Stat(p); err == nil {
			return c.SendFile(p, true)
		}
	}
	for _, p := range absPaths {
		if _, err := os.Stat(p); err == nil {
			return c.SendFile(p, true)
		}
	}
	return c.Status(404).SendString(fmt.Sprintf("Agent binary for %s/%s not available", osParam, archParam))
}

// ServeArtifact serves any file located in the build/ directory by name.
// This is used for serving checksum files like agent-... .sha256 alongside
// binaries.
func ServeArtifact(c *fiber.Ctx) error {
	name := c.Params("name")
	if name == "" {
		return c.Status(400).SendString("artifact name required")
	}
	// look in build/ and common absolute locations
	candidates := []string{
		fmt.Sprintf("build/%s", name),
		fmt.Sprintf("/build/%s", name),
		fmt.Sprintf("/usr/local/bin/build/%s", name),
		fmt.Sprintf("/app/build/%s", name),
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return c.SendFile(p, true)
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
	_ = c.BodyParser(&body)
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
	requestHost := c.Get("Host")
	if requestHost == "" {
		requestHost = "localhost:8080"
	}
	host := fmt.Sprintf("%s://%s", scheme, requestHost)
	cmd := fmt.Sprintf("curl -sSL %s/api/v1/install/agent.sh?token=%s", host, token)
	if installToken.TargetOS != "" {
		cmd = fmt.Sprintf("%s&target_os=%s", cmd, installToken.TargetOS)
	}
	cmd = fmt.Sprintf("%s | sudo bash", cmd)
	return c.JSON(fiber.Map{
		"token":        token,
		"command":      cmd,
		"expires_in":   86400,
		"instructions": "Run this command on your target server as root",
	})
}
