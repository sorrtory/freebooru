ALTER TABLE file ADD COLUMN mime_type TEXT NOT NULL DEFAULT 'application/octet-stream'
    CHECK (mime_type <> '');
ALTER TABLE file ADD COLUMN last_interaction_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP;

CREATE TABLE file_relationship (
    child_file_id INTEGER PRIMARY KEY REFERENCES file(file_id) ON DELETE CASCADE,
    parent_file_id INTEGER NOT NULL REFERENCES file(file_id) ON DELETE CASCADE,
    kind TEXT NOT NULL CHECK (kind IN ('variant', 'alternate', 'derived')),
    comment TEXT NOT NULL DEFAULT '',
    display_order INTEGER CHECK (display_order IS NULL OR display_order >= 0),
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (child_file_id <> parent_file_id)
) STRICT;

CREATE INDEX file_relationship_parent_idx
    ON file_relationship(parent_file_id, display_order, child_file_id);
CREATE INDEX file_mime_type_idx ON file(mime_type);
CREATE INDEX file_last_interaction_idx ON file(last_interaction_at);

CREATE TRIGGER file_tag_insert_touch AFTER INSERT ON file_tag BEGIN
    UPDATE file SET updated_at = CURRENT_TIMESTAMP,
                    last_interaction_at = CURRENT_TIMESTAMP
    WHERE file_id = NEW.file_id;
END;
CREATE TRIGGER file_tag_update_touch AFTER UPDATE ON file_tag BEGIN
    UPDATE file SET updated_at = CURRENT_TIMESTAMP,
                    last_interaction_at = CURRENT_TIMESTAMP
    WHERE file_id = NEW.file_id;
END;
CREATE TRIGGER file_tag_delete_touch AFTER DELETE ON file_tag BEGIN
    UPDATE file SET updated_at = CURRENT_TIMESTAMP,
                    last_interaction_at = CURRENT_TIMESTAMP
    WHERE file_id = OLD.file_id;
END;
CREATE TRIGGER file_tag_value_insert_touch AFTER INSERT ON file_tag_value BEGIN
    UPDATE file SET updated_at = CURRENT_TIMESTAMP,
                    last_interaction_at = CURRENT_TIMESTAMP
    WHERE file_id = (
        SELECT file_id FROM file_tag WHERE file_tag_id = NEW.file_tag_id
    );
END;
CREATE TRIGGER file_tag_value_delete_touch AFTER DELETE ON file_tag_value BEGIN
    UPDATE file SET updated_at = CURRENT_TIMESTAMP,
                    last_interaction_at = CURRENT_TIMESTAMP
    WHERE file_id = (
        SELECT file_id FROM file_tag WHERE file_tag_id = OLD.file_tag_id
    );
END;
CREATE TRIGGER file_storage_insert_touch AFTER INSERT ON file_storage BEGIN
    UPDATE file SET updated_at = CURRENT_TIMESTAMP,
                    last_interaction_at = CURRENT_TIMESTAMP
    WHERE file_id = NEW.file_id;
END;
CREATE TRIGGER file_storage_delete_touch AFTER DELETE ON file_storage BEGIN
    UPDATE file SET updated_at = CURRENT_TIMESTAMP,
                    last_interaction_at = CURRENT_TIMESTAMP
    WHERE file_id = OLD.file_id;
END;
CREATE TRIGGER file_relationship_insert_touch AFTER INSERT ON file_relationship BEGIN
    UPDATE file SET updated_at = CURRENT_TIMESTAMP,
                    last_interaction_at = CURRENT_TIMESTAMP
    WHERE file_id IN (NEW.child_file_id, NEW.parent_file_id);
END;
CREATE TRIGGER file_relationship_update_touch AFTER UPDATE ON file_relationship BEGIN
    UPDATE file SET updated_at = CURRENT_TIMESTAMP,
                    last_interaction_at = CURRENT_TIMESTAMP
    WHERE file_id IN (NEW.child_file_id, NEW.parent_file_id,
                      OLD.child_file_id, OLD.parent_file_id);
END;
CREATE TRIGGER file_relationship_delete_touch AFTER DELETE ON file_relationship BEGIN
    UPDATE file SET updated_at = CURRENT_TIMESTAMP,
                    last_interaction_at = CURRENT_TIMESTAMP
    WHERE file_id IN (OLD.child_file_id, OLD.parent_file_id);
END;

INSERT INTO schema_migrations (version) VALUES (3);
