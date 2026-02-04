# FreeBooru

FreeBooru is a booru-style tagging software with support for multiple file storage backends (praise the rclone).
It is written with tauri and designed to be run as a desktop application.

## What does it do?

FreeBooru allows you to organize and tag your files (images, videos, documents, etc.) in a [booru-style](https://safebooru.org/index.php?page=post&s=list) manner.

Also to be DRY with the tags, you have a tag config that allows you to define exact tag types, tag dependencies (and more) allowed within your booru.

`rclone` is used to support a wide variety of cloud storage backends, allowing you to store your files wherever you want. However, FreeBooru can also be used with local storage only or a mix of local and remote. You can easily control the versions of each file by assigning `storage-tags`.

## Installation

...

## Quick start

See [Configuration Example](docs/config-example.md) for details.

1. Setup cloud storage with `rclone` if you plan to use it.
2. Create a collection in FreeBooru.
3. Edit the collection config file located at `$HOME/.config/freebooru/collections/<collectionName>/config.yaml` to your liking.
4. Add your first file

## Status

Hello world!
