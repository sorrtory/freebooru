package core

import (
	"context"
	"errors"
	"fmt"

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

// TagMutationResult reports persistence and relationship effects.
type TagMutationResult struct {
	Changed    bool
	Evaluation evaluator.Evaluation
}

// SetTag validates a complete proposed file state before replacing one tag.
// Boolean false removes the tag because false represents absence.
func (c *Core) SetTag(
	ctx context.Context,
	request TagMutationRequest,
) (result TagMutationResult, err error) {
	if c.catalog == nil || c.graph == nil {
		return TagMutationResult{}, fmt.Errorf("configuration has not been checked")
	}
	collectionName := request.Collection
	if collectionName == "" {
		collectionName = c.config.DefaultCollection
	}
	references, ok := c.catalog.CollectionReferences(collectionName)
	if !ok {
		return TagMutationResult{}, fmt.Errorf(
			"collection %q is missing or invalid",
			collectionName,
		)
	}
	availability := newCollectionAvailability(references)
	assignment, err := c.prepareTagAssignment(request.Tag, request.Value, availability)
	if err != nil {
		return TagMutationResult{}, err
	}

	database, err := c.OpenCollection(ctx, collectionName)
	if err != nil {
		return TagMutationResult{}, err
	}
	defer func() {
		err = errors.Join(err, closeNamedCollection(database, collectionName))
	}()
	record, err := database.File(ctx, request.SHA256)
	if err != nil {
		return TagMutationResult{}, fmt.Errorf("load file %q: %w", request.SHA256, err)
	}
	values, err := persistedValues(record, c.catalog, availability)
	if err != nil {
		return TagMutationResult{}, fmt.Errorf("reconstruct file %q: %w", request.SHA256, err)
	}
	deleteAssignment(values, assignment.tag.Name)
	if assignment.present {
		values[assignment.tag.Name] = assignment.value
	}
	result.Evaluation, err = c.validateTagMutation(values, record.Storages, references.Required)
	if err != nil {
		return result, err
	}

	var change collection.TagChange
	if assignment.present {
		change, err = database.SetTag(ctx, request.SHA256, assignment.record)
	} else {
		change, err = database.RemoveTag(ctx, request.SHA256, assignment.tag.Name)
	}
	if err != nil {
		return result, fmt.Errorf("persist tag %q for file %q: %w", assignment.tag.Name, request.SHA256, err)
	}
	result.Changed = change.Changed
	return result, nil
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
