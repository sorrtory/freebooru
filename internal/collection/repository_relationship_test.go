package collection

import (
	"strings"
	"testing"
)

func TestDatabaseFileRelationshipsBranchOrderAndRejectCycles(t *testing.T) {
	database := openInitializedTestDatabase(t)
	parent := createRelationshipFile(t, database, "a")
	first := createRelationshipFile(t, database, "b")
	second := createRelationshipFile(t, database, "c")
	grandchild := createRelationshipFile(t, database, "d")
	one := int64(1)
	zero := int64(0)
	for _, relationship := range []FileRelationship{
		{
			ParentSHA256: parent,
			ChildSHA256:  first,
			Kind:         RelationshipVariant,
			Comment:      "cropped version",
			DisplayOrder: &one,
		},
		{
			ParentSHA256: parent,
			ChildSHA256:  second,
			Kind:         RelationshipAlternate,
			DisplayOrder: &zero,
		},
		{
			ParentSHA256: first,
			ChildSHA256:  grandchild,
			Kind:         RelationshipDerived,
		},
	} {
		if err := database.SetFileParent(t.Context(), relationship); err != nil {
			t.Fatalf("SetFileParent(%s) error = %v", relationship.ChildSHA256[:1], err)
		}
	}
	children, err := database.ChildRelationships(t.Context(), parent)
	if err != nil {
		t.Fatalf("ChildRelationships() error = %v", err)
	}
	if len(children) != 2 || children[0].ChildSHA256 != second ||
		children[1].ChildSHA256 != first || children[1].Comment != "cropped version" {
		t.Fatalf("ChildRelationships() = %#v", children)
	}
	if err := database.SetFileParent(t.Context(), FileRelationship{
		ParentSHA256: grandchild,
		ChildSHA256:  parent,
		Kind:         RelationshipVariant,
	}); err == nil || !strings.Contains(err.Error(), "cycle") {
		t.Fatalf("cycle SetFileParent() error = %v", err)
	}
	if err := database.RemoveFileParent(t.Context(), first); err != nil {
		t.Fatalf("RemoveFileParent() error = %v", err)
	}
	children, err = database.ChildRelationships(t.Context(), parent)
	if err != nil || len(children) != 1 || children[0].ChildSHA256 != second {
		t.Fatalf("children after removal = %#v, %v", children, err)
	}
}

func TestDatabaseFileRelationshipValidatesInput(t *testing.T) {
	database := openInitializedTestDatabase(t)
	parent := createRelationshipFile(t, database, "e")
	child := createRelationshipFile(t, database, "f")
	negative := int64(-1)
	for _, relationship := range []FileRelationship{
		{ParentSHA256: parent, ChildSHA256: child, Kind: "sequence"},
		{
			ParentSHA256: parent,
			ChildSHA256:  child,
			Kind:         RelationshipVariant,
			Comment:      "   ",
		},
		{
			ParentSHA256: parent,
			ChildSHA256:  child,
			Kind:         RelationshipVariant,
			DisplayOrder: &negative,
		},
	} {
		if err := database.SetFileParent(t.Context(), relationship); err == nil {
			t.Fatalf("SetFileParent(%#v) error = nil", relationship)
		}
	}
}

func createRelationshipFile(t *testing.T, database *Database, character string) string {
	t.Helper()
	hash := strings.Repeat(character, 64)
	if _, err := database.CreateFile(t.Context(), NewFile{
		SHA256:         hash,
		SizeBytes:      1,
		SourcePath:     "/imports/" + character,
		SourceFilename: character,
		Storages:       []string{"default"},
	}); err != nil {
		t.Fatalf("CreateFile(%s) error = %v", character, err)
	}
	return hash
}
