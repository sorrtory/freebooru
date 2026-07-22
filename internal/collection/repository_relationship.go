package collection

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// RelationshipKind explains how a child file relates to its parent.
type RelationshipKind string

const (
	// RelationshipVariant identifies a minor version of the parent.
	RelationshipVariant RelationshipKind = "variant"
	// RelationshipAlternate identifies an equivalent alternative.
	RelationshipAlternate RelationshipKind = "alternate"
	// RelationshipDerived identifies a work produced from the parent.
	RelationshipDerived RelationshipKind = "derived"
)

// FileRelationship is one canonical parent/child edge.
type FileRelationship struct {
	ParentSHA256 string
	ChildSHA256  string
	Kind         RelationshipKind
	Comment      string
	DisplayOrder *int64
	CreatedAt    string
	UpdatedAt    string
}

// SetFileParent creates or replaces a child's single parent relationship.
func (d *Database) SetFileParent(
	ctx context.Context,
	relationship FileRelationship,
) error {
	if err := validateFileRelationship(relationship); err != nil {
		return err
	}
	return d.writeTransaction(ctx, "set file parent", func(tx *sql.Tx) error {
		parentID, err := queryFileID(ctx, tx, relationship.ParentSHA256)
		if err != nil {
			return fmt.Errorf("resolve parent: %w", err)
		}
		childID, err := queryFileID(ctx, tx, relationship.ChildSHA256)
		if err != nil {
			return fmt.Errorf("resolve child: %w", err)
		}
		if parentID == childID {
			return fmt.Errorf("file cannot be its own parent")
		}
		cycle, err := relationshipCreatesCycle(ctx, tx, parentID, childID)
		if err != nil {
			return err
		}
		if cycle {
			return fmt.Errorf("file relationship would create a cycle")
		}
		if _, err := tx.ExecContext(
			ctx,
			`INSERT INTO file_relationship
			    (child_file_id, parent_file_id, kind, comment, display_order)
			 VALUES (?, ?, ?, ?, ?)
			 ON CONFLICT(child_file_id) DO UPDATE SET
			    parent_file_id = excluded.parent_file_id,
			    kind = excluded.kind,
			    comment = excluded.comment,
			    display_order = excluded.display_order,
			    updated_at = CURRENT_TIMESTAMP`,
			childID,
			parentID,
			relationship.Kind,
			relationship.Comment,
			relationship.DisplayOrder,
		); err != nil {
			return fmt.Errorf("persist file relationship: %w", err)
		}
		return nil
	})
}

// RemoveFileParent removes a child's relationship. Repeated removal is a no-op.
func (d *Database) RemoveFileParent(ctx context.Context, childSHA256 string) error {
	return d.writeTransaction(ctx, "remove file parent", func(tx *sql.Tx) error {
		childID, err := queryFileID(ctx, tx, childSHA256)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(
			ctx,
			"DELETE FROM file_relationship WHERE child_file_id = ?",
			childID,
		); err != nil {
			return fmt.Errorf("remove file relationship: %w", err)
		}
		return nil
	})
}

// ChildRelationships lists direct children in explicit order and then by hash.
func (d *Database) ChildRelationships(
	ctx context.Context,
	parentSHA256 string,
) (relationships []FileRelationship, err error) {
	rows, err := d.db.QueryContext(
		ctx,
		`SELECT parent.sha256, child.sha256, relationship.kind,
		        relationship.comment, relationship.display_order,
		        relationship.created_at, relationship.updated_at
		 FROM file_relationship AS relationship
		 JOIN file AS parent ON parent.file_id = relationship.parent_file_id
		 JOIN file AS child ON child.file_id = relationship.child_file_id
		 WHERE parent.sha256 = ?
		 ORDER BY relationship.display_order IS NULL, relationship.display_order,
		          child.sha256`,
		parentSHA256,
	)
	if err != nil {
		return nil, fmt.Errorf("query children for file %q: %w", parentSHA256, err)
	}
	defer func() {
		err = errors.Join(err, closeRows(rows, "file relationships"))
	}()
	for rows.Next() {
		var relationship FileRelationship
		if err := rows.Scan(
			&relationship.ParentSHA256,
			&relationship.ChildSHA256,
			&relationship.Kind,
			&relationship.Comment,
			&relationship.DisplayOrder,
			&relationship.CreatedAt,
			&relationship.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan file relationship: %w", err)
		}
		relationships = append(relationships, relationship)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate file relationships: %w", err)
	}
	return relationships, nil
}

func validateFileRelationship(relationship FileRelationship) error {
	switch relationship.Kind {
	case RelationshipVariant, RelationshipAlternate, RelationshipDerived:
	default:
		return fmt.Errorf("file relationship kind %q is unsupported", relationship.Kind)
	}
	if relationship.Comment != "" && strings.TrimSpace(relationship.Comment) == "" {
		return fmt.Errorf("file relationship comment must not be blank")
	}
	if relationship.DisplayOrder != nil && *relationship.DisplayOrder < 0 {
		return fmt.Errorf("file relationship display order must be non-negative")
	}
	return nil
}

func relationshipCreatesCycle(
	ctx context.Context,
	tx *sql.Tx,
	parentID int64,
	childID int64,
) (bool, error) {
	var cycle bool
	if err := tx.QueryRowContext(
		ctx,
		`WITH RECURSIVE ancestors(file_id) AS (
		     SELECT ?
		     UNION
		     SELECT relationship.parent_file_id
		     FROM file_relationship AS relationship
		     JOIN ancestors ON relationship.child_file_id = ancestors.file_id
		 )
		 SELECT EXISTS(SELECT 1 FROM ancestors WHERE file_id = ?)`,
		parentID,
		childID,
	).Scan(&cycle); err != nil {
		return false, fmt.Errorf("check file relationship cycle: %w", err)
	}
	return cycle, nil
}
