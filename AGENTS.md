# FreeBooru

FreeBooru is a booru-style tagging software with support for multiple file storage backends (praise the rclone).
It is written with tauri and designed to be run as a desktop application.

## What does it do?

FreeBooru allows you to organize and tag your files (images, videos, documents, etc.) in a [booru-style](https://safebooru.org/index.php?page=post&s=list) manner.

Also to be DRY with the tags, you have a tag config that allows you to define exact tag types, tag dependencies (and more) allowed within your booru.

`rclone` is used to support a wide variety of cloud storage backends, allowing you to store your files wherever you want. However, FreeBooru can also be used with local storage only or a mix of local and remote. You can easily control the versions of each file by assigning `storage-tags`.

## Configuration

The config is done in yaml. It is the key to the minimal tag quantity and quality you want to have in your booru. 

### Config rules



See [Configuration Example](docs/config-example.md) for details.
