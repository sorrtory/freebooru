package collection

import (
	"fmt"
	"strings"
)

// SearchOperator is one repository-supported query operation.
type SearchOperator string

const (
	// SearchPresent requires a boolean assignment.
	SearchPresent SearchOperator = "present"
	// SearchAbsent requires no assignment.
	SearchAbsent SearchOperator = "absent"
	// SearchEqual compares equality or multivalue membership.
	SearchEqual SearchOperator = "equal"
	// SearchLess compares below the operand.
	SearchLess SearchOperator = "less"
	// SearchLessEqual compares below or equal to the operand.
	SearchLessEqual SearchOperator = "less_equal"
	// SearchGreater compares above the operand.
	SearchGreater SearchOperator = "greater"
	// SearchGreaterEqual compares above or equal to the operand.
	SearchGreaterEqual SearchOperator = "greater_equal"
)

// SearchTerm is one canonical, typed repository condition.
type SearchTerm struct {
	Tag      string
	Type     string
	Operator SearchOperator
	Value    any
}

// SearchRequest combines every term with AND and applies explicit pagination.
type SearchRequest struct {
	Terms  []SearchTerm
	Limit  int64
	Offset int64
}

func buildSearchSQL(request SearchRequest) (string, []any, error) {
	if request.Limit < 0 {
		return "", nil, fmt.Errorf("search limit must be non-negative")
	}
	if request.Offset < 0 {
		return "", nil, fmt.Errorf("search offset must be non-negative")
	}
	clauses := make([]string, 0, len(request.Terms))
	arguments := make([]any, 0, len(request.Terms)*3+2)
	for _, term := range request.Terms {
		clause, values, err := searchClause(term)
		if err != nil {
			return "", nil, err
		}
		clauses = append(clauses, clause)
		arguments = append(arguments, values...)
	}
	query := "SELECT f.sha256 FROM file AS f"
	if len(clauses) > 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}
	query += " ORDER BY f.imported_at DESC, f.sha256 ASC LIMIT ? OFFSET ?"
	arguments = append(arguments, request.Limit, request.Offset)
	return query, arguments, nil
}

func searchClause(term SearchTerm) (string, []any, error) {
	if term.Tag == "" {
		return "", nil, fmt.Errorf("search tag is empty")
	}
	if term.Tag == "storage" {
		return storageSearchClause(term)
	}
	switch term.Operator {
	case SearchPresent:
		if term.Type != "bool" {
			return "", nil, fmt.Errorf("presence search for tag %q requires bool", term.Tag)
		}
		return tagExistsClause("ft.tag_type = 'bool'"), []any{term.Tag}, nil
	case SearchAbsent:
		return "NOT EXISTS (SELECT 1 FROM file_tag AS ft " +
			"WHERE ft.file_id = f.file_id AND ft.tag_name = ?)", []any{term.Tag}, nil
	case SearchEqual:
		return equalSearchClause(term)
	case SearchLess, SearchLessEqual, SearchGreater, SearchGreaterEqual:
		return orderedSearchClause(term)
	default:
		return "", nil, fmt.Errorf("search operator %q is unsupported", term.Operator)
	}
}

func storageSearchClause(term SearchTerm) (string, []any, error) {
	switch term.Operator {
	case SearchAbsent:
		return "NOT EXISTS (SELECT 1 FROM file_storage AS fs " +
			"WHERE fs.file_id = f.file_id)", nil, nil
	case SearchEqual:
		value, ok := term.Value.(string)
		if !ok {
			return "", nil, fmt.Errorf("storage equality expects a string")
		}
		return "EXISTS (SELECT 1 FROM file_storage AS fs " +
			"WHERE fs.file_id = f.file_id AND fs.storage_name = ?)", []any{value}, nil
	default:
		return "", nil, fmt.Errorf("storage does not support search operator %q", term.Operator)
	}
}

func equalSearchClause(term SearchTerm) (string, []any, error) {
	switch term.Type {
	case "bool":
		value, ok := term.Value.(bool)
		if !ok {
			return "", nil, fmt.Errorf("boolean equality for tag %q expects bool", term.Tag)
		}
		if value {
			return tagExistsClause("ft.tag_type = 'bool'"), []any{term.Tag}, nil
		}
		return "NOT EXISTS (SELECT 1 FROM file_tag AS ft " +
			"WHERE ft.file_id = f.file_id AND ft.tag_name = ?)", []any{term.Tag}, nil
	case "text", "value":
		value, ok := term.Value.(string)
		if !ok {
			return "", nil, fmt.Errorf("equality for tag %q expects string", term.Tag)
		}
		return tagExistsClause("ft.tag_type = ? AND ft.text_value = ?"),
			[]any{term.Tag, term.Type, value}, nil
	case "date", "datetime":
		value, ok := term.Value.(string)
		if !ok {
			return "", nil, fmt.Errorf("temporal equality for tag %q expects string", term.Tag)
		}
		return tagExistsClause("ft.tag_type = ? AND julianday(ft.text_value) = julianday(?)"),
			[]any{term.Tag, term.Type, value}, nil
	case "int":
		value, ok := term.Value.(int64)
		if !ok {
			return "", nil, fmt.Errorf("integer equality for tag %q expects int64", term.Tag)
		}
		return tagExistsClause("ft.tag_type = 'int' AND ft.integer_value = ?"),
			[]any{term.Tag, value}, nil
	case "multivalue":
		value, ok := term.Value.(string)
		if !ok {
			return "", nil, fmt.Errorf("multivalue equality for tag %q expects string", term.Tag)
		}
		return "EXISTS (SELECT 1 FROM file_tag AS ft " +
			"JOIN file_tag_value AS fv ON fv.file_tag_id = ft.file_tag_id " +
			"WHERE ft.file_id = f.file_id AND ft.tag_name = ? " +
			"AND ft.tag_type = 'multivalue' AND fv.value = ?)", []any{term.Tag, value}, nil
	default:
		return "", nil, fmt.Errorf("tag %q has unsupported search type %q", term.Tag, term.Type)
	}
}

func orderedSearchClause(term SearchTerm) (string, []any, error) {
	comparison, err := comparisonSQL(term.Operator)
	if err != nil {
		return "", nil, err
	}
	switch term.Type {
	case "int":
		value, ok := term.Value.(int64)
		if !ok {
			return "", nil, fmt.Errorf("integer comparison for tag %q expects int64", term.Tag)
		}
		return tagExistsClause("ft.tag_type = 'int' AND ft.integer_value " + comparison + " ?"),
			[]any{term.Tag, value}, nil
	case "date", "datetime":
		value, ok := term.Value.(string)
		if !ok {
			return "", nil, fmt.Errorf("temporal comparison for tag %q expects string", term.Tag)
		}
		condition := "ft.tag_type = ? AND julianday(ft.text_value) " +
			comparison + " julianday(?)"
		return tagExistsClause(condition), []any{term.Tag, term.Type, value}, nil
	default:
		return "", nil, fmt.Errorf("tag %q does not support ordered search", term.Tag)
	}
}

func comparisonSQL(operator SearchOperator) (string, error) {
	switch operator {
	case SearchLess:
		return "<", nil
	case SearchLessEqual:
		return "<=", nil
	case SearchGreater:
		return ">", nil
	case SearchGreaterEqual:
		return ">=", nil
	default:
		return "", fmt.Errorf("search operator %q is not ordered", operator)
	}
}

func tagExistsClause(condition string) string {
	return "EXISTS (SELECT 1 FROM file_tag AS ft WHERE ft.file_id = f.file_id " +
		"AND ft.tag_name = ? AND " + condition + ")"
}
