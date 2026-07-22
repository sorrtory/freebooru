package config

import "fmt"

func compileSetPredicate(target TagConfig, raw Relationship) (Predicate, error) {
	if target.Type != TagTypeValue && target.Type != TagTypeMultivalue {
		return Predicate{}, fmt.Errorf("has and not require a value or multivalue target")
	}
	if raw.Has != nil && target.Type != TagTypeMultivalue {
		return Predicate{}, fmt.Errorf("has requires a multivalue target")
	}
	has, err := compilePredefinedList(target, "has", raw.Has)
	if err != nil {
		return Predicate{}, err
	}
	not, err := compilePredefinedList(target, "not", raw.Not)
	if err != nil {
		return Predicate{}, err
	}
	return Predicate{Has: has, Not: not}, nil
}

func compilePredefinedList(target TagConfig, field string, raw []any) ([]string, error) {
	if raw == nil {
		return nil, nil
	}
	if len(raw) == 0 {
		return nil, fmt.Errorf("%s must not be empty", field)
	}
	values := make([]string, 0, len(raw))
	seen := make(map[string]struct{}, len(raw))
	for _, item := range raw {
		text, ok := item.(string)
		if !ok {
			return nil, fmt.Errorf("%s values must be strings", field)
		}
		canonical, err := canonicalPredefinedValue(target, text)
		if err != nil {
			return nil, err
		}
		key := normalizeName(canonical)
		if _, duplicate := seen[key]; duplicate {
			return nil, fmt.Errorf("%s value %q is duplicated", field, text)
		}
		seen[key] = struct{}{}
		values = append(values, canonical)
	}
	return values, nil
}

func canonicalPredefinedValue(tag TagConfig, value string) (string, error) {
	if canonical, ok := CanonicalPredefinedValue(tag, value); ok {
		return canonical, nil
	}
	return "", fmt.Errorf("target tag %q does not declare value %q", tag.Name, value)
}
