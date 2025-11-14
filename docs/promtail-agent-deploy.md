Promtail agent — deploy guide (Loki-first, minimal resource)

Overview
--------
This file explains how to deploy Promtail on hosts so logs are shipped directly to Loki (recommended minimal-storage, best tailing UX).

Goals
- Minimal resource consumption on hosts
- Reliable resume of tails (positions file)
- Drop noisy logs at source to reduce storage
- Simple TLS/URL configuration to point to central Loki

Config (included)
- `config/promtail-agent.yaml` is included in this repo. It uses:
  - positions file: `/var/lib/promtail/positions.yaml`
  - default Loki url: `http://loki:3100/loki/api/v1/push` (set via `LOKI_URL` env var)
  - pipeline stages that drop debug/heartbeat logs
  - max_line_size: 1MB

Environment & variables
- LOKI_URL: full HTTP(S) URL to Loki push API (example: https://loki.example.com/loki/api/v1/push)
- PROMTAIL_USER: optional user to run promtail as (systemd example uses `promtail`)

Deploy options
------------
1) Systemd (recommended for bare metal / dedicated VMs)

- Create system user and directory:

```bash
sudo useradd --system --no-create-home --shell /usr/sbin/nologin promtail
sudo mkdir -p /etc/promtail /var/lib/promtail
sudo chown promtail:promtail /var/lib/promtail
```

- Install promtail binary (download from Loki releases) and put it in `/usr/local/bin/promtail`.
- Place the config at `/etc/promtail/promtail.yaml` (or symlink from repo `config/promtail-agent.yaml`).

- Example systemd unit (`/etc/systemd/system/promtail.service`):

```ini
[Unit]
Description=Promtail service
After=network.target

[Service]
User=promtail
Group=promtail
Type=simple
ExecStart=/usr/local/bin/promtail --config.file=/etc/promtail/promtail.yaml
Restart=on-failure
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
```

Start and enable:

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now promtail
```

2) Docker (useful for containerized hosts)

```bash
docker run -d --name promtail \
  -v /var/log:/var/log:ro \
  -v /etc/promtail/promtail.yaml:/etc/promtail/promtail.yaml:ro \
  -v /var/lib/promtail:/var/lib/promtail \
  -e LOKI_URL=https://loki.example.com/loki/api/v1/push \
  grafana/promtail:latest -config.file=/etc/promtail/promtail.yaml
```

Resource tuning (keep agent minimal)
- Use small batch sizes (`batchsize: 100` in config).
- Drop debug/noise logs at the pipeline stage.
- Limit number of targets and avoid overly aggressive file globbing.
- Put promtail in a cgroup or Docker container with CPU/memory limits if host resources are tight.

Security
- Always use HTTPS to Loki and enable basic auth or mTLS at the ingress (NGINX, ALB) if Loki is exposed.
- Use network-level ACLs so only your agents can reach Loki.
- Consider proxying Promtail through the backend if you need per-agent authentication tokens — but that adds complexity.

Monitoring & Health
- Promtail exposes a `/metrics` endpoint (if built with it) — scrape it with Prometheus.
- Monitor the positions file size and promtail restarts.

Testing
- After startup, generate a test log line:
  - `logger "promtail test $(hostname)"`
- Query Grafana Explore with label `job="varlogs"` and your host label to verify tail.

Notes and next steps
- If you need durable delivery across network partitions, consider a lightweight disk-buffering shipper (Vector) or run Promtail with an external file buffer (not built-in). That will increase agent complexity.
- If you later need replay or enrichment, add Redpanda with a short retention and a worker that consumes and writes to Loki.

