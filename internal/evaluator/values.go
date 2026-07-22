package evaluator

import (
	"fmt"

	"github.com/sorrtory/freebooru/internal/config"
)

// UnavailableReason explains one new violation caused by assigning a value.
// Edge preserves the originating YAML location and configured reason.
type UnavailableReason struct {
	Edge    config.Edge
	Message string
}

// ValueAvailability describes whether one declared value can be assigned to
// the current state without introducing a new demand or conflict violation.
type ValueAvailability struct {
	Value   config.PredefinedValue
	Allowed bool
	Reasons []UnavailableReason
}

// AllowedValues evaluates every declared value against a hypothetical copy of
// the state. Existing unrelated violations are treated as the baseline and do
// not make every candidate unavailable.
func (e *Evaluator) AllowedValues(
	tagName string,
	state FileState,
) ([]ValueAvailability, error) {
	tag, _, ok := e.catalog.Tag(tagName)
	if !ok {
		return nil, fmt.Errorf("tag %q does not exist", tagName)
	}
	if tag.Type != config.TagTypeValue && tag.Type != config.TagTypeMultivalue {
		return nil, fmt.Errorf("tag %q has no predefined values", tag.Name)
	}
	baseline := violationIDs(e.ValidateFile(state))
	values := e.catalog.DeclaredValues(tag.Name)
	availability := make([]ValueAvailability, 0, len(values))
	for _, value := range values {
		candidate := state.withValue(tag, value.Val)
		reasons := newViolationReasons(e.ValidateFile(candidate), baseline)
		availability = append(availability, ValueAvailability{
			Value:   value,
			Allowed: len(reasons) == 0,
			Reasons: reasons,
		})
	}
	return availability, nil
}

func (s FileState) withValue(tag config.TagConfig, value string) FileState {
	assignments := make(map[string]assignment, len(s.assignments)+1)
	for key, assigned := range s.assignments {
		assigned.value = cloneValue(assigned.value)
		assignments[key] = assigned
	}
	key := normalizeName(tag.Name)
	if tag.Type == config.TagTypeValue {
		assignments[key] = assignment{tag: tag, value: value}
		return FileState{assignments: assignments}
	}
	values := []string{}
	if assigned, ok := assignments[key]; ok {
		values = append(values, assigned.value.([]string)...)
	}
	if !containsNormalized(values, value) {
		values = append(values, value)
	}
	assignments[key] = assignment{tag: tag, value: values}
	return FileState{assignments: assignments}
}

func violationIDs(evaluation Evaluation) map[string]struct{} {
	ids := make(map[string]struct{}, len(evaluation.MissingDemands)+len(evaluation.ActiveConflicts))
	for _, edge := range evaluation.MissingDemands {
		ids[edgeID(edge)] = struct{}{}
	}
	for _, edge := range evaluation.ActiveConflicts {
		ids[edgeID(edge)] = struct{}{}
	}
	return ids
}

func newViolationReasons(
	evaluation Evaluation,
	baseline map[string]struct{},
) []UnavailableReason {
	var reasons []UnavailableReason
	for _, edge := range evaluation.MissingDemands {
		if _, existed := baseline[edgeID(edge)]; !existed {
			reasons = append(reasons, unavailableReason(edge, "requires"))
		}
	}
	for _, edge := range evaluation.ActiveConflicts {
		if _, existed := baseline[edgeID(edge)]; !existed {
			reasons = append(reasons, unavailableReason(edge, "conflicts with"))
		}
	}
	return reasons
}

func unavailableReason(edge config.Edge, fallback string) UnavailableReason {
	message := edge.Reason
	if message == "" {
		message = fmt.Sprintf("%s %s", fallback, edge.TargetTag)
	}
	return UnavailableReason{Edge: edge, Message: message}
}

func edgeID(edge config.Edge) string {
	return fmt.Sprintf(
		"%s\x00%d\x00%s\x00%s",
		edge.Location.File,
		edge.Location.Document,
		edge.Location.Field,
		edge.Kind,
	)
}
