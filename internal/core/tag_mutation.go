package core

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/sorrtory/freebooru/internal/collection"
	"github.com/sorrtory/freebooru/internal/config"
	"github.com/sorrtory/freebooru/internal/evaluator"
)

// TagMutationRequest identifies one complete tag assignment replacement.
// An empty Collection selects the configured default collection.
type TagMutationRequest struct {
	Collection string
	SHA256     string
	Tag        string
	Value      any
}

// TagRemovalRequest identifies one complete tag assignment to remove.
type TagRemovalRequest struct {
	Collection string
	SHA256     string
	Tag        string
}

// TagMutationResult reports persistence and relationship effects.
type TagMutationResult struct {
	Changed    bool
	Evaluation evaluator.Evaluation
}

type tagMutationKind uint8

const (
	tagMutationAdd tagMutationKind = iota
	tagMutationSet
	tagMutationRemove
)

// AddTag validates a proposed addition before persistence. Scalar additions
// cannot replace a different existing value; multivalue additions are merged.
func (c *Core) AddTag(
	ctx context.Context,
	request TagMutationRequest,
) (TagMutationResult, error) {
	collectionName, references, availability, err := c.tagMutationCollection(request.Collection)
	if err != nil {
		return TagMutationResult{}, err
	}
	assignment, err := c.prepareTagAssignment(request.Tag, request.Value, availability)
	if err != nil {
		return TagMutationResult{}, err
	}
	kind := tagMutationAdd
	if !assignment.present {
		kind = tagMutationRemove
	}
	return c.applyTagMutation(
		ctx,
		collectionName,
		request.SHA256,
		references,
		availability,
		assignment,
		kind,
	)
}

// SetTag validates a complete proposed file state before replacing one tag.
// Boolean false removes the tag because false represents absence.
func (c *Core) SetTag(
	ctx context.Context,
	request TagMutationRequest,
) (TagMutationResult, error) {
	collectionName, references, availability, err := c.tagMutationCollection(request.Collection)
	if err != nil {
		return TagMutationResult{}, err
	}
	assignment, err := c.prepareTagAssignment(request.Tag, request.Value, availability)
	if err != nil {
		return TagMutationResult{}, err
	}
	kind := tagMutationSet
	if !assignment.present {
		kind = tagMutationRemove
	}
	return c.applyTagMutation(
		ctx,
		collectionName,
		request.SHA256,
		references,
		availability,
		assignment,
		kind,
	)
}

// RemoveTag validates the state without an assignment before removing it.
func (c *Core) RemoveTag(
	ctx context.Context,
	request TagRemovalRequest,
) (TagMutationResult, error) {
	collectionName, references, availability, err := c.tagMutationCollection(request.Collection)
	if err != nil {
		return TagMutationResult{}, err
	}
	tag, err := c.resolveMutableTag(request.Tag, availability)
	if err != nil {
		return TagMutationResult{}, err
	}
	return c.applyTagMutation(
		ctx,
		collectionName,
		request.SHA256,
		references,
		availability,
		preparedTagAssignment{tag: tag},
		tagMutationRemove,
	)
}

func (c *Core) tagMutationCollection(
	requested string,
) (string, config.ResolvedReferences, collectionAvailability, error) {
	if c.catalog == nil || c.graph == nil {
		return "", config.ResolvedReferences{}, collectionAvailability{},
			fmt.Errorf("configuration has not been checked")
	}
	if requested == "" {
		requested = c.config.DefaultCollection
	}
	references, ok := c.catalog.CollectionReferences(requested)
	if !ok {
		return "", config.ResolvedReferences{}, collectionAvailability{}, fmt.Errorf(
			"collection %q is missing or invalid",
			requested,
		)
	}
	return requested, references, newCollectionAvailability(references), nil
}

