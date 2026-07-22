# FreeBooru frontend goal and implementation TODO

This document defines the frontend work after the import workspace. It refines
the broader [library GUI contract](./todo-gui-library.ai.md) into a small,
collection-scoped application with a coherent visual system.

## Product goal

Provide a fast booru workspace where a user can select or create a collection,
browse or search its files, upload one file, inspect collection information,
and edit typed file assignments. The same Vue application and relative HTTP
contracts must work in a remote browser and in Wails.

The interface favors direct actions, dense useful information, and shallow
navigation. Remove slogans and implementation commentary from normal product
screens. Runtime and configuration details remain available when they help the
user resolve a problem.

## Visual direction: Konata archive

Use the supplied Konata reference image as the palette source:

`https://static.wikia.nocookie.net/luckystar/images/f/fe/Konata-san2.png/revision/latest/scale-to-width-down/268?cb=20240612064611`

The product does not need to display or redistribute the character image. Its
visual identity comes from the image's color relationships:

- clear mid and deep blues from the hair are dominant;
- near-white uniform panels become light surfaces;
- saturated ribbon pink-red marks important actions and conflicts;
- eye green is reserved for success and valid state;
- skin and taupe tones may provide rare warm details, never a competing theme;
- ink-like navy replaces generic black.

Initial tokens sampled and interpreted from the reference image:

```css
:root {
  color-scheme: light;
  --canvas: #f3f7fb;
  --surface: #ffffff;
  --surface-raised: #eaf2f9;
  --text: #192a3a;
  --text-muted: #5d7082;
  --border: #c8d9e8;
  --primary: #2772b2;
  --primary-hover: #1b5688;
  --primary-soft: #dbeaf7;
  --action: #c93b53;
  --action-hover: #a43648;
  --success: #187247;
  --warning: #936763;
  --danger: #b52f47;
  --focus: #2a7cc3;
}

:root[data-theme='dark'] {
  color-scheme: dark;
  --canvas: #101923;
  --surface: #172534;
  --surface-raised: #203246;
  --text: #f5f7f9;
  --text-muted: #afc0cf;
  --border: #36536c;
  --primary: #679dd3;
  --primary-hover: #91bae1;
  --primary-soft: #213f5a;
  --action: #f0526d;
  --action-hover: #ff7188;
  --success: #55b77c;
  --warning: #d6aaa6;
  --danger: #ff7188;
  --focus: #7db1df;
}
```

These are semantic tokens, not a mandate to color every component. Blue owns
navigation, selection, links, focus, and primary structure. Pink is sparse:
use it for the principal creation/upload action or destructive/conflict
attention, but not both simultaneously in the same context. Green means a
successful or valid state and must not become a general accent.

### Theme behavior

- Support `System`, `Light`, and `Dark` choices.
- Default to `System` on first use.
- Persist an explicit preference in local storage; theme preference is local UI
  state and is never sent to Core.
- Apply the resolved theme before Vue mounts to avoid a light/dark flash.
- React to operating-system theme changes while preference is `System`.
- Use the same semantic token names in both themes; components do not branch on
  theme.
- Verify WCAG AA contrast, visible keyboard focus, disabled state, selection,
  errors, and preview media against both themes.

## Content cleanup

When configuration is ready, `/` redirects to the selected or default
collection instead of showing a congratulatory status card.

Remove normal-screen copy such as:

- `One interface · Two runtimes`;
- `FreeBooru is ready`;
- `Personal archive / system check`;
- explanatory text about Wails versus the web server.

Keep `FreeBooru` once in the application shell as product identity. Keep
runtime/configuration information only on an unavailable/configuration problem
screen or a compact diagnostics view. Labels describe user concepts and
actions: `Collection`, `All files`, `Search`, `Upload`, `Tags`, `Save`.

## Information architecture

```text
Application shell
├── collection switcher
│   ├── searchable configured collections
│   └── Add collection…
├── Collection overview
├── All files
├── Search
├── Upload
└── theme control

File result
└── File detail
    ├── content preview / safe fallback
    ├── metadata
    └── typed assignment editor
```

Routes:

| Route | Screen |
| --- | --- |
| `/` | readiness gate, then redirect |
| `/collections` | collection chooser and empty state |
| `/collections/:collection` | collection overview |
| `/collections/:collection/files` | all files |
| `/collections/:collection/search?q=...` | search and help |
| `/collections/:collection/files/:sha256` | file detail and tag editing |
| `/collections/:collection/import` | existing import workspace |

