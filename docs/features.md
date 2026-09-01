# tinear — feature reference

CLI / TUI for [Linear.app](https://linear.app). Two interfaces share the same
config and API layer:

- **CLI** (`ard run main.ard login`) — one-shot API key setup
- **TUI** (default, no command) — interactive terminal dashboard

---

## 1. Startup

1. `main()` dispatches to `commands/tui.ard` (default) or `commands/login.ard`.
2. Config loaded via `config.ard`:
   - `$LINEAR_API_KEY` env var (takes priority)
   - `~/.tinear/config.json` (fallback)
3. If no key found → render `welcome_screen` (login form).
4. If key found → validate with a lightweight query, then render `logged_in_screen`.

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
- **Login button**: calls `login::login(key)` which:
  1. Calls Linear GraphQL `viewer { id name displayName email }` to validate.
  2. On success: saves key to `~/.tinear/config.json`, transitions to
     logged-in screen.
  3. On failure: shows "Invalid API token" error below the button.
- **Quit**: `Ctrl+c` → exit.

---

## 3. Logged-in screen (tabbed shell)

After authentication, the app shows a persistent tabbed interface.

### 3.1 State shape

```ard
struct LoggedInUiState {
  tab: Str,                    // active tab: "Inbox" | "My Issues" | issue-identifier
  issue_tabs: [Issue],         // dynamically opened issue detail tabs
  active_picker: PickerProps?, // non-none when a picker overlay is open
  search_open: Bool,           // search overlay visible
}
```

### 3.2 Tab bar

Horizontal row of tabs across the top of the screen:

```
 Inbox  │  My Issues  │  ENG-123  │  ENG-456
```

- **Inbox** and **My Issues** are permanent (always present, leftmost).
- **Issue tabs** (e.g. `ENG-123`) are added dynamically when an issue is opened
  from Inbox, My Issues, or Search. Max one tab per issue (re-selecting an
  already-open issue switches to its existing tab).
- **Selected tab** is accent-colored and bold; inactive tabs are dim.
- Tab bar is a horizontal row followed by a dim rule. See
  [design-language.md](design-language.md) for the current visual standard.

### 3.3 Tab body

Below the tab bar, a single body region shows the active tab's content. Tabs are
mounted in an `indexed_stack` so inactive tab state survives switches.

| Tab | Content |
|---|---|
| **Inbox** | Scrollable notification list (see §4) |
| **My Issues** | Kanban board grouped by workflow state (see §5) |
| **Issue detail** | Full issue view (see §6) |

Footer bar below the body shows left-aligned, context-sensitive key hints without repeating the active tab title.

### 3.4 Tab switching

| Key | Action |
|---|---|
| `Tab` | Next tab (wrap around) |
| `Shift+Tab` | Previous tab (wrap around) |
| `1` | Jump to Inbox |
| `2` | Jump to My Issues |
| `Escape` | If picker open → close picker. If issue tab active → close it and switch to My Issues. Otherwise ignored. |

After switching tabs, focus is requested on the new tab's content via
`dispatch_after(ctx, 0, ...)` to ensure keyboard events route correctly.

### 3.5 Tab persistence

When tabs change (new issue opened, issue closed), the state is written to
`~/.tinear/cache.json` via `models/cache.ard`. On next launch, the cached
tab set and active tab are restored.

**Cache schema:**

```json
{
  "version": 1,
  "tab": "Inbox",
  "issue_tabs": [
    { "id": "...", "identifier": "ENG-123", "title": "...", ... }
  ]
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
  identifier: Str,      // e.g. "ENG-123"
  title: Str,
  description: Str?,
  state: WorkflowState, // { id, name, type, color }
  cycle: WorkflowState?,
  priority_label: Str,  // "Urgent", "High", "Medium", "Low", "No Priority"
  assignee_name: Str?,
  project_name: Str?,
  team_key: Str?,
  parent_issue: IssueRef?,
  labels: [Str],
  // ...
}
```

### 5.2 Board layout

Horizontal scroll of columns, one per workflow state. Columns are ordered by
`state_type_rank` (triage < backlog < unstarted < started < completed <
canceled). States from different teams merge by rank.

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
- Visible column count adapts to terminal width.

### 5.3 Keys

| Key | Action |
|---|---|
| `j` / `Down` | Move cursor down within column |
| `k` / `Up` | Move cursor up within column |
| `h` / `Left` | Move to previous column |
| `l` / `Right` | Move to next column |
| `Enter` | Open selected issue in a new tab |
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
appear centered over the current view. They trap focus and auto-focus the filter
input on open.

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

`models/cache.ard` handles saving and restoring tab state:

- **Save**: called after tab open, tab close, tab switch. Writes `LoggedInCache`
  (active tab + list of open issue/document tabs) to `~/.tinear/cache.json`.
- **Load**: called on `logged_in_screen` init. Reads cache file. On parse error
  or missing file, returns defaults (Inbox tab, no issue tabs).
- **Version**: schema version `1` — if future versions change the shape, old
  caches are discarded.