func (c *Core) applyTagMutation(
	ctx context.Context,
	collectionName string,
	sha256 string,
	references config.ResolvedReferences,
	availability collectionAvailability,
	assignment preparedTagAssignment,
	kind tagMutationKind,
) (result TagMutationResult, err error) {
	database, err := c.OpenCollection(ctx, collectionName)
	if err != nil {
		return TagMutationResult{}, err
	}
	defer func() {
		err = errors.Join(err, closeNamedCollection(database, collectionName))
	}()
	record, err := database.File(ctx, sha256)
	if err != nil {
		return TagMutationResult{}, fmt.Errorf("load file %q: %w", sha256, err)
	}
	values, err := persistedValues(record, c.catalog, availability)
	if err != nil {
		return TagMutationResult{}, fmt.Errorf("reconstruct file %q: %w", sha256, err)
	}
	if err := proposeTagMutation(values, assignment, kind); err != nil {
		return TagMutationResult{}, err
	}
	result.Evaluation, err = c.validateTagMutation(values, record.Storages, references.Required)
	if err != nil {
		return result, err
	}

	var change collection.TagChange
	switch kind {
	case tagMutationAdd:
		change, err = database.AddTag(ctx, sha256, assignment.record)
	case tagMutationSet:
		change, err = database.SetTag(ctx, sha256, assignment.record)
	case tagMutationRemove:
		change, err = database.RemoveTag(ctx, sha256, assignment.tag.Name)
	default:
		return result, fmt.Errorf("unsupported tag mutation")
	}
	if err != nil {
		return result, fmt.Errorf("persist tag %q for file %q: %w", assignment.tag.Name, sha256, err)
	}
	result.Changed = change.Changed
	return result, nil
}

func proposeTagMutation(
	values map[string]any,
	assignment preparedTagAssignment,
	kind tagMutationKind,
) error {
	existing, assigned := assignmentValue(values, assignment.tag.Name)
	if kind == tagMutationAdd && assigned {
		if assignment.tag.Type == config.TagTypeMultivalue {
			assignment.value = mergeTagValues(existing.([]string), assignment.value.([]string))
		} else if !reflect.DeepEqual(existing, assignment.value) {
			return fmt.Errorf("%w: %s", collection.ErrTagAlreadyAssigned, assignment.tag.Name)
		}
	}
	deleteAssignment(values, assignment.tag.Name)
	if kind != tagMutationRemove {
		values[assignment.tag.Name] = assignment.value
	}
	return nil
}

func assignmentValue(values map[string]any, name string) (any, bool) {
	for current, value := range values {
		if normalizeStateName(current) == normalizeStateName(name) {
			return value, true
		}
	}
	return nil, false
}

func mergeTagValues(existing, added []string) []string {
	merged := append([]string(nil), existing...)
	for _, candidate := range added {
		found := false
		for _, current := range merged {
			if normalizeStateName(current) == normalizeStateName(candidate) {
				found = true
				break
			}
		}
		if !found {
			merged = append(merged, candidate)
		}
	}
	return merged
}

func (c *Core) validateTagMutation(
	values map[string]any,
	storages []string,
	required []config.TagReference,
) (evaluator.Evaluation, error) {
	state, err := evaluator.NewFileState(c.catalog, values)
	if err != nil {
		return evaluator.Evaluation{}, fmt.Errorf("validate proposed tag state: %w", err)
	}
	if err := verifyRequiredState(state, storages, required); err != nil {
		return evaluator.Evaluation{}, err
	}
	checker, err := evaluator.New(c.catalog, c.graph)
	if err != nil {
		return evaluator.Evaluation{}, fmt.Errorf("create tag evaluator: %w", err)
	}
	evaluation := checker.ValidateFile(state)
	if !evaluation.Valid() {
		return evaluation, fmt.Errorf(
			"tag mutation has %d missing demands and %d active conflicts",
			len(evaluation.MissingDemands),
			len(evaluation.ActiveConflicts),
		)
	}
	return evaluation, nil
}

func deleteAssignment(values map[string]any, name string) {
	for current := range values {
		if normalizeStateName(current) == normalizeStateName(name) {
			delete(values, current)
		}
	}
}
