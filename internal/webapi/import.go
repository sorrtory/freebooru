package webapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sorrtory/freebooru/internal/config"
	"github.com/sorrtory/freebooru/internal/core"
)

const maxImportDraftBodyBytes int64 = 1 << 20

// ImportFieldResponse describes one typed control in the import workspace.
type ImportFieldResponse struct {
	Name     string   `json:"name"`
	Type     string   `json:"type"`
	Values   []string `json:"values"`
	Required bool     `json:"required"`
}

// ImportSchemaResponse describes every assignment accepted by a collection.
type ImportSchemaResponse struct {
	Collection string                `json:"collection"`
	Fields     []ImportFieldResponse `json:"fields"`
}

// AssignmentResponse is one canonical typed draft assignment.
type AssignmentResponse struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Value    any    `json:"value"`
	Required bool   `json:"required"`
}

// PredicateResponse is the JSON-safe target condition of a relationship.
type PredicateResponse struct {
	Presence bool     `json:"presence"`
	Has      []string `json:"has"`
	Is       any      `json:"is,omitempty"`
	Not      []string `json:"not"`
	Min      *int64   `json:"min,omitempty"`
	Max      *int64   `json:"max,omitempty"`
	Before   string   `json:"before,omitempty"`
	After    string   `json:"after,omitempty"`
	Regex    string   `json:"regex,omitempty"`
}

// RelationshipResponse describes one active draft relationship.
type RelationshipResponse struct {
	Kind        string            `json:"kind"`
	SourceTag   string            `json:"source_tag"`
	SourceValue string            `json:"source_value"`
	TargetTag   string            `json:"target_tag"`
	Target      PredicateResponse `json:"target"`
	Reason      string            `json:"reason"`
}

// ImportDraftResponse is one atomic canonical import workspace snapshot.
type ImportDraftResponse struct {
	Collection      string                 `json:"collection"`
	Assignments     []AssignmentResponse   `json:"assignments"`
	MissingRequired []ImportFieldResponse  `json:"missing_required"`
	MissingDemands  []RelationshipResponse `json:"missing_demands"`
	ActiveConflicts []RelationshipResponse `json:"active_conflicts"`
	Suggestions     []RelationshipResponse `json:"suggestions"`
	Complete        bool                   `json:"complete"`
}

type importDraftRequest struct {
	Assignments map[string]json.RawMessage `json:"assignments"`
}

type errorEnvelope struct {
	Error errorResponse `json:"error"`
}

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func handleCollectionRequest(response http.ResponseWriter, request *http.Request, app Application) {
	path := strings.TrimPrefix(request.URL.Path, "/api/v1/collections/")
	parts := strings.Split(path, "/")
	if len(parts) != 3 || parts[0] == "" || parts[1] != "imports" {
		writeAPIError(response, http.StatusNotFound, "route.not_found", "API endpoint not found")
		return
	}
	switch parts[2] {
	case "schema":
		handleImportSchema(response, request, app, parts[0])
	case "evaluate":
		handleImportDraft(response, request, app, parts[0])
	default:
		writeAPIError(response, http.StatusNotFound, "route.not_found", "API endpoint not found")
	}
}

func handleImportSchema(
	response http.ResponseWriter,
	request *http.Request,
	app Application,
	collection string,
) {
	if request.Method != http.MethodGet {
		methodNotAllowed(response, http.MethodGet)
		return
	}
	fields, err := app.ImportFields(collection)
	if err != nil {
		writeAPIError(response, http.StatusUnprocessableEntity, "collection.invalid", err.Error())
		return
	}
	writeJSON(response, http.StatusOK, ImportSchemaResponse{
		Collection: collection,
		Fields:     importFieldResponses(fields),
	})
}

func handleImportDraft(
	response http.ResponseWriter,
	request *http.Request,
	app Application,
	collection string,
) {
	if request.Method != http.MethodPost {
		methodNotAllowed(response, http.MethodPost)
		return
	}
	fields, err := app.ImportFields(collection)
	if err != nil {
		writeAPIError(response, http.StatusUnprocessableEntity, "collection.invalid", err.Error())
		return
	}
	request.Body = http.MaxBytesReader(response, request.Body, maxImportDraftBodyBytes)
	var body importDraftRequest
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&body); err != nil {
		writeAPIError(response, http.StatusBadRequest, "request.invalid", "Invalid import draft JSON")
		return
	}
	if err := ensureJSONEnd(decoder); err != nil {
		writeAPIError(response, http.StatusBadRequest, "request.invalid", "Invalid import draft JSON")
		return
	}
	assignments, err := decodeAssignments(body.Assignments, fields)
	if err != nil {
		writeAPIError(response, http.StatusBadRequest, "tag.value_invalid", err.Error())
		return
	}
	draft, err := app.EvaluateImportDraft(request.Context(), core.ImportDraftRequest{
		Collection:  collection,
		Assignments: assignments,
	})
	if err != nil {
		writeAPIError(response, http.StatusUnprocessableEntity, "import.draft_invalid", err.Error())
		return
	}
	writeJSON(response, http.StatusOK, importDraftResponse(draft, fields))
}

