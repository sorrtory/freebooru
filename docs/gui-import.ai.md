# FreeBooru import workspace UX contract

This document refines the one-file import phase in
[todo-gui-library.ai.md](./todo-gui-library.ai.md). It describes a fast,
booru-style tagging workspace that exposes FreeBooru’s typed tags and
relationships without pretending they form a linear kanban workflow.

## UX judgment

The proposed `Assigned | Suggested | New` layout is a strong desktop starting
point because it keeps current state, system guidance, and the tag catalog
visible together. It should not behave as literal kanban:

- an assigned tag is state, not a task that moved to “done”;
- a suggestion is a relationship-derived candidate, not the next mandatory
  pipeline stage;
- a conflict is an edge between assignments, not an item waiting in another
  column;
- required tags without values are incomplete draft fields, not yet assigned.

Use three work areas plus a validation layer:

1. **Assigned** — required placeholders and the current draft assignments;
2. **Suggested** — active demands and recommendations derived from the draft;
3. **Add tag** — searchable imported tags and type-aware editors;
4. **Issues** — conflicts and incomplete requirements shown inline and in a
   compact sticky summary, not as a fourth column.

This retains the speed and spatial memory of the original idea while matching
the actual relationship model.

## Surface choice

Use the route-backed page
`/collections/:collection/import`, not a transient modal.

The workspace contains enough state and explanation to deserve browser
history, refresh behavior, responsive layout, and room for large value sets.
On wide desktop screens it may animate over the file grid like a full-height
sheet, but the URL and component lifecycle remain those of a page. Closing it
returns to the collection route.

The collection header keeps a prominent `Upload` action. Opening the page
places keyboard focus on the file chooser; after a file is chosen, focus moves
to the first unresolved required field.

## Visual direction

The style is a dense personal archive/toolbench rather than a generic project
management board:

- compact tag cards, strong type labels, hash/filename metadata, and restrained
  separators inspired by classic booru sidebars;
- the file preview and final action remain visually dominant;
- status uses shape, label, and icon as well as color;
- completed assignments become quieter, not disabled;
- relationships appear as technical annotations, not decorative animation;
- motion is brief and functional when a card changes section or an edge is
  focused, and is removed under `prefers-reduced-motion`.

Do not imitate a specific booru’s branding or reproduce an old table layout.
Keep the information density and tagging speed while preserving accessible
modern controls.

## Desktop layout

```text
┌──────────────────────────────────────────────────────────────────────────┐
│ ← main / Import     filename.ext · 24 MB              Replace file       │
├────────────────────┬────────────────────┬────────────────────────────────┤
│ ASSIGNED           │ SUGGESTED          │ ADD TAG                        │
│                    │                    │                                │
│ Needs value (2)    │ Required next (1)  │ [ Search imported tags... ]    │
│ ┌ rating  value ┐  │ ┌ title required┐ │ recent / filtered tag cards    │
│ └───────────────┘  │ └───────────────┘ │ selecting opens typed editor   │
│                    │                    │                                │
│ Applied (4)        │ Recommended (3)    │                                │
│ muted, expandable  │ one-click/prefill  │                                │
├────────────────────┴────────────────────┴────────────────────────────────┤
│ 2 required missing · 1 conflict                  [ Import file ]         │
└──────────────────────────────────────────────────────────────────────────┘
```

The file preview is a collapsible header/side panel rather than a fourth
column. Images show a bounded preview; other file types show icon, filename,
size, and detected media type. Preview never delays tag editing.

At widths of at least `64rem`, use three fluid columns. Do not assign fixed
pixel widths; long tag names and translated labels must wrap safely.

## Responsive behavior

- Below `64rem`, Assigned remains visible and Suggested/Add tag become a
  two-tab secondary pane.
- Below `48rem`, use a single-pane segmented control:
  `Assigned | Suggested | Add` with counts in each label.
- The issue summary and Import action remain sticky at the bottom but must not
  cover focused inputs; account for safe-area insets.
- A relation selected in another pane navigates to the target pane/card.
- Connector arrows are omitted on narrow layouts; textual relationship
  explanations remain complete.
