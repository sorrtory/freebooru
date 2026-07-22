package search

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/sorrtory/freebooru/internal/config"
)

// ResolvedTerm contains the canonical tag and typed comparison operand.
type ResolvedTerm struct {
	Tag      string
	Type     config.TagType
	Operator Operator
	Value    any
}

// ResolvedQuery is safe to compile because every term is catalog-validated.
type ResolvedQuery struct {
	Terms []ResolvedTerm
}

// Resolve validates query terms against one collection's effective imports.
func Resolve(
	query Query,
	catalog *config.Catalog,
	references config.ResolvedReferences,
) (ResolvedQuery, error) {
	if catalog == nil {
		return ResolvedQuery{}, fmt.Errorf("catalog is required")
	}
	availability := searchAvailability(references)
	resolved := ResolvedQuery{Terms: make([]ResolvedTerm, 0, len(query.Terms))}
	for _, term := range query.Terms {
		item, err := resolveTerm(term, catalog, availability)
		if err != nil {
			return ResolvedQuery{}, err
		}
		resolved.Terms = append(resolved.Terms, item)
	}
	return resolved, nil
}

type availability struct {
	tags     map[string]struct{}
	storages map[string]struct{}
}

func searchAvailability(references config.ResolvedReferences) availability {
	available := availability{
		tags:     make(map[string]struct{}),
		storages: make(map[string]struct{}),
	}
	for _, reference := range references.Imported {
		if reference.Tag != "" {
			available.tags[normalize(reference.Tag)] = struct{}{}
		}
		if reference.Storage != "" {
			available.storages[normalize(reference.Storage)] = struct{}{}
		}
	}
	if len(available.storages) > 0 {
		available.tags["storage"] = struct{}{}
	}
	return available
}

func resolveTerm(
	term Term,
	catalog *config.Catalog,
	available availability,
) (ResolvedTerm, error) {
	if _, ok := available.tags[normalize(term.Tag)]; !ok {
		return ResolvedTerm{}, fmt.Errorf("tag %q is not imported by the collection", term.Tag)
	}
	tag, _, ok := catalog.Tag(term.Tag)
	if !ok {
		return ResolvedTerm{}, fmt.Errorf("tag %q does not exist", term.Tag)
	}
	resolved := ResolvedTerm{Tag: tag.Name, Type: tag.Type, Operator: term.Operator}
	switch term.Operator {
	case Present:
		if tag.Type != config.TagTypeBool {
			return ResolvedTerm{}, fmt.Errorf("presence term %q requires a boolean tag", term.Tag)
		}
		return resolved, nil
	case Absent:
		return resolved, nil
	case Equal:
		value, err := resolveEqualValue(tag, term.Value, available)
		if err != nil {
			return ResolvedTerm{}, err
		}
		resolved.Value = value
		return resolved, nil
	case Less, LessEqual, Greater, GreaterEqual:
		value, err := resolveOrderedValue(tag, term.Value)
		if err != nil {
			return ResolvedTerm{}, err
		}
		resolved.Value = value
		return resolved, nil
	default:
		return ResolvedTerm{}, fmt.Errorf("search operator %q is not supported", term.Operator)
	}
}

func resolveEqualValue(
	tag config.TagConfig,
	raw string,
	available availability,
) (any, error) {
	var value any = raw
	switch tag.Type {
	case config.TagTypeBool:
		parsed, err := strconv.ParseBool(strings.ToLower(raw))
		if err != nil {
			return nil, fmt.Errorf("tag %q expects true or false: %w", tag.Name, err)
		}
		value = parsed
	case config.TagTypeInt:
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("tag %q expects an integer: %w", tag.Name, err)
		}
		value = parsed
	case config.TagTypeValue:
		value = canonicalDeclaredValue(tag, raw)
	case config.TagTypeMultivalue:
		value = canonicalDeclaredValue(tag, raw)
		if normalize(tag.Name) == "storage" {
			if _, ok := available.storages[normalize(value.(string))]; !ok {
				return nil, fmt.Errorf("storage %q is not imported by the collection", raw)
			}
		}
		if err := config.VerifyTagValue(tag, []string{value.(string)}); err != nil {
			return nil, err
		}
		return value, nil
	}
	if err := config.VerifyTagValue(tag, value); err != nil {
		return nil, err
	}
	return value, nil
}

func resolveOrderedValue(tag config.TagConfig, raw string) (any, error) {
	var value any
	switch tag.Type {
	case config.TagTypeInt:
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("tag %q expects an integer: %w", tag.Name, err)
		}
		value = parsed
	case config.TagTypeDate, config.TagTypeDatetime:
		value = raw
	default:
		return nil, fmt.Errorf("tag %q does not support ordered comparison", tag.Name)
	}
	if err := config.VerifyTagValue(tag, value); err != nil {
		return nil, err
	}
	return value, nil
}

func canonicalDeclaredValue(tag config.TagConfig, raw string) string {
	for _, declared := range tag.Values {
		if strings.EqualFold(declared.Val, raw) {
			return declared.Val
		}
	}
	return raw
}

func normalize(value string) string {
	return strings.ToLower(value)
}
