package config

import (
	"fmt"
	"regexp"
	"time"
)

func compileIntegerBounds(target TagConfig, raw Relationship) (Predicate, error) {
	if target.Type != TagTypeInt {
		return Predicate{}, fmt.Errorf("min and max require an int target")
	}
	if raw.Min != nil && *raw.Min < 0 || raw.Max != nil && *raw.Max < 0 {
		return Predicate{}, fmt.Errorf("min and max must be at least 0")
	}
	if raw.Min != nil && raw.Max != nil && *raw.Min > *raw.Max {
		return Predicate{}, fmt.Errorf("min must not exceed max")
	}
	return Predicate{Min: raw.Min, Max: raw.Max}, nil
}

func compileTimeBounds(target TagConfig, raw Relationship) (Predicate, error) {
	layout := dateLayout
	if target.Type == TagTypeDatetime {
		layout = time.RFC3339
	} else if target.Type != TagTypeDate {
		return Predicate{}, fmt.Errorf("before and after require a date or datetime target")
	}
	predicate := Predicate{}
	var err error
	if raw.Before != "" {
		predicate.Before, err = parseTimePointer(raw.Before, layout)
		if err != nil {
			return Predicate{}, fmt.Errorf("before: %w", err)
		}
	}
	if raw.After != "" {
		predicate.After, err = parseTimePointer(raw.After, layout)
		if err != nil {
			return Predicate{}, fmt.Errorf("after: %w", err)
		}
	}
	if predicate.Before != nil && predicate.After != nil && !predicate.After.Before(*predicate.Before) {
		return Predicate{}, fmt.Errorf("after must be earlier than before")
	}
	return predicate, nil
}

func compileRegex(target TagConfig, pattern string) (Predicate, error) {
	if target.Type != TagTypeText {
		return Predicate{}, fmt.Errorf("regex requires a text target")
	}
	compiled, err := regexp.Compile(pattern)
	if err != nil {
		return Predicate{}, fmt.Errorf("compile regex: %w", err)
	}
	return Predicate{Regex: compiled}, nil
}

func parseTimePointer(value, layout string) (*time.Time, error) {
	parsed, err := time.Parse(layout, value)
	if err != nil {
		return nil, fmt.Errorf("must match %s: %w", layout, err)
	}
	return &parsed, nil
}
