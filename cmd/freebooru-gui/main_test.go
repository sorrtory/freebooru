package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sorrtory/freebooru/internal/config"
	"github.com/sorrtory/freebooru/internal/core"
	"github.com/sorrtory/freebooru/internal/webapi"
)

type fakeApplication struct{}

func (fakeApplication) LoadConfig(context.Context) error {
	return nil
}

func (fakeApplication) CheckConfig(context.Context) config.Diagnostics {
	return nil
}

func (fakeApplication) AppConfig() config.AppConfig {
	return config.DefaultAppConfig()
}

func (fakeApplication) ImportFields(string) ([]core.ImportField, error) {
	return nil, nil
}

func (fakeApplication) EvaluateImportDraft(
	context.Context,
	core.ImportDraftRequest,
) (core.ImportDraft, error) {
	return core.ImportDraft{}, nil
}

func TestAPIMiddlewareRoutesOnlyAPIRequests(t *testing.T) {
	api, err := webapi.New(webapi.ModeDesktop, fakeApplication{})
	if err != nil {
		t.Fatalf("construct API: %v", err)
	}
	next := http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		response.WriteHeader(http.StatusTeapot)
	})
	handler := apiMiddleware(api)(next)

	apiResponse := httptest.NewRecorder()
	handler.ServeHTTP(apiResponse, httptest.NewRequest(http.MethodGet, "/api/v1/hello", nil))
	if apiResponse.Code != http.StatusOK {
		t.Fatalf("API status = %d, want %d", apiResponse.Code, http.StatusOK)
	}
	if !strings.Contains(apiResponse.Body.String(), `"mode":"desktop"`) {
		t.Fatalf("API body = %q, want desktop mode", apiResponse.Body.String())
	}

	statusResponse := httptest.NewRecorder()
	handler.ServeHTTP(statusResponse, httptest.NewRequest(http.MethodGet, "/api/v1/status", nil))
	if statusResponse.Code != http.StatusOK || !strings.Contains(statusResponse.Body.String(), `"ready":true`) {
		t.Fatalf("status response = %d %q", statusResponse.Code, statusResponse.Body.String())
	}

	assetResponse := httptest.NewRecorder()
	handler.ServeHTTP(assetResponse, httptest.NewRequest(http.MethodGet, "/assets/app.js", nil))
	if assetResponse.Code != http.StatusTeapot {
		t.Fatalf("asset status = %d, want %d", assetResponse.Code, http.StatusTeapot)
	}
}
