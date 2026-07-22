package webapi

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"

	"github.com/sorrtory/freebooru/internal/core"
)

const maxCollectionRequestBytes int64 = 16 << 10

type collectionSummaryResponse struct {
	Name      string `json:"name"`
	IsDefault bool   `json:"is_default"`
}

type collectionsResponse struct {
	DefaultCollection string                      `json:"default_collection"`
	Collections       []collectionSummaryResponse `json:"collections"`
}

type collectionInfoResponse struct {
	Name           string   `json:"name"`
	Comment        string   `json:"comment"`
	IsDefault      bool     `json:"is_default"`
	TagCount       int      `json:"tag_count"`
	RequiredCount  int      `json:"required_count"`
	Storages       []string `json:"storages"`
	FileCount      int64    `json:"file_count"`
	TotalSizeBytes int64    `json:"total_size_bytes"`
}

type createCollectionRequest struct {
	Name string `json:"name"`
}

func handleCollections(response http.ResponseWriter, request *http.Request, app Application) {
	switch request.Method {
	case http.MethodGet:
		collections, err := app.ListCollections()
		if err != nil {
			writeAPIError(response, http.StatusServiceUnavailable, "application.not_ready", "Collections are unavailable")
			return
		}
		items := make([]collectionSummaryResponse, 0, len(collections))
		defaultName := app.AppConfig().DefaultCollection
		for _, item := range collections {
			items = append(items, collectionSummaryResponse{Name: item.Name, IsDefault: strings.EqualFold(item.Name, defaultName)})
		}
		sort.Slice(items, func(i, j int) bool { return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name) })
		writeJSON(response, http.StatusOK, collectionsResponse{DefaultCollection: defaultName, Collections: items})
	case http.MethodPost:
		request.Body = http.MaxBytesReader(response, request.Body, maxCollectionRequestBytes)
		decoder := json.NewDecoder(request.Body)
		decoder.DisallowUnknownFields()
		var body createCollectionRequest
		if err := decoder.Decode(&body); err != nil || ensureJSONEnd(decoder) != nil {
			writeAPIError(response, http.StatusBadRequest, "request.invalid", "Invalid collection JSON")
			return
		}
		info, err := app.CreateCollection(request.Context(), body.Name)
		if err != nil {
			writeAPIError(response, http.StatusUnprocessableEntity, "collection.invalid", "Collection name is invalid or already in use")
			return
		}
		response.Header().Set("Location", "/api/v1/collections/"+info.Name)
		writeJSON(response, http.StatusCreated, collectionInfoResponseFromCore(info))
	default:
		response.Header().Set("Allow", http.MethodGet+", "+http.MethodPost)
		writeAPIError(response, http.StatusMethodNotAllowed, "method.not_allowed", "Method not allowed")
	}
}

func handleCollectionInfo(response http.ResponseWriter, request *http.Request, app Application, name string) {
	if request.Method != http.MethodGet {
		methodNotAllowed(response, http.MethodGet)
		return
	}
	info, err := app.DescribeCollection(request.Context(), name)
	if err != nil {
		writeAPIError(response, http.StatusNotFound, "collection.not_found", "Collection is unavailable")
		return
	}
	writeJSON(response, http.StatusOK, collectionInfoResponseFromCore(info))
}

func collectionInfoResponseFromCore(info core.CollectionInfo) collectionInfoResponse {
	return collectionInfoResponse{
		Name: info.Name, Comment: info.Comment, IsDefault: info.Default,
		TagCount: info.TagCount, RequiredCount: info.RequiredCount,
		Storages: append([]string{}, info.Storages...), FileCount: info.FileCount,
		TotalSizeBytes: info.TotalSizeBytes,
	}
}
