# FreeBooru frontend refinement TODO

This contract covers the next frontend iteration after collection browsing,
search, upload, and file editing. It turns the agreed UX changes into small,
independently reviewable phases. The interface remains one Vue application for
the HTTP server and Wails, and every collection-scoped Core call explicitly
names its collection.

## Product direction

FreeBooru should feel like a mature, compact personal booru: collection context
first, files immediately reachable, and configuration available without
competing with daily actions.

- Blue owns primary actions, navigation, focus, and selection.
- Red is used only for destructive actions, active conflicts, and errors.
- Upload uses the same typography, semantic tokens, and surfaces as every other
  route; components must not contain hard-coded dark backgrounds.
- Prefer restrained inline SVG icons with visible text labels. Icons do not
  replace accessible names.
- Avoid duplicate headings, explanatory runtime copy, decorative statistics,
  and actions already present in adjacent navigation.
- Respect reduced motion, keyboard navigation, touch targets of at least
  `2.75rem`, and light/dark contrast.

## Information architecture

Collection navigation, in order: Overview, Files, Search, Tags, Storage,
Upload. Application Settings is opened by a gear outside collection navigation.
Theme selection belongs inside Settings.

| Route | Purpose |
| --- | --- |
| `/collections/:collection` | concise collection overview |
| `/collections/:collection/files` | complete collection file browser |
| `/collections/:collection/search?q=...` | assisted tag search |
| `/collections/:collection/tags?required=true` | imported and available tags |
| `/collections/:collection/storage` | imported and available storage |
| `/collections/:collection/import` | one-file upload workflow |
| `/settings` | appearance and editable application configuration |

Navigation selection must use an exact route match. A nested collection route
must not leave Overview selected.

## Phase 1: visual and navigation corrections

### Theme consistency

- [x] Replace Upload-specific colors and fonts with global semantic tokens.
- [x] Remove hard-coded dark input, panel, and card backgrounds.
- [x] Use the shared body font for controls and compact metadata; retain the
  display face only for deliberate page headings.
- [x] Refine primary blue into a mature mid/deep blue in both themes.
- [x] Change Upload and Assign actions to primary blue.
- [x] Reserve danger red for trash, destructive confirmation, conflicts, and
  errors.
- [x] Verify Upload in System, Light, and Dark themes before and after mount.

### Shell corrections

- [x] Add a magnifying-glass icon to Search and an upload-tray icon to Upload.
- [x] Use inline SVG with `aria-hidden` when adjacent text is present.
- [x] Replace the visible theme selector with a settings gear.
- [x] Move System/Light/Dark selection into the Settings surface.
- [x] Fix Overview active state using exact route matching.
- [x] Rename `All files` to `Files` consistently in navigation and headings.
- [x] Test desktop, compact desktop, and mobile active states.
- [x] Autoreview and commit this phase independently.

## Phase 2: collection resource contracts

Implement read contracts before their pages. Responses expose useful names,
types, comments, counts, and state but never local configuration, database, or
storage paths.

### Tags

- [x] Define a collection tag summary DTO with name, type, comment, required
  state, allowed values, assignment count, and imported state.
- [x] List imported tags first and globally configured but unimported tags
  second.
- [x] Add collection-scoped filtering by name and required state.
- [x] Define an atomic `Import tag into collection` Core operation.
- [x] Reject duplicate imports and references unavailable to the collection.
- [x] Rewrite collection YAML atomically, reload the catalog/graph, and roll
  back the file if validation fails.

### Storage

- [x] Define a safe collection storage summary DTO with name, imported state,
  file count, aggregate size, and availability/status that does not disclose
  credentials or paths.
- [x] List imported storage first and configured but unimported storage second.
- [x] Define an atomic `Import storage into collection` Core operation with the
  same validation and rollback guarantees as tags.
- [x] Keep importing storage distinct from copying a particular file to it.

