package config

import (
	"fmt"
	"strings"
)

type CollectionConfig struct {
	Name     string               `yaml:"name"`
	Location string               `yaml:"location"`
	Storage  []string             `yaml:"storage,omitempty"`
	Tags     CollectionTagImports `yaml:"tags,omitempty"`
}

type CollectionTagImports struct {
	Require []TagReference `yaml:"require,omitempty"`
	Import  []TagReference `yaml:"import,omitempty"`
}

type TagReference struct {
	Tag   string `yaml:"tag,omitempty"`
	Group string `yaml:"group,omitempty"`
}

func VerifyCollectionConfig(collection CollectionConfig) error {
	if strings.TrimSpace(collection.Name) == "" {
		return fmt.Errorf("name is required")
	}
	if strings.TrimSpace(collection.Location) == "" {
		return fmt.Errorf("location is required")
	}
	return nil
}
