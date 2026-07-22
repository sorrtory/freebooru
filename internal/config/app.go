package config

import "fmt"

type AppConfig struct {
	HTTPPort int `yaml:"http_port"`
}

func DefaultAppConfig() AppConfig {
	return AppConfig{HTTPPort: 8080}
}

func VerifyAppConfig(cfg AppConfig) error {
	if cfg.HTTPPort < 1 || cfg.HTTPPort > 65535 {
		return fmt.Errorf("http_port must be between 1 and 65535")
	}
	return nil
}

func LoadApp(path string) (AppConfig, error) {
	return (YAMLFile[AppConfig]{
		Path:     path,
		Validate: VerifyAppConfig,
	}).Read()
}
