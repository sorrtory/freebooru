# freebooru-cli

A tool to manage FreeBooru collections and tags from the command line.
It calls the FreeBooru core.

## Commands

- `init` - initialize the application, default storage, default collection, and
  collection database
- `config` - work with FreeBooru config files
    - `check` - check config validity

There is no `config init`: initialization is an application workflow, not a
configuration-only operation. Successful initialization prints
`FreeBooru initialized`. A failure returns a nonzero status and does not print
the success message.
