#!/usr/bin/env bash
set -euo pipefail

PROMTAIL_USER=${PROMTAIL_USER:-promtail}
PROMTAIL_BIN=${PROMTAIL_BIN:-/usr/local/bin/promtail}
PROMTAIL_CONFIG_SRC="$(pwd)/config/promtail-agent.yaml"
PROMTAIL_CONFIG_DST="/etc/promtail/promtail.yaml"
PROMTAIL_POS_DIR=/var/lib/promtail
SYSTEMD_UNIT_SRC="$(pwd)/packaging/promtail.service"
SYSTEMD_UNIT_DST=/etc/systemd/system/promtail.service

# Optional registration endpoint to fetch per-host Loki URL and token
# Example: REGISTER_URL=https://backend.example.com/api/v1/agents/register
REGISTER_URL=${REGISTER_URL:-}

# SIMULATE mode (non-destructive): set SIMULATE=1 to avoid sudo/systemctl and write to local paths
SIMULATE=${SIMULATE:-0}
if [[ "$SIMULATE" == "1" ]]; then
  echo "SIMULATE mode enabled: installer will not call sudo or systemctl and will write to local dirs"
fi

# When simulating, redirect system paths into a local temp tree to avoid permission errors
if [[ "$SIMULATE" == "1" ]]; then
  SIM_ROOT="$(pwd)/.promtail-sim"
  echo "SIMULATE root: $SIM_ROOT"
  mkdir -p "$SIM_ROOT"
  PROMTAIL_POS_DIR="$SIM_ROOT/var/lib/promtail"
  PROMTAIL_CONFIG_DST="$SIM_ROOT/etc/promtail/promtail.yaml"
  SYSTEMD_UNIT_DST="$SIM_ROOT/etc/systemd/promtail.service"
fi

echo "Installing promtail agent"

if [[ ! -f "$PROMTAIL_CONFIG_SRC" ]]; then
  echo "Config file not found at $PROMTAIL_CONFIG_SRC"
  exit 1
fi

# Create user
if ! id -u "$PROMTAIL_USER" >/dev/null 2>&1; then
  echo "Creating user $PROMTAIL_USER"
  if [[ "$SIMULATE" == "1" ]]; then
    echo "SIMULATE: would create system user $PROMTAIL_USER"
  else
    sudo useradd --system --no-create-home --shell /usr/sbin/nologin "$PROMTAIL_USER"
  fi
fi

# Create directories
# Create directories
if [[ "$SIMULATE" == "1" ]]; then
  mkdir -p "$PROMTAIL_POS_DIR"
  mkdir -p "$(dirname $PROMTAIL_CONFIG_DST)"
  chown -R "$PROMTAIL_USER:$PROMTAIL_USER" "$PROMTAIL_POS_DIR" 2>/dev/null || true
else
  sudo mkdir -p "$PROMTAIL_POS_DIR"
  sudo mkdir -p "$(dirname $PROMTAIL_CONFIG_DST)"
  sudo chown -R "$PROMTAIL_USER:$PROMTAIL_USER" "$PROMTAIL_POS_DIR"
fi

# Install config
echo "Installing config to $PROMTAIL_CONFIG_DST"
if [[ "$SIMULATE" == "1" ]]; then
  cp "$PROMTAIL_CONFIG_SRC" "$PROMTAIL_CONFIG_DST"
  chown "$PROMTAIL_USER:$PROMTAIL_USER" "$PROMTAIL_CONFIG_DST" 2>/dev/null || true
else
  sudo cp "$PROMTAIL_CONFIG_SRC" "$PROMTAIL_CONFIG_DST"
  sudo chown "$PROMTAIL_USER:$PROMTAIL_USER" "$PROMTAIL_CONFIG_DST"
fi

