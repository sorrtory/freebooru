package config

// BuildValidatedGraph runs the complete graph pipeline and returns the compiled
// query snapshot together with every structural and predicate diagnostic.
func BuildValidatedGraph(catalog *Catalog) (*Graph, Diagnostics) {
	raw := BuildGraph(catalog)
	diagnostics := CheckGraph(catalog, raw)
	compiled, compileDiagnostics := CompileGraph(catalog, raw)
	diagnostics = append(diagnostics, compileDiagnostics...)
	diagnostics = append(diagnostics, checkGraphContradictions(compiled)...)
	return compiled, diagnostics
}