- Inputs and actions have at least `2.75rem` touch targets and body text remains
  at least `1rem`.

The mobile experience is not a horizontally scrolling three-column board.

## Assigned area

### Needs value

On schema load, every required non-boolean tag and required storage without a
value appears at the top as an expanded editor labeled `Needs value`. Required
boolean tags may be applied automatically only when Core’s import rules already
apply them automatically; the UI must reflect the returned canonical draft.

An unresolved required card is visually part of Assigned because it describes
the desired file state, but it must say `Not assigned` until it has a valid
value.

After Apply:

1. validate the field locally for basic shape;
2. send the complete draft for authoritative evaluation;
3. replace local draft state with the canonical response;
4. move the completed card into Applied with a short positional transition;
5. keep it selectable and editable.

### Applied

Applied cards are compact and muted, showing tag name, type, and formatted
value. Muted means lower emphasis, never disabled. Selecting a card expands
the same type-aware editor in place with `Apply`, `Cancel`, and `Remove`.

Only one card needs to be expanded at a time. Keyboard navigation must not
depend on drag-and-drop. Cards are ordered deterministically: required first,
then normalized tag name.

## Suggested area

Split this area into two sections.

### Required next

Active missing demands are blocking guidance. Each card states:

- the source assignment that activated the demand;
- the target tag/value condition;
- the configured reason, or a safe generated fallback;
- an action to open a prefilled target editor.

These are not ordinary suggestions and contribute to the blocking count.

### Recommended

Active `suggest` relationships are non-blocking. A recommendation may target:

- a boolean tag presence;
- a tag that still needs user input;
- one exact predefined value;
- one or more allowed values.

If a relationship gives an exact valid value, `Apply` may add it immediately
to the local draft and evaluate. If user input is still needed, the action is
`Choose value` and opens the appropriate editor. Satisfied suggestions
disappear from this section because they are now visible in Applied.

Duplicate suggestions targeting the same condition are grouped into one card
while preserving all source reasons.

## Add tag area

This is a searchable catalog, not a pile of draggable cards.

- Search only tags imported by the explicit collection.
- Exclude tags already represented in Assigned unless the result is used to
  focus/edit that assignment.
- Show type, required/optional state, and a small predefined-value summary.
- Keep recent tags locally for the current page session.
- Selecting a result opens its type-aware editor within this area.
- Applying a valid value moves the tag to Assigned after draft evaluation.

No drag-and-drop is required for the first implementation. Click, keyboard,
and touch must provide the complete workflow. Drag-and-drop may later become a
shortcut, never the only interaction.

## Type-aware fields

| FreeBooru type | Primary control | Notes |
| --- | --- | --- |
| `bool` | checkbox or two-state switch | unchecked means absent, not persisted `false` |
| `text` | text input | preserve user text; show configured validation error |
| `int` | number input | integer step and numeric input mode; reject fractions |
| `date` | date input | submit canonical date string |
| `datetime` | datetime-local input | make timezone/canonical conversion explicit |
| `value` | combobox or radio group | only declared values; show unavailable reasons |
| `multivalue` | multi-select combobox | removable chips plus declared-value search |
| `storage` | storage checklist | visually distinct because it has physical effects |

Large predefined sets use a searchable combobox rather than rendering every
option. Unavailable values remain discoverable when useful, but are disabled
with their relationship reason. Errors are attached to the field and included
in a top issue summary.

## Conflicts and relationship visualization

When a draft contains an active conflict:

- keep both assignments visible;
- mark the involved cards with a conflict badge;
- add one item to the sticky Issues summary;
- show the configured reason and both endpoints in text;
- disable final Import until the conflict is resolved;
- offer `Edit`/`Remove` actions for the involved draft assignments.

On wide screens, selecting or hovering a demand, suggestion, or conflict may
draw a connector between its source and target cards. Draw only the focused
relationship; drawing every edge creates unreadable “spaghetti.”

