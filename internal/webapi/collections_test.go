package webapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sorrtory/freebooru/internal/config"
	"github.com/sorrtory/freebooru/internal/core"
)

func TestCollectionsListCreateAndDescribe(t *testing.T) {
	app := &fakeApplication{
		appConfig:   config.AppConfig{DefaultCollection: "main"},
		collections: []config.CollectionConfig{{Name: "work"}, {Name: "main"}},
		collectionInfo: core.CollectionInfo{
			Name: "pictures", TagCount: 7, RequiredCount: 1, Storages: []string{"default"},
		},
	}
	handler := mustNew(t, ModeServer, app)

	list := httptest.NewRecorder()
	handler.ServeHTTP(list, httptest.NewRequest(http.MethodGet, "/api/v1/collections", nil))
	if list.Code != http.StatusOK || list.Body.String() != "{\"default_collection\":\"main\",\"collections\":[{\"name\":\"main\",\"is_default\":true},{\"name\":\"work\",\"is_default\":false}]}\n" {
		t.Fatalf("list = %d %q", list.Code, list.Body.String())
	}

	create := httptest.NewRecorder()
	handler.ServeHTTP(create, httptest.NewRequest(http.MethodPost, "/api/v1/collections", strings.NewReader(`{"name":"pictures"}`)))
	if create.Code != http.StatusCreated || app.collection != "pictures" || create.Header().Get("Location") != "/api/v1/collections/pictures" {
		t.Fatalf("create = %d %q, collection %q", create.Code, create.Body.String(), app.collection)
	}

	detail := httptest.NewRecorder()
	handler.ServeHTTP(detail, httptest.NewRequest(http.MethodGet, "/api/v1/collections/pictures", nil))
	if detail.Code != http.StatusOK || !strings.Contains(detail.Body.String(), `"tag_count":7`) {
		t.Fatalf("detail = %d %q", detail.Code, detail.Body.String())
	}
}
