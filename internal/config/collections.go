package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
)

// CollectionConfig defines one collection and its available tags.
type CollectionConfig struct {
	Name     string               `yaml:"name"`
	Location string               `yaml:"location,omitempty"`
	Comment  string               `yaml:"comment,omitempty"`
	Tags     CollectionTagImports `yaml:"tags"`
}

// StarterCollectionConfigForName builds a new collection using starter groups.
func StarterCollectionConfigForName(app AppConfig, name string) CollectionConfig {
	collection := StarterCollectionConfig(app)
	collection.Name = name
	collection.Location = "$HOME/.local/share/freebooru/collections/" + name + ".sqlite"
	collection.Comment = "FreeBooru collection"
	return collection
}

// CreateStarterCollection atomically creates one non-overwriting collection YAML.
func CreateStarterCollection(paths Paths, app AppConfig, name string) (CollectionConfig, error) {
	collection := StarterCollectionConfigForName(app, name)
	return CreateCollectionConfig(paths, collection)
}

// CreateCollectionConfig atomically publishes one validated non-overwriting YAML.
func CreateCollectionConfig(paths Paths, collection CollectionConfig) (CollectionConfig, error) {
	name := collection.Name
	if err := VerifyCollectionConfig(collection); err != nil {
		return CollectionConfig{}, err
	}
	data, err := yaml.Marshal(collection)
	if err != nil {
		return CollectionConfig{}, fmt.Errorf("marshal collection %q: %w", name, err)
	}
	if err := os.MkdirAll(paths.Collections, 0o755); err != nil {
		return CollectionConfig{}, fmt.Errorf("create collections directory: %w", err)
	}
	target := filepath.Join(paths.Collections, name+".yaml")
	if _, err := os.Stat(target); err == nil {
		return CollectionConfig{}, fmt.Errorf("collection %q already exists", name)
	} else if !errors.Is(err, fs.ErrNotExist) {
		return CollectionConfig{}, fmt.Errorf("stat collection %q: %w", name, err)
	}
	temporary, err := os.CreateTemp(paths.Collections, ".freebooru-collection-*")
	if err != nil {
		return CollectionConfig{}, fmt.Errorf("create temporary collection %q: %w", name, err)
	}
	temporaryName := temporary.Name()
	defer func() { _ = os.Remove(temporaryName) }()
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return CollectionConfig{}, fmt.Errorf("protect temporary collection %q: %w", name, err)
	}
	if _, err := temporary.Write(data); err != nil {
		_ = temporary.Close()
		return CollectionConfig{}, fmt.Errorf("write temporary collection %q: %w", name, err)
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return CollectionConfig{}, fmt.Errorf("sync temporary collection %q: %w", name, err)
	}
	if err := temporary.Close(); err != nil {
		return CollectionConfig{}, fmt.Errorf("close temporary collection %q: %w", name, err)
	}
	if err := os.Link(temporaryName, target); err != nil {
		if errors.Is(err, fs.ErrExist) {
			return CollectionConfig{}, fmt.Errorf("collection %q already exists", name)
		}
		return CollectionConfig{}, fmt.Errorf("publish collection %q: %w", name, err)
	}
	return collection, nil
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
	collection.Comment = "Default FreeBooru collection"
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
	if err := verifyComment(collection.Comment); err != nil {
		return err
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
