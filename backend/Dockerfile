# Multi-stage build: compile the Go binary, then run in a minimal runtime image
FROM golang:1.25 AS builder

WORKDIR /src
# Copy go.mod and go.sum to speed up dependency download
COPY go.mod go.sum ./
RUN go env -w GOPROXY=https://proxy.golang.org,direct

# Copy the rest of the backend source
COPY . .

# Build a statically linked binary
ENV CGO_ENABLED=0
RUN go build -o /kineticops-backend ./cmd/server

# Runtime image
FROM debian:bookworm-slim

# Set container timezone to IST and install tzdata so the container OS timezone
# matches the application-level `time.Local` (Asia/Kolkata). Use noninteractive
# frontend to avoid interactive timezone prompts during image build.
ENV TZ=Asia/Kolkata
RUN apt-get update && \
	DEBIAN_FRONTEND=noninteractive apt-get install -y ca-certificates tzdata && \
	ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone && \
	rm -rf /var/lib/apt/lists/*

# Create app user for safety (optional)
RUN useradd -m appuser || true

COPY --from=builder /kineticops-backend /usr/local/bin/kineticops-backend
WORKDIR /app

# Expose API port and metrics/pprof port
EXPOSE 8080 9090

USER appuser
ENTRYPOINT ["/usr/local/bin/kineticops-backend"]
