package core

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/sorrtory/freebooru/internal/config"
)

// ParseTagAssignments resolves frontend assignment strings against one
// collection. The first colon separates a tag name from its value.
func (c *Core) ParseTagAssignments(collectionName string, assignments []string) (map[string]any, error) {
	_, _, availability, err := c.tagMutationCollection(collectionName)
	if err != nil {
		return nil, err
	}
	values := make(map[string]any, len(assignments))
	seen := make(map[string]struct{}, len(assignments))
	for _, assignment := range assignments {
		name, raw, hasValue := strings.Cut(assignment, ":")
		if name == "" {
			return nil, fmt.Errorf("tag assignment %q has an empty name", assignment)
		}
		key := normalizeStateName(name)
		if key == "storage" {
			if !hasValue || raw == "" {
				return nil, fmt.Errorf("storage assignment %q requires a value", assignment)
			}
			values["storage"] = appendStringValue(values["storage"], raw)
			continue
		}
		if _, ok := availability.importedTags[key]; !ok {
			return nil, fmt.Errorf("tag %q is not imported by the collection", name)
		}
		tag, _, ok := c.catalog.Tag(name)
		if !ok {
			return nil, fmt.Errorf("tag %q does not exist", name)
		}
		value, repeatable, err := parseAssignmentValue(tag, raw, hasValue)
		if err != nil {
			return nil, err
		}
		if _, duplicate := seen[key]; duplicate && !repeatable {
			return nil, fmt.Errorf("tag %q is assigned more than once", tag.Name)
		}
		seen[key] = struct{}{}
		if repeatable {
			values[tag.Name] = appendStringValue(values[tag.Name], value.(string))
		} else {
			values[tag.Name] = value
		}
	}
	return values, nil
}

func parseAssignmentValue(tag config.TagConfig, raw string, hasValue bool) (any, bool, error) {
	if tag.Type == config.TagTypeBool {
		if hasValue {
			return nil, false, fmt.Errorf("boolean tag %q does not accept a value", tag.Name)
		}
		return true, false, nil
	}
	if !hasValue {
		return nil, false, fmt.Errorf("tag %q requires a value", tag.Name)
	}
	if tag.Type == config.TagTypeInt {
		number, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return nil, false, fmt.Errorf("parse tag %q integer value: %w", tag.Name, err)
		}
		return number, false, nil
	}
	return raw, tag.Type == config.TagTypeMultivalue, nil
}

func appendStringValue(current any, value string) []string {
	if current == nil {
		return []string{value}
	}
	return append(current.([]string), value)
}