The route is the source of truth for the selected collection. A small session
composable may remember the last successfully opened collection only to choose
the next redirect. Every collection-scoped HTTP and Core call still receives
the explicit route collection. Never introduce process-global `OpenCollection`
state.

## Application shell

- Desktop: compact top bar or narrow sidebar with collection switcher and the
  four primary destinations: Overview, All files, Search, Upload.
- Mobile: compact header plus an accessible menu or bottom navigation; do not
  squeeze a desktop sidebar into the viewport.
- Active destination and collection must remain obvious without relying on
  color alone.
- Keep touch targets at least `2.75rem`; keyboard order follows visual order.
- Theme selection belongs in a small settings/menu surface, not a permanent
  three-button toolbar.

## Collection switcher and creation

The collection switcher is a searchable combobox, not a native `<select>`.

- Show the current collection and an optional default marker.
- Filter configured collections as the user types.
- Arrow keys move through results; Enter selects; Escape closes.
- The last row is always `Add collection…`, separated from results.
- Selecting a collection navigates to its overview and closes the combobox.
- If the current route names a missing collection, show a useful not-found
  state and keep the switcher available.

`Add collection…` opens a focused page or small dialog. The minimal form asks
for a valid unique collection name and shows the resulting identifier before
submission. The initial backend operation should create a usable collection
from FreeBooru's starter/system tag catalog and default storage, write config
atomically, validate/reload it, and return the collection summary. It must not
leave a partial YAML file or database after failure.

Before UI implementation, define and test:

```http
GET  /api/v1/collections
POST /api/v1/collections
GET  /api/v1/collections/{collection}
```

Collection information should expose safe, useful data only: name, default
state, imported tag count, required field count, available storages, file
count, and aggregate size when cheap to obtain. Do not expose local database,
config, or storage paths to a remote frontend.

## Collection overview

Clicking the current collection name or `Overview` opens a real collection
page. Keep it concise:

- collection name and default marker;
- file count and total indexed size;
- imported tag and storage counts;
- recent files when available;
- direct actions for Search, All files, and Upload;
- configuration warnings relevant to this collection.

This is an operational summary, not a dashboard full of decorative charts.

## All files and reusable file results

`All files` is a dedicated route and button. It uses the same `FileResults`
component as Search with an empty query and stable newest-first ordering.

- Use a responsive thumbnail grid for visual media and a clear type fallback
  for other files.
- Each card shows a useful filename, a short tag summary, file type/size, and
  selection/open affordance.
- Avoid one detail API request per result; summary DTOs contain card data.
- Provide deterministic pagination or cursor loading with explicit loading,
  empty, end, cancellation, and retry states.
- Do not make infinite scroll the only way to reach or revisit results.

## Search

Search is a separate route with its own prominent action in the shell.

- Keep the query in URL state so refresh, back, forward, and links work.
- Submit deliberately; do not run a remote search on every keystroke.
- Cancel superseded requests and reject stale results.
- Reuse the All files result cards and pagination.
- Distinguish an empty collection from a valid query with no matches.

Search help must be close to the field but visually secondary. Show a short
starter row and an expandable reference with examples generated from the
actual supported grammar, for example tag presence, exact value, negation,
multiple conditions, and quoted text. Do not document syntax the backend does
not implement. API parse errors should identify the problematic portion and
link/focus the relevant help entry.

## File detail and assignment editing

Every file card opens `/collections/:collection/files/:sha256`.

- Show a bounded preview when supported and a useful fallback otherwise.
- Show safe metadata, storages, and all canonical assignments.
- Allow add, edit, and remove operations using the same type-aware field
  components as Import.
- Evaluate prospective changes and display demands/conflicts before mutation
  when the Core contract supports it; final mutation remains authoritative.
- Require revision/ETag protection so two tabs cannot silently overwrite each
  other.
- Keep storage mutation visually distinct because it has physical effects.
- Return to Search or All files with query and pagination state preserved.

## Value and multivalue controls

Replace plain native value selection with one shared accessible combobox.

- Search predefined values.
- Open on click, keyboard, or typing.
- Single `value` commits one option; `multivalue` shows removable chips.
- Selected and unavailable states use text/icon/shape as well as color.
- Large lists virtualize or page only after measurement proves it necessary.
- Native controls may remain as a no-JavaScript/accessibility fallback, not the
  normal presentation.
- Bool, text, integer, date, and datetime controls remain intentionally simple.

## Vue structure

Start without Pinia. Use route state plus focused composables and introduce a
store only if cross-route mutable state becomes difficult to reason about.

