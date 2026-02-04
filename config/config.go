package config

import (
	"fmt"
	"os"
)

type Config struct {
	ClaudeAPIKey string
}

// Loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		ClaudeAPIKey: os.Getenv("ANTHROPIC_API_KEY"),
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Checks if the configuration is valid
func (c *Config) Validate() error {
	if c.ClaudeAPIKey == "" {
		return fmt.Errorf("ANTHROPIC_API_KEY is required")
	}

	return nil
}
