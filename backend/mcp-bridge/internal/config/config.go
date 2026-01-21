package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	BackendURL string      `yaml:"backend_url" mapstructure:"backend_url"`
	AuthToken  string      `yaml:"auth_token" mapstructure:"auth_token"`
	UserID     string      `yaml:"user_id" mapstructure:"user_id"`
	MCPServers []MCPServer `yaml:"mcp_servers" mapstructure:"mcp_servers"`
}

// MCPServer represents a configured MCP server
type MCPServer struct {
	Name        string                 `yaml:"name" mapstructure:"name"`
	Path        string                 `yaml:"path,omitempty" mapstructure:"path"`       // For executable path
	Command     string                 `yaml:"command,omitempty" mapstructure:"command"` // For command-based (e.g., "npx")
	Args        []string               `yaml:"args,omitempty" mapstructure:"args"`       // Command arguments
	URL         string                 `yaml:"url,omitempty" mapstructure:"url"`
	Type        string                 `yaml:"type" mapstructure:"type"` // "stdio" or "sse"
	Config      map[string]interface{} `yaml:"config,omitempty" mapstructure:"config"`
	Enabled     bool                   `yaml:"enabled" mapstructure:"enabled"`
	Description string                 `yaml:"description,omitempty" mapstructure:"description"`
}

var (
	configPath string
	configDir  string
)

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Sprintf("failed to get home directory: %v", err))
	}

	configDir = filepath.Join(home, ".claraverse")
	configPath = filepath.Join(configDir, "mcp-config.yaml")
}

// GetConfigPath returns the path to the config file
func GetConfigPath() string {
	return configPath
}

// GetConfigDir returns the config directory
func GetConfigDir() string {
	return configDir
}

// Load loads the configuration from file
func Load() (*Config, error) {
	// Ensure config directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create default config
		defaultConfig := &Config{
			BackendURL: "ws://localhost:3001/mcp/connect",
			MCPServers: []MCPServer{},
		}
		if err := Save(defaultConfig); err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}
		return defaultConfig, nil
	}

	// Read config file
	viper.SetConfigFile(configPath)
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

// Save saves the configuration to file
func Save(cfg *Config) error {
	// Ensure config directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file with secure permissions
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// AddServer adds a new MCP server to the configuration
func (c *Config) AddServer(server MCPServer) error {
	// Check if server already exists
	for i, s := range c.MCPServers {
		if s.Name == server.Name {
			// Update existing server
			c.MCPServers[i] = server
			return nil
		}
	}

	// Add new server
	c.MCPServers = append(c.MCPServers, server)
	return nil
}

// RemoveServer removes an MCP server by name
func (c *Config) RemoveServer(name string) error {
	for i, s := range c.MCPServers {
		if s.Name == name {
			c.MCPServers = append(c.MCPServers[:i], c.MCPServers[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("server %s not found", name)
}

// GetServer retrieves a server by name
func (c *Config) GetServer(name string) (*MCPServer, error) {
	for _, s := range c.MCPServers {
		if s.Name == name {
			return &s, nil
		}
	}
	return nil, fmt.Errorf("server %s not found", name)
}

// GetEnabledServers returns only enabled servers
func (c *Config) GetEnabledServers() []MCPServer {
	var enabled []MCPServer
	for _, s := range c.MCPServers {
		if s.Enabled {
			enabled = append(enabled, s)
		}
	}
	return enabled
}
