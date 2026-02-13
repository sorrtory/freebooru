# TODO

### General

- fill up issues and project board
- setup ci/cd with github actions
- setup workflow on windows and linux
- setup code linting and formatting

### Features

- Test if rclone freezes on large number of files getting quiried
- update dataabase schema to support tag dependencies

# Issues

## Config crate

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

Similar to obsidian, we want multiple vault/collection support.
That's why we need a first init script that will remember the latest-opened collections and open it

## Collection crate

Must have a button to read config
Should have a button to write config

## Freebooru crate

Main crate that will call business logic, interact with UI etc

## Storage crate

crate that will work with local folders and rclone (via shell or that librclone crate)
