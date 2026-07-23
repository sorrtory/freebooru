package webapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sorrtory/freebooru/internal/collection"
	"github.com/sorrtory/freebooru/internal/config"
	"github.com/sorrtory/freebooru/internal/core"
)

type fakeApplication struct {
	appConfig        config.AppConfig
	diagnostics      config.Diagnostics
	load             func(context.Context) error
	loadCalls        int
	checkCalls       int
	contextErr       error
	importFields     []core.ImportField
	draft            core.ImportDraft
	importErr        error
	collection       string
	draftRequest     core.ImportDraftRequest
	importRequest    core.ImportRequest
	importResult     core.ImportResult
	executeImportErr error
	collections      []config.CollectionConfig
	collectionInfo   core.CollectionInfo
	collectionErr    error
	files            []collection.FileRecord
	file             collection.FileRecord
	tags             []core.CollectionTagInfo
	storages         []core.CollectionStorageInfo
	resource         string
	settings         core.Settings
}

func (f *fakeApplication) Search(_ context.Context, request core.FileSearchRequest) ([]collection.FileRecord, error) {
	f.collection = request.Collection
	return f.files, f.collectionErr
}

func (f *fakeApplication) GetFile(_ context.Context, collectionName, _ string) (collection.FileRecord, error) {
	f.collection = collectionName
	return f.file, f.collectionErr
}

func (f *fakeApplication) OpenFileContent(context.Context, string, string) (core.FileContent, error) {
	return core.FileContent{}, f.collectionErr
}
func (f *fakeApplication) SetTag(context.Context, core.TagMutationRequest) (core.TagMutationResult, error) {
	return core.TagMutationResult{}, f.collectionErr
}
func (f *fakeApplication) RemoveTag(context.Context, core.TagRemovalRequest) (core.TagMutationResult, error) {
	return core.TagMutationResult{}, f.collectionErr
}
func (f *fakeApplication) ListCollectionTagInfo(_ context.Context, name string) ([]core.CollectionTagInfo, error) {
	f.collection = name
	return f.tags, f.collectionErr
}
func (f *fakeApplication) ListCollectionStorageInfo(_ context.Context, name string) ([]core.CollectionStorageInfo, error) {
	f.collection = name
	return f.storages, f.collectionErr
}
func (f *fakeApplication) ImportCollectionTag(_ context.Context, name, resource string) error {
	f.collection, f.resource = name, resource
	return f.collectionErr
}
func (f *fakeApplication) ImportCollectionStorage(_ context.Context, name, resource string) error {
	f.collection, f.resource = name, resource
	return f.collectionErr
}
func (f *fakeApplication) Settings() core.Settings { return f.settings }
func (f *fakeApplication) UpdateSettings(_ context.Context, update core.SettingsUpdate) (core.Settings, error) {
	f.resource = update.ExpectedRevision
	result := f.settings
	result.Language = update.Language
	result.DefaultCollection = update.DefaultCollection
	result.DefaultStorageName = update.DefaultStorageName
	result.HTTPPort = update.HTTPPort
	result.RemoveOnUpload = update.RemoveOnUpload
	return result, f.collectionErr
}

func (f *fakeApplication) ListCollections() ([]config.CollectionConfig, error) {
	return f.collections, f.collectionErr
}

func (f *fakeApplication) DescribeCollection(_ context.Context, name string) (core.CollectionInfo, error) {
	f.collection = name
	return f.collectionInfo, f.collectionErr
}

func (f *fakeApplication) CreateCollection(_ context.Context, name string) (core.CollectionInfo, error) {
	f.collection = name
	return f.collectionInfo, f.collectionErr
}

func (f *fakeApplication) Import(_ context.Context, request core.ImportRequest) (core.ImportResult, error) {
	f.importRequest = request
	return f.importResult, f.executeImportErr
}

func (f *fakeApplication) LoadConfig(ctx context.Context) error {
	f.loadCalls++
	f.contextErr = ctx.Err()
	if f.load != nil {
		return f.load(ctx)
	}
	return nil
}

func (f *fakeApplication) CheckConfig(ctx context.Context) config.Diagnostics {
	f.checkCalls++
	f.contextErr = ctx.Err()
	return f.diagnostics
}

func (f *fakeApplication) AppConfig() config.AppConfig {
	return f.appConfig
}

