package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sorrtory/freebooru/internal/webapi"
)

func TestAPIMiddlewareRoutesOnlyAPIRequests(t *testing.T) {
	api := webapi.New(webapi.ModeDesktop)
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

	assetResponse := httptest.NewRecorder()
	handler.ServeHTTP(assetResponse, httptest.NewRequest(http.MethodGet, "/assets/app.js", nil))
	if assetResponse.Code != http.StatusTeapot {
		t.Fatalf("asset status = %d, want %d", assetResponse.Code, http.StatusTeapot)
	}
}
