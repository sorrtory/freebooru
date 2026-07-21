package config

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/goccy/go-yaml"
)

type AppConfig struct {
	HTTPPort int `yaml:"http_port"`
}

type AppConfigParser struct {
	configFile string
}

var _ Parser = (*AppConfigParser)(nil)

func NewAppConfigParser(log *slog.Logger, configFile string) (*AppConfigParser, error) {
	if configFile == "" {
		return nil, fmt.Errorf("config file path is empty")
	}

	return &AppConfigParser{
		configFile: configFile,
	}, nil
}

func (p *AppConfigParser) GetDefaultConfig() AppConfig {
	return AppConfig{
		HTTPPort: 8080,
	}
}

func (p *AppConfigParser) Read() (ConfigStruct, error) {
	data, err := os.ReadFile(p.configFile)
	if err != nil {
		return AppConfig{}, fmt.Errorf("read config: %w", err)
	}

	var cfg AppConfig

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return AppConfig{}, fmt.Errorf("parse config: %w", err)
	}

	return cfg, nil
}

func (p *AppConfigParser) Write(config ConfigStruct) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(p.configFile, data, 0644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}

func (p *AppConfigParser) Check() (CheckStatus, error) {
	if _, err := os.Stat(p.configFile); os.IsNotExist(err) {
		return NOT_EXIST, nil
	} else if err != nil {
		return READ_ERROR, fmt.Errorf("failed to stat config file: %v", err)
	}

	return OK, nil
}

func (p *AppConfigParser) Validate(config ConfigStruct) error {
