package config

// CheckDomain validates and cross-checks storage, tag, and collection configuration.
func CheckDomain(paths Paths, app AppConfig) Diagnostics {
	_, diagnostics := LoadCatalog(paths, app)
	return diagnostics
}
