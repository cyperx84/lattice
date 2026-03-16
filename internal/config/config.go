package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	LLMCmd        string `yaml:"llm_cmd"`
	DefaultModels int    `yaml:"default_models"`
	VaultPath     string `yaml:"vault_path,omitempty"`
	VaultFolder   string `yaml:"vault_folder,omitempty"`
}

func DefaultConfig() *Config {
	return &Config{
		LLMCmd:        "claude -p",
		DefaultModels: 3,
	}
}

func Load() *Config {
	cfg := DefaultConfig()

	home, err := os.UserHomeDir()
	if err != nil {
		return cfg
	}

	path := filepath.Join(home, ".config", "lattice", "config.yml")
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg
	}

	_ = yaml.Unmarshal(data, cfg)
	if cfg.DefaultModels == 0 {
		cfg.DefaultModels = 3
	}
	if cfg.LLMCmd == "" {
		cfg.LLMCmd = "claude -p"
	}
	return cfg
}

func Save(cfg *Config) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	dir := filepath.Join(home, ".config", "lattice")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(dir, "config.yml"), data, 0644)
}
