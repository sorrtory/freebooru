package config

import (
	"fmt"
	"strings"
)

// CollectionConfig defines one collection and its available tags.
type CollectionConfig struct {
	Name     string               `yaml:"name"`
	Location string               `yaml:"location,omitempty"`
	Tags     CollectionTagImports `yaml:"tags"`
}

// CollectionTagImports separates required and optional tag references.
type CollectionTagImports struct {
	Require []TagReference `yaml:"require,omitempty"`
	Import  []TagReference `yaml:"import,omitempty"`
}

// TagReference selects exactly one tag, group, or storage.
type TagReference struct {
	Tag     string `yaml:"tag,omitempty"`
	Group   string `yaml:"group,omitempty"`
	Storage string `yaml:"storage,omitempty"`
}

// DefaultCollectionConfig builds the collection selected by application defaults.
func DefaultCollectionConfig(app AppConfig) CollectionConfig {
	return CollectionConfig{
		Name:     app.DefaultCollection,
		Location: "$HOME/.local/share/freebooru/collections/" + app.DefaultCollection + ".sqlite",
		Tags: CollectionTagImports{
			Require: []TagReference{{Storage: app.DefaultStorageName}},
		},
	}
}

// StarterCollectionConfig adds the starter tag groups created by init.
func StarterCollectionConfig(app AppConfig) CollectionConfig {
	collection := DefaultCollectionConfig(app)
	collection.Tags.Import = []TagReference{
		{Group: "creator"},
		{Group: "universe"},
		{Group: "character"},
		{Group: "general"},
		{Group: "metadata"},
	}
	return collection
}

// CollectionLocation resolves a collection's explicit or default database path.
func CollectionLocation(collection CollectionConfig) (string, error) {
	location := collection.Location
	if location == "" {
		location = "$HOME/.local/share/freebooru/collections/" + collection.Name + ".sqlite"
	}
	return ExpandPath(location)
}

// VerifyCollectionConfig checks a collection definition.
func VerifyCollectionConfig(collection CollectionConfig) error {
	if err := verifyName("name", collection.Name); err != nil {
		return err
	}
	if _, err := CollectionLocation(collection); err != nil {
		return fmt.Errorf("location: %w", err)
	}
	if len(collection.Tags.Require) == 0 && len(collection.Tags.Import) == 0 {
		return fmt.Errorf("tags must contain require or import")
	}
	storageCount := 0
	for _, ref := range append(collection.Tags.Require, collection.Tags.Import...) {
		if err := VerifyTagReference(ref); err != nil {
			return err
		}
		if ref.Storage != "" {
			storageCount++
		}
	}
	if storageCount == 0 {
		return fmt.Errorf("at least one storage reference is required")
	}
	return nil
}

// VerifyTagReference checks that a reference selects exactly one category.
func VerifyTagReference(ref TagReference) error {
	count := 0
	for _, value := range []string{ref.Tag, ref.Group, ref.Storage} {
		if strings.TrimSpace(value) != "" {
			count++
			if err := verifyName("reference name", value); err != nil {
				return err
			}
		}
	}
	if count != 1 {
		return fmt.Errorf("tag reference must contain exactly one of tag, group, or storage")
	}
	return nil
}
