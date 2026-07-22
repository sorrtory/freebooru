package collection

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

var (
	// ErrFileNotFound means that a collection does not index the requested hash.
	ErrFileNotFound = errors.New("collection file not found")
	// ErrDuplicateFile means that a collection already indexes the content.
	ErrDuplicateFile = errors.New("collection file already exists")
	// ErrTagAlreadyAssigned means add cannot replace an existing scalar value.
	ErrTagAlreadyAssigned = errors.New("collection tag already assigned")
)

// FileRecord is one persisted collection file and all of its assignments.
type FileRecord struct {
	ID         int64
	SHA256     string
	SizeBytes  int64
	ImportedAt string
	CreatedAt  string
	UpdatedAt  string
	Sources    []SourceRecord
	Tags       []TagRecord
	Storages   []string
}

// SourceRecord describes one source path observed during import.
type SourceRecord struct {
	Path       string
	Filename   string
	ObservedAt string
}

// TagRecord preserves one typed tag assignment without coercing its value.
type TagRecord struct {
	Name         string
	Type         string
	TextValue    *string
	IntegerValue *int64
	Values       []string
}

// StorageChange describes the logical result of a storage mutation.
type StorageChange struct {
	Changed     bool
	FileDeleted bool
}

// TagChange describes whether a tag mutation changed persisted state.
type TagChange struct {
	Changed bool
}

// NewFile describes the immutable metadata recorded by an initial import.
type NewFile struct {
	SHA256         string
	SizeBytes      int64
	SourcePath     string
	SourceFilename string
	Tags           []TagRecord
	Storages       []string
}

// DuplicateFileError carries the existing record rejected by CreateFile.
type DuplicateFileError struct {
	Existing FileRecord
}

func (e *DuplicateFileError) Error() string {
	return fmt.Sprintf("%s: %s", ErrDuplicateFile, e.Existing.SHA256)
}

// Unwrap supports errors.Is(err, ErrDuplicateFile).
func (e *DuplicateFileError) Unwrap() error {
	return ErrDuplicateFile
}

// File loads one file and every persisted source, tag, and storage assignment.
func (d *Database) File(ctx context.Context, sha256 string) (FileRecord, error) {
	var file FileRecord
	if err := d.db.QueryRowContext(
		ctx,
		`SELECT file_id, sha256, size_bytes, imported_at, created_at, updated_at
		 FROM file WHERE sha256 = ?`,
		sha256,
	).Scan(
		&file.ID,
		&file.SHA256,
		&file.SizeBytes,
		&file.ImportedAt,
		&file.CreatedAt,
		&file.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return FileRecord{}, fmt.Errorf("%w: %s", ErrFileNotFound, sha256)
		}
		return FileRecord{}, fmt.Errorf("query file %q: %w", sha256, err)
	}
	if err := d.loadSources(ctx, &file); err != nil {
		return FileRecord{}, err
	}
	if err := d.loadTags(ctx, &file); err != nil {
		return FileRecord{}, err
	}
	if err := d.loadStorages(ctx, &file); err != nil {
		return FileRecord{}, err
	}
	return file, nil
}

func (d *Database) loadSources(ctx context.Context, file *FileRecord) (err error) {
	rows, err := d.db.QueryContext(
		ctx,
		`SELECT source_path, filename, observed_at FROM file_source
		 WHERE file_id = ? ORDER BY observed_at, file_source_id`,
		file.ID,
	)
	if err != nil {
		return fmt.Errorf("query sources for file %q: %w", file.SHA256, err)
	}
	defer func() {
		err = errors.Join(err, closeRows(rows, "file sources"))
	}()
	for rows.Next() {
		var source SourceRecord
		if err := rows.Scan(&source.Path, &source.Filename, &source.ObservedAt); err != nil {
			return fmt.Errorf("scan source for file %q: %w", file.SHA256, err)
		}
		file.Sources = append(file.Sources, source)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate sources for file %q: %w", file.SHA256, err)
	}
	return nil
}

func (d *Database) loadTags(ctx context.Context, file *FileRecord) (err error) {
	rows, err := d.db.QueryContext(
		ctx,
		`SELECT tag.file_tag_id, tag.tag_name, tag.tag_type,
		        tag.text_value, tag.integer_value, value.value
		 FROM file_tag AS tag
		 LEFT JOIN file_tag_value AS value ON value.file_tag_id = tag.file_tag_id
		 WHERE tag.file_id = ?
		 ORDER BY tag.tag_name, tag.file_tag_id, value.value`,
		file.ID,
	)
	if err != nil {
		return fmt.Errorf("query tags for file %q: %w", file.SHA256, err)
	}
	defer func() {
		err = errors.Join(err, closeRows(rows, "file tags"))
	}()
	var currentID int64 = -1
	for rows.Next() {
		var (
			id           int64
			name         string
			tagType      string
			textValue    sql.NullString
			integerValue sql.NullInt64
			value        sql.NullString
		)
		if err := rows.Scan(&id, &name, &tagType, &textValue, &integerValue, &value); err != nil {
			return fmt.Errorf("scan tag for file %q: %w", file.SHA256, err)
		}
		if id != currentID {
			file.Tags = append(file.Tags, TagRecord{
				Name:         name,
				Type:         tagType,
				TextValue:    nullableString(textValue),
				IntegerValue: nullableInt64(integerValue),
			})
			currentID = id
		}
		if value.Valid {
			last := len(file.Tags) - 1
			file.Tags[last].Values = append(file.Tags[last].Values, value.String)
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate tags for file %q: %w", file.SHA256, err)
	}
	return nil
}

func (d *Database) loadStorages(ctx context.Context, file *FileRecord) (err error) {
	rows, err := d.db.QueryContext(
		ctx,
		`SELECT storage_name FROM file_storage
		 WHERE file_id = ? ORDER BY storage_name`,
		file.ID,
	)
	if err != nil {
		return fmt.Errorf("query storages for file %q: %w", file.SHA256, err)
	}
	defer func() {
		err = errors.Join(err, closeRows(rows, "file storages"))
	}()
	for rows.Next() {
		var storage string
		if err := rows.Scan(&storage); err != nil {
			return fmt.Errorf("scan storage for file %q: %w", file.SHA256, err)
		}
		file.Storages = append(file.Storages, storage)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate storages for file %q: %w", file.SHA256, err)
	}
	return nil
}

func nullableString(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}

func nullableInt64(value sql.NullInt64) *int64 {
	if !value.Valid {
		return nil
	}
	return &value.Int64
}

func closeRows(rows *sql.Rows, subject string) error {
	if err := rows.Close(); err != nil {
		return fmt.Errorf("close %s rows: %w", subject, err)
	}
	return nil
}
