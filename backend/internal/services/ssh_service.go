package services

import (
	"bytes"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

type SSHClient struct {
	client *ssh.Client
	config *ssh.ClientConfig
}

// TestSSHConnection tests if SSH connection works with password or key
func TestSSHConnection(host string, port int, username, password string) error {
	return TestSSHConnectionWithKey(host, port, username, password, "")
}

// TestSSHConnectionWithKey tests SSH connection with password or private key
func TestSSHConnectionWithKey(host string, port int, username, password, privateKey string) error {
	var authMethods []ssh.AuthMethod
	
	// Always try password first if provided (more reliable)
	if password != "" {
		authMethods = append(authMethods, ssh.Password(password))
	}
	
	// Try SSH key if provided
	if privateKey != "" {
		signer, err := ssh.ParsePrivateKey([]byte(privateKey))
		if err != nil {
			return fmt.Errorf("failed to parse private key: %w", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}
	
	if len(authMethods) == 0 {
		return fmt.Errorf("either password or private key must be provided")
	}

	config := &ssh.ClientConfig{
		User: username,
		Auth: authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	addr := net.JoinHostPort(host, strconv.Itoa(port))
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return fmt.Errorf("SSH connection failed: %w", err)
	}
	defer client.Close()

	return nil
}

// NewSSHClient creates a new SSH client with password or key
func NewSSHClient(host string, port int, username, password string) (*SSHClient, error) {
	return NewSSHClientWithKey(host, port, username, password, "")
}

// NewSSHClientWithKey creates SSH client with password or private key
func NewSSHClientWithKey(host string, port int, username, password, privateKey string) (*SSHClient, error) {
	var authMethods []ssh.AuthMethod
	
	if privateKey != "" {
		signer, err := ssh.ParsePrivateKey([]byte(privateKey))
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	} else if password != "" {
		authMethods = append(authMethods, ssh.Password(password))
	} else {
		return nil, fmt.Errorf("either password or private key must be provided")
	}

	config := &ssh.ClientConfig{
		User: username,
		Auth: authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	addr := net.JoinHostPort(host, strconv.Itoa(port))
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	return &SSHClient{
		client: client,
		config: config,
	}, nil
}

// ExecuteCommand executes a command on remote host
func (s *SSHClient) ExecuteCommand(cmd string) (string, error) {
	// Default ExecuteCommand enforces a reasonable timeout to avoid hanging SSH commands.
	return s.ExecuteCommandTimeout(cmd, 10*time.Second)
}

// ExecuteCommandTimeout runs a command with a hard timeout. If the timeout elapses
// the attempt returns an error. Note: closing the session may not always kill the
// remote process, but it avoids hanging the collector.
func (s *SSHClient) ExecuteCommandTimeout(cmd string, timeout time.Duration) (string, error) {
	type res struct {
		out string
		err error
	}

	ch := make(chan res, 1)

	go func() {
		session, err := s.client.NewSession()
		if err != nil {
			ch <- res{"", fmt.Errorf("failed to create session: %w", err)}
			return
		}
		defer session.Close()

		var stdout bytes.Buffer
		session.Stdout = &stdout

		if err := session.Run(cmd); err != nil {
			ch <- res{"", fmt.Errorf("command execution failed: %w", err)}
			return
		}
		ch <- res{strings.TrimSpace(stdout.String()), nil}
	}()

	select {
	case r := <-ch:
		return r.out, r.err
	case <-time.After(timeout):
		return "", fmt.Errorf("command timeout after %s", timeout)
	}
}

// Close closes the SSH connection
func (s *SSHClient) Close() error {
	if s.client != nil {
		return s.client.Close()
	}
	return nil
}

// CollectCPUUsage gets CPU usage percentage
func (s *SSHClient) CollectCPUUsage() (float64, error) {
	// Use /proc/stat to compute CPU usage over a 1-second interval (more reliable)
	// This awk reads /proc/stat twice with a sleep and computes the percentage of non-idle time.
	cmd := `awk 'BEGIN{getline a < "/proc/stat"; split(a, t); tot1=0; for(i=2;i<=8;i++) tot1+=t[i]; idle1=t[5]; system("sleep 1"); getline b < "/proc/stat"; split(b, s); tot2=0; for(i=2;i<=8;i++) tot2+=s[i]; idle2=s[5]; usage=(1-(idle2-idle1)/(tot2-tot1))*100; if(usage<0) usage=0; printf "%.2f", usage}'`
	output, err := s.ExecuteCommandTimeout(cmd, 10*time.Second)
	if err != nil {
		// On command failure, return 0 but do not fail the whole collection.
		return 0, nil
	}

	cpu, err := strconv.ParseFloat(output, 64)
	if err != nil {
		// Parsing error — treat as missing metric rather than fatal.
		return 0, nil
	}

	return cpu, nil
}

// CollectMemoryUsage gets memory usage percentage
func (s *SSHClient) CollectMemoryUsage() (used, total, percentage float64, err error) {
	// Use /proc/meminfo to calculate used and total memory reliably
	cmd := `awk '/MemTotal/ {total=$2} /MemAvailable/ {avail=$2} END {used=total-avail; if(total>0) perc=used*100/total; else perc=0; printf "%d %d %.2f", used, total, perc}' /proc/meminfo`
	output, err := s.ExecuteCommandTimeout(cmd, 10*time.Second)
	if err != nil {
		return 0, 0, 0, nil
	}

	parts := strings.Fields(output)
	if len(parts) < 3 {
		return 0, 0, 0, nil
	}

	// meminfo reports kB; convert to MB for consistency with frontend where applicable
	usedKb, err1 := strconv.ParseFloat(parts[0], 64)
	totalKb, err2 := strconv.ParseFloat(parts[1], 64)
	perc, err3 := strconv.ParseFloat(parts[2], 64)
	if err1 != nil || err2 != nil || err3 != nil {
		return 0, 0, 0, nil
	}

	used = usedKb / 1024.0
	total = totalKb / 1024.0
	percentage = perc

	return used, total, percentage, nil
}

// CollectDiskUsage gets disk usage percentage
func (s *SSHClient) CollectDiskUsage() (used, total, percentage float64, err error) {
	// Use df -B1 to get bytes (avoid human-readable parsing ambiguities)
	cmd := `df -B1 / | awk 'NR==2{gsub(/%/,"",$5); printf "%s %s %s", $3,$2,$5}'`
	output, err := s.ExecuteCommandTimeout(cmd, 10*time.Second)
	if err != nil {
		return 0, 0, 0, nil
	}

	parts := strings.Fields(output)
	if len(parts) < 3 {
		return 0, 0, 0, nil
	}

	usedBytes, err1 := strconv.ParseFloat(parts[0], 64)
	totalBytes, err2 := strconv.ParseFloat(parts[1], 64)
	percentage, err3 := strconv.ParseFloat(parts[2], 64)
	if err1 != nil || err2 != nil || err3 != nil {
		return 0, 0, 0, nil
	}

	// Convert bytes to GB for consistent display in frontend
	if totalBytes > 0 {
		used = usedBytes / (1024 * 1024 * 1024)
		total = totalBytes / (1024 * 1024 * 1024)
	}

	return used, total, percentage, nil
}

// CollectNetworkStats gets network I/O
func (s *SSHClient) CollectNetworkStats() (rxBytes, txBytes float64, err error) {
	// Sum rx and tx across all non-loopback interfaces for more accurate totals
	cmd := `awk '/:/ {iface=$1; gsub(/:/,"",iface); if(iface!="lo") {rx+=$2; tx+=$10}} END {printf "%d %d", rx, tx}' /proc/net/dev`
	output, err := s.ExecuteCommandTimeout(cmd, 10*time.Second)
	if err != nil {
		return 0, 0, nil
	}

	parts := strings.Fields(output)
	if len(parts) < 2 {
		return 0, 0, nil // No network data
	}

	rx, err1 := strconv.ParseFloat(parts[0], 64)
	tx, err2 := strconv.ParseFloat(parts[1], 64)
	if err1 != nil || err2 != nil {
		return 0, 0, nil
	}

	// Convert to MB
	return rx / 1024 / 1024, tx / 1024 / 1024, nil
}

// CollectUptime gets system uptime in seconds
func (s *SSHClient) CollectUptime() (int64, error) {
	output, err := s.ExecuteCommandTimeout("cat /proc/uptime | awk '{print $1}'", 5*time.Second)
	if err != nil {
		return 0, nil
	}

	uptime, err := strconv.ParseFloat(output, 64)
	if err != nil {
		return 0, nil
	}

	return int64(uptime), nil
}

// CollectLoadAverage gets system load average
func (s *SSHClient) CollectLoadAverage() (string, error) {
	output, err := s.ExecuteCommandTimeout("uptime | awk -F'load average:' '{print $2}' | xargs", 5*time.Second)
	if err != nil {
		return "", nil
	}
	return output, nil
}

// SSHService provides high-level SSH operations
type SSHService struct{}

func NewSSHService() *SSHService {
	return &SSHService{}
}

// ExecuteScript executes a script on remote host with password auth
func (s *SSHService) ExecuteScript(host, username, password, script string, port int) error {
	if port == 0 {
		port = 22
	}
	
	client, err := NewSSHClient(host, port, username, password)
	if err != nil {
		return fmt.Errorf("failed to connect: %v", err)
	}
	defer client.Close()
	
	// Upload and execute script
	cmd := fmt.Sprintf("cat > /tmp/install.sh << 'EOF'\n%s\nEOF\nchmod +x /tmp/install.sh && bash /tmp/install.sh", script)
	_, err = client.ExecuteCommandTimeout(cmd, 5*time.Minute)
	return err
}

// ExecuteScriptWithKey executes a script on remote host with SSH key auth
func (s *SSHService) ExecuteScriptWithKey(host, username, privateKey, script string, port int) error {
	if port == 0 {
		port = 22
	}
	
	client, err := NewSSHClientWithKey(host, port, username, "", privateKey)
	if err != nil {
		return fmt.Errorf("failed to connect: %v", err)
	}
	defer client.Close()
	
	// Upload and execute script
	cmd := fmt.Sprintf("cat > /tmp/install.sh << 'EOF'\n%s\nEOF\nchmod +x /tmp/install.sh && bash /tmp/install.sh", script)
	_, err = client.ExecuteCommandTimeout(cmd, 5*time.Minute)
	return err
}
