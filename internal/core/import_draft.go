package core

import (
	"context"
	"fmt"

	"github.com/sorrtory/freebooru/internal/evaluator"
)

// ImportDraftRequest identifies one side-effect-free import assignment draft.
type ImportDraftRequest struct {
	Collection  string
	Assignments map[string]any
}

// ImportDraft is one canonical snapshot of an incomplete or complete import form.
type ImportDraft struct {
	Collection      string
	Assignments     map[string]any
	MissingRequired []ImportField
	Evaluation      evaluator.Evaluation
	Complete        bool
}

// EvaluateImportDraft validates and evaluates import assignments without
// opening a collection database or touching source/storage files.
func (c *Core) EvaluateImportDraft(
	ctx context.Context,
	request ImportDraftRequest,
) (ImportDraft, error) {
	if err := ctx.Err(); err != nil {
		return ImportDraft{}, err
	}
	if request.Collection == "" {
		return ImportDraft{}, fmt.Errorf("collection is required")
	}

	c.sessionMu.Lock()
	defer c.sessionMu.Unlock()
	if c.catalog == nil || c.graph == nil {
		return ImportDraft{}, fmt.Errorf("configuration has not been checked")
	}
	collectionConfig, _, ok := c.catalog.Collection(request.Collection)
	if !ok {
		return ImportDraft{}, fmt.Errorf("collection %q is missing or invalid", request.Collection)
	}
	references, ok := c.catalog.CollectionReferences(collectionConfig.Name)
	if !ok {
		return ImportDraft{}, fmt.Errorf(
			"collection %q references are unavailable",
			collectionConfig.Name,
		)
	}

	availability := newCollectionAvailability(references)
	assignments, storageNames, err := c.collectImportAssignments(
		request.Assignments,
		availability,
	)
	if err != nil {
		return ImportDraft{}, err
	}
	missingNames, err := c.applyRequiredAssignments(
		assignments,
		storageNames,
		references.Required,
	)
	if err != nil {
		return ImportDraft{}, err
	}
	_, storages, err := c.resolveImportStorages(storageNames, availability)
	if err != nil {
		return ImportDraft{}, err
	}
	assignments["storage"] = storages
	missing, err := c.missingImportFields(collectionConfig.Name, missingNames)
	if err != nil {
		return ImportDraft{}, err
	}

	state, err := evaluator.NewFileState(c.catalog, assignments)
	if err != nil {
		return ImportDraft{}, fmt.Errorf("validate import draft tags: %w", err)
	}
	checker, err := evaluator.New(c.catalog, c.graph)
	if err != nil {
		return ImportDraft{}, fmt.Errorf("create import draft evaluator: %w", err)
	}
	evaluation := checker.ValidateFile(state)
	return ImportDraft{
		Collection:      collectionConfig.Name,
		Assignments:     assignments,
		MissingRequired: missing,
		Evaluation:      evaluation,
		Complete:        len(missing) == 0 && evaluation.Valid(),
	}, nil
}

func (c *Core) missingImportFields(
	collectionName string,
	names []string,
) ([]ImportField, error) {
	missingNames := make(map[string]struct{}, len(names))
	for _, name := range names {
		missingNames[normalizeStateName(name)] = struct{}{}
	}
	fields, err := c.importFields(collectionName)
	if err != nil {
		return nil, err
	}
	missing := make([]ImportField, 0)
	for _, field := range fields {
		if _, ok := missingNames[normalizeStateName(field.Name)]; ok {
			missing = append(missing, field)
		}
	}
	return missing, nil
}
