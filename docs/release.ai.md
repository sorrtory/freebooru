# MVP release evidence

Run `go tool task release` from the repository root. The target includes the
authoritative `check` task and fails if formatting, lint, tests, races, the
clean-home CLI workflow, or module verification fails.

| MVP criterion                          | Executable evidence                                                                       |
| -------------------------------------- | ----------------------------------------------------------------------------------------- |
| 1. Safe initialization                 | `TestInitCommand`, `TestEnsureDefaultsIsIdempotent`                                       |
| 2. Complete config validation          | config catalog, YAML, duplicate, reference, and graph test suites                         |
| 3. Default collection schema           | `TestMVPCleanHomeWorkflow`, `TestDatabaseLifecycle`                                       |
| 4. Transactional multi-storage import  | `TestImportPersistsContentAndTypedState`, per-copy failure/cancellation tests             |
| 5. Typed tag persistence and lookup    | `TestSetTagPersistsEveryTagType`, `TestTagCommandsMutateAndReadTypedAssignments`          |
| 6. Required tags                       | `TestOpenCollectionRejectsMissingRequiredPersistedTag` and required mutation/import tests |
| 7. Storage copy consistency            | storage mutation tests, concurrent duplicate import test, final-copy deletion test        |
| 8. Independent GUI/CLI-style access    | `TestIndependentDatabaseConnectionsShareCommittedState` and contention test               |
| 9. Failure and recovery automation     | Phase 22 rollback, source modification, orphan restart, and race tests                    |
| 10. Every MVP predicate and graph rule | predicate compiler/checker suites and persisted relationship test                         |
| 11. Authoritative quality gate         | `go tool task check`, invoked by `go tool task release`                                   |

The end-to-end `TestMVPCleanHomeWorkflow` uses a fresh home and fresh command
tree for every invocation. It covers init, config check, import, tag lookup and
mutation, search, adding and removing storage copies, final-copy deletion, and
reopening persisted state through later CLI invocations.

Platform limitations and fail-safe behavior are recorded in `mvp.ai.md`.
