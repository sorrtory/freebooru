package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreateAndDescribeCollectionPublishesUsableExplicitCollection(t *testing.T) {
	app := newRealImportTestCore(t)
	created, err := app.CreateCollection(t.Context(), "pictures")
	if err != nil {
		t.Fatalf("CreateCollection() error = %v", err)
	}
	if created.Name != "pictures" || created.Default || created.TagCount == 0 || len(created.Storages) != 1 {
		t.Fatalf("created = %#v", created)
	}
	if _, err := os.Stat(filepath.Join(app.paths.Collections, "pictures.yaml")); err != nil {
		t.Fatalf("stat collection config: %v", err)
	}
	info, err := app.DescribeCollection(t.Context(), "pictures")
	if err != nil {
		t.Fatalf("DescribeCollection() error = %v", err)
	}
	if info.Name != "pictures" || info.FileCount != 0 || info.TotalSizeBytes != 0 {
		t.Fatalf("info = %#v", info)
	}
	if _, err := app.CreateCollection(t.Context(), "pictures"); err == nil {
		t.Fatal("duplicate CreateCollection() error = nil")
	}
}
