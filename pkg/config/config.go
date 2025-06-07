package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all configuration for the fly-mcp server
type Config struct {
	// Server configuration
	Server ServerConfig `mapstructure:"server"`
	
	// Fly.io configuration
	Fly FlyConfig `mapstructure:"fly"`
	
	// MCP configuration
	MCP MCPConfig `mapstructure:"mcp"`
	
	// Security configuration
	Security SecurityConfig `mapstructure:"security"`
	
	// Logging configuration
	Logging LoggingConfig `mapstructure:"logging"`
	
	// Environment (local, staging, production)
	Environment string `mapstructure:"environment"`
}

// ServerConfig contains HTTP server settings
type ServerConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
	IdleTimeout  int    `mapstructure:"idle_timeout"`
}

// FlyConfig contains Fly.io API settings
type FlyConfig struct {
	APIToken     string `mapstructure:"api_token"`
	Organization string `mapstructure:"organization"`
	BaseURL      string `mapstructure:"base_url"`
	Timeout      int    `mapstructure:"timeout"`
}

// MCPConfig contains MCP protocol settings
type MCPConfig struct {
	Version     string            `mapstructure:"version"`
	ServerInfo  MCPServerInfo     `mapstructure:"server_info"`
	Capabilities MCPCapabilities `mapstructure:"capabilities"`
}

// MCPServerInfo contains server identification
type MCPServerInfo struct {
	Name    string `mapstructure:"name"`
	Version string `mapstructure:"version"`
}

// MCPCapabilities defines what the server can do
type MCPCapabilities struct {
	Tools     MCPToolsCapability     `mapstructure:"tools"`
	Resources MCPResourcesCapability `mapstructure:"resources"`
	Prompts   MCPPromptsCapability   `mapstructure:"prompts"`
}

type MCPToolsCapability struct {
	ListChanged bool `mapstructure:"list_changed"`
}

type MCPResourcesCapability struct {
	Subscribe   bool `mapstructure:"subscribe"`
	ListChanged bool `mapstructure:"list_changed"`
}

type MCPPromptsCapability struct {
	ListChanged bool `mapstructure:"list_changed"`
}

// SecurityConfig contains security settings
type SecurityConfig struct {
	RateLimitEnabled bool              `mapstructure:"rate_limit_enabled"`
	RateLimitRPS     int               `mapstructure:"rate_limit_rps"`
	AuditLogEnabled  bool              `mapstructure:"audit_log_enabled"`
	AllowedOrigins   []string          `mapstructure:"allowed_origins"`
	Permissions      map[string][]string `mapstructure:"permissions"`
}

// LoggingConfig contains logging settings
type LoggingConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"` // json or text
	Output     string `mapstructure:"output"` // stdout, stderr, or file path
	Structured bool   `mapstructure:"structured"`
}

// Load loads configuration from various sources
func Load() (*Config, error) {
	v := viper.New()
	
	// Set defaults
	setDefaults(v)
	
	// Set config name and paths
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./configs")
	v.AddConfigPath("/etc/fly-mcp")
	
	// Environment variable support
	v.SetEnvPrefix("FLY_MCP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	
	// Try to read config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found is OK, we'll use defaults and env vars
	}
	
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}
	
	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}
	
	return &config, nil
}

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.read_timeout", 30)
	v.SetDefault("server.write_timeout", 30)
	v.SetDefault("server.idle_timeout", 120)
	
	// Fly.io defaults
	v.SetDefault("fly.base_url", "https://api.machines.dev")
	v.SetDefault("fly.timeout", 30)
	
	// MCP defaults
	v.SetDefault("mcp.version", "2024-11-05")
	v.SetDefault("mcp.server_info.name", "fly-mcp")
	v.SetDefault("mcp.server_info.version", "0.1.0")
	v.SetDefault("mcp.capabilities.tools.list_changed", true)
	v.SetDefault("mcp.capabilities.resources.subscribe", false)
	v.SetDefault("mcp.capabilities.resources.list_changed", true)
	v.SetDefault("mcp.capabilities.prompts.list_changed", false)
	
	// Security defaults
	v.SetDefault("security.rate_limit_enabled", true)
	v.SetDefault("security.rate_limit_rps", 10)
	v.SetDefault("security.audit_log_enabled", true)
	v.SetDefault("security.allowed_origins", []string{"*"})
	
	// Logging defaults
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")
	v.SetDefault("logging.output", "stdout")
	v.SetDefault("logging.structured", true)
	
	// Environment default
	v.SetDefault("environment", getEnvironment())
}

// getEnvironment determines the current environment
func getEnvironment() string {
	if env := os.Getenv("FLY_MCP_ENVIRONMENT"); env != "" {
		return env
	}
	if os.Getenv("FLY_APP_NAME") != "" {
		return "production"
	}
	return "local"
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate Fly.io configuration
	if c.Fly.APIToken == "" {
		return fmt.Errorf("fly.api_token is required")
	}
	
	// Validate server configuration
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("server.port must be between 1 and 65535")
	}
	
	// Validate logging configuration
	validLevels := []string{"debug", "info", "warn", "error"}
	if !contains(validLevels, c.Logging.Level) {
		return fmt.Errorf("logging.level must be one of: %v", validLevels)
	}
	
	validFormats := []string{"json", "text"}
	if !contains(validFormats, c.Logging.Format) {
		return fmt.Errorf("logging.format must be one of: %v", validFormats)
	}
	
	return nil
}

// IsLocal returns true if running in local development environment
func (c *Config) IsLocal() bool {
	return c.Environment == "local"
}

// IsProduction returns true if running in production environment
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// LoadFromFile loads configuration from a specific file
func LoadFromFile(configFile string) (*Config, error) {
	v := viper.New()

	// Set defaults
	setDefaults(v)

	// Set config file
	v.SetConfigFile(configFile)

	// Environment variable support
	v.SetEnvPrefix("FLY_MCP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file %s: %w", configFile, err)
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
