package core

import (
	"sort"

	"github.com/sorrtory/freebooru/internal/collection"
	"github.com/sorrtory/freebooru/internal/config"
)

func importTagRecords(catalog *config.Catalog, values map[string]any) []collection.TagRecord {
	records := make([]collection.TagRecord, 0, len(values)-1)
	names := make([]string, 0, len(values))
	for name := range values {
		if normalizeStateName(name) != "storage" {
			names = append(names, name)
		}
	}
	sort.Slice(names, func(left, right int) bool {
		return normalizeStateName(names[left]) < normalizeStateName(names[right])
	})
	for _, name := range names {
		value := values[name]
		tag, _, _ := catalog.Tag(name)
		if tag.Type == config.TagTypeBool && !value.(bool) {
			continue
		}
		records = append(records, tagRecord(tag, value))
	}
	return records
}

func storageNamesFromProviders(providers []config.StorageProvider) []string {
	names := make([]string, 0, len(providers))
	for _, provider := range providers {
		names = append(names, provider.Name)
	}
	return names
}