```text
App.vue
└── AppShell.vue
    ├── CollectionSwitcher.vue
    ├── PrimaryNavigation.vue
    └── ThemeMenu.vue

pages/
├── CollectionsPage.vue
├── CollectionOverviewPage.vue
├── FilesPage.vue
├── SearchPage.vue
├── FileDetailPage.vue
└── ImportPage.vue

components/
├── FileResults.vue
├── FileCard.vue
├── SearchHelp.vue
├── ValueCombobox.vue
└── TagField.vue

composables/
├── useTheme.ts
├── useCollections.ts
├── useFileSearch.ts
└── useFileDetail.ts
```

Components receive props and emit typed events. API calls, cancellation, and
replaceable response snapshots live in composables or page orchestration, not
in presentation components.

## Ordered implementation checklist

### Phase 1: visual foundation and shell

- [x] Add semantic light/dark tokens derived from the reference image.
- [x] Implement `System`, `Light`, and `Dark` preference without theme flash.
- [x] Restyle existing status/import surfaces in both themes.
- [x] Remove slogans and runtime exposition from the ready experience.
- [x] Add responsive `AppShell`, primary navigation, and route-aware states.
- [x] Add keyboard, contrast, mobile, and theme persistence tests.
- [x] Build embedded assets and commit this phase independently.

### Phase 2: collections

- [ ] Define Core collection list/info/create request and result types.
- [ ] Define atomic starter-backed collection creation and rollback behavior.
- [ ] Add collection list, create, and detail HTTP endpoints.
- [ ] Implement searchable `CollectionSwitcher` with final `Add collection…`
  row and complete keyboard behavior.
- [ ] Implement collection chooser, creation flow, and overview page.
- [ ] Prove every request passes the route collection explicitly.
- [ ] Test duplicate/invalid creation, unavailable config, and stale routes.
- [ ] Commit this phase independently.

### Phase 3: all files and search

- [ ] Complete summary DTO, content URL, stable pagination, and cancellation
  contracts from `todo-gui-library.ai.md`.
- [ ] Implement responsive reusable `FileResults` and `FileCard` components.
- [ ] Implement dedicated All files route and empty-collection state.
- [ ] Implement URL-backed Search route, deliberate submit, and syntax errors.
- [ ] Add concise and expanded search help matching the supported grammar.
- [ ] Test non-image files, broken previews, no matches, paging, and stale
  responses.
- [ ] Commit this phase independently.

### Phase 4: file detail and tag editing

- [ ] Complete file detail, ETag/revision, assignment mutation, relationship
  error, and storage confirmation contracts.
- [ ] Implement preview, metadata, assignment list, and edit modes.
- [ ] Reuse and refine the typed `TagField` controls from Import.
- [ ] Implement accessible `ValueCombobox` for value and multivalue tags.
- [ ] Preserve originating search/all-files navigation state.
- [ ] Test every tag type, demands, conflicts, stale revisions, and last-copy
  storage removal.
- [ ] Commit this phase independently.

### Phase 5: integration and release evidence

- [ ] Verify every route in browser and Wails with light and dark themes.
- [ ] Verify mobile widths, keyboard-only use, reduced motion, and screen-reader
  names for comboboxes and file cards.
- [ ] Verify remote mode exposes no local paths and uses only relative API URLs.
- [ ] Run `go tool task check` and record any manual Wails evidence.
- [ ] Update `docs/gui.md`, `docs/http-server.md`, `docs/navigation.md`, and the
  relevant completed checklists.
- [ ] Commit the final integration separately.

## Explicit non-goals for this sequence

- visual tag/config schema editing;
- batch, directory, resumable, or background uploads;
- social features, accounts, moderation, votes, or comments;
- decorative dashboards and charts;
- downloading or bundling the Konata reference image;
- a Pinia store added only for convention;
- connector arrows before the text-first relationship UI is complete.

## Completion criteria

- A first-time user can create or select a collection without leaving the GUI.
- The URL and every backend call identify the same explicit collection.
- Search and All files share one fast, responsive result experience.
- Any result can be opened and its typed assignments safely edited.
- Import remains functional and visually integrated.
- Theme defaults to the operating system and explicit choice persists without
  flashing the wrong theme.
- Both themes clearly reflect the reference palette and meet accessibility
  contrast/focus requirements.
- Ready screens contain user actions and useful state, not implementation
  slogans.
- The web and Wails builds use the same Vue routes and HTTP contracts.
