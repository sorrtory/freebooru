package collection

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// AddTag adds a new assignment or merges values into a multivalue assignment.
// An identical scalar assignment is a no-op; replacing it requires SetTag.
func (d *Database) AddTag(
	ctx context.Context,
	sha256 string,
	tag TagRecord,
) (TagChange, error) {
	if err := validateTagRecord(tag); err != nil {
		return TagChange{}, err
	}
	var change TagChange
	err := d.writeTransaction(ctx, "add file tag", func(tx *sql.Tx) error {
		fileID, err := queryFileID(ctx, tx, sha256)
		if err != nil {
			return err
		}
		result, err := tx.ExecContext(
			ctx,
			`INSERT INTO file_tag
			    (file_id, tag_name, tag_type, text_value, integer_value)
			 VALUES (?, ?, ?, ?, ?)
			 ON CONFLICT(file_id, tag_name) DO NOTHING`,
			fileID,
			tag.Name,
			tag.Type,
			tag.TextValue,
			tag.IntegerValue,
		)
		if err != nil {
			return fmt.Errorf("insert tag %q: %w", tag.Name, err)
		}
		inserted, err := statementChanged(result)
		if err != nil {
			return fmt.Errorf("read add tag %q result: %w", tag.Name, err)
		}
		if inserted {
			change.Changed = true
			tagID, err := result.LastInsertId()
			if err != nil {
				return fmt.Errorf("read inserted tag %q id: %w", tag.Name, err)
			}
			return insertTagValues(ctx, tx, tagID, tag.Name, tag.Values)
		}
		existing, tagID, err := queryTagAssignment(ctx, tx, fileID, tag.Name)
		if err != nil {
			return err
		}
		if existing.Type == "multivalue" && tag.Type == "multivalue" {
			for _, value := range tag.Values {
				result, err := tx.ExecContext(
					ctx,
					"INSERT OR IGNORE INTO file_tag_value (file_tag_id, value) VALUES (?, ?)",
					tagID,
					value,
				)
				if err != nil {
					return fmt.Errorf("insert tag %q value %q: %w", tag.Name, value, err)
				}
				added, err := statementChanged(result)
				if err != nil {
					return fmt.Errorf("read add tag %q value %q result: %w", tag.Name, value, err)
				}
				change.Changed = change.Changed || added
			}
			return nil
		}
		if sameScalarTag(existing, tag) {
			return nil
		}
		return fmt.Errorf("%w: %s", ErrTagAlreadyAssigned, tag.Name)
	})
	if err != nil {
		return TagChange{}, err
	}
	return change, nil
}

// SetTag atomically replaces the complete assignment for one tag.
func (d *Database) SetTag(
	ctx context.Context,
	sha256 string,
	tag TagRecord,
) (TagChange, error) {
	if err := validateTagRecord(tag); err != nil {
		return TagChange{}, err
	}
	err := d.writeTransaction(ctx, "set file tag", func(tx *sql.Tx) error {
		fileID, err := queryFileID(ctx, tx, sha256)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(
			ctx,
			"DELETE FROM file_tag WHERE file_id = ? AND tag_name = ?",
			fileID,
			tag.Name,
		); err != nil {
			return fmt.Errorf("delete previous tag %q: %w", tag.Name, err)
		}
		return insertTag(ctx, tx, fileID, tag)
	})
	if err != nil {
		return TagChange{}, err
	}
	return TagChange{Changed: true}, nil
}

// RemoveTag removes a complete tag assignment and its multivalue children.
// Repeated removals from an existing file are no-ops.
func (d *Database) RemoveTag(
	ctx context.Context,
	sha256 string,
	name string,
) (TagChange, error) {
	var change TagChange
	err := d.writeTransaction(ctx, "remove file tag", func(tx *sql.Tx) error {
		fileID, err := queryFileID(ctx, tx, sha256)
		if err != nil {
			return err
		}
		result, err := tx.ExecContext(
			ctx,
			"DELETE FROM file_tag WHERE file_id = ? AND tag_name = ?",
			fileID,
			name,
		)
		if err != nil {
			return fmt.Errorf("delete tag %q: %w", name, err)
		}
		changed, err := statementChanged(result)
		if err != nil {
			return fmt.Errorf("read remove tag %q result: %w", name, err)
		}
		change.Changed = changed
		return nil
	})
	if err != nil {
		return TagChange{}, err
	}
	return change, nil
}

