// Package webapi exposes FreeBooru workflows over versioned JSON endpoints.
package webapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"sync"

	"github.com/sorrtory/freebooru/internal/config"
	"github.com/sorrtory/freebooru/internal/core"
)

// Mode describes the process serving the API.
type Mode string

const (
	// ModeServer identifies the standalone HTTP server.
	ModeServer Mode = "server"
	// ModeDesktop identifies the Wails asset server.
	ModeDesktop Mode = "desktop"
)

// Application is the Core behavior consumed by the current API route groups.
type Application interface {
	LoadConfig(context.Context) error
	CheckConfig(context.Context) config.Diagnostics
	AppConfig() config.AppConfig
	ImportFields(string) ([]core.ImportField, error)
	EvaluateImportDraft(context.Context, core.ImportDraftRequest) (core.ImportDraft, error)
}

// DiagnosticResponse is one configuration problem exposed to the frontend.
type DiagnosticResponse struct {
	Severity string `json:"severity"`
	Code     string `json:"code"`
	Message  string `json:"message"`
	File     string `json:"file"`
	Document int    `json:"document"`
	Field    string `json:"field"`
}

// StatusResponse describes whether FreeBooru configuration is usable.
type StatusResponse struct {
	Ready             bool                 `json:"ready"`
	Mode              Mode                 `json:"mode"`
	DefaultCollection string               `json:"default_collection"`
	Diagnostics       []DiagnosticResponse `json:"diagnostics"`
}

// New returns the versioned FreeBooru API handler.
func New(mode Mode, app Application) (http.Handler, error) {
	if mode != ModeServer && mode != ModeDesktop {
		return nil, fmt.Errorf("invalid API mode %q", mode)
	}
	if app == nil || isNilApplication(app) {
		return nil, fmt.Errorf("application is required")
	}

	mux := http.NewServeMux()
	var statusMu sync.Mutex
	mux.HandleFunc("/api/v1/hello", func(response http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			response.Header().Set("Allow", http.MethodGet)
			writeJSON(response, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		writeJSON(response, http.StatusOK, struct {
			Message string `json:"message"`
			Mode    Mode   `json:"mode"`
		}{
			Message: "Hello FreeBooru",
			Mode:    mode,
		})
	})
	mux.HandleFunc("/api/v1/status", func(response http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			response.Header().Set("Allow", http.MethodGet)
			writeJSON(response, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		statusMu.Lock()
		defer statusMu.Unlock()
		writeJSON(response, http.StatusOK, applicationStatus(request.Context(), mode, app))
	})
	mux.HandleFunc("/api/v1/collections/", func(response http.ResponseWriter, request *http.Request) {
		handleCollectionRequest(response, request, app)
	})
	mux.HandleFunc("/api/", func(response http.ResponseWriter, _ *http.Request) {
		writeJSON(response, http.StatusNotFound, map[string]string{"error": "API endpoint not found"})
	})
	return mux, nil
}

func isNilApplication(app Application) bool {
	value := reflect.ValueOf(app)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

func applicationStatus(ctx context.Context, mode Mode, app Application) StatusResponse {
	status := StatusResponse{
		Mode:        mode,
		Diagnostics: make([]DiagnosticResponse, 0),
	}
	if err := app.LoadConfig(ctx); err != nil {
		status.Diagnostics = append(status.Diagnostics, DiagnosticResponse{
			Severity: string(config.SeverityError),
			Code:     "application.config_load",
			Message:  err.Error(),
		})
		return status
	}

	diagnostics := app.CheckConfig(ctx)
	status.Ready = !diagnostics.HasErrors()
	status.DefaultCollection = app.AppConfig().DefaultCollection
	for _, diagnostic := range diagnostics {
		status.Diagnostics = append(status.Diagnostics, DiagnosticResponse{
			Severity: string(diagnostic.Severity),
			Code:     diagnostic.Code,
			Message:  diagnostic.Message,
			File:     diagnostic.File,
			Document: diagnostic.Document,
			Field:    diagnostic.Field,
		})
	}
	return status
}

func writeJSON(response http.ResponseWriter, status int, value any) {
	data, err := json.Marshal(value)
	if err != nil {
		http.Error(response, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(status)
	if _, err := response.Write(append(data, '\n')); err != nil {
		return
	}
}