# If a REGISTER_URL is provided, call it to fetch per-host Loki URL and token.
if [[ -n "$REGISTER_URL" ]]; then
  echo "Registering with backend to fetch Loki URL/token: $REGISTER_URL"
  HOSTNAME=$(hostname)

  # Allow passing REG_SECRET and CREATE_HOST flags to request host creation when needed
  REG_SECRET_VAL=${REG_SECRET:-}
  CREATE_HOST_FLAG=${CREATE_HOST:-0}
  payload="{\"hostname\": \"$HOSTNAME\"}"
  if [[ "$CREATE_HOST_FLAG" == "1" && -n "$REG_SECRET_VAL" ]]; then
    payload="{\"hostname\": \"$HOSTNAME\", \"create_if_missing\": true, \"reg_secret\": \"$REG_SECRET_VAL\"}"
  fi

  # Function to parse JSON without jq: try python, then fallback to grep/awk
  parse_json_field() {
    field=$1
    file=$2
    if command -v jq >/dev/null 2>&1; then
      jq -r ".${field} // empty" "$file" 2>/dev/null || true
    elif command -v python3 >/dev/null 2>&1; then
      python3 -c "import sys, json
try:
    obj=json.load(open('$file'))
    print(obj.get('$field',''))
except Exception:
    sys.exit(0)" 2>/dev/null || true
    else
      # last-resort simple grep (fragile) - returns empty on failure
      grep -oP '"${field}"\s*:\s*"\K[^"]+' "$file" 2>/dev/null || true
    fi
  }

  # Try POST with exponential backoff
  resp=$(mktemp)
  max_attempts=5
  attempt=0
  success=0
  if command -v curl >/dev/null 2>&1; then
    while [[ $attempt -lt $max_attempts ]]; do
      attempt=$((attempt + 1))
      echo "Registration attempt $attempt/$max_attempts..."
      if curl -sS -X POST -H "Content-Type: application/json" -d "$payload" "$REGISTER_URL" -o "$resp"; then
        # Basic content sanity check
        if [[ -s "$resp" ]]; then
          LOKI_URL=$(parse_json_field loki_url "$resp")
          TOKEN=$(parse_json_field token "$resp")
          if [[ -n "$LOKI_URL" ]]; then
            success=1
            break
          fi
        fi
      fi
      sleep $((2 ** attempt))
    done

    if [[ $success -eq 1 ]]; then
      echo "Patching Promtail config with Loki URL from registration"
      # Replace any existing clients block and append a new clients block after 'server:'
      if [[ "$SIMULATE" == "1" ]]; then
        sed -i.bak -e '/^clients:/,/^\s*- url:/d' "$PROMTAIL_CONFIG_DST" || true
        sed -i "/^server:/a clients:\n  - url: '$LOKI_URL'\n    batchsize: 100\n    backoff_config:\n      max_period: 5s" "$PROMTAIL_CONFIG_DST"
        if [[ -n "$TOKEN" ]]; then
          sed -i "/^clients:/,0 s/\(backoff_config:\)/\1\n    headers:\n      Authorization: 'Bearer $TOKEN'/" "$PROMTAIL_CONFIG_DST" || true
        fi
      else
        sudo sed -i.bak -e '/^clients:/,/^\s*- url:/d' "$PROMTAIL_CONFIG_DST" || true
        sudo sed -i "/^server:/a clients:\n  - url: '$LOKI_URL'\n    batchsize: 100\n    backoff_config:\n      max_period: 5s" "$PROMTAIL_CONFIG_DST"
        if [[ -n "$TOKEN" ]]; then
          sudo sed -i "/^clients:/,0 s/\(backoff_config:\)/\1\n    headers:\n      Authorization: 'Bearer $TOKEN'/" "$PROMTAIL_CONFIG_DST" || true
        fi
      fi
    else
      echo "Registration failed after $max_attempts attempts or did not return loki_url; skipping patch"
    fi
  else
    echo "curl not installed; cannot call registration endpoint"
  fi
  rm -f "$resp"
fi

# Install systemd unit
if [[ -f "$SYSTEMD_UNIT_SRC" ]]; then
  echo "Installing systemd unit to $SYSTEMD_UNIT_DST"
  if [[ "$SIMULATE" == "1" ]]; then
    cp "$SYSTEMD_UNIT_SRC" "$SYSTEMD_UNIT_DST"
  else
    sudo cp "$SYSTEMD_UNIT_SRC" "$SYSTEMD_UNIT_DST"
  fi
else
  echo "Warning: systemd unit not found at $SYSTEMD_UNIT_SRC"
fi

# Install binary
if [[ -x "$PROMTAIL_BIN" ]]; then
  echo "Promtail binary already exists at $PROMTAIL_BIN"
else
  echo "Promtail binary not found. Attempting to extract from Docker image grafana/promtail:latest (requires docker)."
  if command -v docker >/dev/null 2>&1; then
    CONTAINER_ID=$(docker create grafana/promtail:latest /bin/true)
    echo "Created temp container $CONTAINER_ID"
    # try common binary paths
    if docker cp "$CONTAINER_ID":/usr/bin/promtail "$PROMTAIL_BIN" 2>/dev/null; then
      echo "Copied /usr/bin/promtail"
    elif docker cp "$CONTAINER_ID":/bin/promtail "$PROMTAIL_BIN" 2>/dev/null; then
      echo "Copied /bin/promtail"
    else
      echo "Could not find promtail binary inside image. You may need to download a release binary manually and place it at $PROMTAIL_BIN"
      docker rm "$CONTAINER_ID" >/dev/null || true
      exit 1
    fi
    docker rm "$CONTAINER_ID" >/dev/null || true
    if [[ "$SIMULATE" == "1" ]]; then
      echo "SIMULATE: would chmod +x $PROMTAIL_BIN"
    else
      sudo chmod +x "$PROMTAIL_BIN"
      echo "Installed promtail to $PROMTAIL_BIN"
    fi
  else
    echo "Docker not available and no promtail binary found. Please download promtail and place it at $PROMTAIL_BIN"
    echo "See https://grafana.com/docs/loki/latest/installation/clients/#promtail for downloads"
    exit 1
  fi
