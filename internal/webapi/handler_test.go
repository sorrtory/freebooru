package webapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHello(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/v1/hello", nil)
	response := httptest.NewRecorder()

	New(ModeDesktop).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if contentType := response.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", contentType)
	}
	var body struct {
		Message string `json:"message"`
		Mode    Mode   `json:"mode"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Message != "Hello FreeBooru" || body.Mode != ModeDesktop {
		t.Fatalf("body = %#v", body)
	}
}

func TestHelloRejectsOtherMethods(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/api/v1/hello", nil)
	response := httptest.NewRecorder()

	New(ModeServer).ServeHTTP(response, request)

	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusMethodNotAllowed)
	}
	if allow := response.Header().Get("Allow"); allow != http.MethodGet {
		t.Fatalf("Allow = %q, want GET", allow)
	}
	if !strings.Contains(response.Body.String(), `"error"`) {
		t.Fatalf("body = %q, want JSON error", response.Body.String())
	}
}

func TestUnknownAPIPathReturnsJSON(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/v1/missing", nil)
	response := httptest.NewRecorder()

	New(ModeServer).ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}
	if contentType := response.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", contentType)
	}
}
