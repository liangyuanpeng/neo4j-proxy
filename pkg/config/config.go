package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config represents the proxy configuration
type Config struct {
	ProxyPort int                      `json:"proxy_port"`
	Tenants   map[string]TenantConfig  `json:"tenants"`
}

// TenantConfig represents configuration for a single tenant
type TenantConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

// Load loads configuration from environment variables and config file
func Load() (*Config, error) {
	cfg := &Config{
		ProxyPort: 7687, // Default Neo4j Bolt port
		Tenants:   make(map[string]TenantConfig),
	}

	// Load from config file if exists
	if configFile := os.Getenv("CONFIG_FILE"); configFile != "" {
		if err := loadFromFile(cfg, configFile); err != nil {
			return nil, fmt.Errorf("failed to load config file: %w", err)
		}
	} else {
		// Load default configuration for testing
		loadDefaultConfig(cfg)
	}

	return cfg, nil
}

func loadFromFile(cfg *Config, filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, cfg)
}

func loadDefaultConfig(cfg *Config) {
	// Default configuration for your test environment
	cfg.Tenants = map[string]TenantConfig{
		"tenant1": {
			Host: "yunhorn187",
			Port: 17687,
		},
		"tenant2": {
			Host: "yunhorn187", 
			Port: 27687,
		},
	}
}