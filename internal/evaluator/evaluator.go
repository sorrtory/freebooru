package evaluator

import (
	"fmt"

	"github.com/sorrtory/freebooru/internal/config"
)

// Evaluator applies one immutable compiled graph to file states.
type Evaluator struct {
	catalog *config.Catalog
	graph   *config.Graph
}

// Evaluation is one consistent snapshot of active relationship results.
type Evaluation struct {
	MissingDemands  []config.Edge
	ActiveConflicts []config.Edge
	Suggestions     []config.Edge
}

// Valid reports whether all demands are satisfied and no conflict is active.
func (e Evaluation) Valid() bool {
	return len(e.MissingDemands) == 0 && len(e.ActiveConflicts) == 0
}

// New constructs an evaluator from validated immutable configuration.
func New(catalog *config.Catalog, graph *config.Graph) (*Evaluator, error) {
	if catalog == nil {
		return nil, fmt.Errorf("catalog is required")
	}
	if graph == nil {
		return nil, fmt.Errorf("graph is required")
	}
	return &Evaluator{catalog: catalog, graph: graph}, nil
}

// ValidateFile evaluates every edge activated by the file's assigned tags.
// Suggestions are returned only while their target condition is unsatisfied.
func (e *Evaluator) ValidateFile(state FileState) Evaluation {
	var result Evaluation
	for _, source := range state.ActiveSources() {
		for _, edge := range e.graph.Outgoing(source) {
			matches := e.matchesPredicate(state, edge)
			switch edge.Kind {
			case config.RelationshipSuggest:
				if !matches {
					result.Suggestions = append(result.Suggestions, edge)
				}
			case config.RelationshipDemand:
				if !matches {
					result.MissingDemands = append(result.MissingDemands, edge)
				}
			case config.RelationshipConflict:
				if matches {
					result.ActiveConflicts = append(result.ActiveConflicts, edge)
				}
			}
		}
	}
	return result
}

// MissingDemands returns currently unsatisfied mandatory relationships.
func (e *Evaluator) MissingDemands(state FileState) []config.Edge {
	return e.ValidateFile(state).MissingDemands
}

// ActiveConflicts returns forbidden relationships currently matching the file.
func (e *Evaluator) ActiveConflicts(state FileState) []config.Edge {
	return e.ValidateFile(state).ActiveConflicts
}

// Suggestions returns active, currently unsatisfied recommendations.
func (e *Evaluator) Suggestions(state FileState) []config.Edge {
	return e.ValidateFile(state).Suggestions
}
