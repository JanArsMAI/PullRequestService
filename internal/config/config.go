package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	App AppConfig `yaml:"app"`
}

type AppConfig struct {
	Server  ServerConfig  `yaml:"server"`
	Logging LoggingConfig `yaml:"logging"`
}

type ServerConfig struct {
	Port int `yaml:"port"`
}

type LoggingConfig struct {
	Level string `yaml:"level"`
}

func MustLoad(path string) (*AppConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal yaml: %w", err)
	}

	return &cfg.App, nil
}