fi

# Permissions
if [[ "$SIMULATE" == "1" ]]; then
  echo "SIMULATE: would set ownership and permissions on $PROMTAIL_BIN"
else
  sudo chown root:root "$PROMTAIL_BIN"
  sudo chmod 0755 "$PROMTAIL_BIN"
fi

# Reload systemd
if [[ "$SIMULATE" == "1" ]]; then
  echo "SIMULATE: skipping systemctl actions (would daemon-reload, enable/start promtail)"
else
  if command -v systemctl >/dev/null 2>&1; then
    echo "Reloading systemd and enabling promtail service"
    sudo systemctl daemon-reload
    sudo systemctl enable --now promtail || sudo systemctl restart promtail || true
    echo "Service status:"
    sudo systemctl status --no-pager promtail || true
  else
    echo "systemctl not found. You will need to start promtail manually:"
    echo "$PROMTAIL_BIN --config.file=$PROMTAIL_CONFIG_DST"
  fi
fi

echo "Promtail install script finished. Configure LOKI_URL or TLS as needed and verify logs in Grafana Explore."

# Disable existing agent's log shipping if present (set modules.logs.enabled = false)
AGENT_CFG=/etc/kineticops-agent/config.yaml
if [[ -f "$AGENT_CFG" ]]; then
  echo "Found existing agent config at $AGENT_CFG — disabling agent log module to avoid duplicate shipping"
  if [[ "$SIMULATE" == "1" ]]; then
    cp "$AGENT_CFG" "$AGENT_CFG.bak-$(date +%s)" || true
  else
    sudo cp "$AGENT_CFG" "$AGENT_CFG.bak-$(date +%s)" || true
  fi

  # Try to replace 'enabled: true' under the 'logs:' section with 'enabled: false'.
  # This is a best-effort, indentation-aware sed script.
  awk '
  BEGIN{in_logs=0}
  { if ($0 ~ /^\s*logs:\s*$/) {print; in_logs=1; next} 
    if (in_logs) {
      if ($0 ~ /^\s*[a-zA-Z]/) { in_logs=0; print; next }
      if ($0 ~ /enabled:\s*(true|false)/) {
        sub(/enabled:\s*(true|false)/,"enabled: false")
        print; next
      }
    }
    print
  }' "$AGENT_CFG" | sudo tee "$AGENT_CFG.tmp" >/dev/null && sudo mv "$AGENT_CFG.tmp" "$AGENT_CFG"
  if [[ "$SIMULATE" == "1" ]]; then
    awk '
    BEGIN{in_logs=0}
    { if ($0 ~ /^\s*logs:\s*$/) {print; in_logs=1; next} 
      if (in_logs) {
        if ($0 ~ /^\s*[a-zA-Z]/) { in_logs=0; print; next }
        if ($0 ~ /enabled:\s*(true|false)/) {
          sub(/enabled:\s*(true|false)/,"enabled: false")
          print; next
        }
      }
      print
    }' "$AGENT_CFG" | tee "$AGENT_CFG.tmp" >/dev/null && mv "$AGENT_CFG.tmp" "$AGENT_CFG"
  else
    awk '
    BEGIN{in_logs=0}
    { if ($0 ~ /^\s*logs:\s*$/) {print; in_logs=1; next} 
      if (in_logs) {
        if ($0 ~ /^\s*[a-zA-Z]/) { in_logs=0; print; next }
        if ($0 ~ /enabled:\s*(true|false)/) {
          sub(/enabled:\s*(true|false)/,"enabled: false")
          print; next
        }
      }
      print
    }' "$AGENT_CFG" | sudo tee "$AGENT_CFG.tmp" >/dev/null && sudo mv "$AGENT_CFG.tmp" "$AGENT_CFG"
  fi

  echo "Agent config updated (backup at ${AGENT_CFG}.bak-*)"

  # Restart agent service if present
  if systemctl list-units --full -all | grep -q kineticops-agent; then
    echo "Restarting kineticops-agent service to apply config"
    if [[ "$SIMULATE" == "1" ]]; then
      echo "SIMULATE: would restart kineticops-agent"
    else
      sudo systemctl restart kineticops-agent || true
    fi
  else
    echo "kineticops-agent systemd unit not found; please restart agent manually to apply config change"
  fi
fi
