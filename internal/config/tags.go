package config

import (
	"fmt"
	"strings"
)

// TagConfig contains fields required for initial loading and verification.
// Tag relationship semantics can be added without changing document loading.
type TagConfig struct {
	Name  string `yaml:"name"`
	Type  string `yaml:"type"`
	Group string `yaml:"group,omitempty"`
}

// VerifyTagConfig checks the currently implemented tag fields.
func VerifyTagConfig(tag TagConfig) error {
	if strings.TrimSpace(tag.Name) == "" {
		return fmt.Errorf("name is required")
	}
	if strings.TrimSpace(tag.Type) == "" {
		return fmt.Errorf("type is required")
	}
	return nil
}