func (f *fakeApplication) ImportFields(collection string) ([]core.ImportField, error) {
	f.collection = collection
	return f.importFields, f.importErr
}

func (f *fakeApplication) EvaluateImportDraft(
	_ context.Context,
	request core.ImportDraftRequest,
) (core.ImportDraft, error) {
	f.draftRequest = request
	return f.draft, f.importErr
}

func TestNewRejectsInvalidDependencies(t *testing.T) {
	var typedNil *fakeApplication
	tests := []struct {
		name string
		mode Mode
		app  Application
	}{
		{name: "invalid mode", mode: Mode("invalid"), app: &fakeApplication{}},
		{name: "missing application", mode: ModeServer},
		{name: "typed nil application", mode: ModeServer, app: typedNil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := New(test.mode, test.app); err == nil {
				t.Fatal("New() error = nil")
			}
		})
	}
}

func TestHello(t *testing.T) {
	app := &fakeApplication{}
	request := httptest.NewRequest(http.MethodGet, "/api/v1/hello", nil)
	response := httptest.NewRecorder()

	mustNew(t, ModeDesktop, app).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if contentType := response.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", contentType)
	}
	if response.Body.String() != "{\"message\":\"Hello FreeBooru\",\"mode\":\"desktop\"}\n" {
		t.Fatalf("body = %q", response.Body.String())
	}
	if app.loadCalls != 0 || app.checkCalls != 0 {
		t.Fatal("hello endpoint inspected application status")
	}
}

func TestStatus(t *testing.T) {
	tests := []struct {
		name      string
		mode      Mode
		app       *fakeApplication
		wantReady bool
		wantCode  string
	}{
		{
			name:      "ready",
			mode:      ModeServer,
			app:       &fakeApplication{appConfig: config.DefaultAppConfig()},
			wantReady: true,
		},
		{
			name: "warning remains ready",
			mode: ModeDesktop,
			app: &fakeApplication{
				appConfig: config.DefaultAppConfig(),
				diagnostics: config.Diagnostics{{
					Severity: config.SeverityWarning,
					Code:     "tag.unused",
					Message:  "tag is not imported",
					File:     "/config/tags/example.yaml",
					Document: 2,
					Field:    "name",
				}},
			},
			wantReady: true,
			wantCode:  "tag.unused",
		},
		{
			name: "error is unready",
			mode: ModeServer,
			app: &fakeApplication{
				appConfig: config.DefaultAppConfig(),
				diagnostics: config.Diagnostics{{
					Severity: config.SeverityError,
					Code:     "reference.missing",
					Message:  "tag reference is missing",
				}},
			},
			wantCode: "reference.missing",
		},
		{
			name: "load failure is unready",
			mode: ModeDesktop,
			app: &fakeApplication{load: func(context.Context) error {
				return errors.New("load application configuration: file is missing")
			}},
			wantCode: "application.config_load",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			mustNew(t, test.mode, test.app).ServeHTTP(
				response,
				httptest.NewRequest(http.MethodGet, "/api/v1/status", nil),
			)

			if response.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
			}
			var status StatusResponse
			if err := json.NewDecoder(response.Body).Decode(&status); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if status.Ready != test.wantReady {
				t.Fatalf("ready = %t, want %t", status.Ready, test.wantReady)
			}
			if status.Mode != test.mode {
				t.Fatalf("mode = %q, want %q", status.Mode, test.mode)
			}
			if test.wantCode == "" {
				if status.Diagnostics == nil || len(status.Diagnostics) != 0 {
					t.Fatalf("diagnostics = %#v, want empty array", status.Diagnostics)
				}
			} else if len(status.Diagnostics) != 1 || status.Diagnostics[0].Code != test.wantCode {
				t.Fatalf("diagnostics = %#v, want code %q", status.Diagnostics, test.wantCode)
			}
			if test.wantCode == "application.config_load" {
				if status.DefaultCollection != "" || test.app.checkCalls != 0 {
					t.Fatalf("load failure status = %#v, check calls = %d", status, test.app.checkCalls)
				}
			} else if status.DefaultCollection != "main" || test.app.checkCalls != 1 {
				t.Fatalf("status = %#v, check calls = %d", status, test.app.checkCalls)
			}
		})
	}
}

