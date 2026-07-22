package core

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/sorrtory/freebooru/internal/collection"
	"github.com/sorrtory/freebooru/internal/config"
	"github.com/sorrtory/freebooru/internal/evaluator"
)

func validateCollectionRecords(
	ctx context.Context,
	database CollectionDatabase,
	collectionName string,
	catalog *config.Catalog,
	graph *config.Graph,
) error {
	references, ok := catalog.CollectionReferences(collectionName)
	if !ok {
		return fmt.Errorf("collection references are unavailable")
	}
	checker, err := evaluator.New(catalog, graph)
	if err != nil {
		return fmt.Errorf("create persisted state evaluator: %w", err)
	}
	availability := newCollectionAvailability(references)
	var invalid []error
	err = database.ForEachFile(ctx, func(file collection.FileRecord) error {
		if err := validateCollectionRecord(file, catalog, checker, availability); err != nil {
			invalid = append(invalid, fmt.Errorf("file %s: %w", file.SHA256, err))
		}
		return nil
	})
	if err != nil {
		return err
	}
	return errors.Join(invalid...)
}

type collectionAvailability struct {
	importedTags     map[string]struct{}
	importedStorages map[string]struct{}
	required         []config.TagReference
}

func newCollectionAvailability(references config.ResolvedReferences) collectionAvailability {
	availability := collectionAvailability{
		importedTags:     make(map[string]struct{}),
		importedStorages: make(map[string]struct{}),
		required:         references.Required,
	}
	for _, reference := range references.Imported {
		if reference.Tag != "" {
			availability.importedTags[normalizeStateName(reference.Tag)] = struct{}{}
		}
		if reference.Storage != "" {
			availability.importedStorages[normalizeStateName(reference.Storage)] = struct{}{}
		}
	}
	return availability
}

func validateCollectionRecord(
	file collection.FileRecord,
	catalog *config.Catalog,
	checker *evaluator.Evaluator,
	availability collectionAvailability,
) error {
	values, err := persistedValues(file, catalog, availability)
	if err != nil {
		return err
	}
	state, err := evaluator.NewFileState(catalog, values)
	if err != nil {
		return fmt.Errorf("reconstruct tag state: %w", err)
	}
	if err := verifyRequiredState(state, file.Storages, availability.required); err != nil {
		return err
	}
	result := checker.ValidateFile(state)
	if !result.Valid() {
		return fmt.Errorf(
			"relationship state has %d missing demands and %d active conflicts",
			len(result.MissingDemands),
			len(result.ActiveConflicts),
		)
	}
	return nil
}

func persistedValues(
	file collection.FileRecord,
	catalog *config.Catalog,
	availability collectionAvailability,
) (map[string]any, error) {
	values := make(map[string]any, len(file.Tags)+1)
	for _, persisted := range file.Tags {
		key := normalizeStateName(persisted.Name)
		if key == "storage" {
			return nil, fmt.Errorf("built-in storage tag is duplicated in file tags")
		}
		if _, ok := availability.importedTags[key]; !ok {
			return nil, fmt.Errorf("tag %q is not imported by the collection", persisted.Name)
		}
		tag, _, ok := catalog.Tag(persisted.Name)
		if !ok {
			return nil, fmt.Errorf("tag %q no longer exists", persisted.Name)
		}
		if string(tag.Type) != persisted.Type {
			return nil, fmt.Errorf(
				"tag %q persisted type %q does not match configured type %q",
				persisted.Name,
				persisted.Type,
				tag.Type,
			)
		}
		value, err := persistedTagValue(persisted)
		if err != nil {
			return nil, err
		}
		values[persisted.Name] = value
	}
	for _, storage := range file.Storages {
		if _, ok := availability.importedStorages[normalizeStateName(storage)]; !ok {
			return nil, fmt.Errorf("storage %q is not imported by the collection", storage)
		}
	}
	if len(file.Storages) > 0 {
		values["storage"] = append([]string(nil), file.Storages...)
	}
	return values, nil
}

func persistedTagValue(tag collection.TagRecord) (any, error) {
	switch config.TagType(tag.Type) {
	case config.TagTypeBool:
		if tag.TextValue == nil && tag.IntegerValue == nil && len(tag.Values) == 0 {
			return true, nil
		}
	case config.TagTypeText, config.TagTypeDate, config.TagTypeDatetime, config.TagTypeValue:
		if tag.TextValue != nil && tag.IntegerValue == nil && len(tag.Values) == 0 {
			return *tag.TextValue, nil
		}
	case config.TagTypeInt:
		if tag.TextValue == nil && tag.IntegerValue != nil && len(tag.Values) == 0 {
			return *tag.IntegerValue, nil
		}
	case config.TagTypeMultivalue:
		if tag.TextValue == nil && tag.IntegerValue == nil {
			return append([]string(nil), tag.Values...), nil
		}
	}
	return nil, fmt.Errorf("tag %q has an invalid persisted %q value", tag.Name, tag.Type)
}

func verifyRequiredState(
	state evaluator.FileState,
	storages []string,
	required []config.TagReference,
) error {
	assignedStorages := make(map[string]struct{}, len(storages))
	for _, storage := range storages {
		assignedStorages[normalizeStateName(storage)] = struct{}{}
	}
	for _, reference := range required {
		if reference.Tag != "" {
			if _, ok := state.Value(reference.Tag); !ok {
				return fmt.Errorf("required tag %q is missing", reference.Tag)
			}
		}
		if reference.Storage != "" {
			if _, ok := assignedStorages[normalizeStateName(reference.Storage)]; !ok {
				return fmt.Errorf("required storage %q is missing", reference.Storage)
			}
		}
	}
	return nil
}

func normalizeStateName(name string) string {
	return strings.ToLower(name)
}
