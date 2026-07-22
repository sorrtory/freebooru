package core

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/sorrtory/freebooru/internal/collection"
	"github.com/sorrtory/freebooru/internal/config"
)

// CollectionInfo is safe collection metadata for application frontends.
type CollectionInfo struct {
	Name           string
	Comment        string
	Default        bool
	TagCount       int
	RequiredCount  int
	Storages       []string
	FileCount      int64
	TotalSizeBytes int64
}

// CreateCollection creates and publishes one starter-backed collection.
func (c *Core) CreateCollection(ctx context.Context, name string) (CollectionInfo, error) {
	c.sessionMu.Lock()
	defer c.sessionMu.Unlock()
	if err := ctx.Err(); err != nil {
		return CollectionInfo{}, err
	}
	if c.catalog == nil || c.graph == nil {
		return CollectionInfo{}, fmt.Errorf("configuration has not been checked")
	}
	if c.session != nil {
		return CollectionInfo{}, fmt.Errorf("close collection %q before creating a collection", c.session.name)
	}
	if _, _, exists := c.catalog.Collection(name); exists {
		return CollectionInfo{}, fmt.Errorf("collection %q already exists", name)
	}
	template, _, ok := c.catalog.Collection(c.config.DefaultCollection)
	if !ok {
		return CollectionInfo{}, fmt.Errorf("default collection %q is unavailable", c.config.DefaultCollection)
	}
	createdConfig := template
	createdConfig.Name = name
	createdConfig.Location = "$HOME/.local/share/freebooru/collections/" + name + ".sqlite"
	createdConfig.Comment = "FreeBooru collection"
	created, err := config.CreateCollectionConfig(c.paths, createdConfig)
	if err != nil {
		return CollectionInfo{}, err
	}
	path := filepath.Join(c.paths.Collections, created.Name+".yaml")
	previousCatalog, previousGraph := c.catalog, c.graph
	catalog, diagnostics := config.LoadCatalog(c.paths, c.config)
	graph, graphDiagnostics := config.BuildValidatedGraph(catalog)
	diagnostics = append(diagnostics, graphDiagnostics...)
	if diagnostics.HasErrors() {
		if removeErr := os.Remove(path); removeErr != nil {
			return CollectionInfo{}, fmt.Errorf("new collection is invalid and rollback failed: %w", removeErr)
		}
		c.catalog, c.graph = previousCatalog, previousGraph
		return CollectionInfo{}, fmt.Errorf("new collection %q is invalid: %s", name, diagnostics[0].Message)
	}
	c.catalog, c.graph = catalog, graph
	return c.collectionInfoWithoutStats(created.Name)
}

// DescribeCollection returns safe metadata and aggregate persisted file stats.
func (c *Core) DescribeCollection(ctx context.Context, name string) (info CollectionInfo, err error) {
	c.sessionMu.Lock()
	info, err = c.collectionInfoWithoutStats(name)
	c.sessionMu.Unlock()
	if err != nil {
		return CollectionInfo{}, err
	}
	session, release, err := c.acquireCollectionSession(ctx, info.Name)
	if err != nil {
		return CollectionInfo{}, err
	}
	defer release(&err)
	err = session.database.ForEachFile(ctx, func(file collection.FileRecord) error {
		info.FileCount++
		info.TotalSizeBytes += file.SizeBytes
		return ctx.Err()
	})
	if err != nil {
		return CollectionInfo{}, fmt.Errorf("summarize collection %q: %w", info.Name, err)
	}
	return info, nil
}

func (c *Core) collectionInfoWithoutStats(name string) (CollectionInfo, error) {
	if c.catalog == nil {
		return CollectionInfo{}, fmt.Errorf("configuration has not been checked")
	}
	collectionConfig, _, ok := c.catalog.Collection(name)
	if !ok {
		return CollectionInfo{}, fmt.Errorf("collection %q is missing or invalid", name)
	}
	references, ok := c.catalog.CollectionReferences(collectionConfig.Name)
	if !ok {
		return CollectionInfo{}, fmt.Errorf("collection %q references are unavailable", collectionConfig.Name)
	}
	storages := make([]string, 0)
	tags := make(map[string]struct{})
	for _, reference := range references.Imported {
		if reference.Storage != "" {
			storages = append(storages, reference.Storage)
		}
		if reference.Tag != "" {
			tags[normalizeStateName(reference.Tag)] = struct{}{}
		}
	}
	sort.Slice(storages, func(i, j int) bool { return normalizeStateName(storages[i]) < normalizeStateName(storages[j]) })
	return CollectionInfo{
		Name: collectionConfig.Name, Comment: collectionConfig.Comment,
		Default:  strings.EqualFold(collectionConfig.Name, c.config.DefaultCollection),
		TagCount: len(tags) + len(config.SystemTags()), RequiredCount: len(references.Required),
		Storages: storages,
	}, nil
}

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
