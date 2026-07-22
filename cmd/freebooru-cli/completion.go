package main

import (
	"strings"

	"github.com/sorrtory/freebooru/internal/config"
	"github.com/sorrtory/freebooru/internal/core"
	"github.com/spf13/cobra"
)

func completeImportAssignment(
	root *rootOptions,
	collectionName string,
) cobra.CompletionFunc {
	return func(cmd *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		app, err := loadCore(cmd.Context(), root)
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}
		return completeAssignment(cmd, app, collectionName, "", toComplete)
	}
}

func completeFileAssignment(
	root *rootOptions,
	collectionName string,
	sha256 string,
) cobra.CompletionFunc {
	return func(cmd *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		app, err := loadCore(cmd.Context(), root)
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}
		return completeAssignment(cmd, app, collectionName, sha256, toComplete)
	}
}

func completeAssignment(
	cmd *cobra.Command,
	app *core.Core,
	collectionName string,
	sha256 string,
	toComplete string,
) ([]string, cobra.ShellCompDirective) {
	fields, err := app.ImportFields(collectionName)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	name, valuePrefix, hasValue := strings.Cut(toComplete, ":")
	if !hasValue {
		var suggestions []string
		for _, field := range fields {
			if !strings.HasPrefix(strings.ToLower(field.Name), strings.ToLower(name)) {
				continue
			}
			suggestion := field.Name
			if field.Type != config.TagTypeBool {
				suggestion += ":"
			}
			suggestions = append(suggestions, suggestion)
		}
		return suggestions, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveNoSpace
	}
	for _, field := range fields {
		if !strings.EqualFold(field.Name, name) {
			continue
		}
		values := field.Values
		if sha256 != "" && field.Name != "storage" && len(values) > 0 {
			available, err := app.AllowedValues(cmd.Context(), collectionName, sha256, field.Name)
			if err != nil {
				return nil, cobra.ShellCompDirectiveError
			}
			values = values[:0]
			for _, candidate := range available {
				if candidate.Allowed {
					values = append(values, candidate.Value.Val)
				}
			}
		}
		var suggestions []string
		for _, value := range values {
			if strings.HasPrefix(strings.ToLower(value), strings.ToLower(valuePrefix)) {
				suggestions = append(suggestions, field.Name+":"+value)
			}
		}
		return suggestions, cobra.ShellCompDirectiveNoFileComp
	}
	return nil, cobra.ShellCompDirectiveNoFileComp
}
