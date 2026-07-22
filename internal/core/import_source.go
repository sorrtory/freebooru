package core

import (
	"context"
	"fmt"
	"os"

	contentstorage "github.com/sorrtory/freebooru/internal/storage"
)

func removeImportedSource(
	ctx context.Context,
	sourcePath string,
	imported contentstorage.StoredFile,
	copies []storedCopy,
) error {
	for _, stored := range copies {
		same, err := sameFile(sourcePath, stored.file.ContentPath)
		if err != nil {
			return err
		}
		if same {
			return nil
		}
	}
	inspected, err := copies[0].backend.Inspect(ctx, sourcePath)
	if err != nil {
		return fmt.Errorf("verify source before removal: %w", err)
	}
	if inspected.SHA256 != imported.SHA256 || inspected.SizeBytes != imported.SizeBytes {
		return fmt.Errorf("source changed after import; source was not removed")
	}
	if err := os.Remove(sourcePath); err != nil {
		return fmt.Errorf("remove imported source %q: %w", sourcePath, err)
	}
	return nil
}

func sameFile(left, right string) (bool, error) {
	leftInfo, err := os.Stat(left)
	if err != nil {
		return false, fmt.Errorf("stat import source %q: %w", left, err)
	}
	rightInfo, err := os.Stat(right)
	if err != nil {
		return false, fmt.Errorf("stat stored content %q: %w", right, err)
	}
	return os.SameFile(leftInfo, rightInfo), nil
}
