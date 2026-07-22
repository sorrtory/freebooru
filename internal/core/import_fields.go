package core

import (
	"fmt"
	"sort"

	"github.com/sorrtory/freebooru/internal/config"
)

// ImportField describes one assignment accepted by a collection import.
type ImportField struct {
	Name     string
	Type     config.TagType
	Values   []string
	Required bool
}

// ImportFields returns a deterministic, side-effect-free import form schema.
func (c *Core) ImportFields(collectionName string) ([]ImportField, error) {
	c.sessionMu.Lock()
	defer c.sessionMu.Unlock()
	if c.catalog == nil {
		return nil, fmt.Errorf("configuration has not been checked")
	}
	if collectionName == "" {
		collectionName = c.config.DefaultCollection
	}
	references, ok := c.catalog.CollectionReferences(collectionName)
	if !ok {
		return nil, fmt.Errorf("collection %q is missing or invalid", collectionName)
	}
	required := make(map[string]struct{}, len(references.Required))
	for _, reference := range references.Required {
		required[referenceKey(reference)] = struct{}{}
	}
	fields := make([]ImportField, 0, len(references.Imported))
	storageValues := make([]string, 0)
	storageRequired := false
	for _, reference := range references.Imported {
		_, isRequired := required[referenceKey(reference)]
		if reference.Storage != "" {
			storageValues = append(storageValues, reference.Storage)
			storageRequired = storageRequired || isRequired
			continue
		}
		tag, _, ok := c.catalog.Tag(reference.Tag)
		if !ok {
			return nil, fmt.Errorf("tag %q is unavailable", reference.Tag)
		}
		values := make([]string, 0, len(tag.Values))
		for _, value := range tag.Values {
			values = append(values, value.Val)
		}
		fields = append(fields, ImportField{
			Name:     tag.Name,
			Type:     tag.Type,
			Values:   values,
			Required: isRequired,
		})
	}
	if len(storageValues) > 0 {
		sort.Strings(storageValues)
		fields = append(fields, ImportField{
			Name:     "storage",
			Type:     config.TagTypeMultivalue,
			Values:   storageValues,
			Required: storageRequired,
		})
	}
	sort.Slice(fields, func(left, right int) bool {
		return normalizeStateName(fields[left].Name) < normalizeStateName(fields[right].Name)
	})
	return fields, nil
}

func referenceKey(reference config.TagReference) string {
	if reference.Storage != "" {
		return "storage:" + normalizeStateName(reference.Storage)
	}
	return "tag:" + normalizeStateName(reference.Tag)
}
