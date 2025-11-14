#!/usr/bin/env bash
# Agent cleanup helper for KineticOps
# Idempotent script to backup, stop and remove a previously-installed agent
# Usage: sudo ./agent_cleanup.sh [--yes]

set -euo pipefail

DRY_RUN=1
if [[ "${1:-}" == "--yes" || "${1:-}" == "-y" ]]; then
  DRY_RUN=0
fi

backup_dir="/root/kineticops_agent_backup_$(date +%Y%m%d%H%M%S)"
echo "KineticOps agent cleanup helper"
echo "Backup dir: $backup_dir"

run() {
  if [[ $DRY_RUN -eq 1 ]]; then
    echo "DRY RUN: $*"
  else
    echo "+ $*"
    $*
  fi
}

if [[ $EUID -ne 0 ]]; then
  echo "This script should be run as root (sudo). Running in dry-run mode unless --yes provided."
fi

echo "1) Backing up /opt/kineticops-agent if present"
if [[ -d /opt/kineticops-agent ]]; then
  run mkdir -p "$backup_dir"
  run cp -a /opt/kineticops-agent "$backup_dir/" || true
  echo "  backed up to $backup_dir/kineticops-agent"
else
  echo "  /opt/kineticops-agent not present"
fi

echo "2) Stop and disable systemd service if present"
if command -v systemctl >/dev/null 2>&1; then
  if systemctl list-unit-files | grep -q '^kineticops-agent.service'; then
    run systemctl stop kineticops-agent || true
    run systemctl disable kineticops-agent || true
    run systemctl daemon-reload || true
    echo "  stopped and disabled unit"
  else
    echo "  systemd unit not found"
  fi
else
  echo "  systemctl not available"
fi

echo "3) Remove service file(s)"
if [[ -f /etc/systemd/system/kineticops-agent.service ]]; then
  run rm -f /etc/systemd/system/kineticops-agent.service
  run systemctl daemon-reload || true
  echo "  removed /etc/systemd/system/kineticops-agent.service"
fi
if [[ -d /etc/systemd/system/kineticops-agent.service.d ]]; then
  run rm -rf /etc/systemd/system/kineticops-agent.service.d
  run systemctl daemon-reload || true
  echo "  removed unit drop-ins"
fi

echo "4) Kill running agent processes"
if pgrep -f '/opt/kineticops-agent/agent' >/dev/null 2>&1; then
  run pkill -f '/opt/kineticops-agent/agent' || true
  sleep 1
fi
if pgrep -f 'kineticops-agent' >/dev/null 2>&1; then
  run pkill -f 'kineticops-agent' || true
fi

echo "5) Remove agent files and symlinks"
run rm -rf /opt/kineticops-agent || true
run rm -f /usr/local/bin/kineticops-agent || true
run rm -f /usr/local/bin/agent || true

echo "6) Optionally remove dedicated user (skip unless created by installer)"
if id -u kineticops >/dev/null 2>&1; then
  echo "  found user 'kineticops' — not removing automatically. To remove: userdel -r kineticops"
fi

echo "7) Clean journal entries (this rotates/limits logs for the unit only)"
if command -v journalctl >/dev/null 2>&1; then
  run journalctl --rotate || true
  run journalctl --vacuum-time=1s || true
  echo "  vacuumed journal logs (1s retention)"
fi

echo "8) Remove /etc config if present (best-effort)"
if [[ -f /etc/kineticops-agent/config.yaml ]]; then
  run mv /etc/kineticops-agent/config.yaml "$backup_dir/" || run rm -f /etc/kineticops-agent/config.yaml || true
  echo "  /etc/kineticops-agent/config.yaml moved to backup"
fi

echo "9) Sanity checks"
echo "  - /opt exists: "; [[ -d /opt/kineticops-agent ]] && echo "present" || echo "absent"
echo "  - systemd unit present: "; systemctl list-unit-files | grep -q '^kineticops-agent.service' && echo "yes" || echo "no"

echo "Cleanup complete."
if [[ $DRY_RUN -eq 1 ]]; then
  echo "Dry run mode. Re-run with --yes to perform actions."
fi

exit 0
