package core

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/sorrtory/freebooru/internal/collection"
)

func TestOpenCollectionBindsValidatedSessionAndRejectsSecondOpen(t *testing.T) {
	paths, _ := provisionTestConfig(t)
	database := &fakeDatabase{}
	openCalls := 0
	app, err := New(testLogger(), paths, func(context.Context, string) (CollectionDatabase, error) {
		openCalls++
		return database, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := app.LoadConfig(t.Context()); err != nil {
		t.Fatal(err)
	}
	if diagnostics := app.CheckConfig(t.Context()); diagnostics.HasErrors() {
		t.Fatalf("CheckConfig() diagnostics = %#v", diagnostics)
	}
	if err := app.OpenCollection(t.Context(), "MAIN"); err != nil {
		t.Fatalf("OpenCollection() error = %v", err)
	}
	if app.session == nil || app.session.name != "main" || app.session.database != database ||
		app.session.evaluator == nil || len(app.session.storages) != 1 {
		t.Fatalf("session = %#v", app.session)
	}
	if err := app.OpenCollection(t.Context(), "main"); err == nil ||
		!strings.Contains(err.Error(), "already open") {
		t.Fatalf("second OpenCollection() error = %v", err)
	}
	if openCalls != 1 {
		t.Fatalf("opener calls = %d, want 1", openCalls)
	}
	if err := app.CloseCollection(); err != nil {
		t.Fatalf("CloseCollection() error = %v", err)
	}
	if err := app.CloseCollection(); err != nil {
		t.Fatalf("repeated CloseCollection() error = %v", err)
	}
	if database.closeCalls != 1 {
		t.Fatalf("database close calls = %d, want 1", database.closeCalls)
	}
}

func TestCloseCollectionDetachesSessionAfterCloseFailure(t *testing.T) {
	paths, _ := provisionTestConfig(t)
	database := &fakeDatabase{closeErr: errors.New("close failure")}
	app, err := New(testLogger(), paths, func(context.Context, string) (CollectionDatabase, error) {
		return database, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := app.LoadConfig(t.Context()); err != nil {
		t.Fatal(err)
	}
	if diagnostics := app.CheckConfig(t.Context()); diagnostics.HasErrors() {
		t.Fatalf("CheckConfig() diagnostics = %#v", diagnostics)
	}
	if err := app.OpenCollection(t.Context(), "main"); err != nil {
		t.Fatal(err)
	}
	if err := app.CloseCollection(); err == nil || !strings.Contains(err.Error(), "close failure") {
		t.Fatalf("CloseCollection() error = %v", err)
	}
	if _, ok := app.OpenCollectionName(); ok {
		t.Fatal("failed close left a published session")
	}
	if err := app.CloseCollection(); err != nil {
		t.Fatalf("repeated CloseCollection() error = %v", err)
	}
	if database.closeCalls != 1 {
		t.Fatalf("database close calls = %d, want 1", database.closeCalls)
	}
}

func TestOpenCollectionPropagatesCancellationWithoutPublishingSession(t *testing.T) {
	paths, _ := provisionTestConfig(t)
	app, err := New(testLogger(), paths, func(ctx context.Context, _ string) (CollectionDatabase, error) {
		return nil, ctx.Err()
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := app.LoadConfig(t.Context()); err != nil {
		t.Fatal(err)
	}
	if diagnostics := app.CheckConfig(t.Context()); diagnostics.HasErrors() {
		t.Fatalf("CheckConfig() diagnostics = %#v", diagnostics)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if err := app.OpenCollection(ctx, "main"); !errors.Is(err, context.Canceled) {
		t.Fatalf("OpenCollection() error = %v, want context.Canceled", err)
	}
	if _, ok := app.OpenCollectionName(); ok {
		t.Fatal("canceled open published a session")
	}
}

func TestOpenCollectionFreezesConfigurationSnapshot(t *testing.T) {
	paths, _ := provisionTestConfig(t)
	database := &fakeDatabase{}
	app, err := New(testLogger(), paths, func(context.Context, string) (CollectionDatabase, error) {
		return database, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := app.LoadConfig(t.Context()); err != nil {
		t.Fatal(err)
	}
	if diagnostics := app.CheckConfig(t.Context()); diagnostics.HasErrors() {
		t.Fatalf("CheckConfig() diagnostics = %#v", diagnostics)
	}
	if err := app.OpenCollection(t.Context(), "main"); err != nil {
		t.Fatal(err)
	}
	if err := app.LoadConfig(t.Context()); err == nil || !strings.Contains(err.Error(), "before loading") {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	diagnostics := app.CheckConfig(t.Context())
	if !diagnostics.HasErrors() || diagnostics[0].Code != "application.collection_open" {
		t.Fatalf("CheckConfig() diagnostics = %#v", diagnostics)
	}
	if err := app.Init(t.Context()); err == nil || !strings.Contains(err.Error(), "before initializing") {
		t.Fatalf("Init() error = %v", err)
	}
	if err := app.CloseCollection(); err != nil {
		t.Fatal(err)
	}
}

func TestOperationsReusePersistentSessionAndRejectCollectionSwitch(t *testing.T) {
	paths, _ := provisionTestConfig(t)
	hash := strings.Repeat("a", 64)
	database := &fakeDatabase{files: []collection.FileRecord{{
		SHA256: hash, Storages: []string{"default"},
	}}}
	openCalls := 0
	app, err := New(testLogger(), paths, func(context.Context, string) (CollectionDatabase, error) {
		openCalls++
		return database, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := app.LoadConfig(t.Context()); err != nil {
		t.Fatal(err)
	}
	if diagnostics := app.CheckConfig(t.Context()); diagnostics.HasErrors() {
		t.Fatalf("CheckConfig() diagnostics = %#v", diagnostics)
	}
	if err := app.OpenCollection(t.Context(), "main"); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if _, err := app.GetFile(t.Context(), "main", hash); err != nil {
			t.Fatalf("GetFile() error = %v", err)
		}
	}
	if database.closed {
		t.Fatal("operation closed persistent session")
	}
	if _, err := app.GetFile(t.Context(), "archive", hash); err == nil ||
		!strings.Contains(err.Error(), "close it before using") {
		t.Fatalf("GetFile(archive) error = %v", err)
	}
	if openCalls != 1 {
		t.Fatalf("opener calls = %d, want 1", openCalls)
	}
	if err := app.CloseCollection(); err != nil {
		t.Fatal(err)
	}
	if database.closeCalls != 1 {
		t.Fatalf("database close calls = %d, want 1", database.closeCalls)
	}
}
