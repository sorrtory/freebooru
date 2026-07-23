package webui

import (
	"context"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/sorrtory/freebooru/internal/collection"
	"github.com/sorrtory/freebooru/internal/config"
	"github.com/sorrtory/freebooru/internal/core"
	"github.com/sorrtory/freebooru/internal/webapi"
)

type staticApplication struct{}

func (staticApplication) LoadConfig(context.Context) error {
	return nil
}

func (staticApplication) CheckConfig(context.Context) config.Diagnostics {
	return nil
}

func (staticApplication) AppConfig() config.AppConfig {
	return config.DefaultAppConfig()
}

func (staticApplication) ImportFields(string) ([]core.ImportField, error) {
	return nil, nil
}

func (staticApplication) EvaluateImportDraft(
	context.Context,
	core.ImportDraftRequest,
) (core.ImportDraft, error) {
	return core.ImportDraft{}, nil
}

func (staticApplication) Import(context.Context, core.ImportRequest) (core.ImportResult, error) {
	return core.ImportResult{}, nil
}

func (staticApplication) ListCollections() ([]config.CollectionConfig, error) { return nil, nil }
func (staticApplication) DescribeCollection(context.Context, string) (core.CollectionInfo, error) {
	return core.CollectionInfo{}, nil
}

func (staticApplication) CreateCollection(context.Context, string) (core.CollectionInfo, error) {
	return core.CollectionInfo{}, nil
}
func (staticApplication) Search(context.Context, core.FileSearchRequest) ([]collection.FileRecord, error) {
	return nil, nil
}
func (staticApplication) GetFile(context.Context, string, string) (collection.FileRecord, error) {
	return collection.FileRecord{}, nil
}
func (staticApplication) OpenFileContent(context.Context, string, string) (core.FileContent, error) {
	return core.FileContent{}, nil
}
func (staticApplication) SetTag(context.Context, core.TagMutationRequest) (core.TagMutationResult, error) {
	return core.TagMutationResult{}, nil
}
func (staticApplication) RemoveTag(context.Context, core.TagRemovalRequest) (core.TagMutationResult, error) {
	return core.TagMutationResult{}, nil
}
func (staticApplication) ListCollectionTagInfo(context.Context, string) ([]core.CollectionTagInfo, error) {
	return nil, nil
}
func (staticApplication) ListCollectionStorageInfo(context.Context, string) ([]core.CollectionStorageInfo, error) {
	return nil, nil
}
func (staticApplication) ImportCollectionTag(context.Context, string, string) error     { return nil }
func (staticApplication) ImportCollectionStorage(context.Context, string, string) error { return nil }
func (staticApplication) Settings() core.Settings                                       { return core.Settings{} }
func (staticApplication) UpdateSettings(context.Context, core.SettingsUpdate) (core.Settings, error) {
	return core.Settings{}, nil
}

func TestHandlerServesAssetsAndSPAFallback(t *testing.T) {
	assets := fstest.MapFS{
		"index.html":    &fstest.MapFile{Data: []byte("<main>FreeBooru</main>")},
		"assets/app.js": &fstest.MapFile{Data: []byte("console.log('hello')")},
	}
	handler := NewHandler(http.NotFoundHandler(), assets)

	for _, requestPath := range []string{"/", "/collection/main"} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, requestPath, nil))
		if response.Code != http.StatusOK {
			t.Fatalf("GET %s status = %d, want %d", requestPath, response.Code, http.StatusOK)
		}
		if response.Body.String() != "<main>FreeBooru</main>" {
			t.Fatalf("GET %s body = %q", requestPath, response.Body.String())
		}
	}

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/assets/app.js", nil))
	if response.Code != http.StatusOK || response.Body.String() != "console.log('hello')" {
		t.Fatalf("asset response = %d %q", response.Code, response.Body.String())
	}
}

func TestHandlerNeverFallsBackForAPI(t *testing.T) {
	assets := fstest.MapFS{"index.html": &fstest.MapFile{Data: []byte("index")}}
	handler := NewHandler(http.NotFoundHandler(), assets)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/missing", nil))

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}
	if response.Body.String() == "index" {
		t.Fatal("API response used SPA fallback")
	}
}

func TestEmbeddedAssetsAreRootedAtDist(t *testing.T) {
	assets, err := Assets()
	if err != nil {
		t.Fatalf("get embedded assets: %v", err)
	}
	if _, err := fs.Stat(assets, "index.html"); err != nil {
		t.Fatalf("stat embedded index: %v", err)
	}
}

func TestHTTPServerServesHelloAndApplication(t *testing.T) {
	assets := fstest.MapFS{"index.html": &fstest.MapFile{Data: []byte("FreeBooru app")}}
	api, err := webapi.New(webapi.ModeServer, staticApplication{})
	if err != nil {
		t.Fatalf("construct API: %v", err)
	}
	server := httptest.NewServer(NewHandler(api, assets))
	t.Cleanup(server.Close)

	for _, test := range []struct {
		path string
		want string
	}{
		{path: "/", want: "FreeBooru app"},
		{path: "/api/v1/hello", want: `{"message":"Hello FreeBooru","mode":"server"}`},
		{path: "/api/v1/status", want: `{"ready":true,"mode":"server"`},
	} {
		response, err := server.Client().Get(server.URL + test.path)
		if err != nil {
			t.Fatalf("GET %s: %v", test.path, err)
		}
		body, readErr := io.ReadAll(response.Body)
		closeErr := response.Body.Close()
		if readErr != nil {
			t.Fatalf("read GET %s: %v", test.path, readErr)
		}
		if closeErr != nil {
			t.Fatalf("close GET %s: %v", test.path, closeErr)
		}
		if response.StatusCode != http.StatusOK {
			t.Fatalf("GET %s status = %d", test.path, response.StatusCode)
		}
		if !strings.Contains(string(body), test.want) {
			t.Fatalf("GET %s body = %q, want %q", test.path, body, test.want)
		}
	}
}
