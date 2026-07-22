package main

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/sorrtory/freebooru/internal/config"
	"github.com/sorrtory/freebooru/internal/core"
	"github.com/spf13/cobra"
)

type importOptions struct {
	tags        []string
	interactive bool
}

func newImportCommand(root *rootOptions, collectionName string) *cobra.Command {
	options := &importOptions{}
	command := &cobra.Command{
		Use:   "import <file>",
		Short: "Import one regular file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := loadCore(cmd.Context(), root)
			if err != nil {
				return err
			}
			tags := append([]string(nil), options.tags...)
			if options.interactive {
				tags, err = promptImportAssignments(cmd, app, collectionName, tags)
				if err != nil {
					return err
				}
			}
			assignments, err := app.ParseTagAssignments(collectionName, tags)
			if err != nil {
				return err
			}
			result, err := app.Import(cmd.Context(), core.ImportRequest{
				Collection: collectionName,
				SourcePath: args[0],
				Tags:       assignments,
			})
			if err != nil {
				return err
			}
			if _, err := fmt.Fprintln(cmd.OutOrStdout(), result.SHA256); err != nil {
				return fmt.Errorf("write imported SHA-256: %w", err)
			}
			return nil
		},
	}
	command.Flags().StringArrayVar(
		&options.tags,
		"tag",
		nil,
		"assign a tag as name[:value] (repeatable)",
	)
	command.Flags().BoolVar(
		&options.interactive,
		"interactive",
		false,
		"prompt for missing and optional tags",
	)
	if err := command.RegisterFlagCompletionFunc(
		"tag",
		completeImportAssignment(root, collectionName),
	); err != nil {
		panic(err)
	}
	return command
}

func promptImportAssignments(
	cmd *cobra.Command,
	app *core.Core,
	collectionName string,
	assignments []string,
) ([]string, error) {
	fields, err := app.ImportFields(collectionName)
	if err != nil {
		return nil, err
	}
	reader := bufio.NewScanner(cmd.InOrStdin())
	for _, required := range []bool{true, false} {
		for _, field := range fields {
			needsInput := field.Required && field.Name != "storage"
			if needsInput != required || hasAssignment(assignments, field.Name) {
				continue
			}
			// Core supplies required booleans and storage references directly
			// from the collection contract; only valued requirements are missing.
			if required && field.Type == config.TagTypeBool {
				continue
			}
			prompt := "optional"
			if required {
				prompt = "required"
			}
			if _, err := fmt.Fprintf(
				cmd.ErrOrStderr(),
				"%s %s (%s%s): ",
				prompt,
				field.Name,
				field.Type,
				formatAllowedValues(field.Values),
			); err != nil {
				return nil, fmt.Errorf("write import prompt: %w", err)
			}
			if !reader.Scan() {
				if err := reader.Err(); err != nil {
					return nil, fmt.Errorf("read import prompt: %w", err)
				}
				if required {
					return nil, fmt.Errorf("required tag %q was not provided", field.Name)
				}
				return assignments, nil
			}
			entered := strings.TrimSpace(reader.Text())
			found, err := promptedAssignments(field, entered, required)
			if err != nil {
				return nil, err
			}
			assignments = append(assignments, found...)
		}
	}
	return assignments, nil
}

func hasAssignment(assignments []string, name string) bool {
	for _, assignment := range assignments {
		current, _, _ := strings.Cut(assignment, ":")
		if strings.EqualFold(current, name) {
			return true
		}
	}
	return false
}

func formatAllowedValues(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return "; " + strings.Join(values, "|")
}

func promptedAssignments(field core.ImportField, entered string, required bool) ([]string, error) {
	if entered == "" {
		if required {
			return nil, fmt.Errorf("required tag %q needs a value", field.Name)
		}
		return nil, nil
	}
	if field.Type == config.TagTypeBool {
		switch strings.ToLower(entered) {
		case "true", "yes", "y":
			return []string{field.Name}, nil
		case "false", "no", "n":
			return nil, nil
		default:
			return nil, fmt.Errorf("boolean tag %q expects true, false, yes, or no", field.Name)
		}
	}
	if field.Type != config.TagTypeMultivalue {
		return []string{field.Name + ":" + entered}, nil
	}
	var assignments []string
	for _, value := range strings.Split(entered, ",") {
		value = strings.TrimSpace(value)
		if value == "" {
			return nil, fmt.Errorf("multivalue tag %q contains an empty value", field.Name)
		}
		assignments = append(assignments, field.Name+":"+value)
	}
	return assignments, nil
}
