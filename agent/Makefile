# KineticOps Agent Makefile

# Build variables
BINARY_NAME=kineticops-agent
VERSION?=1.0.0
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.gitCommit=$(GIT_COMMIT)"

# Go variables
GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)

# Directories
BUILD_DIR=build
DIST_DIR=dist
CONFIG_DIR=/etc/kineticops-agent
STATE_DIR=/var/lib/kineticops-agent
LOG_DIR=/var/log

.PHONY: all build clean test install uninstall package deps

all: clean deps test build

# Build the agent
build:
	@echo "Building $(BINARY_NAME) for $(GOOS)/$(GOARCH)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./main.go

# Build for multiple platforms
build-all:
	@echo "Building for multiple platforms..."
	@mkdir -p $(DIST_DIR)
	GOOS=linux GOARCH=amd64 make build && mv $(BUILD_DIR)/$(BINARY_NAME) $(DIST_DIR)/$(BINARY_NAME)-linux-amd64
	GOOS=linux GOARCH=arm64 make build && mv $(BUILD_DIR)/$(BINARY_NAME) $(DIST_DIR)/$(BINARY_NAME)-linux-arm64
	GOOS=darwin GOARCH=amd64 make build && mv $(BUILD_DIR)/$(BINARY_NAME) $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64
	GOOS=darwin GOARCH=arm64 make build && mv $(BUILD_DIR)/$(BINARY_NAME) $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64
	GOOS=windows GOARCH=amd64 make build && mv $(BUILD_DIR)/$(BINARY_NAME) $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Test configuration
test-config:
	@echo "Testing configuration..."
	./$(BUILD_DIR)/$(BINARY_NAME) -test -c config/config.yaml

# Install the agent (requires sudo)
install: build
	@echo "Installing $(BINARY_NAME)..."
	sudo mkdir -p $(CONFIG_DIR)
	sudo mkdir -p $(STATE_DIR)
	sudo mkdir -p $(LOG_DIR)
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	sudo cp config/config.yaml $(CONFIG_DIR)/
	sudo chmod +x /usr/local/bin/$(BINARY_NAME)
	sudo chown -R root:root $(CONFIG_DIR)
	sudo chown -R kineticops:kineticops $(STATE_DIR) 2>/dev/null || sudo chown -R nobody:nobody $(STATE_DIR)
	@echo "Installation complete. Edit $(CONFIG_DIR)/config.yaml and start the service."

# Install systemd service
install-service: install
	@echo "Installing systemd service..."
	sudo cp scripts/kineticops-agent.service /etc/systemd/system/
	sudo systemctl daemon-reload
	sudo systemctl enable kineticops-agent
	@echo "Service installed. Start with: sudo systemctl start kineticops-agent"

# Uninstall the agent
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	sudo systemctl stop kineticops-agent 2>/dev/null || true
	sudo systemctl disable kineticops-agent 2>/dev/null || true
	sudo rm -f /etc/systemd/system/kineticops-agent.service
	sudo systemctl daemon-reload
	sudo rm -f /usr/local/bin/$(BINARY_NAME)
	sudo rm -rf $(CONFIG_DIR)
	@echo "Uninstallation complete. State directory $(STATE_DIR) preserved."

# Create installation package
package: build-all
	@echo "Creating installation packages..."
	@mkdir -p $(DIST_DIR)/packages
	
	# Linux package
	@mkdir -p $(DIST_DIR)/linux-package/usr/local/bin
	@mkdir -p $(DIST_DIR)/linux-package/etc/kineticops-agent
	@mkdir -p $(DIST_DIR)/linux-package/etc/systemd/system
	@cp $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 $(DIST_DIR)/linux-package/usr/local/bin/$(BINARY_NAME)
	@cp config/config.yaml $(DIST_DIR)/linux-package/etc/kineticops-agent/
	@cp scripts/kineticops-agent.service $(DIST_DIR)/linux-package/etc/systemd/system/
	@cp scripts/install.sh $(DIST_DIR)/linux-package/
	@chmod +x $(DIST_DIR)/linux-package/install.sh
	@tar -czf $(DIST_DIR)/packages/$(BINARY_NAME)-$(VERSION)-linux-amd64.tar.gz -C $(DIST_DIR)/linux-package .
	
	@echo "Packages created in $(DIST_DIR)/packages/"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)
	rm -rf $(DIST_DIR)

# Run the agent in development mode
run: build
	@echo "Running $(BINARY_NAME) in development mode..."
	./$(BUILD_DIR)/$(BINARY_NAME) -c config/config.yaml -v

# Show version
version:
	@echo "$(BINARY_NAME) version $(VERSION)"
	@echo "Build time: $(BUILD_TIME)"
	@echo "Git commit: $(GIT_COMMIT)"

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Lint code
lint:
	@echo "Linting code..."
	golangci-lint run

# Generate documentation
docs:
	@echo "Generating documentation..."
	@mkdir -p docs
	@echo "# KineticOps Agent" > docs/README.md
	@echo "" >> docs/README.md
	@echo "Version: $(VERSION)" >> docs/README.md
	@echo "Build time: $(BUILD_TIME)" >> docs/README.md
	@echo "" >> docs/README.md
	@echo "## Configuration" >> docs/README.md
	@echo "" >> docs/README.md
	@echo "\`\`\`yaml" >> docs/README.md
	@cat config/config.yaml >> docs/README.md
	@echo "\`\`\`" >> docs/README.md

help:
	@echo "Available targets:"
	@echo "  build        - Build the agent binary"
	@echo "  build-all    - Build for multiple platforms"
	@echo "  deps         - Install dependencies"
	@echo "  test         - Run tests"
	@echo "  test-config  - Test configuration"
	@echo "  install      - Install the agent"
	@echo "  install-service - Install with systemd service"
	@echo "  uninstall    - Uninstall the agent"
	@echo "  package      - Create installation packages"
	@echo "  clean        - Clean build artifacts"
	@echo "  run          - Run in development mode"
	@echo "  fmt          - Format code"
	@echo "  lint         - Lint code"
	@echo "  docs         - Generate documentation"
	@echo "  version      - Show version information"
	@echo "  help         - Show this help"