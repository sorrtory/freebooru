# TODO


use frontend from here: https://github.com/immich-app/immich - ?


### General

- fill up issues and project board
- setup ci/cd with github actions (build, test, lint)
- setup workflow on windows and linux
- setup code linting and formatting


### Features

- 1. Test if rclone freezes on large number of files getting queried -- telegram: rclone conf + ssh creds
- 2. update database schema to support tag dependencies -- add new table for many tags support; add new table for tag collisions support (trigger?)
- 3. Add [parent:child](https://danbooru.donmai.us/wiki_pages/help:post_relationships) for multiple versions of the same pic

# Issues

ddd ?

## Frontend

- 1. Schema for frontend
- Minimal buttons vue.js


## Config crate

pattern?

- parser
  - unix support
  - windows support (different paths and encoding?)
  - example config parsing
- tester
  - checks name collisions
  - no underscore in name/value
  - storage tester for local and rclone to be alive
- todo: api for UI that allows to add tags to config

## Superviser crate

remote sources chart?

Similar to obsidian, we want multiple vault/collection support.
That's why we need a first init script that will remember the latest-opened collections and open it

## Collection crate

Must have a button to read config
Should have a button to write config

## Freebooru crate

Main crate that will call business logic, interact with UI etc

## Database crate

sqlite interactions

## Storage crate

crate that will work with local folders and rclone (via shell or that librclone crate)
