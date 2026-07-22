package webapi

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/sorrtory/freebooru/internal/collection"
	"github.com/sorrtory/freebooru/internal/config"
	"github.com/sorrtory/freebooru/internal/core"
	"github.com/sorrtory/freebooru/internal/evaluator"
)

func TestImportUploadStreamsFileAndCleansTemporarySource(t *testing.T) {
	app := &fakeApplication{
		importFields: testImportFields(),
		importResult: core.ImportResult{
			SHA256: "abc123", SizeBytes: 7, Storages: []string{"default"},
			RecordCreated: true, CreatedCopies: []string{"default"},
		},
	}
	request := multipartImportRequest(t, `{"flag":true,"score":7}`, "folder\\picture.png", []byte("content"))
	response := httptest.NewRecorder()

	mustNew(t, ModeServer, app).ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %q", response.Code, response.Body.String())
	}
	if app.importRequest.Collection != "Main" || app.importRequest.SourceFilename != "picture.png" {
		t.Fatalf("import request = %#v", app.importRequest)
	}
	if app.importRequest.Tags["score"] != int64(7) {
		t.Fatalf("score = %#v", app.importRequest.Tags["score"])
	}
	if _, err := os.Stat(app.importRequest.SourcePath); !os.IsNotExist(err) {
		t.Fatalf("temporary source still exists: %v", err)
	}
	if response.Header().Get("Location") != "/api/v1/collections/Main/files/abc123" {
		t.Fatalf("Location = %q", response.Header().Get("Location"))
	}
}

func TestImportUploadReturnsDuplicateHash(t *testing.T) {
	app := &fakeApplication{
		importFields:     testImportFields(),
		importResult:     core.ImportResult{SHA256: "existing"},
		executeImportErr: collection.ErrDuplicateFile,
	}
	response := httptest.NewRecorder()
	mustNew(t, ModeServer, app).ServeHTTP(
		response,
		multipartImportRequest(t, `{}`, "same.jpg", []byte("same")),
	)

	if response.Code != http.StatusConflict || !strings.Contains(response.Body.String(), `"code":"import.duplicate"`) || !strings.Contains(response.Body.String(), "existing") {
		t.Fatalf("response = %d %q", response.Code, response.Body.String())
	}
}

func multipartImportRequest(t *testing.T, assignments, filename string, content []byte) *http.Request {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	file, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("create file part: %v", err)
	}
	if _, err := file.Write(content); err != nil {
		t.Fatalf("write file part: %v", err)
	}
	if err := writer.WriteField("assignments", assignments); err != nil {
		t.Fatalf("write assignments: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart: %v", err)
	}
	request := httptest.NewRequest(http.MethodPost, "/api/v1/collections/Main/imports", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	return request
}

func TestImportSchema(t *testing.T) {
	app := &fakeApplication{importFields: testImportFields()}
	response := httptest.NewRecorder()

	mustNew(t, ModeServer, app).ServeHTTP(
		response,
		httptest.NewRequest(http.MethodGet, "/api/v1/collections/Main/imports/schema", nil),
	)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %q", response.Code, response.Body.String())
	}
	if app.collection != "Main" {
		t.Fatalf("collection = %q, want explicit route collection", app.collection)
	}
	var schema ImportSchemaResponse
	if err := json.NewDecoder(response.Body).Decode(&schema); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if schema.Collection != "Main" || len(schema.Fields) != 4 {
		t.Fatalf("schema = %#v", schema)
	}
	if schema.Fields[0].Name != "flag" || schema.Fields[0].Values == nil {
		t.Fatalf("first field = %#v, want flag with empty values array", schema.Fields[0])
	}
}

