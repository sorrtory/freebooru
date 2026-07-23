package webapi

import (
	"encoding/json"
	"net/http"

	"github.com/sorrtory/freebooru/internal/core"
)

const maxSettingsRequestBytes int64 = 32 << 10

type settingsResponse struct {
	Revision           string   `json:"revision"`
	Language           string   `json:"language"`
	DefaultCollection  string   `json:"default_collection"`
	DefaultStorageName string   `json:"default_storage_name"`
	HTTPPort           int      `json:"http_port"`
	RemoveOnUpload     bool     `json:"remove_on_upload"`
	RestartRequired    bool     `json:"restart_required"`
	Collections        []string `json:"collections"`
	Storages           []string `json:"storages"`
}

type settingsRequest struct {
	Revision           string `json:"revision"`
	Language           string `json:"language"`
	DefaultCollection  string `json:"default_collection"`
	DefaultStorageName string `json:"default_storage_name"`
	HTTPPort           int    `json:"http_port"`
	RemoveOnUpload     bool   `json:"remove_on_upload"`
}

func handleSettings(response http.ResponseWriter, request *http.Request, app Application) {
	switch request.Method {
	case http.MethodGet:
		writeJSON(response, http.StatusOK, settingsResponseFromCore(app.Settings(), false))
	case http.MethodPut:
		request.Body = http.MaxBytesReader(response, request.Body, maxSettingsRequestBytes)
		decoder := json.NewDecoder(request.Body)
		decoder.DisallowUnknownFields()
		var body settingsRequest
		if err := decoder.Decode(&body); err != nil || ensureJSONEnd(decoder) != nil {
			writeAPIError(response, http.StatusBadRequest, "request.invalid", "Invalid settings JSON")
			return
		}
		before := app.Settings()
		updated, err := app.UpdateSettings(request.Context(), core.SettingsUpdate{
			ExpectedRevision: body.Revision, Language: body.Language,
			DefaultCollection:  body.DefaultCollection,
			DefaultStorageName: body.DefaultStorageName, HTTPPort: body.HTTPPort,
			RemoveOnUpload: body.RemoveOnUpload,
		})
		if err != nil {
			writeAPIError(response, http.StatusConflict, "settings.invalid", "Settings could not be saved; reload and try again")
			return
		}
		writeJSON(response, http.StatusOK, settingsResponseFromCore(updated, before.HTTPPort != updated.HTTPPort))
	default:
		response.Header().Set("Allow", http.MethodGet+", "+http.MethodPut)
		writeAPIError(response, http.StatusMethodNotAllowed, "method.not_allowed", "Method not allowed")
	}
}

func settingsResponseFromCore(settings core.Settings, restart bool) settingsResponse {
	return settingsResponse{
		Revision: settings.Revision, Language: settings.Language,
		DefaultCollection:  settings.DefaultCollection,
		DefaultStorageName: settings.DefaultStorageName, HTTPPort: settings.HTTPPort,
		RemoveOnUpload: settings.RemoveOnUpload, RestartRequired: restart,
		Collections: append([]string{}, settings.Collections...),
		Storages:    append([]string{}, settings.Storages...),
	}
}
