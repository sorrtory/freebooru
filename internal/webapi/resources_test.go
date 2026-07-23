package webapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sorrtory/freebooru/internal/config"
	"github.com/sorrtory/freebooru/internal/core"
)

func TestCollectionResourcesUseExplicitScopeAndSafeDTOs(t *testing.T) {
	app := &fakeApplication{
		tags: []core.CollectionTagInfo{{
			Name: "rating", Type: config.TagTypeValue, Comment: "Safety",
			Values: []config.PredefinedValue{{Val: "safe", Comment: "Safe content"}}, Groups: []string{"general"},
			Imported: true, Required: true, AssignmentCount: 3,
		}, {Name: "sha256", Type: config.TagTypeText, Imported: true, System: true}},
		storages: []core.CollectionStorageInfo{{
			Name: "archive", Type: "local", Comment: "Archive",
			Imported: true, FileCount: 2, TotalSizeBytes: 42,
		}},
	}
	handler := mustNew(t, ModeServer, app)

	for _, path := range []string{
		"/api/v1/collections/main/tags",
		"/api/v1/collections/main/storages",
	} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusOK || strings.Contains(response.Body.String(), "path") {
			t.Fatalf("GET %s = %d %q", path, response.Code, response.Body.String())
		}
		if app.collection != "main" {
			t.Fatalf("collection = %q", app.collection)
		}
	}

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/api/v1/collections/main/tags/artist/import", nil))
	if response.Code != http.StatusNoContent || app.collection != "main" || app.resource != "artist" {
		t.Fatalf("import = %d, %q, %q", response.Code, app.collection, app.resource)
	}
}

func TestCollectionTagsSerializeEmptyGroupsAsArray(t *testing.T) {
	app := &fakeApplication{tags: []core.CollectionTagInfo{{Name: "sha256", Type: config.TagTypeText, Imported: true, System: true}}}
	handler := mustNew(t, ModeServer, app)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/collections/main/tags", nil))
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"groups":[]`) {
		t.Fatalf("tags response = %d %s", response.Code, response.Body.String())
	}
}
