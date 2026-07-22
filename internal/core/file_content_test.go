package core

import (
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestOpenFileContentReturnsSeekableStoredCopyWithoutPath(t *testing.T) {
	app := newRealImportTestCore(t)
	source := filepath.Join(t.TempDir(), "picture.txt")
	if err := os.WriteFile(source, []byte("content"), 0o600); err != nil {
		t.Fatal(err)
	}
	result, err := app.Import(t.Context(), ImportRequest{SourcePath: source, SourceFilename: "visible.txt", Tags: map[string]any{"rating": "safe"}})
	if err != nil {
		t.Fatal(err)
	}
	content, err := app.OpenFileContent(t.Context(), "main", result.SHA256)
	if err != nil {
		t.Fatalf("OpenFileContent() error = %v", err)
	}
	defer content.Reader.Close()
	data, err := io.ReadAll(content.Reader)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "content" || content.Filename != "visible.txt" || content.SizeBytes != 7 {
		t.Fatalf("content = %#v, data = %q", content, data)
	}
}
