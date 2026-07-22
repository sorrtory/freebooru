package config

import "strings"

// Severity classifies the effect of a configuration diagnostic.
type Severity string

const (
	// SeverityError marks configuration that cannot be used.
	SeverityError Severity = "error"
	// SeverityWarning marks usable configuration that needs attention.
	SeverityWarning Severity = "warning"
)

// Diagnostic describes one configuration problem at a stable source location.
type Diagnostic struct {
	Severity Severity
	Code     string
	Message  string
	File     string
	Document int
	Field    string
}

// Diagnostics is a collection of configuration problems.
type Diagnostics []Diagnostic

// HasErrors reports whether any diagnostic prevents configuration use.
func (d Diagnostics) HasErrors() bool {
	for _, diagnostic := range d {
		if diagnostic.Severity == SeverityError {
			return true
		}
	}
	return false
}

func newDiagnostic(code, message, file string, document int) Diagnostic {
	return Diagnostic{
		Severity: SeverityError,
		Code:     code,
		Message:  message,
		File:     file,
		Document: document,
	}
}

func validationField(err error) string {
	field, _, found := strings.Cut(err.Error(), " ")
	if !found {
		return ""
	}
	return field
}
