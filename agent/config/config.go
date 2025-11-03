package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

// Config represents the agent configuration
type Config struct {
	Agent    AgentConfig    `yaml:"agent"`
	Output   OutputConfig   `yaml:"output"`
	Modules  ModulesConfig  `yaml:"modules"`
	Security SecurityConfig `yaml:"security"`
	Logging  LoggingConfig  `yaml:"logging"`
}

// AgentConfig contains agent-specific settings
type AgentConfig struct {
	Name     string        `yaml:"name"`
	Hostname string        `yaml:"hostname"`
	Period   time.Duration `yaml:"period"`
	Tags     []string      `yaml:"tags"`
}

// OutputConfig defines where to send data
type OutputConfig struct {
	KineticOps KineticOpsOutput `yaml:"kineticops"`
}

// KineticOpsOutput configuration
type KineticOpsOutput struct {
	Hosts    []string      `yaml:"hosts"`
	Token    string        `yaml:"token"`
	Timeout  time.Duration `yaml:"timeout"`
	MaxRetry int           `yaml:"max_retry"`
	TLS      TLSConfig     `yaml:"tls"`
}

// TLSConfig for secure connections
type TLSConfig struct {
	Enabled            bool   `yaml:"enabled"`
	VerificationMode   string `yaml:"verification_mode"`
	CertificateAuthorities []string `yaml:"certificate_authorities"`
	Certificate        string `yaml:"certificate"`
	Key                string `yaml:"key"`
}

// ModulesConfig enables/disables data collection modules
type ModulesConfig struct {
	System SystemModule `yaml:"system"`
	Logs   LogsModule   `yaml:"logs"`
	Docker DockerModule `yaml:"docker"`
}

// SystemModule collects system metrics
type SystemModule struct {
	Enabled    bool          `yaml:"enabled"`
	Period     time.Duration `yaml:"period"`
	Processes  bool          `yaml:"processes"`
	CPU        CPUConfig     `yaml:"cpu"`
	Memory     MemoryConfig  `yaml:"memory"`
	Network    NetworkConfig `yaml:"network"`
	Filesystem FSConfig      `yaml:"filesystem"`
}

type CPUConfig struct {
	Enabled   bool `yaml:"enabled"`
	PerCPU    bool `yaml:"percpu"`
	TotalCPU  bool `yaml:"totalcpu"`
}

type MemoryConfig struct {
	Enabled bool `yaml:"enabled"`
}

type NetworkConfig struct {
	Enabled bool `yaml:"enabled"`
}

type FSConfig struct {
	Enabled bool `yaml:"enabled"`
}

// LogsModule collects log files
type LogsModule struct {
	Enabled bool       `yaml:"enabled"`
	Inputs  []LogInput `yaml:"inputs"`
}

type LogInput struct {
	Type        string            `yaml:"type"`
	Paths       []string          `yaml:"paths"`
	Exclude     []string          `yaml:"exclude"`
	Fields      map[string]string `yaml:"fields"`
	Multiline   MultilineConfig   `yaml:"multiline"`
	Processors  []ProcessorConfig `yaml:"processors"`
}

type MultilineConfig struct {
	Pattern string `yaml:"pattern"`
	Negate  bool   `yaml:"negate"`
	Match   string `yaml:"match"`
}

type ProcessorConfig struct {
	Name   string                 `yaml:"name"`
	Config map[string]interface{} `yaml:"config"`
}

// DockerModule collects Docker container data
type DockerModule struct {
	Enabled bool `yaml:"enabled"`
	Period  time.Duration `yaml:"period"`
}

// SecurityConfig for authentication and encryption
type SecurityConfig struct {
	Token string `yaml:"token"`
}

// LoggingConfig for agent logging
type LoggingConfig struct {
	Level  string `yaml:"level"`
	ToFile bool   `yaml:"to_file"`
	File   string `yaml:"file"`
}

// Load reads and parses the configuration file
func Load(path string) (*Config, error) {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return getDefaultConfig(), nil
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Apply defaults
	applyDefaults(&config)

	// Validate configuration
	if err := validate(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// getDefaultConfig returns a default configuration
func getDefaultConfig() *Config {
	hostname, _ := os.Hostname()
	
	config := &Config{
		Agent: AgentConfig{
			Name:     "kineticops-agent",
			Hostname: hostname,
			Period:   30 * time.Second,
			Tags:     []string{},
		},
		Output: OutputConfig{
			KineticOps: KineticOpsOutput{
				Hosts:    []string{"http://localhost:8080"},
				Timeout:  30 * time.Second,
				MaxRetry: 3,
				TLS: TLSConfig{
					Enabled:          false,
					VerificationMode: "full",
				},
			},
		},
		Modules: ModulesConfig{
			System: SystemModule{
				Enabled: true,
				Period:  30 * time.Second,
				CPU: CPUConfig{
					Enabled:  true,
					PerCPU:   false,
					TotalCPU: true,
				},
				Memory: MemoryConfig{
					Enabled: true,
				},
				Network: NetworkConfig{
					Enabled: true,
				},
				Filesystem: FSConfig{
					Enabled: true,
				},
			},
			Logs: LogsModule{
				Enabled: false,
				Inputs: []LogInput{
					{
						Type:  "log",
						Paths: []string{"/var/log/*.log"},
					},
				},
			},
			Docker: DockerModule{
				Enabled: false,
				Period:  30 * time.Second,
			},
		},
		Security: SecurityConfig{},
		Logging: LoggingConfig{
			Level:  "info",
			ToFile: false,
		},
	}

	return config
}

// applyDefaults fills in missing values with defaults
func applyDefaults(config *Config) {
	if config.Agent.Period == 0 {
		config.Agent.Period = 30 * time.Second
	}
	
	if config.Output.KineticOps.Timeout == 0 {
		config.Output.KineticOps.Timeout = 30 * time.Second
	}
	
	if config.Output.KineticOps.MaxRetry == 0 {
		config.Output.KineticOps.MaxRetry = 3
	}
	
	if config.Modules.System.Period == 0 {
		config.Modules.System.Period = 30 * time.Second
	}
	
	if config.Modules.Docker.Period == 0 {
		config.Modules.Docker.Period = 30 * time.Second
	}
	
	if config.Logging.Level == "" {
		config.Logging.Level = "info"
	}
}

// validate checks if the configuration is valid
func validate(config *Config) error {
	if len(config.Output.KineticOps.Hosts) == 0 {
		return fmt.Errorf("at least one output host must be specified")
	}
	
	if config.Agent.Period < time.Second {
		return fmt.Errorf("agent period must be at least 1 second")
	}
	
	return nil
}

// Save writes the configuration to a file
func (c *Config) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	
	return ioutil.WriteFile(path, data, 0644)
}