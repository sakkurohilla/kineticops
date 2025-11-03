# KineticOps Agent

A lightweight, ELK-inspired monitoring agent for collecting system metrics and logs.

## Features

- **System Metrics Collection**: CPU, memory, disk, network, and load metrics
- **Log Collection**: File-based log collection with pattern matching
- **ELK-Compatible**: Uses Elastic Common Schema (ECS) format
- **Lightweight**: Minimal resource footprint
- **Reliable**: Built-in retry logic and state persistence
- **Secure**: TLS support and token-based authentication

## Quick Start

### Installation

```bash
# Download and install
curl -sSL https://install.kineticops.com/agent.sh | sudo bash

# Or build from source
git clone https://github.com/sakkurohilla/kineticops.git
cd kineticops/agent
make build
sudo make install
```

### Configuration

Edit `/etc/kineticops-agent/config.yaml`:

```yaml
# Basic configuration
agent:
  name: kineticops-agent
  hostname: ${HOSTNAME}
  period: 30s

# Backend connection
output:
  kineticops:
    hosts:
      - "https://your-kineticops-server.com"
    token: "your-auth-token"

# Enable modules
modules:
  system:
    enabled: true
    period: 30s
  logs:
    enabled: true
    inputs:
      - type: log
        paths:
          - /var/log/*.log
```

### Start the Agent

```bash
# Start service
sudo systemctl start kineticops-agent

# Enable auto-start
sudo systemctl enable kineticops-agent

# Check status
sudo systemctl status kineticops-agent

# View logs
sudo journalctl -u kineticops-agent -f
```

## Architecture

The agent follows a modular architecture similar to Elastic Agent:

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Collectors    │───▶│    Pipeline      │───▶│     Output      │
│                 │    │                  │    │                 │
│ • System        │    │ • Batching       │    │ • KineticOps    │
│ • Logs          │    │ • Processing     │    │ • HTTP          │
│ • Docker        │    │ • Retry Logic    │    │ • File          │
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

### Components

- **Collectors**: Gather data from various sources
- **Pipeline**: Batch, process, and route data
- **Outputs**: Send data to configured destinations
- **State Manager**: Persist file positions and checkpoints

## Configuration Reference

### Agent Settings

```yaml
agent:
  name: kineticops-agent          # Agent name
  hostname: ${HOSTNAME}           # Override hostname
  period: 30s                     # Heartbeat interval
  tags:                           # Global tags
    - production
    - web-server
```

### Output Configuration

```yaml
output:
  kineticops:
    hosts:
      - "https://primary.example.com"
      - "https://backup.example.com"
    token: "${KINETICOPS_TOKEN}"
    timeout: 30s
    max_retry: 3
    tls:
      enabled: true
      verification_mode: full
```

### System Metrics Module

```yaml
modules:
  system:
    enabled: true
    period: 30s
    processes: false              # Collect process metrics
    cpu:
      enabled: true
      percpu: false               # Per-CPU metrics
      totalcpu: true              # Total CPU metrics
    memory:
      enabled: true
    network:
      enabled: true
    filesystem:
      enabled: true
```

### Log Collection Module

```yaml
modules:
  logs:
    enabled: true
    inputs:
      - type: log
        paths:
          - /var/log/*.log
          - /var/log/messages
        exclude:
          - /var/log/btmp
        fields:
          service: web
          environment: production
        multiline:
          pattern: '^\d{4}-\d{2}-\d{2}'
          negate: true
          match: after
```

## Data Format

The agent sends data in ECS (Elastic Common Schema) format:

### Metric Event

```json
{
  "@timestamp": "2024-01-15T10:30:00.000Z",
  "agent": {
    "name": "kineticops-agent",
    "type": "metricbeat",
    "version": "1.0.0"
  },
  "host": {
    "hostname": "web-server-01"
  },
  "event": {
    "kind": "metric",
    "category": "host",
    "type": "info"
  },
  "system": {
    "cpu": {
      "total": {
        "pct": 0.25
      }
    }
  }
}
```

### Log Event

```json
{
  "@timestamp": "2024-01-15T10:30:00.000Z",
  "agent": {
    "name": "kineticops-agent",
    "type": "filebeat",
    "version": "1.0.0"
  },
  "host": {
    "hostname": "web-server-01"
  },
  "event": {
    "kind": "event",
    "category": "file",
    "type": "info"
  },
  "log": {
    "file": {
      "path": "/var/log/nginx/access.log"
    },
    "level": "info"
  },
  "message": "192.168.1.100 - - [15/Jan/2024:10:30:00 +0000] \"GET / HTTP/1.1\" 200 1234"
}
```

## Development

### Building

```bash
# Install dependencies
make deps

# Build binary
make build

# Build for all platforms
make build-all

# Run tests
make test

# Format code
make fmt
```

### Testing

```bash
# Test configuration
make test-config

# Run in development mode
make run

# View logs
tail -f /var/log/kineticops-agent.log
```

## Troubleshooting

### Common Issues

1. **Agent not starting**
   ```bash
   # Check service status
   sudo systemctl status kineticops-agent
   
   # Check logs
   sudo journalctl -u kineticops-agent -f
   
   # Test configuration
   sudo kineticops-agent -test -c /etc/kineticops-agent/config.yaml
   ```

2. **No data in backend**
   ```bash
   # Check connectivity
   curl -v https://your-backend.com/health
   
   # Verify token
   grep token /etc/kineticops-agent/config.yaml
   
   # Check agent logs for errors
   sudo journalctl -u kineticops-agent | grep ERROR
   ```

3. **High resource usage**
   ```bash
   # Reduce collection frequency
   # Edit config.yaml and increase period values
   
   # Disable unnecessary modules
   # Set enabled: false for unused modules
   ```

### Log Levels

- `DEBUG`: Detailed debugging information
- `INFO`: General information messages
- `WARN`: Warning messages
- `ERROR`: Error messages

Set log level in configuration:

```yaml
logging:
  level: info
  to_file: true
  file: /var/log/kineticops-agent.log
```

## Security

### Best Practices

1. **Use TLS**: Always enable TLS for production
2. **Rotate Tokens**: Regularly rotate authentication tokens
3. **Limit Permissions**: Run agent with minimal required permissions
4. **Monitor Logs**: Watch for authentication failures
5. **Network Security**: Use firewalls to restrict agent communication

### File Permissions

```bash
# Configuration files
sudo chmod 644 /etc/kineticops-agent/config.yaml
sudo chown root:root /etc/kineticops-agent/config.yaml

# State directory
sudo chmod 755 /var/lib/kineticops-agent
sudo chown kineticops:kineticops /var/lib/kineticops-agent

# Log files
sudo chmod 644 /var/log/kineticops-agent.log
sudo chown kineticops:kineticops /var/log/kineticops-agent.log
```

## Support

- Documentation: https://docs.kineticops.com/agent
- Issues: https://github.com/sakkurohilla/kineticops/issues
- Community: https://community.kineticops.com