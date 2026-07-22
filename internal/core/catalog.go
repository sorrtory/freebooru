package core

import (
	"fmt"
	"sort"

	"github.com/sorrtory/freebooru/internal/config"
)

// ListCollections returns every usable configured collection.
func (c *Core) ListCollections() ([]config.CollectionConfig, error) {
	c.sessionMu.Lock()
	defer c.sessionMu.Unlock()
	if c.catalog == nil {
		return nil, fmt.Errorf("configuration has not been checked")
	}
	return c.catalog.Collections(), nil
}

// ListCollectionStorages returns storages imported by one collection.
func (c *Core) ListCollectionStorages(collectionName string) ([]config.StorageProvider, error) {
	c.sessionMu.Lock()
	defer c.sessionMu.Unlock()
	references, err := c.collectionReferences(collectionName)
	if err != nil {
		return nil, err
	}
	storages := make([]config.StorageProvider, 0)
	seen := make(map[string]struct{})
	for _, reference := range references.Imported {
		if reference.Storage == "" {
			continue
		}
		key := normalizeStateName(reference.Storage)
		if _, exists := seen[key]; exists {
			continue
		}
		storage, _, ok := c.catalog.Storage(reference.Storage)
		if ok {
			seen[key] = struct{}{}
			storages = append(storages, storage)
		}
	}
	sort.Slice(storages, func(i, j int) bool {
		return normalizeStateName(storages[i].Name) < normalizeStateName(storages[j].Name)
	})
	return storages, nil
}

// ListCollectionTags returns tags available to one collection, including storage.
func (c *Core) ListCollectionTags(collectionName string) ([]config.TagConfig, error) {
	tags, err := c.SearchCollectionTags(collectionName, "")
	if err != nil {
		return nil, err
	}
	tags = append(tags, config.SystemTags()...)
	sort.Slice(tags, func(i, j int) bool {
		return normalizeStateName(tags[i].Name) < normalizeStateName(tags[j].Name)
	})
	return tags, nil
}

// StorageConfigSource returns the global storage configuration path.
func (c *Core) StorageConfigSource() (string, error) {
	c.sessionMu.Lock()
	defer c.sessionMu.Unlock()
	if c.catalog == nil {
		return "", fmt.Errorf("configuration has not been checked")
	}
	return c.paths.Storage, nil
}

// CollectionConfigSource returns the YAML file defining a collection.
func (c *Core) CollectionConfigSource(name string) (string, error) {
	c.sessionMu.Lock()
	defer c.sessionMu.Unlock()
	if c.catalog == nil {
		return "", fmt.Errorf("configuration has not been checked")
	}
	_, source, ok := c.catalog.Collection(name)
	if !ok {
		return "", fmt.Errorf("collection %q is missing or invalid", name)
	}
	return source.File, nil
}

// TagConfigSource returns the YAML file defining an imported collection tag.
func (c *Core) TagConfigSource(collectionName, tagName string) (string, error) {
	c.sessionMu.Lock()
	defer c.sessionMu.Unlock()
	references, err := c.collectionReferences(collectionName)
	if err != nil {
		return "", err
	}
	available := newCollectionAvailability(references)
	key := normalizeStateName(tagName)
	if _, ok := available.importedTags[key]; !ok {
		return "", fmt.Errorf("tag %q is not imported by the collection", tagName)
	}
	_, source, ok := c.catalog.Tag(tagName)
	if !ok || key == "storage" {
		return "", fmt.Errorf("tag %q has no editable tag YAML source", tagName)
	}
	return source.File, nil
}

func (c *Core) collectionReferences(name string) (config.ResolvedReferences, error) {
	if c.catalog == nil {
		return config.ResolvedReferences{}, fmt.Errorf("configuration has not been checked")
	}
	if name == "" {
		name = c.config.DefaultCollection
	}
	references, ok := c.catalog.CollectionReferences(name)
	if !ok {
		return config.ResolvedReferences{}, fmt.Errorf("collection %q is missing or invalid", name)
	}
	return references, nil
}
