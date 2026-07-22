package config

import (
	"fmt"
	"strings"
)

// VerifyTagConfig checks facts that can be validated within one tag document.
// Reference targets and predicate types require the complete catalog and graph.
func VerifyTagConfig(tag TagConfig) error {
	if err := verifyName("name", tag.Name); err != nil {
		return err
	}
	if strings.EqualFold(tag.Name, "storage") {
		return fmt.Errorf("name %q is reserved", tag.Name)
	}
	if !validTagType(tag.Type) {
		return fmt.Errorf("type %q is not supported", tag.Type)
	}
	if err := verifyUniqueNames("groups", tag.Groups); err != nil {
		return err
	}
	if err := verifyValues(tag); err != nil {
		return err
	}
	return verifyRelationships(tag)
}

func validTagType(tagType TagType) bool {
	switch tagType {
	case TagTypeBool,
		TagTypeText,
		TagTypeInt,
		TagTypeDate,
		TagTypeDatetime,
		TagTypeValue,
		TagTypeMultivalue:
		return true
	default:
		return false
	}
}

func verifyValues(tag TagConfig) error {
	requiresValues := tag.Type == TagTypeValue || tag.Type == TagTypeMultivalue
	if requiresValues && len(tag.Values) == 0 {
		return fmt.Errorf("values are required for type %s", tag.Type)
	}
	if !requiresValues && len(tag.Values) != 0 {
		return fmt.Errorf("values are only allowed for value or multivalue tags")
	}

	names := make([]string, len(tag.Values))
	for index, value := range tag.Values {
		if err := verifyName("val", value.Val); err != nil {
			return fmt.Errorf("values[%d].%w", index, err)
		}
		names[index] = value.Val
		if err := verifyRelationshipLists(value.Suggest, value.Demand, value.Conflict); err != nil {
			return fmt.Errorf("values[%d].%w", index, err)
		}
	}
	return verifyUniqueNames("values", names)
}

func verifyUniqueNames(field string, names []string) error {
	seen := make(map[string]struct{}, len(names))
	for index, name := range names {
		if err := verifyName(field, name); err != nil {
			return fmt.Errorf("%s[%d]: %w", field, index, err)
		}
		key := normalizeName(name)
		if _, ok := seen[key]; ok {
			return fmt.Errorf("%s[%d] duplicates %q", field, index, name)
		}
		seen[key] = struct{}{}
	}
	return nil
}

func verifyRelationships(tag TagConfig) error {
	return verifyRelationshipLists(tag.Suggest, tag.Demand, tag.Conflict)
}

func verifyRelationshipLists(lists ...[]Relationship) error {
	for _, relationships := range lists {
		for _, relationship := range relationships {
			if err := verifyName("relationship tag", relationship.Tag); err != nil {
				return err
			}
			if relationship.Reason != "" && strings.TrimSpace(relationship.Reason) == "" {
				return fmt.Errorf("relationship reason must not be blank")
			}
		}
	}
	return nil
}
