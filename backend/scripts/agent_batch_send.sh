#!/usr/bin/env bash
# Example script to batch-send metrics from an agent to /api/v1/metrics/bulk
# Usage: ./agent_batch_send.sh <server_url> <token>

SERVER_URL=${1:-http://localhost:8080}
TOKEN=${2:-agent-demo-token}

# Build a JSON array of metrics
cat > /tmp/metrics_batch.json <<EOF
[
  {"host_id": 1, "tenant_id": 1, "name": "cpu_usage", "value": 12.3, "labels": {"env":"dev"}},
  {"host_id": 1, "tenant_id": 1, "name": "memory_usage", "value": 42.7, "labels": {"env":"dev"}}
]
EOF

curl -s -X POST "${SERVER_URL}/api/v1/metrics/bulk" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN}" \
  --data-binary @/tmp/metrics_batch.json | jq '.'

rm -f /tmp/metrics_batch.json
