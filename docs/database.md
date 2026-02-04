# FreeBooru database

FreeBooru uses a SQLite database to store metadata about the files and tags.
The database file is located at:

```
$HOME/.config/freebooru/collections/<collectionName>/database.db
```

Then it can be easily backed up on any cloud storage using `rclone` or other means.

> TODO: Add back up example here

## Database schema

| Entity            | Description                                                         |
| ----------------- | ------------------------------------------------------------------- |
| FILE              | Metadata for an individual file                                     |
| TAG               | Tag metadata                                                        |
| REPOSITORY        | Storage location info for files (splited from tags for consistency) |
| FILE_TAGS         | Join table linking files to tags                                    |
| FILE_REPOSITORIES | Join table linking files to repositories                            |

```mermaid
erDiagram
    FILE {
        int file_id
        %% --
        text filename
        datetime uploaded_at
        datetime downloaded_at
        %% --
        datetime created_at
        datetime updated_at
    }

    TAG {
        int tag_id
        %% --
        text name
        %% --
        datetime created_at
        datetime updated_at
    }

    REPOSITORY {
        int repository_id
        %% --
        text name
        %% type = local|rclone
        text type
        datetime uploaded_at
        datetime downloaded_at
        %% --
        datetime created_at
        datetime updated_at
    }

    FILE_TAGS {
        int file_id
        int tag_id
        %% --
        datetime created_at
        datetime updated_at
    }

    FILE_REPOSITORIES {
        int file_id
        int repository_id
        %% --
        datetime created_at
        datetime updated_at
    }

    FILE ||--o{ FILE_TAGS : has
    TAG ||--o{ FILE_TAGS : has

    FILE ||--o{ FILE_REPOSITORIES : stored_in
    REPOSITORY ||--o{ FILE_REPOSITORIES : contains
```