```http
GET  /api/v1/collections/{collection}/tags
POST /api/v1/collections/{collection}/tags/{tag}/import
GET  /api/v1/collections/{collection}/storages
POST /api/v1/collections/{collection}/storages/{storage}/import
```

- [x] Test explicit collection propagation on every Core call.
- [x] Test safe remote DTOs for path and credential disclosure.
- [x] Test atomic success, invalid graph rollback, duplicate import, and
  concurrent update behavior.
- [x] Document and commit this phase independently.

## Phase 3: Tags, Storage, and linked overview

### Overview navigation

- [x] Make the file count open Files.
- [x] Make the tag count open Tags.
- [x] Make the required count open Tags with `required=true`.
- [x] Make the storage summary open Storage.
- [x] Make each useful row/card one accessible click target.

### Tags page

- [x] Show imported collection tags first.
- [x] Support URL-backed name search and a Required-only filter.
- [x] Show type, comment, values, usage count, and required state without
  duplicating labels.
- [x] Place `Other configured tags` after the imported section.
- [x] Initially collapse or bound the secondary section and reveal more on
  explicit interaction or scroll; never make it undiscoverable.
- [x] Add `Import into collection` with pending, success, and error states.

### Storage page

- [x] Show collection storage first with file count, size, and safe status.
- [x] Open compact storage information when an imported row is selected.
- [x] Place `Other configured storage` after the collection section.
- [x] Add `Import into collection` where valid.
- [x] Keep destructive per-file storage operations out of this phase.

- [ ] Test loading, empty, filtered, missing, successful import, and failed
  validation states.
- [x] Autoreview responsive hierarchy and commit independently.

## Phase 4: assisted booru search

Desktop Search uses a compact left sidebar and file results on the right.
Mobile Search collapses it into a filter/search drawer. The useful pattern is
booru-like tag discovery, not a visual copy of legacy markup.

### Search hint contracts

- [x] Define collection-aware tag-name prefix hints.
- [x] Define valid value hints after a resolved `tag:` prefix.
- [x] Return type and short comment where useful.
- [x] Define hottest tags as the most frequently assigned tags in the current
  collection, with deterministic tie ordering.
- [x] Bound displayed responses and avoid per-keystroke requests.

```http
GET /api/v1/collections/{collection}/search/hints?term=rat
GET /api/v1/collections/{collection}/search/hints?tag=rating&value=sa
GET /api/v1/collections/{collection}/tags/hot?limit=20
```

### Search interaction

- [x] Put search and a compact syntax hint at the top of the sidebar.
- [x] Show hot tags with counts; clicking one inserts it without replacing the
  existing query.
- [x] Show active terms as removable compact chips.
- [x] Suggest tag names for a partial token and values after `tag:`.
- [x] Tab accepts the highlighted completion without submitting.
- [x] Arrow keys navigate, Escape closes, and Enter searches.
- [x] Preserve quoted values and URL-backed `q` state.
- [x] Retain current results while a replacement request loads.
- [x] Avoid one metadata request per result.

- [x] Test completion, keyboard behavior, hot tags, duplicates, quoting, stale
  responses, and mobile disclosure.
- [x] Autoreview accessibility and commit independently.

## Phase 5: lower-click Upload workflow

### Assignment cards

- [x] Clicking a collapsed card opens its editor.
- [x] Clicking the open card header closes it without applying draft changes.
- [x] Rename the primary action to `Assign`.
- [x] `Cancel` discards local edits and closes the editor.
- [x] Add a danger-colored trash icon for an existing assignment with an
  accessible name containing the tag name.
- [x] Inputs and buttons inside a card must not toggle the card.

### Demands, suggestions, and conflicts

- [x] Keep optional recommendations in Suggested.
- [x] Move blocking demanded tags into the Assigned workflow.
- [x] Render a demand directly beneath the assignment that activated it.
- [x] Open the demanded editor when no unambiguous default exists.
- [x] Use indentation, a shared accent, or a restrained connector to show
  causality without requiring an arrow library.
