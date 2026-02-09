package config

import (
	"fmt"
	"os"
)

type Config struct {
	ClaudeAPIKey       string
	XAPIKey            string
	XAPISecret         string
	XAccessToken       string
	XAccessTokenSecret string
}

func Load() (*Config, error) {
	cfg := &Config{
		ClaudeAPIKey:       os.Getenv("ANTHROPIC_API_KEY"),
		XAPIKey:            os.Getenv("X_API_KEY"),
		XAPISecret:         os.Getenv("X_API_SECRET"),
		XAccessToken:       os.Getenv("X_ACCESS_TOKEN"),
		XAccessTokenSecret: os.Getenv("X_ACCESS_TOKEN_SECRET"),
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

// Checks if X API credentials are configured
func (c *Config) HasXAPICredentials() bool {
	return c.XAPIKey != "" &&
		c.XAPISecret != "" &&
		c.XAccessToken != "" &&
		c.XAccessTokenSecret != ""
}
