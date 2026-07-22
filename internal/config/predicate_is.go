package config

import (
	"fmt"
	"math"
	"time"
)

func compileIs(target TagConfig, raw any) (any, error) {
	switch target.Type {
	case TagTypeBool:
		value, ok := raw.(bool)
		if !ok {
			return nil, fmt.Errorf("is for bool must be true or false")
		}
		return value, nil
	case TagTypeText:
		return requireStringPredicate(target, raw)
	case TagTypeInt:
		value, ok := predicateInt64(raw)
		if !ok || value < 0 {
			return nil, fmt.Errorf("is for int must be between 0 and %d", int64(math.MaxInt64))
		}
		return value, nil
	case TagTypeDate:
		return parsePredicateTime(raw, dateLayout)
	case TagTypeDatetime:
		return parsePredicateTime(raw, time.RFC3339)
	case TagTypeValue:
		value, err := requireStringPredicate(target, raw)
		if err != nil {
			return nil, err
		}
		return canonicalPredefinedValue(target, value)
	case TagTypeMultivalue:
		return nil, fmt.Errorf("is is not supported for multivalue; use has or not")
	default:
		return nil, fmt.Errorf("unsupported target type %q", target.Type)
	}
}

func requireStringPredicate(target TagConfig, raw any) (string, error) {
	value, ok := raw.(string)
	if !ok {
		return "", fmt.Errorf("is for %s must be a string", target.Type)
	}
	return value, nil
}

func parsePredicateTime(raw any, layout string) (time.Time, error) {
	value, ok := raw.(string)
	if !ok {
		return time.Time{}, fmt.Errorf("is must be a string matching %s", layout)
	}
	parsed, err := time.Parse(layout, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("is must match %s: %w", layout, err)
	}
	return parsed, nil
}

func predicateInt64(value any) (int64, bool) {
	switch number := value.(type) {
	case int:
		return int64(number), true
	case int64:
		return number, true
	case uint64:
		if number <= math.MaxInt64 {
			return int64(number), true
		}
	}
	return 0, false
}
