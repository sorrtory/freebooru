package webapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sorrtory/freebooru/internal/core"
)

func TestSettingsNeverExposePathsAndRequireRevision(t *testing.T) {
	app := &fakeApplication{settings: core.Settings{
		Revision: "revision", Language: "en", DefaultCollection: "main",
		DefaultStorageName: "default", HTTPAddress: "0.0.0.0", HTTPPort: 52800,
	}}
	handler := mustNew(t, ModeServer, app)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/settings", nil))
	if response.Code != http.StatusOK || strings.Contains(response.Body.String(), "path") {
		t.Fatalf("GET settings = %d %q", response.Code, response.Body.String())
	}

	body := `{"revision":"revision","language":"en","default_collection":"main","default_storage_name":"default","http_address":"127.0.0.1","http_port":52801,"remove_on_upload":false}`
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodPut, "/api/v1/settings", strings.NewReader(body)))
	if response.Code != http.StatusOK || app.resource != "revision" || !strings.Contains(response.Body.String(), `"restart_required":true`) {
		t.Fatalf("PUT settings = %d %q", response.Code, response.Body.String())
	}
}
