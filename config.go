package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the configuration file structure
type Config struct {
	Git struct {
		RepoURL    string `yaml:"repoUrl"`
		CommitHash string `yaml:"commitHash"`
	} `yaml:"git"`
	Files []struct {
		Source      string `yaml:"source"`
		Destination string `yaml:"destination"`
	} `yaml:"files"`
	SSHPrivateKeyPath string `yaml:"sshPrivateKeyPath"`
}

// LoadConfig loads and validates configuration from the specified file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// Validate checks if all required fields are present
func (c *Config) Validate() error {
	if c.Git.RepoURL == "" {
		return fmt.Errorf("git.repoUrl is required")
	}
	if c.Git.CommitHash == "" {
		return fmt.Errorf("git.commitHash is required")
	}
	if len(c.Files) == 0 {
		return fmt.Errorf("at least one file must be specified")
	}
	if c.SSHPrivateKeyPath == "" {
		return fmt.Errorf("sshPrivateKeyPath is required")
	}
	return nil
}
