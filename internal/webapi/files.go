package webapi

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/sorrtory/freebooru/internal/collection"
	"github.com/sorrtory/freebooru/internal/config"
	"github.com/sorrtory/freebooru/internal/core"
)

const defaultFilePageLimit int64 = 24
const maxFilePageLimit int64 = 100

type fileAssignmentResponse struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value any    `json:"value"`
}

type fileSourceResponse struct {
	Filename   string `json:"filename"`
	ObservedAt string `json:"observed_at"`
}

type fileResponse struct {
	SHA256            string                   `json:"sha256"`
	Filename          string                   `json:"filename"`
	SizeBytes         int64                    `json:"size_bytes"`
	MIMEType          string                   `json:"mime_type"`
	ImportedAt        string                   `json:"imported_at"`
	UpdatedAt         string                   `json:"updated_at"`
	LastInteractionAt string                   `json:"last_interaction_at"`
	Assignments       []fileAssignmentResponse `json:"assignments"`
	Storages          []string                 `json:"storages"`
	Sources           []fileSourceResponse     `json:"sources"`
	ContentURL        string                   `json:"content_url"`
}

type filesResponse struct {
	Files      []fileResponse `json:"files"`
	Limit      int64          `json:"limit"`
	Offset     int64          `json:"offset"`
	HasMore    bool           `json:"has_more"`
	NextOffset *int64         `json:"next_offset"`
}

func handleFiles(response http.ResponseWriter, request *http.Request, app Application, collectionName string, parts []string) {
	if len(parts) == 2 {
		handleFileSearch(response, request, app, collectionName)
		return
	}
	hash := parts[2]
	if !validSHA256(hash) {
		writeAPIError(response, http.StatusBadRequest, "request.invalid", "File hash must be a lowercase SHA-256")
		return
	}
	if len(parts) == 3 {
		handleFileDetail(response, request, app, collectionName, hash)
		return
	}
	if len(parts) == 4 && parts[3] == "content" {
		handleFileContent(response, request, app, collectionName, hash)
		return
	}
	if len(parts) == 5 && parts[3] == "tags" && parts[4] != "" {
		handleFileTag(response, request, app, collectionName, hash, parts[4])
		return
	}
	writeAPIError(response, http.StatusNotFound, "route.not_found", "API endpoint not found")
}

type fileTagRequest struct {
	Value json.RawMessage `json:"value"`
}

func handleFileTag(response http.ResponseWriter, request *http.Request, app Application, collectionName, hash, tagName string) {
	if request.Method != http.MethodPut && request.Method != http.MethodDelete {
		response.Header().Set("Allow", http.MethodPut+", "+http.MethodDelete)
		writeAPIError(response, http.StatusMethodNotAllowed, "method.not_allowed", "Method not allowed")
		return
	}
	if request.Method == http.MethodPut {
		var body fileTagRequest
		decoder := json.NewDecoder(http.MaxBytesReader(response, request.Body, 1<<20))
		if err := decoder.Decode(&body); err != nil || len(body.Value) == 0 {
			writeAPIError(response, http.StatusBadRequest, "request.invalid", "A typed tag value is required")
			return
		}
		fields, err := app.ImportFields(collectionName)
		if err != nil {
			writeAPIError(response, http.StatusBadRequest, "tag.invalid", "Tag is unavailable")
			return
		}
		var tagType config.TagType
		for _, field := range fields {
			if field.Name == tagName {
				tagType = field.Type
				break
			}
		}
		if tagType == "" {
			writeAPIError(response, http.StatusBadRequest, "tag.invalid", "Tag is unavailable")
			return
		}
		value, err := decodeAssignmentValue(tagType, body.Value)
		if err != nil {
			writeAPIError(response, http.StatusBadRequest, "request.invalid", "Tag value is invalid")
			return
		}
		if _, err := app.SetTag(request.Context(), core.TagMutationRequest{Collection: collectionName, SHA256: hash, Tag: tagName, Value: value}); err != nil {
			writeAPIError(response, http.StatusConflict, "tag.invalid", err.Error())
			return
		}
	} else if _, err := app.RemoveTag(request.Context(), core.TagRemovalRequest{Collection: collectionName, SHA256: hash, Tag: tagName}); err != nil {
		writeAPIError(response, http.StatusConflict, "tag.invalid", err.Error())
		return
	}
	file, err := app.GetFile(request.Context(), collectionName, hash)
	if err != nil {
		writeAPIError(response, http.StatusNotFound, "file.not_found", "File is unavailable")
		return
	}
	response.Header().Set("ETag", fileETag(file))
	writeJSON(response, http.StatusOK, fileResponseFromRecord(collectionName, file, true))
}

