package webapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sorrtory/freebooru/internal/collection"
)

const testFileHash = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func TestFilesBrowseAndDetailUseSafeCompleteDTO(t *testing.T) {
	record := collection.FileRecord{
		SHA256: testFileHash, SizeBytes: 42, MIMEType: "image/png", ImportedAt: "2026-07-23T10:00:00Z",
		UpdatedAt: "2026-07-23T10:00:00Z", LastInteractionAt: "2026-07-23T10:00:00Z",
		Sources:  []collection.SourceRecord{{Path: "/secret/path.png", Filename: "visible.png", ObservedAt: "now"}},
		Tags:     []collection.TagRecord{{Name: "rating", Type: "value", TextValue: stringPointer("safe")}},
		Storages: []string{"default"},
	}
	app := &fakeApplication{files: []collection.FileRecord{record}, file: record}
	handler := mustNew(t, ModeServer, app)

	list := httptest.NewRecorder()
	handler.ServeHTTP(list, httptest.NewRequest(http.MethodGet, "/api/v1/collections/main/files?limit=1", nil))
	if list.Code != http.StatusOK || strings.Contains(list.Body.String(), "/secret/") || !strings.Contains(list.Body.String(), `"filename":"visible.png"`) {
		t.Fatalf("list = %d %q", list.Code, list.Body.String())
	}
	if app.collection != "main" {
		t.Fatalf("collection = %q", app.collection)
	}

	detail := httptest.NewRecorder()
	handler.ServeHTTP(detail, httptest.NewRequest(http.MethodGet, "/api/v1/collections/main/files/"+testFileHash, nil))
	if detail.Code != http.StatusOK || detail.Header().Get("ETag") == "" || strings.Contains(detail.Body.String(), "/secret/") || !strings.Contains(detail.Body.String(), `"value":"safe"`) {
		t.Fatalf("detail = %d %q", detail.Code, detail.Body.String())
	}
}

func stringPointer(value string) *string { return &value }
