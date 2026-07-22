package core

import (
	"context"
	"errors"
	"fmt"

	"github.com/sorrtory/freebooru/internal/collection"
	"github.com/sorrtory/freebooru/internal/config"
	"github.com/sorrtory/freebooru/internal/evaluator"
	contentstorage "github.com/sorrtory/freebooru/internal/storage"
)

// StorageMutationRequest identifies one physical and logical storage change.
type StorageMutationRequest struct {
	Collection string
	SHA256     string
	Storage    string
}

// StorageMutationResult reports committed logical and physical effects.
type StorageMutationResult struct {
	Changed     bool
	FileDeleted bool
	CopyCreated bool
	CopyDeleted bool
	Evaluation  evaluator.Evaluation
}

// AddStorage copies or verifies content before committing its logical row.
func (c *Core) AddStorage(
	ctx context.Context,
	request StorageMutationRequest,
) (result StorageMutationResult, err error) {
	collectionName, references, availability, err := c.tagMutationCollection(request.Collection)
	if err != nil {
		return StorageMutationResult{}, err
	}
	target, err := c.resolveMutableStorage(request.Storage, availability)
	if err != nil {
		return StorageMutationResult{}, err
	}
	database, err := c.openCollectionDatabase(ctx, collectionName)
	if err != nil {
		return StorageMutationResult{}, err
	}
	defer func() {
		err = errors.Join(err, closeNamedCollection(database, collectionName))
	}()
	record, err := database.File(ctx, request.SHA256)
	if err != nil {
		return StorageMutationResult{}, fmt.Errorf("load file %q: %w", request.SHA256, err)
	}
	proposed := append([]string(nil), record.Storages...)
	if !containsStorage(proposed, target.provider.Name) {
		proposed = append(proposed, target.provider.Name)
	}
	result.Evaluation, err = c.validateStorageMutation(record, proposed, references.Required, availability)
	if err != nil {
		return result, err
	}
	sourcePath, err := c.storageCopySource(record, target, availability)
	if err != nil {
		return result, err
	}
	stored, err := target.backend.Store(ctx, sourcePath)
	if err != nil {
		return result, fmt.Errorf("store content in %q: %w", target.provider.Name, err)
	}
	if stored.SHA256 != record.SHA256 || stored.SizeBytes != record.SizeBytes {
		return result, errors.Join(
			fmt.Errorf("stored content identity does not match file %q", record.SHA256),
			cleanupStorageCopy(target.backend, stored),
		)
	}
	change, err := database.AddStorage(ctx, record.SHA256, target.provider.Name)
	if err != nil {
		return result, errors.Join(
			fmt.Errorf("persist storage %q for file %q: %w", target.provider.Name, record.SHA256, err),
			cleanupStorageCopy(target.backend, stored),
		)
	}
	result.Changed = change.Changed
	result.CopyCreated = stored.Created
	return result, nil
}

// RemoveStorage commits logical removal before deleting its physical copy.
func (c *Core) RemoveStorage(
	ctx context.Context,
	request StorageMutationRequest,
) (result StorageMutationResult, err error) {
	collectionName, references, availability, err := c.tagMutationCollection(request.Collection)
	if err != nil {
		return StorageMutationResult{}, err
	}
	target, err := c.resolveMutableStorage(request.Storage, availability)
	if err != nil {
		return StorageMutationResult{}, err
	}
	database, err := c.openCollectionDatabase(ctx, collectionName)
	if err != nil {
		return StorageMutationResult{}, err
	}
	defer func() {
		err = errors.Join(err, closeNamedCollection(database, collectionName))
	}()
	record, err := database.File(ctx, request.SHA256)
	if err != nil {
		return StorageMutationResult{}, fmt.Errorf("load file %q: %w", request.SHA256, err)
	}
	if !containsStorage(record.Storages, target.provider.Name) {
		return result, nil
	}
	proposed := removeStorageName(record.Storages, target.provider.Name)
	result.Evaluation, err = c.validateStorageMutation(record, proposed, references.Required, availability)
	if err != nil {
		return result, err
	}
	change, err := database.RemoveStorage(ctx, record.SHA256, target.provider.Name)
	if err != nil {
		return result, fmt.Errorf(
			"persist removal of storage %q for file %q: %w",
			target.provider.Name,
			record.SHA256,
			err,
		)
	}
	result.Changed = change.Changed
	result.FileDeleted = change.FileDeleted
	if !change.Changed {
		return result, nil
	}
	deleted, deleteErr := target.backend.Delete(record.SHA256)
	result.CopyDeleted = deleted
	if deleteErr != nil {
		return result, fmt.Errorf(
			"storage %q was removed from file %q but its copy could not be deleted: %w",
			target.provider.Name,
			record.SHA256,
			deleteErr,
		)
	}
	return result, nil
}

func (c *Core) validateStorageMutation(
	record collection.FileRecord,
	storages []string,
	required []config.TagReference,
	availability collectionAvailability,
) (evaluator.Evaluation, error) {
	values, err := persistedValues(record, c.catalog, availability)
	if err != nil {
		return evaluator.Evaluation{}, fmt.Errorf("reconstruct file %q: %w", record.SHA256, err)
	}
	deleteAssignment(values, "storage")
	if len(storages) > 0 {
		values["storage"] = append([]string(nil), storages...)
	}
	return c.validateTagMutation(values, storages, required)
}

func (c *Core) storageCopySource(
	record collection.FileRecord,
	target localStorage,
	availability collectionAvailability,
) (string, error) {
	targetPath, err := target.backend.ContentPath(record.SHA256)
	if err != nil {
		return "", err
	}
	var sourcePath string
	targetAssigned := false
	for _, name := range record.Storages {
		current, err := c.resolveMutableStorage(name, availability)
		if err != nil {
			return "", err
		}
		path, err := current.backend.ContentPath(record.SHA256)
		if err != nil {
			return "", err
		}
		if path == targetPath && normalizeStateName(name) != normalizeStateName(target.provider.Name) {
			return "", fmt.Errorf(
				"storages %q and %q resolve to the same content path",
				name,
				target.provider.Name,
			)
		}
		if normalizeStateName(name) == normalizeStateName(target.provider.Name) {
			targetAssigned = true
			continue
		}
		if sourcePath == "" {
			sourcePath = path
		}
	}
	if targetAssigned {
		return targetPath, nil
	}
	if sourcePath == "" {
		return "", fmt.Errorf("file %q has no source storage", record.SHA256)
	}
	return sourcePath, nil
}

func cleanupStorageCopy(backend *contentstorage.Local, stored contentstorage.StoredFile) error {
	if !stored.Created {
		return nil
	}
	if _, err := backend.Delete(stored.SHA256); err != nil {
		return fmt.Errorf("remove new storage copy: %w", err)
	}
	return nil
}
