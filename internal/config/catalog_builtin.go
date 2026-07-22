package config

import "sort"

// addStorageTag exposes usable storage providers through the reserved built-in
// multivalue tag. Invalid or duplicate providers were already omitted from the
// storage index, so they cannot become assignable values.
func (c *Catalog) addStorageTag(sourcePath string) {
	storageKeys := make([]string, 0, len(c.storages))
	for key := range c.storages {
		storageKeys = append(storageKeys, key)
	}
	sort.Strings(storageKeys)

	values := make([]PredefinedValue, 0, len(storageKeys))
	for _, key := range storageKeys {
		values = append(values, PredefinedValue{Val: c.storages[key].value.Name})
	}
	c.tags[normalizeName("storage")] = catalogEntry[TagConfig]{
		value: TagConfig{
			Name:   "storage",
			Type:   TagTypeMultivalue,
			Values: values,
		},
		source: Source{File: sourcePath, Document: 1},
	}
}
