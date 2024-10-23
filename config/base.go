package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type ServerConfig struct {
	Server
	Pow
}

type ClientConfig struct {
	Client
	Pow
}

func LoadServerConfig() (*ServerConfig, error) {
	cfg := &ServerConfig{}
	if err := envconfig.Process("", cfg); err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}
	return cfg, nil
}

func LoadClientConfig() (*ClientConfig, error) {
	cfg := &ClientConfig{}
	if err := envconfig.Process("", cfg); err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}
	return cfg, nil
}