func handleFileSearch(response http.ResponseWriter, request *http.Request, app Application, collectionName string) {
	if request.Method != http.MethodGet {
		methodNotAllowed(response, http.MethodGet)
		return
	}
	limit, err := queryInt(request, "limit", defaultFilePageLimit)
	if err != nil || limit < 1 || limit > maxFilePageLimit {
		writeAPIError(response, http.StatusBadRequest, "request.invalid", "Limit must be between 1 and 100")
		return
	}
	offset, err := queryInt(request, "offset", 0)
	if err != nil || offset < 0 {
		writeAPIError(response, http.StatusBadRequest, "request.invalid", "Offset must be non-negative")
		return
	}
	fetchLimit := limit + 1
	files, err := app.Search(request.Context(), core.FileSearchRequest{Collection: collectionName, Terms: request.URL.Query()["term"], Limit: &fetchLimit, Offset: offset})
	if err != nil {
		writeAPIError(response, http.StatusBadRequest, "search.invalid", err.Error())
		return
	}
	hasMore := int64(len(files)) > limit
	if hasMore {
		files = files[:limit]
	}
	items := make([]fileResponse, 0, len(files))
	for _, file := range files {
		items = append(items, fileResponseFromRecord(collectionName, file, false))
	}
	var next *int64
	if hasMore {
		value := offset + limit
		next = &value
	}
	writeJSON(response, http.StatusOK, filesResponse{Files: items, Limit: limit, Offset: offset, HasMore: hasMore, NextOffset: next})
}

func handleFileDetail(response http.ResponseWriter, request *http.Request, app Application, collectionName, hash string) {
	if request.Method != http.MethodGet {
		methodNotAllowed(response, http.MethodGet)
		return
	}
	file, err := app.GetFile(request.Context(), collectionName, hash)
	if err != nil {
		writeAPIError(response, http.StatusNotFound, "file.not_found", "File is unavailable")
		return
	}
	response.Header().Set("ETag", fileETag(file))
	writeJSON(response, http.StatusOK, fileResponseFromRecord(collectionName, file, true))
}

func handleFileContent(response http.ResponseWriter, request *http.Request, app Application, collectionName, hash string) {
	if request.Method != http.MethodGet && request.Method != http.MethodHead {
		response.Header().Set("Allow", http.MethodGet+", "+http.MethodHead)
		writeAPIError(response, http.StatusMethodNotAllowed, "method.not_allowed", "Method not allowed")
		return
	}
	content, err := app.OpenFileContent(request.Context(), collectionName, hash)
	if err != nil {
		writeAPIError(response, http.StatusNotFound, "file.content_unavailable", "File content is unavailable")
		return
	}
	defer func() { _ = content.Reader.Close() }()
	mediaType := content.MIMEType
	if mediaType == "" {
		mediaType = "application/octet-stream"
	}
	disposition := "attachment"
	if safeInlineMediaType(mediaType) {
		disposition = "inline"
	}
	response.Header().Set("Content-Type", mediaType)
	response.Header().Set("Content-Disposition", mime.FormatMediaType(disposition, map[string]string{"filename": content.Filename}))
	response.Header().Set("X-Content-Type-Options", "nosniff")
	response.Header().Set("ETag", `"`+content.SHA256+`"`)
	http.ServeContent(response, request, content.Filename, time.Time{}, content.Reader)
}

func safeInlineMediaType(mediaType string) bool {
	base, _, err := mime.ParseMediaType(mediaType)
	if err != nil {
		return false
	}
	if strings.HasPrefix(base, "audio/") || strings.HasPrefix(base, "video/") {
		return true
	}
	switch base {
	case "image/avif", "image/gif", "image/jpeg", "image/png", "image/webp", "application/pdf", "text/plain":
		return true
	default:
		return false
	}
}

func fileResponseFromRecord(collectionName string, file collection.FileRecord, includeSources bool) fileResponse {
	filename := file.SHA256
	if len(file.Sources) > 0 && file.Sources[len(file.Sources)-1].Filename != "" {
		filename = file.Sources[len(file.Sources)-1].Filename
	}
	assignments := make([]fileAssignmentResponse, 0, len(file.Tags))
	for _, tag := range file.Tags {
		assignments = append(assignments, fileAssignmentResponse{Name: tag.Name, Type: tag.Type, Value: tagRecordValue(tag)})
	}
	sources := make([]fileSourceResponse, 0)
	if includeSources {
		for _, source := range file.Sources {
			sources = append(sources, fileSourceResponse{Filename: source.Filename, ObservedAt: source.ObservedAt})
		}
	}
	return fileResponse{
		SHA256: file.SHA256, Filename: filename, SizeBytes: file.SizeBytes, MIMEType: file.MIMEType,
		ImportedAt: file.ImportedAt, UpdatedAt: file.UpdatedAt, LastInteractionAt: file.LastInteractionAt,
		Assignments: assignments, Storages: append([]string{}, file.Storages...), Sources: sources,
		ContentURL: "/api/v1/collections/" + collectionName + "/files/" + file.SHA256 + "/content",
	}
}

func tagRecordValue(tag collection.TagRecord) any {
	if tag.IntegerValue != nil {
		return *tag.IntegerValue
	}
	if tag.TextValue != nil {
		return *tag.TextValue
	}
	if tag.Type == "bool" {
		return true
	}
	return append([]string{}, tag.Values...)
}

func fileETag(file collection.FileRecord) string {
	digest := sha256.Sum256([]byte(file.UpdatedAt))
	return `"` + hex.EncodeToString(digest[:]) + `"`
}

func queryInt(request *http.Request, name string, fallback int64) (int64, error) {
	raw := request.URL.Query().Get(name)
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", name, err)
	}
	return value, nil
}

func validSHA256(value string) bool {
	if len(value) != 64 || strings.ToLower(value) != value {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}
