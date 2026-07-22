package config

import (
	"math"
	"testing"
)

func TestVerifyTagValueAcceptsTypeBoundaries(t *testing.T) {
	tests := []struct {
		name  string
		tag   TagConfig
		value any
	}{
		{name: "bool false", tag: TagConfig{Name: "reviewed", Type: TagTypeBool}, value: false},
		{name: "text empty", tag: TagConfig{Name: "caption", Type: TagTypeText}, value: ""},
		{name: "int zero", tag: TagConfig{Name: "score", Type: TagTypeInt}, value: int64(0)},
		{name: "int max", tag: TagConfig{Name: "score", Type: TagTypeInt}, value: int64(math.MaxInt64)},
		{name: "date", tag: TagConfig{Name: "published", Type: TagTypeDate}, value: "2026-07-22"},
		{
			name:  "datetime",
			tag:   TagConfig{Name: "created_at", Type: TagTypeDatetime},
			value: "2026-07-22T12:30:00+03:00",
		},
		{
			name:  "value case insensitive",
			tag:   valueTag(TagTypeValue),
			value: "SAFE",
		},
		{
			name:  "empty multivalue",
			tag:   valueTag(TagTypeMultivalue),
			value: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := VerifyTagValue(tt.tag, tt.value); err != nil {
				t.Fatalf("VerifyTagValue() error = %v", err)
			}
		})
	}
}

func TestVerifyTagValueRejectsInvalidValues(t *testing.T) {
	tests := []struct {
		name  string
		tag   TagConfig
		value any
	}{
		{name: "wrong bool type", tag: TagConfig{Name: "reviewed", Type: TagTypeBool}, value: "true"},
		{name: "negative int", tag: TagConfig{Name: "score", Type: TagTypeInt}, value: int64(-1)},
		{name: "wrong int type", tag: TagConfig{Name: "score", Type: TagTypeInt}, value: 1},
		{name: "invalid date", tag: TagConfig{Name: "published", Type: TagTypeDate}, value: "2026-02-30"},
		{name: "datetime is date", tag: TagConfig{Name: "created_at", Type: TagTypeDatetime}, value: "2026-07-22"},
		{name: "unknown value", tag: valueTag(TagTypeValue), value: "explicit"},
		{name: "duplicate multivalue", tag: valueTag(TagTypeMultivalue), value: []string{"safe", "SAFE"}},
		{name: "unknown multivalue", tag: valueTag(TagTypeMultivalue), value: []string{"explicit"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := VerifyTagValue(tt.tag, tt.value); err == nil {
				t.Fatal("VerifyTagValue() error = nil")
			}
		})
	}
}

func valueTag(tagType TagType) TagConfig {
	return TagConfig{
		Name:   "rating",
		Type:   tagType,
		Values: []PredefinedValue{{Val: "safe"}},
	}
}
