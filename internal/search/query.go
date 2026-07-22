// Package search parses and resolves frontend-independent collection queries.
package search

import (
	"fmt"
	"strings"
	"unicode"
)

// Operator is one supported MVP search operation.
type Operator string

const (
	// Present requires an assignment.
	Present Operator = "present"
	// Absent requires no assignment.
	Absent Operator = "absent"
	// Equal compares equality or multivalue membership.
	Equal Operator = "equal"
	// Less compares typed values strictly below the operand.
	Less Operator = "less"
	// LessEqual compares typed values below or equal to the operand.
	LessEqual Operator = "less_equal"
	// Greater compares typed values strictly above the operand.
	Greater Operator = "greater"
	// GreaterEqual compares typed values above or equal to the operand.
	GreaterEqual Operator = "greater_equal"
)

// Term is one raw tag condition. Value remains text until catalog resolution.
type Term struct {
	Tag      string
	Operator Operator
	Value    string
}

// Query combines every term with AND.
type Query struct {
	Terms []Term
}

// Parse converts shell-preserved query terms into the strict MVP query model.
func Parse(input []string) (Query, error) {
	if len(input) == 0 {
		return Query{}, fmt.Errorf("search query is empty")
	}
	query := Query{Terms: make([]Term, 0, len(input))}
	for _, raw := range input {
		term, err := parseTerm(raw)
		if err != nil {
			return Query{}, err
		}
		query.Terms = append(query.Terms, term)
	}
	return query, nil
}

func parseTerm(raw string) (Term, error) {
	if raw == "" {
		return Term{}, fmt.Errorf("search term is empty")
	}
	if strings.EqualFold(raw, "or") || strings.Contains(raw, "||") {
		return Term{}, fmt.Errorf("OR search is not supported")
	}
	if strings.ContainsAny(raw, "()") {
		return Term{}, fmt.Errorf("grouped search is not supported")
	}
	if strings.HasPrefix(raw, "!") {
		name := raw[1:]
		if err := verifyTagName(name); err != nil {
			return Term{}, fmt.Errorf("invalid absent search term %q: %w", raw, err)
		}
		return Term{Tag: name, Operator: Absent}, nil
	}
	operatorIndex := strings.IndexAny(raw, ":<>")
	if operatorIndex < 0 {
		if strings.Contains(raw, "!") || strings.Contains(raw, "=") {
			return Term{}, fmt.Errorf("unsupported search operator in %q", raw)
		}
		if err := verifyTagName(raw); err != nil {
			return Term{}, fmt.Errorf("invalid presence search term %q: %w", raw, err)
		}
		return Term{Tag: raw, Operator: Present}, nil
	}
	name := raw[:operatorIndex]
	if err := verifyTagName(name); err != nil {
		return Term{}, fmt.Errorf("invalid search term %q: %w", raw, err)
	}
	operator, width, err := parseOperator(raw[operatorIndex:])
	if err != nil {
		return Term{}, fmt.Errorf("invalid search term %q: %w", raw, err)
	}
	value := raw[operatorIndex+width:]
	if value == "" {
		return Term{}, fmt.Errorf("search term %q has an empty value", raw)
	}
	if strings.HasPrefix(value, "!") {
		return Term{}, fmt.Errorf("negated search values are not supported")
	}
	return Term{Tag: name, Operator: operator, Value: value}, nil
}

func parseOperator(raw string) (Operator, int, error) {
	switch {
	case strings.HasPrefix(raw, ":"):
		return Equal, 1, nil
	case strings.HasPrefix(raw, "<="):
		return LessEqual, 2, nil
	case strings.HasPrefix(raw, ">="):
		return GreaterEqual, 2, nil
	case strings.HasPrefix(raw, "<"):
		return Less, 1, nil
	case strings.HasPrefix(raw, ">"):
		return Greater, 1, nil
	default:
		return "", 0, fmt.Errorf("operator is not supported")
	}
}

func verifyTagName(name string) error {
	if name == "" {
		return fmt.Errorf("tag name is empty")
	}
	for _, character := range name {
		if character != '_' && !unicode.IsLetter(character) && !unicode.IsDigit(character) {
			return fmt.Errorf("tag name %q contains an unsupported character", name)
		}
	}
	return nil
}
