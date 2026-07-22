package core

import (
	"fmt"
	"strings"

	"github.com/sorrtory/freebooru/internal/collection"
	"github.com/sorrtory/freebooru/internal/config"
)

type preparedTagAssignment struct {
	tag     config.TagConfig
	value   any
	record  collection.TagRecord
	present bool
}

func (c *Core) prepareTagAssignment(
	name string,
	value any,
	availability collectionAvailability,
) (preparedTagAssignment, error) {
	key := normalizeStateName(name)
	if key == "storage" {
		return preparedTagAssignment{}, fmt.Errorf("storage assignments use the storage workflow")
	}
	if _, ok := availability.importedTags[key]; !ok {
		return preparedTagAssignment{}, fmt.Errorf("tag %q is not imported by the collection", name)
	}
	tag, _, ok := c.catalog.Tag(name)
	if !ok {
		return preparedTagAssignment{}, fmt.Errorf("tag %q does not exist", name)
	}
	value = canonicalTagValue(tag, value)
	if err := config.VerifyTagValue(tag, value); err != nil {
		return preparedTagAssignment{}, err
	}
	if tag.Type == config.TagTypeBool && !value.(bool) {
		return preparedTagAssignment{tag: tag, value: false}, nil
	}
	return preparedTagAssignment{
		tag:     tag,
		value:   value,
		record:  tagRecord(tag, value),
		present: true,
	}, nil
}

func canonicalTagValue(tag config.TagConfig, value any) any {
	canonical := func(value string) string {
		for _, declared := range tag.Values {
			if strings.EqualFold(declared.Val, value) {
				return declared.Val
			}
		}
		return value
	}
	switch tag.Type {
	case config.TagTypeValue:
		if text, ok := value.(string); ok {
			return canonical(text)
		}
	case config.TagTypeMultivalue:
		if items, ok := value.([]string); ok {
			result := make([]string, 0, len(items))
			for _, item := range items {
				result = append(result, canonical(item))
			}
			return result
		}
	}
	return value
}

func tagRecord(tag config.TagConfig, value any) collection.TagRecord {
	record := collection.TagRecord{Name: tag.Name, Type: string(tag.Type)}
	switch tag.Type {
	case config.TagTypeBool:
	case config.TagTypeInt:
		number := value.(int64)
		record.IntegerValue = &number
	case config.TagTypeMultivalue:
		record.Values = append([]string(nil), value.([]string)...)
	default:
		text := value.(string)
		record.TextValue = &text
	}
	return record
}