- [x] Keep circularly demanding assignments visually linked to their causes.
- [x] Show both involved tags on conflicts and focus the editable cause.
- [x] Never auto-assign unless Core returns one unambiguous value.

### Blocked import focus

- [x] Keep Import operable for validation when blocked; use `aria-disabled`
  rather than a non-interactive disabled button.
- [x] On activation, scroll to and open the first problem in stable order:
  missing required, unmet demand, then conflict.
- [x] Focus the first relevant control.
- [x] Pulse the card once and announce the problem in a live region.
- [x] Disable the pulse under reduced motion while preserving focus and the
  announcement.

- [ ] Test card boundaries, Assign, Cancel, trash, nested demands, cycles,
  conflicts, focus order, and reduced motion.
- [ ] Compare click count for a representative import before and after.
- [x] Autoreview and commit independently.

## Phase 6: Settings and `freebooru.yaml`

Settings is application-scoped. Appearance is local frontend state;
application configuration is authoritative backend state.

### Configuration contract

- [x] Inventory supported fields and classify each as editable, read-only,
  sensitive, or restart-required.
- [x] Return explicit typed values and options; do not infer controls from raw YAML.
- [x] Do not expose unrelated paths, credentials, YAML comments, or unknown
  fields to a remote client.
- [x] Validate a complete candidate through Core before persistence.
- [x] Write atomically, reload, and retain the previous file on failure.
- [x] Preserve YAML fields not owned by the form.
- [x] Return restart requirements without exposing backend errors.
- [x] Prevent two Settings tabs from silently overwriting each other.

```http
GET /api/v1/settings
PUT /api/v1/settings
```

### Settings UI

- [x] Put System/Light/Dark first.
- [x] Render typed controls for approved application fields.
- [x] Distinguish saved state, unsaved edits, errors, and restart notices.
- [x] Require explicit Save rather than persisting per keystroke.
- [x] Warn before navigation with unsaved configuration.
- [x] Keep sensitive filesystem values outside the remote form.

- [x] Test rollback, server-owned-field preservation, stale revision, diagnostics,
  theme persistence, and restart notices.
- [x] Autoreview remote disclosure and commit independently.

## Phase 7: integration and release evidence

- [ ] Verify every route in browser and Wails.
- [ ] Verify System, Light, and Dark, including Upload before Vue mounts.
- [ ] Verify keyboard-only search, settings, resource import, and blocked Upload
  recovery.
- [ ] Verify mobile widths from `20rem`, touch targets, and reduced motion.
- [ ] Verify useful screen-reader names.
- [ ] Verify relative API URLs and absence of remote path/credential leaks.
- [ ] Run frontend tests/build and focused Go tests after every phase.
- [ ] Run `go tool task check` before completion.
- [ ] Update GUI, HTTP, navigation, and machine-readable contracts.
- [ ] Autoreview and commit final integration independently.

## Explicit non-goals

- raw YAML editing in the browser;
- editing tag definition files or relationship graphs;
- deleting configured tags or storage definitions;
- copying/removing a particular file's storage assignment on the Storage page;
- arbitrary directory browsing in the remote frontend;
- a global frontend store without demonstrated need;
- decorative charts, social booru features, or a visual Safebooru clone;
- a connector-arrow dependency before the CSS relationship layout proves
  insufficient.

## Completion criteria

- Upload belongs visually to the same light/dark application and uses blue for
  constructive actions.
- Exact navigation state, icons, and Settings placement are correct.
- Overview statistics navigate to useful filtered resource pages.
- Imported tags/storage precede resources available to the collection.
- Search supports keyboard tag/value completion and one-click hot tags.
- Upload demands require fewer clicks and blocked submission focuses the next
  actionable problem.
- Settings safely edits the supported `freebooru.yaml` subset with atomic
  validation and concurrency protection.
- Browser and Wails remain equivalent, and collection scope stays explicit.
