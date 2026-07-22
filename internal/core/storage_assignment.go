package core

import (
	"fmt"

	"github.com/sorrtory/freebooru/internal/config"
	contentstorage "github.com/sorrtory/freebooru/internal/storage"
)

type localStorage struct {
	provider config.StorageProvider
	backend  *contentstorage.Local
}

func (c *Core) resolveMutableStorage(
	name string,
	availability collectionAvailability,
) (localStorage, error) {
	if _, ok := availability.importedStorages[normalizeStateName(name)]; !ok {
		return localStorage{}, fmt.Errorf("storage %q is not imported by the collection", name)
	}
	provider, _, ok := c.catalog.Storage(name)
	if !ok {
		return localStorage{}, fmt.Errorf("storage %q does not exist", name)
	}
	root, err := config.ExpandPath(provider.Path)
	if err != nil {
		return localStorage{}, fmt.Errorf("resolve storage %q path: %w", provider.Name, err)
	}
	backend, err := contentstorage.NewLocal(root)
	if err != nil {
		return localStorage{}, fmt.Errorf("create storage %q backend: %w", provider.Name, err)
	}
	return localStorage{provider: provider, backend: backend}, nil
}

func containsStorage(storages []string, name string) bool {
	for _, storage := range storages {
		if normalizeStateName(storage) == normalizeStateName(name) {
			return true
		}
	}
	return false
}

func removeStorageName(storages []string, name string) []string {
	result := make([]string, 0, len(storages))
	for _, storage := range storages {
		if normalizeStateName(storage) != normalizeStateName(name) {
			result = append(result, storage)
		}
	}
	return result
}
