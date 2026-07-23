package core

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/sorrtory/freebooru/internal/collection"
	"github.com/sorrtory/freebooru/internal/config"
)

// CollectionTagInfo is safe collection-aware tag metadata for frontends.
type CollectionTagInfo struct {
	Name            string
	Type            config.TagType
	Comment         string
	Groups          []string
	Values          []config.PredefinedValue
	Required        bool
	Imported        bool
	System          bool
	AssignmentCount int64
}

// CollectionStorageInfo is safe collection-aware storage metadata.
type CollectionStorageInfo struct {
	Name           string
	Type           string
	Comment        string
	Imported       bool
	FileCount      int64
	TotalSizeBytes int64
}

// ListCollectionTagInfo returns imported tags first, then other configured tags.
func (c *Core) ListCollectionTagInfo(
	ctx context.Context,
	collectionName string,
) (items []CollectionTagInfo, err error) {
	session, release, err := c.acquireCollectionSession(ctx, collectionName)
	if err != nil {
		return nil, err
	}
	defer release(&err)

	imported, required := referenceSets(session.references)
	for _, tag := range c.catalog.SearchTags("") {
		key := normalizeStateName(tag.Name)
		items = append(items, CollectionTagInfo{
			Name: tag.Name, Type: tag.Type, Comment: tag.Comment,
			Groups:   append([]string(nil), tag.Groups...),
			Values:   append([]config.PredefinedValue(nil), tag.Values...),
			Imported: imported[key], Required: required[key],
		})
	}
	for _, tag := range config.SystemTags() {
		items = append(items, CollectionTagInfo{
			Name: tag.Name, Type: tag.Type, Comment: tag.Comment,
			Imported: true, System: true,
		})
	}
	indexes := make(map[string]int, len(items))
	valueSets := make([]map[string]bool, len(items))
	for index := range items {
		indexes[normalizeStateName(items[index].Name)] = index
		valueSets[index] = make(map[string]bool, len(items[index].Values))
		for _, value := range items[index].Values {
			valueSets[index][normalizeStateName(value.Val)] = true
		}
	}
	err = session.database.ForEachFile(ctx, func(file collection.FileRecord) error {
		for _, tag := range file.Tags {
			if index, ok := indexes[normalizeStateName(tag.Name)]; ok {
				items[index].AssignmentCount++
				for _, value := range persistedHintValues(tag) {
					key := normalizeStateName(value)
					if !valueSets[index][key] && len(valueSets[index]) < 128 {
						items[index].Values = append(items[index].Values, config.PredefinedValue{Val: value})
						valueSets[index][key] = true
					}
				}
			}
		}
		return ctx.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("count collection tag usage: %w", err)
	}
	for index := range items {
		sort.SliceStable(items[index].Values, func(left, right int) bool {
			return strings.ToLower(items[index].Values[left].Val) < strings.ToLower(items[index].Values[right].Val)
		})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Imported != items[j].Imported {
			return items[i].Imported
		}
		return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
	})
	return items, nil
}

func persistedHintValues(tag collection.TagRecord) []string {
	if tag.TextValue != nil {
		return []string{*tag.TextValue}
	}
	if tag.IntegerValue != nil {
		return []string{strconv.FormatInt(*tag.IntegerValue, 10)}
	}
	return append([]string(nil), tag.Values...)
}

// ListCollectionStorageInfo returns imported storage first, then other configured storage.
func (c *Core) ListCollectionStorageInfo(
	ctx context.Context,
	collectionName string,
) (items []CollectionStorageInfo, err error) {
	session, release, err := c.acquireCollectionSession(ctx, collectionName)
	if err != nil {
		return nil, err
	}
	defer release(&err)
	_, importedStorages := resourceReferenceSets(session.references)
	for _, storage := range c.catalog.Storages() {
		items = append(items, CollectionStorageInfo{
			Name: storage.Name, Type: storage.Type, Comment: storage.Comment,
			Imported: importedStorages[normalizeStateName(storage.Name)],
		})
	}
	indexes := make(map[string]int, len(items))
	for index := range items {
		indexes[normalizeStateName(items[index].Name)] = index
	}
	err = session.database.ForEachFile(ctx, func(file collection.FileRecord) error {
		for _, storage := range file.Storages {
			if index, ok := indexes[normalizeStateName(storage)]; ok {
				items[index].FileCount++
				items[index].TotalSizeBytes += file.SizeBytes
			}
		}
		return ctx.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("count collection storage usage: %w", err)
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Imported != items[j].Imported {
			return items[i].Imported
		}
		return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
	})
	return items, nil
}

