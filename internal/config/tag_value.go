package config

import (
	"fmt"
	"time"
)

const dateLayout = "2006-01-02"

// VerifyTagValue checks a typed value against one tag definition.
// Dates and datetimes use strings at the config boundary and are parsed here
// strictly; persisted representations belong to the collection package.
func VerifyTagValue(tag TagConfig, value any) error {
	switch tag.Type {
	case TagTypeBool:
		return requireValueType[bool](tag, value)
	case TagTypeText:
		return requireValueType[string](tag, value)
	case TagTypeInt:
		return verifyIntValue(tag, value)
	case TagTypeDate:
		return verifyTimeValue(tag, value, dateLayout)
	case TagTypeDatetime:
		return verifyTimeValue(tag, value, time.RFC3339)
	case TagTypeValue:
		return verifyPredefinedValue(tag, value)
	case TagTypeMultivalue:
		return verifyMultivalue(tag, value)
	default:
		return fmt.Errorf("tag %q has unsupported type %q", tag.Name, tag.Type)
	}
}

func requireValueType[T any](tag TagConfig, value any) error {
	if _, ok := value.(T); !ok {
		return fmt.Errorf("tag %q expects a %s value", tag.Name, tag.Type)
	}
	return nil
}

func verifyIntValue(tag TagConfig, value any) error {
	number, ok := value.(int64)
	if !ok {
		return fmt.Errorf("tag %q expects an int64 value", tag.Name)
	}
	if number < 0 {
		return fmt.Errorf("tag %q value must be at least 0", tag.Name)
	}
	return nil
}

func verifyTimeValue(tag TagConfig, value any, layout string) error {
	text, ok := value.(string)
	if !ok {
		return fmt.Errorf("tag %q expects a %s string", tag.Name, tag.Type)
	}
	if _, err := time.Parse(layout, text); err != nil {
		return fmt.Errorf("tag %q value must match %s: %w", tag.Name, layout, err)
	}
	return nil
}

func verifyPredefinedValue(tag TagConfig, value any) error {
	text, ok := value.(string)
	if !ok {
		return fmt.Errorf("tag %q expects a predefined string value", tag.Name)
	}
	if !hasPredefinedValue(tag, text) {
		return fmt.Errorf("tag %q does not declare value %q", tag.Name, text)
	}
	return nil
}

func verifyMultivalue(tag TagConfig, value any) error {
	values, ok := value.([]string)
	if !ok {
		return fmt.Errorf("tag %q expects a string slice", tag.Name)
	}
	seen := make(map[string]struct{}, len(values))
	for _, item := range values {
		key := normalizeName(item)
		if _, duplicate := seen[key]; duplicate {
			return fmt.Errorf("tag %q value %q is duplicated", tag.Name, item)
		}
		if !hasPredefinedValue(tag, item) {
			return fmt.Errorf("tag %q does not declare value %q", tag.Name, item)
		}
		seen[key] = struct{}{}
	}
	return nil
}

func hasPredefinedValue(tag TagConfig, value string) bool {
	_, ok := CanonicalPredefinedValue(tag, value)
	return ok
}

// CanonicalPredefinedValue resolves a canonical value or alias to its val.
func CanonicalPredefinedValue(tag TagConfig, value string) (string, bool) {
	want := normalizeName(value)
	for _, declared := range tag.Values {
		if normalizeName(declared.Val) == want {
			return declared.Val, true
		}
		for _, alias := range declared.Aliases {
			if normalizeName(alias) == want {
				return declared.Val, true
			}
		}
	}
	return "", false
}
