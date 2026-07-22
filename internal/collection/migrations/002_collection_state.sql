CREATE TABLE file (
    file_id INTEGER PRIMARY KEY,
    sha256 TEXT NOT NULL UNIQUE
        CHECK (
            length(sha256) = 64
            AND sha256 = lower(sha256)
            AND sha256 NOT GLOB '*[^0-9a-f]*'
        ),
    size_bytes INTEGER NOT NULL CHECK (size_bytes >= 0),
    imported_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
) STRICT;

CREATE TABLE file_source (
    file_source_id INTEGER PRIMARY KEY,
    file_id INTEGER NOT NULL REFERENCES file(file_id) ON DELETE CASCADE,
    source_path TEXT NOT NULL CHECK (source_path <> ''),
    filename TEXT NOT NULL CHECK (filename <> ''),
    observed_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (file_id, source_path)
) STRICT;

CREATE TABLE file_tag (
    file_tag_id INTEGER PRIMARY KEY,
    file_id INTEGER NOT NULL REFERENCES file(file_id) ON DELETE CASCADE,
    tag_name TEXT NOT NULL CHECK (tag_name <> ''),
    tag_type TEXT NOT NULL CHECK (
        tag_type IN ('bool', 'text', 'int', 'date', 'datetime', 'value', 'multivalue')
    ),
    text_value TEXT,
    integer_value INTEGER,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (file_id, tag_name),
    CHECK (
        (tag_type = 'bool' AND text_value IS NULL AND integer_value IS NULL)
        OR (
            tag_type IN ('text', 'value', 'date', 'datetime')
            AND text_value IS NOT NULL
            AND integer_value IS NULL
        )
        OR (
            tag_type = 'int'
            AND text_value IS NULL
            AND integer_value >= 0
        )
        OR (
            tag_type = 'multivalue'
            AND text_value IS NULL
            AND integer_value IS NULL
        )
    )
) STRICT;

CREATE TABLE file_tag_value (
    file_tag_value_id INTEGER PRIMARY KEY,
    file_tag_id INTEGER NOT NULL REFERENCES file_tag(file_tag_id) ON DELETE CASCADE,
    value TEXT NOT NULL CHECK (value <> ''),
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (file_tag_id, value)
) STRICT;

CREATE TABLE file_storage (
    file_storage_id INTEGER PRIMARY KEY,
    file_id INTEGER NOT NULL REFERENCES file(file_id) ON DELETE CASCADE,
    storage_name TEXT NOT NULL CHECK (storage_name <> ''),
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (file_id, storage_name)
) STRICT;

CREATE INDEX file_source_path_idx ON file_source(source_path);
CREATE INDEX file_tag_name_text_idx ON file_tag(tag_name, text_value);
CREATE INDEX file_tag_name_integer_idx ON file_tag(tag_name, integer_value);
CREATE INDEX file_tag_value_value_idx ON file_tag_value(value);
CREATE INDEX file_storage_name_idx ON file_storage(storage_name);

INSERT INTO schema_migrations (version) VALUES (2);
