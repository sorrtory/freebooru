package webapi

import (
	"net/http"
	"strings"
)

type collectionTagResponse struct {
	Name            string            `json:"name"`
	Type            string            `json:"type"`
	Comment         string            `json:"comment"`
	Values          []predefinedValue `json:"values"`
	Required        bool              `json:"required"`
	Imported        bool              `json:"imported"`
	System          bool              `json:"system"`
	AssignmentCount int64             `json:"assignment_count"`
}

type predefinedValue struct {
	Value   string `json:"value"`
	Comment string `json:"comment"`
}

type collectionStorageResponse struct {
	Name           string `json:"name"`
	Type           string `json:"type"`
	Comment        string `json:"comment"`
	Imported       bool   `json:"imported"`
	FileCount      int64  `json:"file_count"`
	TotalSizeBytes int64  `json:"total_size_bytes"`
}

func handleCollectionResources(
	response http.ResponseWriter,
	request *http.Request,
	app Application,
	parts []string,
) {
	if len(parts) == 2 && request.Method == http.MethodGet {
		if parts[1] == "tags" {
			handleCollectionTags(response, request, app, parts[0])
			return
		}
		handleCollectionStorages(response, request, app, parts[0])
		return
	}
	if len(parts) == 4 && parts[3] == "import" && request.Method == http.MethodPost {
		var err error
		if parts[1] == "tags" {
			err = app.ImportCollectionTag(request.Context(), parts[0], parts[2])
		} else {
			err = app.ImportCollectionStorage(request.Context(), parts[0], parts[2])
		}
		if err != nil {
			writeAPIError(response, http.StatusUnprocessableEntity, "resource.invalid", "Resource could not be imported into this collection")
			return
		}
		response.WriteHeader(http.StatusNoContent)
		return
	}
	if len(parts) == 2 {
		methodNotAllowed(response, http.MethodGet)
		return
	}
	writeAPIError(response, http.StatusNotFound, "route.not_found", "API endpoint not found")
}

func handleCollectionTags(response http.ResponseWriter, request *http.Request, app Application, collectionName string) {
	items, err := app.ListCollectionTagInfo(request.Context(), collectionName)
	if err != nil {
		writeAPIError(response, http.StatusNotFound, "collection.not_found", "Collection tags are unavailable")
		return
	}
	result := make([]collectionTagResponse, 0, len(items))
	query := strings.ToLower(strings.TrimSpace(request.URL.Query().Get("q")))
	requiredOnly := request.URL.Query().Get("required") == "true"
	for _, item := range items {
		if query != "" && !strings.Contains(strings.ToLower(item.Name), query) {
			continue
		}
		if requiredOnly && !item.Required {
			continue
		}
		values := make([]predefinedValue, 0, len(item.Values))
		for _, value := range item.Values {
			values = append(values, predefinedValue{Value: value.Val, Comment: value.Comment})
		}
		result = append(result, collectionTagResponse{
			Name: item.Name, Type: string(item.Type), Comment: item.Comment,
			Values: values, Required: item.Required, Imported: item.Imported,
			System: item.System, AssignmentCount: item.AssignmentCount,
		})
	}
	writeJSON(response, http.StatusOK, map[string]any{"tags": result})
}

func handleCollectionStorages(response http.ResponseWriter, request *http.Request, app Application, collectionName string) {
	items, err := app.ListCollectionStorageInfo(request.Context(), collectionName)
	if err != nil {
		writeAPIError(response, http.StatusNotFound, "collection.not_found", "Collection storage is unavailable")
		return
	}
	result := make([]collectionStorageResponse, 0, len(items))
	for _, item := range items {
		result = append(result, collectionStorageResponse{
			Name: item.Name, Type: item.Type, Comment: item.Comment,
			Imported: item.Imported, FileCount: item.FileCount,
			TotalSizeBytes: item.TotalSizeBytes,
		})
	}
	writeJSON(response, http.StatusOK, map[string]any{"storages": result})
}
