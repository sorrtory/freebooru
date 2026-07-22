package collection

import (
	"reflect"
	"strings"
	"testing"
)

func TestBuildSearchSQLKeepsValuesAndNamesInParameters(t *testing.T) {
	maliciousName := "tag') OR 1=1 --"
	maliciousValue := "value') OR 1=1 --"
	query, arguments, err := buildSearchSQL(SearchRequest{
		Terms: []SearchTerm{{
			Tag: maliciousName, Type: "text", Operator: SearchEqual, Value: maliciousValue,
		}},
		Limit: 10,
	})
	if err != nil {
		t.Fatalf("buildSearchSQL() error = %v", err)
	}
	if strings.Contains(query, maliciousName) || strings.Contains(query, maliciousValue) {
		t.Fatalf("query contains untrusted input: %s", query)
	}
	want := []any{maliciousName, "text", maliciousValue, int64(10), int64(0)}
	if !reflect.DeepEqual(arguments, want) {
		t.Fatalf("arguments = %#v, want %#v", arguments, want)
	}
}

func TestBuildSearchSQLRejectsInvalidPaginationAndShapes(t *testing.T) {
	tests := []struct {
		name    string
		request SearchRequest
	}{
		{name: "negative limit", request: SearchRequest{Limit: -1}},
		{name: "negative offset", request: SearchRequest{Offset: -1}},
		{
			name: "wrong operand type",
			request: SearchRequest{Terms: []SearchTerm{{
				Tag: "score", Type: "int", Operator: SearchEqual, Value: "10",
			}}},
		},
		{
			name: "unsupported operation",
			request: SearchRequest{Terms: []SearchTerm{{
				Tag: "title", Type: "text", Operator: "fuzzy", Value: "x",
			}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, _, err := buildSearchSQL(tt.request); err == nil {
				t.Fatalf("buildSearchSQL(%#v) error = nil", tt.request)
			}
		})
	}
}
