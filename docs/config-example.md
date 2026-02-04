# Freebooru Configuration Example

FreeBooru divides to collections.
Each collection has its own configuration files stored in the following path:

```
$HOME/.config/freebooru/collections/<collectionName>/
```

```yaml
storage:
    folder1:
        type: local
        # location: $HOME/Downloads/freebooru is reserved and used by default `Download` location
        location: $HOME/Pictures/freebooru
tags:
    collisions:
		required:
            - [character, anime]
            - [character_cirno, cute]
    required:
		language:
			default: ru
			oneof:
				- ru
				- en
    optional:
        cute:
        anime:
        character:
			manyof:
				- unknown
				- cirno
                - asuka
```

## Configuration file structure

FreeBooru has a per collection configuration
The following file is loaded on program start up:

```
$HOME/.config/freebooru/collections/<collectionName>/config.yaml
```

You can edit it manually (todo: or through the UI).

```yaml
storage:
	folder1: # this will be tag
		type: local
		location: $HOME/Downloads/freebooru

	yandex-disk: # and this will be tag.
		type: rclone
        name: yandex-disk # rclone remote name. Based on rclone config. It's suggested to be the same as tag

		# rclone has only one config, we just point to what we use
		# config: $HOME/.config/rclone/...

	# we can skip minio, aws-s3, all that is supported by rclone
tags:
	# tags from storage names will be created: storage_<name> ...
	# if no storage tags are assigned, then file is removed

	collisions:
		required:
            # [A, B] means if A is assigned, then B must be assigned as well but not necessarily vice versa

			- [anime, anime-title] # requires anime to have a title
			- [anime-title, anime] # vice versa; anime-title unwraps to any of oneofs
			- [edit, video]
		conflicting:
            # [A, B] means that if A is assigned, then B must NOT be assigned and vice versa

			- [document, text-on-image_no-text] # document tag may not be with no-text

	# ---
    # Underscore is is not allowed in tag names and values.
    # We have two types of tags: valued and non-valued.
    # Non-valued tags.
    #   <name>:
    #
    # Valued tags. <name>_<value> will be created on assigning.
    #   <name>:
    #       default?: <default value>
    #       oneof: # for single value tags
    #           - <default value>
    #           - ... <allowed value>
    #       manyof: # for multi value tags
    #           - <default value>
    #           - ... <allowed value>
    # ---


	# you will have to set one of this on upload
	required:
		language:
			default: ru
			oneof:
				- ru
				- en
				- jpn
		file-type:
			default: img
			oneof:
				- img
				- video
				- gif
				- website? # todo
				- document
                # any other type you need: word, gpg, txt
                # because it is just a tag
		text-on-image:
			default: text
			oneof:
				- text
				- no-text
		privacy:
			default: public
			oneof:
				- public
				- private
		sanity:
			default: sfw
			oneof:
				- sfw
				- nsfw
	optional:
		ecnrypted:
		cosplay:
		cute:
		meme:
		edit:
		chatting: # for very "important" conversations
		anime:
		anime-title:
			# intentionally no default to force user to choose
			oneof:
				- unknown
				- ... # allowed titles
		character:
			manyof:
				- unknown
				- ... # allowed characters; eg cirno, asmongold
		mood:
			oneof:
				- funny
				- sad
		topic:
            oneof:
                - it
                - vibe
                - game
```

> We may also add global configs like `/home/z/.config/freebooru/tags.yaml`
