package core

import (
	"reflect"
	"testing"

	"github.com/sorrtory/freebooru/internal/config"
)

func TestImportFieldsReturnsRequiredAndOptionalSchema(t *testing.T) {
	app := newImportTestCore(t)
	fields, err := app.ImportFields("")
	if err != nil {
		t.Fatalf("ImportFields() error = %v", err)
	}
	byName := make(map[string]ImportField, len(fields))
	for _, field := range fields {
		byName[field.Name] = field
	}
	if field := byName["rating"]; !field.Required || field.Type != config.TagTypeValue ||
		!reflect.DeepEqual(field.Values, []string{"safe", "questionable"}) {
		t.Fatalf("rating field = %#v", field)
	}
	if field := byName["score"]; field.Required || field.Type != config.TagTypeInt {
		t.Fatalf("score field = %#v", field)
	}
	if field := byName["storage"]; !field.Required ||
		!reflect.DeepEqual(field.Values, []string{"default"}) {
		t.Fatalf("storage field = %#v", field)
	}
}
