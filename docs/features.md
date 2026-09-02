# tinear — feature reference

Retained terminal UI for [Linear.app](https://linear.app). Tinear has one
no-argument interactive entrypoint; authentication is completed in the welcome
screen when no API key is configured.

---

## 1. Startup

1. `main()` constructs the Cooper application, root shell, modal host, and toast host before startup.
2. Config loaded via `config.ard`:
   - `$LINEAR_API_KEY` env var (takes priority)
   - `~/.tinear/config.json` (fallback)
3. If no key is found, install the retained `WelcomeController` login form.
4. If a key is found, install the retained logged-in shell and start Inbox/board loading.

**Config schema (`~/.tinear/config.json`):**

```json
{ "api_key": "lin_api_..." }
```

---

## 2. Welcome screen (unauthenticated)

ASCII logo + form:

```
▀█▀ ▀ █▀█ █▀▀ ▄▀█ █▀█
 █  █ █ █ ██▄ █▀█ █▀▄

Enter your Linear API key

[___________________________]   ← text field, paste API key

[  Login  ]                     ← button
```

- **Text field**: on change updates local draft. Enter submits. Obscured=false
  (so user can verify pasted key).
- **Login button**: calls the asynchronous authentication service, which:
  1. Calls Linear GraphQL `viewer { id name displayName email }` to validate.
  2. On success: saves key to `~/.tinear/config.json`, transitions to
     logged-in screen.
  3. On failure: shows "Invalid API token" error below the button.
- **Quit**: `Ctrl+c` → exit.

---

## 3. Logged-in screen (tabbed shell)

After authentication, the app shows a persistent retained tabbed interface.

### 3.1 State shape

```ard
struct Model {
  tab_index: Int,       // 0 = Inbox, 1 = My Issues, 2+ = dynamic tabs
  tabs: [TabRef],       // issue/document identity + cached display title
}
```

### 3.2 Tab bar

Horizontal row of tabs across the top of the screen:

```
 Inbox  │  My Issues  │  ENG-123  │  ENG-456
```

- **Inbox** and **My Issues** are permanent (always present, leftmost).
- **Issue and document tabs** are added dynamically from Inbox, My Issues, or
  Search. Tabs are deduplicated by `(kind, id)`; opening an existing identity
  selects its retained controller.
- **Selected tab** is accent-colored and bold; inactive tabs are dim.
- Tab bar is a horizontal row followed by a dim rule. See
  [design-language.md](design-language.md) for the current visual standard.

### 3.3 Tab body

Below the tab bar, a single body region shows the active tab's content. Each
screen has one stable retained root; inactive roots remain mounted with
`Display::none` so local state survives switches.

| Tab | Content |
|---|---|
| **Inbox** | Scrollable notification list (see §4) |
| **My Issues** | Kanban board grouped by workflow state (see §5) |
| **Issue/document detail** | Full retained detail view (see §6) |

Footer bar below the body shows left-aligned, context-sensitive key hints without repeating the active tab title.

### 3.4 Tab switching

| Key | Action |
|---|---|
| `Tab` | Next tab (wrap around) |
| `Shift+Tab` | Previous tab (wrap around) |
| `1` | Jump to Inbox |
| `2` | Jump to My Issues |
| `Escape` | If picker open → close picker. If issue tab active → close it and switch to My Issues. Otherwise ignored. |

After switching tabs, the shell focuses the selected retained screen. Only the
active screen receives delegated commands.

### 3.5 Tab persistence

When tabs change, the state is serialized/coalesced to
`~/.tinear/data.json` via `models/cache.ard` and `services/cache_writer.ard`.
On next launch, the active index and issue/document tab references are restored.

**Cache schema:**

```json
{
  "version": 2,
  "logged_in": {
    "tab_index": 2,
    "tabs": [
      { "kind": "issue", "id": "ENG-123", "title": "ENG-123" },
      { "kind": "doc", "id": "doc-slug", "title": "Architecture" }
    ]
  }
}
```

---

## 4. Inbox tab

Scrollable list of Linear notifications. Fetched from the API on mount and on
manual refresh.

### 4.1 Data

Each notification (`models/inbox::Item`):

```ard
struct Item {
  notification_id: Str,
  notification_type: Str,   // e.g. "issueAssignedToYou"
  type_label: Str,          // human-readable: "assigned", "commented", etc.
  actor_name: Str,
  target_id: Str,           // e.g. "ENG-123"
  title: Str,
}
```

`notification_type` is mapped to a short label via `format_type()`:
assigned, commented, mentioned, reacted, changed status, resolved, unassigned,
approved, requested changes, checks failed, requested review, etc.

### 4.2 Rendering

Each row uses a comfortable title/subtitle rhythm:
```
  Update the authentication flow
  Alice assigned · ENG-123
```

- Primary line: title, ellipsized to the pane width.
- Secondary line: actor + action and identifier/team context, dim.
- Hovered row: host gray background with contrasting text.
- Selected row: accent background with contrasting text and a bold primary line.

### 4.3 Keys

| Key | Action |
|---|---|
| `j` / `Down` | Move cursor down |
| `k` / `Up` | Move cursor up |
| `Enter` | Open selected notification's issue in a new tab |
| `Backspace` | Delete (archive) the selected notification |
| `r` | Refresh (re-fetch inbox from API) |
| `Tab` | Switch to next tab |

---

## 5. My Issues tab

Kanban-style board showing the user's assigned issues grouped by workflow state.

### 5.1 Data

Issues are fetched from the Linear API via `models/issues.ard`. Each issue:

```ard
struct Issue {
  id: Str,
  identifier: Str,
  title: Str,
  description: Str?,
  priority: Int?,
  state_id: Str,
  state_name: Str,
  state_type: Str,
  team_id: Str,
  team_name: Str,
  team_key: Str,
  cycle_id: Str,
  cycle_label: Str,
  assignee_id: Str,
  assignee_name: Str,
  project_id: Str,
  project_name: Str,
  url: Str,
}
```

### 5.2 Board layout

Horizontal scroll of columns, one per workflow state. Columns are ordered by
`state_type_rank` (triage < backlog < unstarted < started < completed <
canceled). States with the same raw name share a column across teams.

Each column:
```
   STATE NAME (count)
   ─────────────────
   ENG-123  Title of the issue
   ENG-456  Another issue title
   ...
```

- Column header: bold state name, dim count, and a dim rule.
- Cards: colored priority glyph, bold identifier, title up to two lines, and dim
  cycle/project metadata.
- Hovered card: host gray background with contrasting text.
- Selected card: accent background with contrasting text and a bold identifier.
- Dragging lifts a full card surface that follows the pointer and highlights a
  destination column with an accent heavy rule. Dropping moves the card
  optimistically and updates its Linear status; failures restore the original
  column, mark the card briefly, and show an error toast.
- Multi-team moves resolve the equivalent state for the dragged issue's team and
  never reuse another team's state ID.
- Visible column count adapts to terminal width.

### 5.3 Keys

| Key | Action |
|---|---|
| `j` / `Down` | Move cursor down within column |
| `k` / `Up` | Move cursor up within column |
| `h` / `Left` | Move to previous column |
| `l` / `Right` | Move to next column |
| `Enter` | Open selected issue in a new tab |
| Mouse drag | Drop an issue on another visible column to change its status |
| `s` | Open **state picker** (move issue to different workflow state) |
| `y` | Open **cycle picker** (move issue to different cycle) |
| `p` | Open **priority picker** |
| `a` | Open **assignee picker** |
| `Shift+p` | Open **project picker** |
| `r` | Refresh board from API |
| `?` | Open search overlay |

---

## 6. Issue detail tab

Full view of a single issue. Opened dynamically in a new tab.

### 6.1 Content

```
ENG-123 · Team Name · Project Name
──────────────────────────────────
Title of the Issue
──────────────────────────────────
State: In Progress    Cycle: Cycle 3
Priority: High        Assignee: Alice
──────────────────────────────────
Description text with markdown-like rendering.
Supports bold (**), italic (*), inline code (`),
links, and lists. Mojibake from Linear's API is
cleaned up (Windows-1252 → UTF-8).
──────────────────────────────────
Labels: bug, frontend
Parent: ENG-100
```

- Descriptions and comments use retained Markdown controls after mojibake cleanup.
- Metadata rows show state, priority, assignee, cycle, project, and team.
- Comments retain Linear's one-level threading, use friendly local timestamps,
  and share the board's hover/focus surface treatment.

### 6.2 Keys

| Key | Action |
|---|---|
| `d` / `c` | Show Description / Comments section |
| `j` / `k` | Scroll description or move the comment cursor |
| `n` | Compose a top-level comment |
| `r` | Reply to the selected comment |
| `e` | Edit the description |
| `Ctrl+Enter` | Submit a comment or save a description from its editor |
| `s` / `y` / `p` | Change state / cycle / priority |
| `a` / `Shift+p` | Change assignee / project |
| `o` | Open issue in the browser |
| `Escape` | Dismiss an editor/picker, then close the issue tab |
| `?` | Open search overlay |

### 6.3 Document tabs

Document search results and restored document tabs load the latest Linear
content into a retained detail controller. The tab shows title, project or
initiative, creator, a friendly local update time, and rendered Markdown.
Fresh titles retitle the tab and persisted cache. `j`/`k` scroll, `o` opens the
Linear URL, and closing the tab cancels its periodic refresh worker.

---

## 7. Pickers (modal overlays)

State, cycle, priority, assignee, and project pickers are modal overlays that
appear centered over the current view. They trap focus; lists longer than five
auto-focus a filter input, while smaller lists focus direct keyboard navigation.

### 7.1 Picker controller (`tui/picker_controller.ard`)

The reusable retained picker accepts options with stable IDs, labels, and
optional searchable descriptions. It reconciles rows by ID as the filter
changes.

- Rendered as a bordered modal with a scrollable result list; lists longer than five options add a filter input.
- The selected option uses the accent background; hover uses the host gray background.
- The issue's current value seeds the initial cursor and scroll position.
- Async option loads are stale-safe; dismissing a loading modal cancels its local request ownership.
- Mutations show a non-dismissible progress state and refresh the board on success.

### 7.2 Keys

| Key | Action |
|---|---|
| Typing | Filter by label or description when the filter is visible |
| `Down` / `j` | Move cursor down (`j` is available without a filter) |
| `Up` / `k` | Move cursor up (`k` is available without a filter) |
| `Enter` / `Space` | Confirm selection (`Space` is available without a filter) |
| `Escape` | Dismiss the picker or an in-progress option load |

### 7.3 Picker types

- **State picker**: fetched from API (workflow states for the issue's team).
  Selecting a state calls `issues::update_state()`.
- **Cycle picker**: fetched from API (cycles for the issue's team).
  Selecting a cycle calls `issues::update_cycle()`.
- **Priority picker**: static list (Urgent, High, Medium, Low, No Priority).
  Selecting calls `issues::update_priority()`.
- **Assignee picker**: active members for the issue's team, plus Unassigned.
  Selecting calls `issues::update_assignee()`.
- **Project picker**: active projects for the issue's team, plus No project.
  Selecting calls `issues::update_project()`.

After a successful board mutation, My Issues refreshes while preserving the
selected issue by stable ID.

---

## 8. Search overlay

Opened with `?` from any tab.

- Text field searches Linear issues and documents after a 300 ms debounce.
- Issue and document results share one navigable list; stale responses are ignored.
- If one search source fails, surviving results remain visible with a warning.
- Selecting a result opens or reuses its issue/document tab and closes search.
- Escape dismisses search and cancels pending work.

### 8.1 Issue creation

Press `c` from Inbox or My Issues to open the retained creation flow. Tinear
loads workspace teams first: one team advances directly to the form, multiple
teams show a picker, and no teams reports an error. The form requires a title
and accepts an optional multiline description. `Tab` switches fields and
`Ctrl+Enter` creates the issue. Submission guards duplicate input and modal
dismissal; success shows a toast and opens the new issue tab.

---

## 9. Global keys (logged-in screen)

| Key | Action |
|---|---|
| `Ctrl+c` | Quit the app |
| `?` | Open search overlay (if no picker is open) |
| `Tab` | Next tab |
| `Shift+Tab` | Previous tab |
| `1` | Inbox tab |
| `2` | My Issues tab |
| `Escape` | Close modal, or close issue/document tab, or ignored |

---

## 10. Persistence

`models/cache.ard` and `services/cache_writer.ard` save and restore tab state:

- **Save**: tab open, close, and selection changes enqueue a `LoggedInCache`
  snapshot. The writer serializes and coalesces pending writes so an older
  snapshot cannot overwrite a newer one.
- **Load**: startup reads `~/.tinear/data.json`. Missing, malformed, unknown-version,
  or unknown-tab-kind data safely falls back to Inbox with no dynamic tabs.
- **Version**: schema version `2`, storing an active index and typed issue/document
  `TabRef` identities rather than stale full API records.
