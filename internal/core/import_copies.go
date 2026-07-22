package core

import (
	"context"
	"errors"
	"fmt"

	"github.com/sorrtory/freebooru/internal/config"
	contentstorage "github.com/sorrtory/freebooru/internal/storage"
)

type storedCopy struct {
	storage string
	backend *contentstorage.Local
	file    contentstorage.StoredFile
}

func storeImportCopies(
	ctx context.Context,
	sourcePath string,
	providers []config.StorageProvider,
) ([]storedCopy, error) {
	return storeImportCopiesWith(ctx, sourcePath, providers, func(
		ctx context.Context,
		backend *contentstorage.Local,
		path string,
	) (contentstorage.StoredFile, error) {
		return backend.Store(ctx, path)
	})
}

type storeContent func(
	context.Context,
	*contentstorage.Local,
	string,
) (contentstorage.StoredFile, error)

func storeImportCopiesWith(
	ctx context.Context,
	sourcePath string,
	providers []config.StorageProvider,
	store storeContent,
) ([]storedCopy, error) {
	copies := make([]storedCopy, 0, len(providers))
	destinations := make(map[string]string, len(providers))
	copySource := sourcePath
	for _, provider := range providers {
		root, err := config.ExpandPath(provider.Path)
		if err != nil {
			return copies, fmt.Errorf("resolve storage %q path: %w", provider.Name, err)
		}
		backend, err := contentstorage.NewLocal(root)
		if err != nil {
			return copies, fmt.Errorf("create storage %q backend: %w", provider.Name, err)
		}
		stored, err := store(ctx, backend, copySource)
		if err != nil {
			return copies, fmt.Errorf("store content in %q: %w", provider.Name, err)
		}
		if len(copies) > 0 {
			first := copies[0].file
			if stored.SHA256 != first.SHA256 || stored.SizeBytes != first.SizeBytes {
				if stored.Created {
					copies = append(copies, storedCopy{storage: provider.Name, backend: backend, file: stored})
				}
				return copies, fmt.Errorf("source changed while creating storage copies")
			}
		}
		if previous, duplicate := destinations[stored.ContentPath]; duplicate {
			if stored.Created {
				copies = append(copies, storedCopy{storage: provider.Name, backend: backend, file: stored})
			}
			return copies, fmt.Errorf(
				"storages %q and %q resolve to the same content path",
				previous,
				provider.Name,
			)
		}
		destinations[stored.ContentPath] = provider.Name
		copies = append(copies, storedCopy{storage: provider.Name, backend: backend, file: stored})
		copySource = copies[0].file.ContentPath
	}
	return copies, nil
}

func cleanupStoredCopies(copies []storedCopy) error {
	var cleanup []error
	for index := len(copies) - 1; index >= 0; index-- {
		stored := copies[index]
		if !stored.file.Created {
			continue
		}
		if _, err := stored.backend.Delete(stored.file.SHA256); err != nil {
			cleanup = append(cleanup, fmt.Errorf("remove failed import copy: %w", err))
		}
	}
	return errors.Join(cleanup...)
}

func createdStorageNames(copies []storedCopy) []string {
	names := make([]string, 0, len(copies))
	for _, stored := range copies {
		if stored.file.Created {
			names = append(names, stored.storage)
		}
	}
	return names
}