func insertTag(ctx context.Context, tx *sql.Tx, fileID int64, tag TagRecord) error {
	if err := validateTagRecord(tag); err != nil {
		return err
	}
	result, err := tx.ExecContext(
		ctx,
		`INSERT INTO file_tag
		    (file_id, tag_name, tag_type, text_value, integer_value)
		 VALUES (?, ?, ?, ?, ?)`,
		fileID,
		tag.Name,
		tag.Type,
		tag.TextValue,
		tag.IntegerValue,
	)
	if err != nil {
		return fmt.Errorf("insert tag %q: %w", tag.Name, err)
	}
	tagID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("read inserted tag %q id: %w", tag.Name, err)
	}
	return insertTagValues(ctx, tx, tagID, tag.Name, tag.Values)
}

func insertTagValues(
	ctx context.Context,
	tx *sql.Tx,
	tagID int64,
	name string,
	values []string,
) error {
	for _, value := range values {
		if _, err := tx.ExecContext(
			ctx,
			"INSERT INTO file_tag_value (file_tag_id, value) VALUES (?, ?)",
			tagID,
			value,
		); err != nil {
			return fmt.Errorf("insert tag %q value %q: %w", name, value, err)
		}
	}
	return nil
}

func queryTagAssignment(
	ctx context.Context,
	tx *sql.Tx,
	fileID int64,
	name string,
) (TagRecord, int64, error) {
	var tag TagRecord
	var tagID int64
	var textValue sql.NullString
	var integerValue sql.NullInt64
	if err := tx.QueryRowContext(
		ctx,
		`SELECT file_tag_id, tag_name, tag_type, text_value, integer_value
		 FROM file_tag WHERE file_id = ? AND tag_name = ?`,
		fileID,
		name,
	).Scan(&tagID, &tag.Name, &tag.Type, &textValue, &integerValue); err != nil {
		return TagRecord{}, 0, fmt.Errorf("query tag %q: %w", name, err)
	}
	tag.TextValue = nullableString(textValue)
	tag.IntegerValue = nullableInt64(integerValue)
	return tag, tagID, nil
}

func validateTagRecord(tag TagRecord) error {
	if strings.TrimSpace(tag.Name) == "" {
		return fmt.Errorf("tag name is empty")
	}
	scalarText := tag.TextValue != nil
	scalarInteger := tag.IntegerValue != nil
	hasValues := len(tag.Values) > 0
	valid := false
	switch tag.Type {
	case "bool":
		valid = !scalarText && !scalarInteger && !hasValues
	case "text", "value", "date", "datetime":
		valid = scalarText && !scalarInteger && !hasValues
	case "int":
		valid = !scalarText && scalarInteger && *tag.IntegerValue >= 0 && !hasValues
	case "multivalue":
		valid = !scalarText && !scalarInteger
	}
	if !valid {
		return fmt.Errorf("tag %q has invalid %q persistence value", tag.Name, tag.Type)
	}
	seen := make(map[string]struct{}, len(tag.Values))
	for _, value := range tag.Values {
		if value == "" {
			return fmt.Errorf("tag %q has an empty multivalue", tag.Name)
		}
		if _, duplicate := seen[value]; duplicate {
			return fmt.Errorf("tag %q repeats value %q", tag.Name, value)
		}
		seen[value] = struct{}{}
	}
	return nil
}

func sameScalarTag(left, right TagRecord) bool {
	if left.Type != right.Type {
		return false
	}
	return equalStringPointers(left.TextValue, right.TextValue) &&
		equalInt64Pointers(left.IntegerValue, right.IntegerValue)
}

func equalStringPointers(left, right *string) bool {
	return left == nil && right == nil || left != nil && right != nil && *left == *right
}

func equalInt64Pointers(left, right *int64) bool {
	return left == nil && right == nil || left != nil && right != nil && *left == *right
}
