package core

import (
	"strings"
	"testing"

	"github.com/sorrtory/freebooru/internal/config"
	"github.com/sorrtory/freebooru/internal/evaluator"
)

func TestEvaluationErrorIncludesConfiguredReasons(t *testing.T) {
	err := (&EvaluationError{
		Operation: "tag mutation",
		Evaluation: evaluator.Evaluation{
			MissingDemands:  []config.Edge{{Reason: "a title is required"}},
			ActiveConflicts: []config.Edge{{Reason: "reviewed files cannot be drafts"}},
		},
	}).Error()
	for _, want := range []string{
		"1 missing demands and 1 active conflicts",
		"missing demand: a title is required",
		"active conflict: reviewed files cannot be drafts",
	} {
		if !strings.Contains(err, want) {
			t.Fatalf("Error() = %q, want %q", err, want)
		}
	}
}
