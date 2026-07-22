package config

import "fmt"

// compilePredicate classifies the YAML predicate before dispatching to the
// compiler for the target tag type. Only documented combinations reach the
// specialized compilers.
func compilePredicate(target TagConfig, raw Relationship) (Predicate, error) {
	has := raw.Has != nil
	is := raw.Is != nil
	not := raw.Not != nil
	integerBounds := raw.Min != nil || raw.Max != nil
	timeBounds := raw.Before != "" || raw.After != ""
	regexPattern := raw.Regex != ""

	if !has && !is && !not && !integerBounds && !timeBounds && !regexPattern {
		return Predicate{Presence: true}, nil
	}
	if is && (has || not || integerBounds || timeBounds || regexPattern) {
		return Predicate{}, fmt.Errorf("is cannot be combined with another predicate")
	}
	if is {
		value, err := compileIs(target, raw.Is)
		return Predicate{Is: value}, err
	}
	if has || not {
		if integerBounds || timeBounds || regexPattern {
			return Predicate{}, fmt.Errorf("has or not cannot be combined with this predicate")
		}
		return compileSetPredicate(target, raw)
	}
	if integerBounds {
		if timeBounds || regexPattern {
			return Predicate{}, fmt.Errorf("integer bounds cannot be combined with another predicate")
		}
		return compileIntegerBounds(target, raw)
	}
	if timeBounds {
		if regexPattern {
			return Predicate{}, fmt.Errorf("time bounds cannot be combined with regex")
		}
		return compileTimeBounds(target, raw)
	}
	return compileRegex(target, raw.Regex)
}
