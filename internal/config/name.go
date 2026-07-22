package config

import (
	"fmt"
	"strings"
	"unicode"
)

func verifyComment(comment string) error {
	if comment != "" && strings.TrimSpace(comment) == "" {
		return fmt.Errorf("comment must not be blank")
	}
	return nil
}

func verifyName(field, value string) error {
	if value == "" {
		return fmt.Errorf("%s is required", field)
	}
	for _, r := range value {
		if r != '_' && !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return fmt.Errorf("%s contains invalid character %q", field, r)
		}
	}
	return nil
}
