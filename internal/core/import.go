package core

import (
	"fmt"
	"sort"

	"github.com/sorrtory/freebooru/internal/config"
	"github.com/sorrtory/freebooru/internal/evaluator"
)

// ImportRequest contains frontend-independent input for one file import.
// An empty Collection selects the configured default collection.
type ImportRequest struct {
	Collection string
	SourcePath string
	Tags       map[string]any
}

type preparedImport struct {
	collection config.CollectionConfig
	values     map[string]any
	storages   []config.StorageProvider
}

func (c *Core) prepareImport(request ImportRequest) (preparedImport, error) {
	if c.catalog == nil || c.graph == nil {
		return preparedImport{}, fmt.Errorf("configuration has not been checked")
	}
	collectionName := request.Collection
	if collectionName == "" {
		collectionName = c.config.DefaultCollection
	}
	collectionConfig, _, ok := c.catalog.Collection(collectionName)
	if !ok {
		return preparedImport{}, fmt.Errorf("collection %q is missing or invalid", collectionName)
	}
	references, ok := c.catalog.CollectionReferences(collectionConfig.Name)
	if !ok {
		return preparedImport{}, fmt.Errorf("collection %q references are unavailable", collectionConfig.Name)
	}
	availability := newCollectionAvailability(references)
	values, storageNames, err := c.collectImportAssignments(request.Tags, availability)
	if err != nil {
		return preparedImport{}, err
	}
	missingRequired, err := c.applyRequiredAssignments(values, storageNames, references.Required)
	if err != nil {
		return preparedImport{}, err
	}
	if len(missingRequired) > 0 {
		return preparedImport{}, fmt.Errorf(
			"required tag %q needs an explicit value",
			missingRequired[0],
		)
	}
	if len(storageNames) == 0 {
		key := normalizeStateName(c.config.DefaultStorageName)
		if _, ok := availability.importedStorages[key]; !ok {
			return preparedImport{}, fmt.Errorf(
				"default storage %q is not available to the collection; assign storage explicitly",
				c.config.DefaultStorageName,
			)
		}
		storageNames[key] = c.config.DefaultStorageName
	}
	storages, canonicalStorageNames, err := c.resolveImportStorages(storageNames, availability)
	if err != nil {
		return preparedImport{}, err
	}
	values["storage"] = canonicalStorageNames
	if err := c.validateImportState(values, canonicalStorageNames, references.Required); err != nil {
		return preparedImport{}, err
	}
	return preparedImport{
		collection: collectionConfig,
		values:     values,
		storages:   storages,
	}, nil
}

func (c *Core) collectImportAssignments(
	requested map[string]any,
	availability collectionAvailability,
) (map[string]any, map[string]string, error) {
	values := make(map[string]any, len(requested)+1)
	storages := make(map[string]string)
	seen := make(map[string]struct{}, len(requested))
	for name, value := range requested {
		key := normalizeStateName(name)
		if _, duplicate := seen[key]; duplicate {
			return nil, nil, fmt.Errorf("import tag %q is assigned more than once", name)
		}
		seen[key] = struct{}{}
		if key == "storage" {
			names, ok := value.([]string)
			if !ok {
				return nil, nil, fmt.Errorf("storage expects a string slice")
			}
			for _, storage := range names {
				storages[normalizeStateName(storage)] = storage
			}
			continue
		}
		if _, ok := availability.importedTags[key]; !ok {
			return nil, nil, fmt.Errorf("tag %q is not imported by the collection", name)
		}
		tag, _, ok := c.catalog.Tag(name)
		if !ok {
			return nil, nil, fmt.Errorf("tag %q does not exist", name)
		}
		value = canonicalTagValue(tag, value)
		if tag.Type == config.TagTypeBool {
			if assigned, ok := value.(bool); ok && !assigned {
				continue
			}
		}
		values[tag.Name] = value
	}
	return values, storages, nil
}

func (c *Core) applyRequiredAssignments(
	values map[string]any,
	storages map[string]string,
	required []config.TagReference,
) ([]string, error) {
	missing := make([]string, 0)
	for _, reference := range required {
		if reference.Storage != "" {
			storages[normalizeStateName(reference.Storage)] = reference.Storage
			continue
		}
		if _, assigned := values[reference.Tag]; assigned {
			continue
		}
		tag, _, ok := c.catalog.Tag(reference.Tag)
		if !ok {
			return nil, fmt.Errorf("required tag %q does not exist", reference.Tag)
		}
		if tag.Type == config.TagTypeBool {
			values[tag.Name] = true
			continue
		}
		missing = append(missing, tag.Name)
	}
	return missing, nil
}

func (c *Core) resolveImportStorages(
	names map[string]string,
	availability collectionAvailability,
) ([]config.StorageProvider, []string, error) {
	keys := make([]string, 0, len(names))
	for key := range names {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	providers := make([]config.StorageProvider, 0, len(keys))
	canonicalNames := make([]string, 0, len(keys))
	for _, key := range keys {
		name := names[key]
		if _, ok := availability.importedStorages[key]; !ok {
			return nil, nil, fmt.Errorf("storage %q is not imported by the collection", name)
		}
		provider, _, ok := c.catalog.Storage(name)
		if !ok {
			return nil, nil, fmt.Errorf("storage %q does not exist", name)
		}
		providers = append(providers, provider)
		canonicalNames = append(canonicalNames, provider.Name)
	}
	return providers, canonicalNames, nil
}

func (c *Core) validateImportState(
	values map[string]any,
	storages []string,
	required []config.TagReference,
) error {
	state, err := evaluator.NewFileState(c.catalog, values)
	if err != nil {
		return fmt.Errorf("validate import tags: %w", err)
	}
	if err := verifyRequiredState(state, storages, required); err != nil {
		return err
	}
	checker, err := evaluator.New(c.catalog, c.graph)
	if err != nil {
		return fmt.Errorf("create import evaluator: %w", err)
	}
	result := checker.ValidateFile(state)
	if !result.Valid() {
		return &EvaluationError{Operation: "import", Evaluation: result}
	}
	return nil
}