func TestStatusMapsDiagnosticSource(t *testing.T) {
	app := &fakeApplication{
		appConfig: config.DefaultAppConfig(),
		diagnostics: config.Diagnostics{{
			Severity: config.SeverityWarning,
			Code:     "test.warning",
			Message:  "review this value",
			File:     "/config/example.yaml",
			Document: 3,
			Field:    "tags.import[1]",
		}},
	}
	response := httptest.NewRecorder()
	mustNew(t, ModeServer, app).ServeHTTP(
		response,
		httptest.NewRequest(http.MethodGet, "/api/v1/status", nil),
	)

	var status StatusResponse
	if err := json.NewDecoder(response.Body).Decode(&status); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	want := DiagnosticResponse{
		Severity: "warning",
		Code:     "test.warning",
		Message:  "review this value",
		File:     "/config/example.yaml",
		Document: 3,
		Field:    "tags.import[1]",
	}
	if len(status.Diagnostics) != 1 || status.Diagnostics[0] != want {
		t.Fatalf("diagnostics = %#v, want %#v", status.Diagnostics, want)
	}
}

func TestStatusPreservesMultipleDiagnostics(t *testing.T) {
	app := &fakeApplication{
		appConfig: config.DefaultAppConfig(),
		diagnostics: config.Diagnostics{
			{Severity: config.SeverityError, Code: "first", Message: "first problem"},
			{Severity: config.SeverityError, Code: "second", Message: "second problem"},
		},
	}
	response := httptest.NewRecorder()
	mustNew(t, ModeServer, app).ServeHTTP(
		response,
		httptest.NewRequest(http.MethodGet, "/api/v1/status", nil),
	)

	var status StatusResponse
	if err := json.NewDecoder(response.Body).Decode(&status); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(status.Diagnostics) != 2 ||
		status.Diagnostics[0].Code != "first" ||
		status.Diagnostics[1].Code != "second" {
		t.Fatalf("diagnostics = %#v, want first then second", status.Diagnostics)
	}
}

func TestStatusReloadsConfiguration(t *testing.T) {
	app := &fakeApplication{appConfig: config.DefaultAppConfig()}
	app.load = func(context.Context) error {
		if app.loadCalls == 1 {
			return errors.New("configuration unavailable")
		}
		return nil
	}
	handler := mustNew(t, ModeServer, app)

	first := httptest.NewRecorder()
	handler.ServeHTTP(first, httptest.NewRequest(http.MethodGet, "/api/v1/status", nil))
	second := httptest.NewRecorder()
	handler.ServeHTTP(second, httptest.NewRequest(http.MethodGet, "/api/v1/status", nil))

	var status StatusResponse
	if err := json.NewDecoder(second.Body).Decode(&status); err != nil {
		t.Fatalf("decode second response: %v", err)
	}
	if app.loadCalls != 2 || !status.Ready {
		t.Fatalf("load calls = %d, second status = %#v", app.loadCalls, status)
	}
}

func TestStatusPropagatesRequestContext(t *testing.T) {
	app := &fakeApplication{appConfig: config.DefaultAppConfig()}
	handler := mustNew(t, ModeServer, app)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/status", nil).WithContext(ctx))

	if app.loadCalls != 1 || !errors.Is(app.contextErr, context.Canceled) {
		t.Fatalf("load calls = %d, context error = %v", app.loadCalls, app.contextErr)
	}
}

func TestRoutesRejectOtherMethodsAndUnknownPaths(t *testing.T) {
	handler := mustNew(t, ModeServer, &fakeApplication{})
	for _, route := range []string{"/api/v1/hello", "/api/v1/status"} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodPost, route, nil))
		if response.Code != http.StatusMethodNotAllowed {
			t.Fatalf("POST %s status = %d", route, response.Code)
		}
		if response.Header().Get("Allow") != http.MethodGet || !strings.Contains(response.Body.String(), `"error"`) {
			t.Fatalf("POST %s headers/body = %#v %q", route, response.Header(), response.Body.String())
		}
	}

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/missing", nil))
	if response.Code != http.StatusNotFound || response.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("unknown route response = %d %#v", response.Code, response.Header())
	}
}

func mustNew(t *testing.T, mode Mode, app Application) http.Handler {
	t.Helper()
	handler, err := New(mode, app)
	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	return handler
}
