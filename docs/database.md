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
    %% Current FILE
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

    %% mb divide to tag required and tag conficting
    COLLISIONS {
        int collision_id
        %% --
        int tag_id_1
        int tag_id_2
        %% false - conflicting, true - required
        bool is_required
        %% --
        datetime created_at
        datetime updated_at
    }

    %% Allowed TAG
    TAG {
        int tag_id
        %% --
        text name
        bool is_required
        %% --
        datetime created_at
        datetime updated_at
    }

    %% Alowed VALUES for every TAG (may be 0/1/N)
    TAG_VALUES {
        int tag_value_id
        %% --
        int tag_id
        text value
        %% --
        datetime created_at
        datetime updated_at
    }

    %% Current VALUE for a FILE
    FILE_VALUES {
        int file_value_id
        %% --
        int file_id
        int tag_value_id
        %% --
        datetime created_at
        datetime updated_at
    }

    %% Current TAG for a FILE
    FILE_TAGS {
        int file_tag_id
        %% --
        int file_id
        int tag_id
        %% --
        datetime created_at
        datetime updated_at
    }

    %% Alowed Storage providers
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

    %% Current STORAGE for FILE
    FILE_REPOSITORIES {
        int file_repository_id
        %% --
        int file_id
        int repository_id
        %% --
        datetime created_at
        datetime updated_at
    }

    FILE ||--o{ FILE_TAGS : has
    FILE ||--o{ FILE_VALUES : has

    TAG ||--o{ FILE_TAGS : has
    TAG ||--o{ TAG_VALUES : has

    FILE_VALUES ||--o{ TAG_VALUES : is_value_of

    FILE ||--o{ FILE_REPOSITORIES : stored_in
    REPOSITORY ||--o{ FILE_REPOSITORIES : contains

    COLLISIONS |o--|{ TAG : involves


```
