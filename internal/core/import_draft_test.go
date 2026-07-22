package core

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestEvaluateImportDraftReportsIncompleteRequiredFields(t *testing.T) {
	app := newImportTestCore(t)
	app.open = func(context.Context, string) (CollectionDatabase, error) {
		t.Fatal("EvaluateImportDraft() opened a collection database")
		return nil, nil
	}

	draft, err := app.EvaluateImportDraft(t.Context(), ImportDraftRequest{
		Collection:  "MAIN",
		Assignments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("EvaluateImportDraft() error = %v", err)
	}
	if draft.Collection != "main" {
		t.Fatalf("collection = %q, want main", draft.Collection)
	}
	if draft.Complete {
		t.Fatal("draft is complete without required rating")
	}
	if len(draft.MissingRequired) != 1 || draft.MissingRequired[0].Name != "rating" {
		t.Fatalf("missing required = %#v, want rating", draft.MissingRequired)
	}
	if reviewed, ok := draft.Assignments["reviewed"].(bool); !ok || !reviewed {
		t.Fatalf("reviewed = %#v, want automatic true", draft.Assignments["reviewed"])
	}
	storages, ok := draft.Assignments["storage"].([]string)
	if !ok || len(storages) != 1 || storages[0] != "default" {
		t.Fatalf("storage = %#v, want default", draft.Assignments["storage"])
	}
}

func TestEvaluateImportDraftReturnsCanonicalCompleteSnapshot(t *testing.T) {
	app := newImportTestCore(t)

	draft, err := app.EvaluateImportDraft(t.Context(), ImportDraftRequest{
		Collection: "main",
		Assignments: map[string]any{
			"RATING": "SAFE",
			"score":  int64(7),
		},
	})
	if err != nil {
		t.Fatalf("EvaluateImportDraft() error = %v", err)
	}
	if !draft.Complete || len(draft.MissingRequired) != 0 {
		t.Fatalf("draft = %#v, want complete", draft)
	}
	if draft.Assignments["rating"] != "safe" || draft.Assignments["score"] != int64(7) {
		t.Fatalf("assignments = %#v, want canonical values", draft.Assignments)
	}
}

func TestEvaluateImportDraftReturnsRelationships(t *testing.T) {
	app := newImportTestCore(t)

	draft, err := app.EvaluateImportDraft(t.Context(), ImportDraftRequest{
		Collection: "main",
		Assignments: map[string]any{
			"rating":  "safe",
			"trigger": true,
		},
	})
	if err != nil {
		t.Fatalf("EvaluateImportDraft() error = %v", err)
	}
	if draft.Complete {
		t.Fatal("draft with demand and conflict is complete")
	}
	if len(draft.Evaluation.MissingDemands) != 1 ||
		draft.Evaluation.MissingDemands[0].TargetTag != "title" {
		t.Fatalf("missing demands = %#v", draft.Evaluation.MissingDemands)
	}
	if len(draft.Evaluation.ActiveConflicts) != 1 ||
		draft.Evaluation.ActiveConflicts[0].TargetTag != "reviewed" {
		t.Fatalf("active conflicts = %#v", draft.Evaluation.ActiveConflicts)
	}
	if len(draft.Evaluation.Suggestions) != 1 ||
		draft.Evaluation.Suggestions[0].TargetTag != "score" {
		t.Fatalf("suggestions = %#v", draft.Evaluation.Suggestions)
	}
}

func TestEvaluateImportDraftRejectsInvalidInput(t *testing.T) {
	app := newImportTestCore(t)
	tests := []struct {
		name    string
		request ImportDraftRequest
		want    string
	}{
		{name: "missing collection", request: ImportDraftRequest{}, want: "collection is required"},
		{
			name: "unavailable tag",
			request: ImportDraftRequest{
				Collection:  "main",
				Assignments: map[string]any{"missing": true},
			},
			want: "not imported",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := app.EvaluateImportDraft(t.Context(), test.request)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("EvaluateImportDraft() error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestEvaluateImportDraftPropagatesCancellation(t *testing.T) {
	app := newImportTestCore(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := app.EvaluateImportDraft(ctx, ImportDraftRequest{Collection: "main"})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("EvaluateImportDraft() error = %v, want context canceled", err)
	}
}
