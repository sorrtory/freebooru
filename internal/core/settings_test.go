package core

import (
	"testing"

	"github.com/sorrtory/freebooru/internal/config"
)

func TestUpdateSettingsPreservesSensitivePathAndRejectsStaleRevision(t *testing.T) {
	app := newRealImportTestCore(t)
	before := app.config
	current := app.Settings()
	invalid := SettingsUpdate{
		ExpectedRevision: current.Revision, Language: current.Language,
		DefaultCollection: "missing", DefaultStorageName: current.DefaultStorageName,
		HTTPAddress: current.HTTPAddress, HTTPPort: current.HTTPPort,
		RemoveOnUpload: current.RemoveOnUpload,
	}
	if _, err := app.UpdateSettings(t.Context(), invalid); err == nil {
		t.Fatal("invalid UpdateSettings() error = nil")
	}
	rolledBack, err := config.LoadApp(app.paths.App)
	if err != nil || rolledBack.DefaultCollection != before.DefaultCollection {
		t.Fatalf("rolled back settings = %#v, %v", rolledBack, err)
	}
	updated, err := app.UpdateSettings(t.Context(), SettingsUpdate{
		ExpectedRevision: current.Revision, Language: "ru",
		DefaultCollection:  current.DefaultCollection,
		DefaultStorageName: current.DefaultStorageName,
		HTTPAddress:        "127.0.0.1",
		HTTPPort:           52801, RemoveOnUpload: true,
	})
	if err != nil {
		t.Fatalf("UpdateSettings() error = %v", err)
	}
	if updated.Language != "ru" || updated.HTTPAddress != "127.0.0.1" || updated.HTTPPort != 52801 || !updated.RemoveOnUpload {
		t.Fatalf("updated = %#v", updated)
	}
	stored, err := config.LoadApp(app.paths.App)
	if err != nil {
		t.Fatal(err)
	}
	if stored.DefaultStoragePath != before.DefaultStoragePath {
		t.Fatalf("default storage path = %q", stored.DefaultStoragePath)
	}
	if _, err := app.UpdateSettings(t.Context(), SettingsUpdate{ExpectedRevision: current.Revision}); err == nil {
		t.Fatal("stale UpdateSettings() error = nil")
	}
}