The first implementation should use an application-owned SVG overlay anchored
to card elements and updated with `ResizeObserver`. Do not add a connector
library until a prototype proves the overlay insufficient. Connector lines
are progressive enhancement: textual cards and `aria-describedby`
relationships are authoritative, and reduced-motion/mobile modes work without
lines.

Suggested visual semantics:

- suggestion: thin dashed accent line;
- demand: solid amber line with arrowhead;
- conflict: solid red line with stop marker;
- focused endpoints receive matching outlines.

Color alone never communicates relationship kind.

## Draft evaluation contract

The dynamic workspace cannot call the mutating import workflow merely to learn
suggestions. Add a side-effect-free Core operation:

```go
type ImportDraftRequest struct {
    Collection  string
    Assignments map[string]any
}

type ImportDraft struct {
    Assignments     map[string]any
    MissingRequired []ImportField
    Evaluation      evaluator.Evaluation
    Complete        bool
}

func (c *Core) EvaluateImportDraft(
    ctx context.Context,
    request ImportDraftRequest,
) (ImportDraft, error)
```

The exact exported result should remain frontend-independent. It must:

- require an explicit collection;
- validate and canonicalize every supplied typed value;
- apply Core-owned automatic required boolean/storage defaults exactly as the
  final import would;
- tolerate an incomplete draft and report missing required fields rather than
  returning a generic failure;
- return missing demands, active conflicts, and suggestions;
- perform no database, source-file, or storage I/O;
- preserve deterministic ordering;
- share validation logic with final `Import` so preview and commit cannot
  drift.

Expose it as:

```http
POST /api/v1/collections/{collection}/imports/evaluate
Content-Type: application/json

{"assignments":{"rating":"safe","storage":["default"]}}
```

The API response maps assignments into ordered typed DTOs and maps every graph
edge into a stable relationship DTO containing kind, source, target, reason,
and source configuration location when safe to expose.

Evaluation occurs after an explicit field Apply/Remove action, not on every
keystroke. A superseded evaluation request is cancelled. Vue replaces the
complete draft/evaluation snapshot atomically and never merges pieces from
out-of-order responses.

## Submission behavior

The sticky footer shows a live summary such as:

```text
2 required values missing · 1 demand unresolved · 0 conflicts
```

`Import file` is enabled only when:

- exactly one source file is selected;
- Core’s latest draft response has `complete: true`;
- there are no missing demands or active conflicts;
- at least one valid storage is assigned;
- no evaluation or upload request is pending.

The multipart import still performs full authoritative validation. Evaluation
success is not a security or consistency bypass. On success, navigate directly
to the new file detail route. On duplicate, offer `Open existing file`. On an
uncertain network failure, do not retry automatically.

## Component boundaries

Suggested first structure:

```text
ImportPage.vue                 route orchestration and final submit
├── ImportFileHeader.vue       chooser, preview, filename, size
├── AssignedTags.vue           required placeholders and applied cards
│   └── TagField.vue           type-dispatched editor
├── ImportGuidance.vue         demands and suggestions
├── TagCatalog.vue             search and new-tag editor
├── RelationshipOverlay.vue   desktop focused SVG connector
└── ImportActionBar.vue        issue counts and submit action

useImportDraft.ts              one draft snapshot, evaluate/cancel/apply/remove
```

Props flow down and typed events flow up. Child components do not call the API
or mutate shared props. The composable owns replaceable response snapshots,
request cancellation, and stale-response protection.

## Acceptance scenarios

- A new upload shows every unresolved required field before optional tags.
- Filling a required value moves it into the quiet Applied section and it
  remains editable.
- Adding an assignment immediately reveals newly activated demands,
  suggestions, and conflicts without uploading the file.
- An exact suggestion can be applied in one action; an incomplete suggestion
  opens the correct typed field.
- Conflict text remains understandable without connector lines.
- Import cannot proceed with missing requirements or conflicts.
- All seven FreeBooru tag types and storage have appropriate keyboard- and
  touch-accessible controls.
- Desktop uses the three-area workspace; tablet/mobile never require horizontal
  board scrolling.
- Cancelling or superseding evaluation never applies stale guidance.
- The same page and HTTP evaluation/upload contract work in browser and Wails.
