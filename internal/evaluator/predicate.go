package evaluator

import (
	"time"

	"github.com/sorrtory/freebooru/internal/config"
)

func (e *Evaluator) matchesPredicate(state FileState, edge config.Edge) bool {
	tag, _, ok := e.catalog.Tag(edge.TargetTag)
	if !ok {
		return false
	}
	value, present := state.Value(edge.TargetTag)
	predicate := edge.Predicate
	if predicate.Presence {
		return present
	}
	if predicate.Is != nil {
		return matchesIs(tag, value, present, predicate.Is)
	}
	if predicate.Has != nil || predicate.Not != nil {
		if predicate.Has != nil && !matchesHas(value, present, predicate.Has) {
			return false
		}
		if predicate.Not != nil && !matchesNot(tag, value, present, predicate.Not) {
			return false
		}
		return true
	}
	if predicate.Min != nil || predicate.Max != nil {
		return matchesIntegerBounds(value, present, predicate)
	}
	if predicate.Before != nil || predicate.After != nil {
		return matchesTimeBounds(tag, value, present, predicate)
	}
	if predicate.Regex != nil {
		text, ok := value.(string)
		return present && ok && predicate.Regex.MatchString(text)
	}
	return false
}

func matchesIs(tag config.TagConfig, value any, present bool, want any) bool {
	if tag.Type == config.TagTypeBool {
		expected := want.(bool)
		return present == expected
	}
	if !present {
		return false
	}
	switch tag.Type {
	case config.TagTypeValue:
		actual, ok := value.(string)
		expected, expectedOK := want.(string)
		return ok && expectedOK && normalizeName(actual) == normalizeName(expected)
	case config.TagTypeDate:
		return matchesTimeIs(value, want, "2006-01-02")
	case config.TagTypeDatetime:
		return matchesTimeIs(value, want, time.RFC3339)
	default:
		return value == want
	}
}

func matchesHas(value any, present bool, wanted []string) bool {
	values, ok := value.([]string)
	if !present || !ok {
		return false
	}
	for _, actual := range values {
		for _, want := range wanted {
			if normalizeName(actual) == normalizeName(want) {
				return true
			}
		}
	}
	return false
}

func matchesNot(tag config.TagConfig, value any, present bool, forbidden []string) bool {
	if !present {
		return true
	}
	if tag.Type == config.TagTypeValue {
		actual := value.(string)
		return !containsNormalized(forbidden, actual)
	}
	for _, actual := range value.([]string) {
		if containsNormalized(forbidden, actual) {
			return false
		}
	}
	return true
}

func matchesIntegerBounds(value any, present bool, predicate config.Predicate) bool {
	number, ok := value.(int64)
	if !present || !ok {
		return false
	}
	if predicate.Min != nil && number < *predicate.Min {
		return false
	}
	if predicate.Max != nil && number > *predicate.Max {
		return false
	}
	return true
}

func matchesTimeBounds(
	tag config.TagConfig,
	value any,
	present bool,
	predicate config.Predicate,
) bool {
	text, ok := value.(string)
	if !present || !ok {
		return false
	}
	layout := "2006-01-02"
	if tag.Type == config.TagTypeDatetime {
		layout = time.RFC3339
	}
	parsed, err := time.Parse(layout, text)
	if err != nil {
		return false
	}
	if predicate.Before != nil && !parsed.Before(*predicate.Before) {
		return false
	}
	if predicate.After != nil && !parsed.After(*predicate.After) {
		return false
	}
	return true
}

func matchesTimeIs(value, want any, layout string) bool {
	text, ok := value.(string)
	expected, expectedOK := want.(time.Time)
	if !ok || !expectedOK {
		return false
	}
	actual, err := time.Parse(layout, text)
	return err == nil && actual.Equal(expected)
}

func containsNormalized(values []string, want string) bool {
	want = normalizeName(want)
	for _, value := range values {
		if normalizeName(value) == want {
			return true
		}
	}
	return false
}
