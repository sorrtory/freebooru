package main

import (
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/sorrtory/freebooru/internal/collection"
	"github.com/sorrtory/freebooru/internal/core"
	"github.com/spf13/cobra"
)

func newTagCommand(options *rootOptions, collectionName string) *cobra.Command {
	return &cobra.Command{
		Use:   "tag list | edit <name> | <sha256> <operation>",
		Short: "List or edit tag definitions, or mutate file tags",
		Long: "List tags imported by the selected collection, edit a tag's source " +
			"YAML, or read and mutate the tags assigned to one indexed file.",
		Example: "  freebooru-cli tag list\n" +
			"  freebooru-cli tag edit rating\n" +
			"  freebooru-cli tag <sha256> get",
		Args:               tagCommandArgs,
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if isHelpArgument(args[0]) {
				return cmd.Help()
			}
			if args[0] == "list" || args[0] == "edit" {
				return executeNestedCommand(
					cmd,
					newTagCatalogCommand(options, collectionName),
					args,
				)
			}
			sha256, err := canonicalSHA256(args[0])
			if err != nil {
				return err
			}
			scope := newTagScopeCommand(options, collectionName, sha256)
			return executeNestedCommand(cmd, scope, args[1:])
		},
	}
}

func tagCommandArgs(cmd *cobra.Command, args []string) error {
	if len(args) == 1 && (isHelpArgument(args[0]) || args[0] == "list") {
		return nil
	}
	return cobra.MinimumNArgs(2)(cmd, args)
}

func newTagScopeCommand(options *rootOptions, collectionName, sha256 string) *cobra.Command {
	command := &cobra.Command{Use: sha256, SilenceUsage: true, SilenceErrors: true}
	command.AddCommand(
		newTagAssignmentCommand(options, collectionName, sha256, "add"),
		newTagAssignmentCommand(options, collectionName, sha256, "set"),
		newTagRemoveCommand(options, collectionName, sha256),
		newTagGetCommand(options, collectionName, sha256),
	)
	return command
}

func newTagAssignmentCommand(
	options *rootOptions,
	collectionName string,
	sha256 string,
	operation string,
) *cobra.Command {
	return &cobra.Command{
		Use:   operation + " <name[:value]>",
		Short: operation + " one typed tag assignment",
		Args:  cobra.ExactArgs(1),
		ValidArgsFunction: completeFileAssignment(
			options,
			collectionName,
			sha256,
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := loadCore(cmd.Context(), options)
			if err != nil {
				return err
			}
			assignments, err := app.ParseTagAssignments(collectionName, args)
			if err != nil {
				return err
			}
			name, value, err := onlyAssignment(assignments)
			if err != nil {
				return err
			}
			if strings.EqualFold(name, "storage") {
				if operation == "set" {
					return fmt.Errorf("storage values support add and remove, not set")
				}
				_, err = app.AddStorage(cmd.Context(), core.StorageMutationRequest{
					Collection: collectionName,
					SHA256:     sha256,
					Storage:    value.([]string)[0],
				})
			} else {
				request := core.TagMutationRequest{
					Collection: collectionName,
					SHA256:     sha256,
					Tag:        name,
					Value:      value,
				}
				if operation == "add" {
					_, err = app.AddTag(cmd.Context(), request)
				} else {
					_, err = app.SetTag(cmd.Context(), request)
				}
			}
			if err != nil {
				return err
			}
			return writeSHA256(cmd, sha256)
		},
	}
}

func newTagRemoveCommand(options *rootOptions, collectionName, sha256 string) *cobra.Command {
	return &cobra.Command{
		Use:   "remove <name[:value]>",
		Short: "Remove one tag or storage assignment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := loadCore(cmd.Context(), options)
			if err != nil {
				return err
			}
			name, value, hasValue := strings.Cut(args[0], ":")
			if strings.EqualFold(name, "storage") {
				if !hasValue || value == "" {
					return fmt.Errorf("storage removal requires storage:<name>")
				}
				_, err = app.RemoveStorage(cmd.Context(), core.StorageMutationRequest{
					Collection: collectionName,
					SHA256:     sha256,
					Storage:    value,
				})
			} else {
				if hasValue {
					return fmt.Errorf("tag removal accepts a tag name without a value")
				}
				_, err = app.RemoveTag(cmd.Context(), core.TagRemovalRequest{
					Collection: collectionName,
					SHA256:     sha256,
					Tag:        name,
				})
			}
			if err != nil {
				return err
			}
			return writeSHA256(cmd, sha256)
		},
	}
}

func newTagGetCommand(options *rootOptions, collectionName, sha256 string) *cobra.Command {
	return &cobra.Command{
		Use:   "get",
		Short: "Print every assigned tag and storage",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			app, err := loadCore(cmd.Context(), options)
			if err != nil {
				return err
			}
			record, err := app.GetFile(cmd.Context(), collectionName, sha256)
			if err != nil {
				return err
			}
			for _, assignment := range formatFileAssignments(record) {
				if _, err := fmt.Fprintln(cmd.OutOrStdout(), assignment); err != nil {
					return fmt.Errorf("write tag assignment: %w", err)
				}
			}
			return nil
		},
	}
}

func canonicalSHA256(value string) (string, error) {
	if len(value) != sha256HexLength {
		return "", fmt.Errorf("SHA-256 must contain exactly %d hexadecimal characters", sha256HexLength)
	}
	if _, err := hex.DecodeString(value); err != nil {
		return "", fmt.Errorf("SHA-256 must be hexadecimal: %w", err)
	}
	return strings.ToLower(value), nil
}

const sha256HexLength = 64

func onlyAssignment(assignments map[string]any) (string, any, error) {
	for name, value := range assignments {
		return name, value, nil
	}
	return "", nil, fmt.Errorf("tag assignment is required")
}

func writeSHA256(cmd *cobra.Command, sha256 string) error {
	if _, err := fmt.Fprintln(cmd.OutOrStdout(), sha256); err != nil {
		return fmt.Errorf("write SHA-256: %w", err)
	}
	return nil
}

func formatFileAssignments(record collection.FileRecord) []string {
	assignments := make([]string, 0, len(record.Tags)+len(record.Storages))
	for _, tag := range record.Tags {
		switch {
		case tag.IntegerValue != nil:
			assignments = append(assignments, tag.Name+":"+strconv.FormatInt(*tag.IntegerValue, 10))
		case tag.TextValue != nil:
			assignments = append(assignments, tag.Name+":"+*tag.TextValue)
		case len(tag.Values) > 0:
			for _, value := range tag.Values {
				assignments = append(assignments, tag.Name+":"+value)
			}
		default:
			assignments = append(assignments, tag.Name)
		}
	}
	for _, storage := range record.Storages {
		assignments = append(assignments, "storage:"+storage)
	}
	sort.Slice(assignments, func(left, right int) bool {
		return strings.ToLower(assignments[left]) < strings.ToLower(assignments[right])
	})
	return assignments
}