// ImportCollectionTag makes one configured tag available to a collection.
func (c *Core) ImportCollectionTag(ctx context.Context, collectionName, tagName string) error {
	return c.importCollectionReference(ctx, collectionName, config.TagReference{Tag: tagName})
}

// ImportCollectionStorage makes one configured storage available to a collection.
func (c *Core) ImportCollectionStorage(ctx context.Context, collectionName, storageName string) error {
	return c.importCollectionReference(ctx, collectionName, config.TagReference{Storage: storageName})
}

func (c *Core) importCollectionReference(
	ctx context.Context,
	collectionName string,
	reference config.TagReference,
) error {
	c.sessionMu.Lock()
	defer c.sessionMu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	if c.session != nil {
		return fmt.Errorf("close collection %q before changing its configuration", c.session.name)
	}
	if c.catalog == nil || c.graph == nil {
		return fmt.Errorf("configuration has not been checked")
	}
	collectionConfig, _, ok := c.catalog.Collection(collectionName)
	if !ok {
		return fmt.Errorf("collection %q is unavailable", collectionName)
	}
	references, ok := c.catalog.CollectionReferences(collectionName)
	if !ok {
		return fmt.Errorf("collection %q references are unavailable", collectionName)
	}
	importedTags, importedStorages := resourceReferenceSets(references)
	if reference.Tag != "" {
		tag, _, exists := c.catalog.Tag(reference.Tag)
		if !exists {
			return fmt.Errorf("tag %q is unavailable", reference.Tag)
		}
		reference.Tag = tag.Name
		if importedTags[normalizeStateName(tag.Name)] {
			return fmt.Errorf("tag %q is already imported", tag.Name)
		}
	} else {
		storage, _, exists := c.catalog.Storage(reference.Storage)
		if !exists {
			return fmt.Errorf("storage %q is unavailable", reference.Storage)
		}
		reference.Storage = storage.Name
		if importedStorages[normalizeStateName(storage.Name)] {
			return fmt.Errorf("storage %q is already imported", storage.Name)
		}
	}
	previous := collectionConfig
	collectionConfig.Tags.Import = append(collectionConfig.Tags.Import, reference)
	if err := config.ReplaceCollectionConfig(c.paths, collectionConfig); err != nil {
		return err
	}
	catalog, diagnostics := config.LoadCatalog(c.paths, c.config)
	graph, graphDiagnostics := config.BuildValidatedGraph(catalog)
	diagnostics = append(diagnostics, graphDiagnostics...)
	if diagnostics.HasErrors() {
		if rollbackErr := config.ReplaceCollectionConfig(c.paths, previous); rollbackErr != nil {
			return fmt.Errorf("new collection reference is invalid and rollback failed: %w", rollbackErr)
		}
		return fmt.Errorf("new collection reference is invalid: %s", diagnostics[0].Message)
	}
	c.catalog, c.graph = catalog, graph
	return nil
}

func referenceSets(references config.ResolvedReferences) (map[string]bool, map[string]bool) {
	importedTags, importedStorages := resourceReferenceSets(references)
	if len(importedStorages) > 0 {
		importedTags["storage"] = true
	}
	required := make(map[string]bool)
	for _, reference := range references.Required {
		if reference.Tag != "" {
			required[normalizeStateName(reference.Tag)] = true
		}
		if reference.Storage != "" {
			required["storage"] = true
		}
	}
	return importedTags, required
}

func resourceReferenceSets(references config.ResolvedReferences) (map[string]bool, map[string]bool) {
	tags := make(map[string]bool)
	storages := make(map[string]bool)
	for _, reference := range references.Imported {
		if reference.Tag != "" {
			tags[normalizeStateName(reference.Tag)] = true
		}
		if reference.Storage != "" {
			storages[normalizeStateName(reference.Storage)] = true
		}
	}
	return tags, storages
}
