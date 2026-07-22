package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVerifyTagConfigAcceptsSupportedTypes(t *testing.T) {
	tests := []TagConfig{
		{Name: "reviewed", Type: TagTypeBool},
		{Name: "caption", Type: TagTypeText},
		{Name: "score", Type: TagTypeInt},
		{Name: "published", Type: TagTypeDate},
		{Name: "created_at", Type: TagTypeDatetime},
		{Name: "rating", Type: TagTypeValue, Values: []PredefinedValue{{Val: "safe"}}},
		{Name: "character", Type: TagTypeMultivalue, Values: []PredefinedValue{{Val: "cirno"}}},
	}
	for _, tag := range tests {
		t.Run(string(tag.Type), func(t *testing.T) {
			if err := VerifyTagConfig(tag); err != nil {
				t.Fatalf("VerifyTagConfig() error = %v", err)
			}
		})
	}
}

func TestVerifyTagConfigRejectsInvalidLocalRules(t *testing.T) {
	tests := []struct {
		name string
		tag  TagConfig
		want string
	}{
		{name: "reserved", tag: TagConfig{Name: "Storage", Type: TagTypeBool}, want: "reserved"},
		{name: "system reserved", tag: TagConfig{Name: "filetype", Type: TagTypeText}, want: "reserved"},
		{name: "blank comment", tag: TagConfig{Name: "artist", Type: TagTypeText, Comment: "  "}, want: "comment"},
		{name: "type", tag: TagConfig{Name: "artist", Type: "number"}, want: "not supported"},
		{name: "missing values", tag: TagConfig{Name: "rating", Type: TagTypeValue}, want: "values are required"},
		{
			name: "forbidden values",
			tag: TagConfig{
				Name:   "artist",
				Type:   TagTypeText,
				Values: []PredefinedValue{{Val: "reimu"}},
			},
			want: "only allowed",
		},
		{
			name: "duplicate groups",
			tag:  TagConfig{Name: "artist", Type: TagTypeText, Groups: []string{"Meta", "meta"}},
			want: "duplicates",
		},
		{
			name: "duplicate values",
			tag: TagConfig{
				Name: "rating",
				Type: TagTypeValue,
				Values: []PredefinedValue{
					{Val: "Safe"},
					{Val: "safe"},
				},
			},
			want: "duplicates",
		},
		{
			name: "alias collides with canonical value",
			tag: TagConfig{
				Name: "character",
				Type: TagTypeMultivalue,
				Values: []PredefinedValue{
					{Val: "cirno", Aliases: []string{"chiruno"}},
					{Val: "Chiruno"},
				},
			},
			want: "duplicates",
		},
		{
			name: "aliases collide across values",
			tag: TagConfig{
				Name: "character",
				Type: TagTypeMultivalue,
				Values: []PredefinedValue{
					{Val: "cirno", Aliases: []string{"ice_fairy"}},
					{Val: "reimu", Aliases: []string{"ICE_FAIRY"}},
				},
			},
			want: "duplicates",
		},
		{
			name: "invalid alias",
			tag: TagConfig{
				Name:   "character",
				Type:   TagTypeValue,
				Values: []PredefinedValue{{Val: "cirno", Aliases: []string{"ice fairy"}}},
			},
			want: "invalid character",
		},
		{
			name: "blank value comment",
			tag: TagConfig{
				Name:   "character",
				Type:   TagTypeValue,
				Values: []PredefinedValue{{Val: "cirno", Comment: "\t"}},
			},
			want: "comment",
		},
		{
			name: "missing relationship target",
			tag: TagConfig{
				Name:    "artist",
				Type:    TagTypeText,
				Suggest: []Relationship{{}},
			},
			want: "relationship tag is required",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := VerifyTagConfig(tt.tag)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("VerifyTagConfig() error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestTagYAMLDecodesRelationshipFalsePredicate(t *testing.T) {
	path := filepath.Join(t.TempDir(), "reviewed.yaml")
	content := "name: reviewed\ntype: bool\nconflict:\n  - tag: approved\n    is: false\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	documents, diagnostics := loadYAMLFile(path, "tag", true, VerifyTagConfig)
	if diagnostics.HasErrors() || len(documents) != 1 {
		t.Fatalf("loadYAMLFile() = %#v, %#v", documents, diagnostics)
	}
	got, ok := documents[0].Value.Conflict[0].Is.(bool)
	if !ok || got {
		t.Fatalf("is predicate = %#v, want bool(false)", documents[0].Value.Conflict[0].Is)
	}
}
