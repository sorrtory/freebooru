// Package webapi exposes FreeBooru workflows over versioned JSON endpoints.
package webapi

import (
	"encoding/json"
	"net/http"
)

// Mode describes the process serving the API.
type Mode string

const (
	// ModeServer identifies the standalone HTTP server.
	ModeServer Mode = "server"
	// ModeDesktop identifies the Wails asset server.
	ModeDesktop Mode = "desktop"
)

// New returns the versioned FreeBooru API handler.
func New(mode Mode) http.Handler {
	mux := http.NewServeMux()
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
	mux.HandleFunc("/api/", func(response http.ResponseWriter, _ *http.Request) {
		writeJSON(response, http.StatusNotFound, map[string]string{"error": "API endpoint not found"})
	})
	return mux
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
