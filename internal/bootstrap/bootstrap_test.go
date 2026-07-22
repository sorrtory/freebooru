package bootstrap

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestLoadConfigRequiresValidAppConfig(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	app, err := NewCore(testLogger())
	if err != nil {
		t.Fatalf("NewCore() error = %v", err)
	}
	if err := app.LoadConfig(context.Background()); err == nil {
		t.Fatal("LoadConfig() without freebooru.yaml succeeded")
	}
	configDir := filepath.Join(dir, "freebooru")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "freebooru.yaml"), []byte("http_port: nope\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := app.LoadConfig(context.Background()); err == nil {
		t.Fatal("LoadConfig() with invalid freebooru.yaml succeeded")
	}
}

func TestCoreForInitAndDomainCheck(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	app, err := NewCore(testLogger())
	if err != nil {
		t.Fatalf("NewCore() error = %v", err)
	}
	if err := app.InitConfig(context.Background()); err != nil {
		t.Fatalf("InitConfig() error = %v", err)
	}
	app, err = NewCore(testLogger())
	if err != nil {
		t.Fatalf("NewCore() after init error = %v", err)
	}
	if err := app.LoadConfig(context.Background()); err != nil {
		t.Fatalf("LoadConfig() after init error = %v", err)
	}
	if err := app.CheckConfig(context.Background()); err != nil {
		t.Fatalf("CheckConfig() error = %v", err)
	}
}
