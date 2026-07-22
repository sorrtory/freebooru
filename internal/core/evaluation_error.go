package core

import (
	"fmt"
	"strings"

	"github.com/sorrtory/freebooru/internal/config"
	"github.com/sorrtory/freebooru/internal/evaluator"
)

// EvaluationError reports every relationship that makes a proposed state invalid.
type EvaluationError struct {
	Operation  string
	Evaluation evaluator.Evaluation
}

func (e *EvaluationError) Error() string {
	var message strings.Builder
	fmt.Fprintf(
		&message,
		"%s has %d missing demands and %d active conflicts",
		e.Operation,
		len(e.Evaluation.MissingDemands),
		len(e.Evaluation.ActiveConflicts),
	)
	for _, edge := range e.Evaluation.MissingDemands {
		fmt.Fprintf(&message, "\nmissing demand: %s", evaluationReason(edge))
	}
	for _, edge := range e.Evaluation.ActiveConflicts {
		fmt.Fprintf(&message, "\nactive conflict: %s", evaluationReason(edge))
	}
	return message.String()
}

func evaluationReason(edge config.Edge) string {
	if edge.Reason != "" {
		return edge.Reason
	}
	return fmt.Sprintf("%s requires relationship with %s", edge.Source.Tag, edge.TargetTag)
}
