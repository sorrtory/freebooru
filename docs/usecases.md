# Some usecases to illustrate Freebooru capabilities

## Usecase: add new file

1. Open freeboru
2. Open collection
3. Press `Add file`
4. Fill required tags
    - add storage tags
    - set other required tags for this collection
5. Commit

File is uploaded to selected storage backends via rclone

## Usecase: query files by tags

1. Query by tag
2. See previews
3. Add more tags
4. Add "no_\<name\>" tags to exclude
5. Get less images
6. Select image
7. Press download

File is downloaded to ~/Downloads/freebooru

## Usecase: change storage backends

Assume we have two storage backends configured: A and B

1. Query files with storage tag A
2. Press `Edit tags`
3. Remove tag storage_A
4. Add tag storage_B
5. Commit

Files are moved with rclone

If no storage tags left, the file is removed obviously