func TestEvaluateImportDraft(t *testing.T) {
	app := &fakeApplication{
		importFields: testImportFields(),
		draft: core.ImportDraft{
			Collection: "main",
			Assignments: map[string]any{
				"flag":    true,
				"score":   int64(7),
				"labels":  []string{"first"},
				"storage": []string{"default"},
			},
			Evaluation: evaluator.Evaluation{
				Suggestions: []config.Edge{{
					Kind:      config.RelationshipSuggest,
					Source:    config.SourceCondition{Tag: "flag"},
					TargetTag: "labels",
					Predicate: config.Predicate{Has: []string{"first"}},
					Reason:    "label this file",
				}},
			},
			Complete: true,
		},
	}
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/collections/Main/imports/evaluate",
		strings.NewReader(`{"assignments":{"flag":true,"score":7,"labels":["FIRST"]}}`),
	)
	response := httptest.NewRecorder()

	mustNew(t, ModeDesktop, app).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %q", response.Code, response.Body.String())
	}
	if app.draftRequest.Collection != "Main" {
		t.Fatalf("draft collection = %q, want Main", app.draftRequest.Collection)
	}
	if app.draftRequest.Assignments["score"] != int64(7) {
		t.Fatalf("score = %#v, want int64(7)", app.draftRequest.Assignments["score"])
	}
	labels, ok := app.draftRequest.Assignments["labels"].([]string)
	if !ok || len(labels) != 1 || labels[0] != "FIRST" {
		t.Fatalf("labels = %#v", app.draftRequest.Assignments["labels"])
	}
	var draft ImportDraftResponse
	if err := json.NewDecoder(response.Body).Decode(&draft); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !draft.Complete || len(draft.Assignments) != 4 || len(draft.Suggestions) != 1 {
		t.Fatalf("draft = %#v", draft)
	}
	suggestion := draft.Suggestions[0]
	if suggestion.SourceTag != "flag" || suggestion.TargetTag != "labels" ||
		len(suggestion.Target.Has) != 1 || suggestion.Target.Has[0] != "first" {
		t.Fatalf("suggestion = %#v", suggestion)
	}
	if draft.MissingRequired == nil || draft.MissingDemands == nil || draft.ActiveConflicts == nil {
		t.Fatalf("empty response arrays encoded as nil: %#v", draft)
	}
}

func TestImportRoutesRejectInvalidRequests(t *testing.T) {
	app := &fakeApplication{importFields: testImportFields()}
	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		wantStatus int
		wantAllow  string
		wantCode   string
	}{
		{
			name:       "schema method",
			method:     http.MethodPost,
			path:       "/api/v1/collections/main/imports/schema",
			wantStatus: http.StatusMethodNotAllowed,
			wantAllow:  http.MethodGet,
			wantCode:   "method.not_allowed",
		},
		{
			name:       "evaluate method",
			method:     http.MethodGet,
			path:       "/api/v1/collections/main/imports/evaluate",
			wantStatus: http.StatusMethodNotAllowed,
			wantAllow:  http.MethodPost,
			wantCode:   "method.not_allowed",
		},
		{
			name:       "fractional integer",
			method:     http.MethodPost,
			path:       "/api/v1/collections/main/imports/evaluate",
			body:       `{"assignments":{"score":1.5}}`,
			wantStatus: http.StatusBadRequest,
			wantCode:   "tag.value_invalid",
		},
		{
			name:       "unknown tag",
			method:     http.MethodPost,
			path:       "/api/v1/collections/main/imports/evaluate",
			body:       `{"assignments":{"missing":true}}`,
			wantStatus: http.StatusBadRequest,
			wantCode:   "tag.value_invalid",
		},
		{
			name:       "unknown route",
			method:     http.MethodGet,
			path:       "/api/v1/collections/main/missing",
			wantStatus: http.StatusNotFound,
			wantCode:   "route.not_found",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			mustNew(t, ModeServer, app).ServeHTTP(
				response,
				httptest.NewRequest(test.method, test.path, strings.NewReader(test.body)),
			)
			if response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d; body = %q", response.Code, test.wantStatus, response.Body.String())
			}
			if response.Header().Get("Allow") != test.wantAllow {
				t.Fatalf("Allow = %q, want %q", response.Header().Get("Allow"), test.wantAllow)
			}
			if !strings.Contains(response.Body.String(), `"code":"`+test.wantCode+`"`) {
				t.Fatalf("body = %q, want code %q", response.Body.String(), test.wantCode)
			}
		})
	}
}

func testImportFields() []core.ImportField {
	return []core.ImportField{
		{Name: "flag", Type: config.TagTypeBool, Required: true},
		{Name: "labels", Type: config.TagTypeMultivalue, Values: []string{"first", "second"}},
		{Name: "score", Type: config.TagTypeInt},
		{Name: "storage", Type: config.TagTypeMultivalue, Values: []string{"default"}, Required: true},
	}
}
