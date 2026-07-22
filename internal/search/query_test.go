package search

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseSupportsEveryMVPFormAndPreservesQuotedValue(t *testing.T) {
	query, err := Parse([]string{
		"reviewed",
		"!blocked",
		"rating:safe",
		"title:hello world",
		"score<10",
		"score<=20",
		"score>30",
		"score>=40",
	})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	want := []Term{
		{Tag: "reviewed", Operator: Present},
		{Tag: "blocked", Operator: Absent},
		{Tag: "rating", Operator: Equal, Value: "safe"},
		{Tag: "title", Operator: Equal, Value: "hello world"},
		{Tag: "score", Operator: Less, Value: "10"},
		{Tag: "score", Operator: LessEqual, Value: "20"},
		{Tag: "score", Operator: Greater, Value: "30"},
		{Tag: "score", Operator: GreaterEqual, Value: "40"},
	}
	if !reflect.DeepEqual(query.Terms, want) {
		t.Fatalf("Parse() terms = %#v, want %#v", query.Terms, want)
	}
}

func TestParseRejectsUnsupportedAndMalformedTerms(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  string
	}{
		{name: "empty query", want: "empty"},
		{name: "empty term", input: []string{""}, want: "empty"},
		{name: "or keyword", input: []string{"tag", "OR", "other"}, want: "OR"},
		{name: "or operator", input: []string{"tag||other"}, want: "OR"},
		{name: "grouping", input: []string{"(tag)"}, want: "grouped"},
		{name: "negated value", input: []string{"tag:!value"}, want: "negated"},
		{name: "not equal", input: []string{"tag!=value"}, want: "operator"},
		{name: "bare equal", input: []string{"tag=value"}, want: "operator"},
		{name: "empty absence", input: []string{"!"}, want: "empty"},
		{name: "empty value", input: []string{"tag:"}, want: "empty value"},
		{name: "invalid name", input: []string{"bad-name"}, want: "character"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.input)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Parse(%#v) error = %v, want %q", tt.input, err, tt.want)
			}
		})
	}
}