func decodeAssignments(
	values map[string]json.RawMessage,
	fields []core.ImportField,
) (map[string]any, error) {
	known := make(map[string]core.ImportField, len(fields))
	for _, field := range fields {
		known[strings.ToLower(field.Name)] = field
	}
	assignments := make(map[string]any, len(values))
	for name, raw := range values {
		field, ok := known[strings.ToLower(name)]
		if !ok {
			return nil, fmt.Errorf("tag %q is not imported by the collection", name)
		}
		value, err := decodeAssignmentValue(field.Type, raw)
		if err != nil {
			return nil, fmt.Errorf("tag %q: %w", field.Name, err)
		}
		assignments[field.Name] = value
	}
	return assignments, nil
}

func decodeAssignmentValue(tagType config.TagType, raw json.RawMessage) (any, error) {
	var target any
	switch tagType {
	case config.TagTypeBool:
		target = new(bool)
	case config.TagTypeInt:
		target = new(int64)
	case config.TagTypeText, config.TagTypeDate, config.TagTypeDatetime, config.TagTypeValue:
		target = new(string)
	case config.TagTypeMultivalue:
		target = new([]string)
	default:
		return nil, fmt.Errorf("unsupported tag type %q", tagType)
	}
	if err := json.Unmarshal(raw, target); err != nil {
		return nil, fmt.Errorf("expected %s value", tagType)
	}
	switch value := target.(type) {
	case *bool:
		return *value, nil
	case *int64:
		return *value, nil
	case *string:
		return *value, nil
	case *[]string:
		return *value, nil
	default:
		return nil, fmt.Errorf("unsupported tag type %q", tagType)
	}
}

func importDraftResponse(draft core.ImportDraft, fields []core.ImportField) ImportDraftResponse {
	assignments := make([]AssignmentResponse, 0, len(draft.Assignments))
	for _, field := range fields {
		value, ok := assignment(draft.Assignments, field.Name)
		if !ok {
			continue
		}
		assignments = append(assignments, AssignmentResponse{
			Name:     field.Name,
			Type:     string(field.Type),
			Value:    value,
			Required: field.Required,
		})
	}
	return ImportDraftResponse{
		Collection:      draft.Collection,
		Assignments:     assignments,
		MissingRequired: importFieldResponses(draft.MissingRequired),
		MissingDemands:  relationshipResponses(draft.Evaluation.MissingDemands),
		ActiveConflicts: relationshipResponses(draft.Evaluation.ActiveConflicts),
		Suggestions:     relationshipResponses(draft.Evaluation.Suggestions),
		Complete:        draft.Complete,
	}
}

func assignment(assignments map[string]any, name string) (any, bool) {
	for current, value := range assignments {
		if strings.EqualFold(current, name) {
			return value, true
		}
	}
	return nil, false
}

func importFieldResponses(fields []core.ImportField) []ImportFieldResponse {
	responses := make([]ImportFieldResponse, 0, len(fields))
	for _, field := range fields {
		responses = append(responses, ImportFieldResponse{
			Name:     field.Name,
			Type:     string(field.Type),
			Values:   append([]string{}, field.Values...),
			Required: field.Required,
		})
	}
	return responses
}

func relationshipResponses(edges []config.Edge) []RelationshipResponse {
	responses := make([]RelationshipResponse, 0, len(edges))
	for _, edge := range edges {
		responses = append(responses, RelationshipResponse{
			Kind:        string(edge.Kind),
			SourceTag:   edge.Source.Tag,
			SourceValue: edge.Source.Value,
			TargetTag:   edge.TargetTag,
			Target:      predicateResponse(edge.Predicate),
			Reason:      edge.Reason,
		})
	}
	return responses
}

func predicateResponse(predicate config.Predicate) PredicateResponse {
	response := PredicateResponse{
		Presence: predicate.Presence,
		Has:      append([]string{}, predicate.Has...),
		Is:       predicate.Is,
		Not:      append([]string{}, predicate.Not...),
		Min:      predicate.Min,
		Max:      predicate.Max,
	}
	if predicate.Before != nil {
		response.Before = predicate.Before.Format(time.RFC3339)
	}
	if predicate.After != nil {
		response.After = predicate.After.Format(time.RFC3339)
	}
	if predicate.Regex != nil {
		response.Regex = predicate.Regex.String()
	}
	return response
}

func methodNotAllowed(response http.ResponseWriter, allowed string) {
	response.Header().Set("Allow", allowed)
	writeAPIError(response, http.StatusMethodNotAllowed, "method.not_allowed", "Method not allowed")
}

func writeAPIError(response http.ResponseWriter, status int, code, message string) {
	writeJSON(response, status, errorEnvelope{Error: errorResponse{Code: code, Message: message}})
}

func ensureJSONEnd(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); err == nil {
		return fmt.Errorf("multiple JSON values")
	} else if !errors.Is(err, io.EOF) {
		return err
	}
	return nil
}
