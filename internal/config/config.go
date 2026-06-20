package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds the application configuration.
type Config struct {
	ScanFolders []string `json:"scan_folders"`
	Editor      string   `json:"editor,omitempty"`
}

// defaultConfigPath returns the default config file location.
func defaultConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine home directory: %w", err)
	}
	return filepath.Join(home, ".config", "delta", "config.json"), nil
}

// Load reads the config from the given path, or the default path if empty.
// If no config file exists, a default one is created.
func Load(path string) (*Config, error) {
	if path == "" {
		var err error
		path, err = defaultConfigPath()
		if err != nil {
			return nil, err
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return createDefaultConfig(path)
		}
		return nil, fmt.Errorf("could not read config file %s: %w", path, err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("could not parse config file: %w", err)
	}

	return &cfg, nil
}

// createDefaultConfig writes a default config file and returns it.
func createDefaultConfig(path string) (*Config, error) {
	cfg := &Config{
		ScanFolders: []string{},
		Editor:      "code",
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("could not create config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("could not marshal default config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return nil, fmt.Errorf("could not write default config: %w", err)
	}

	return cfg, nil
}

// Save writes the config to the given path, or the default path if empty.
func (c *Config) Save(path string) error {
	if path == "" {
		var err error
		path, err = defaultConfigPath()
		if err != nil {
			return err
		}
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("could not create config directory: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("could not marshal config: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

// AddFolder adds a scan folder to the config if not already present.
func (c *Config) AddFolder(folder string) error {
	cleaned := filepath.Clean(folder)
	for _, existing := range c.ScanFolders {
		if filepath.Clean(existing) == cleaned {
			return fmt.Errorf("folder already in config: %s", cleaned)
		}
	}
	c.ScanFolders = append(c.ScanFolders, cleaned)
	return nil
}
