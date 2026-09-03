# Issue Search UX Proposal

> **Status: implemented and superseded as an implementation plan.** The active
> Cooper controller searches both issues and documents, preserves partial
> results when one source fails, and opens typed dynamic tabs. Widget/state
> sketches below describe the original vaxis/ui proposal and are retained only
> as design history; see [features.md](features.md#8-search-overlay) for current
> behavior.

## Goal

Add a global issue search flow opened with `?`. Search should let users find
Linear issues by text from anywhere in the TUI and open a selected result as an
issue detail tab.

## Interaction Summary

- `?` opens search from Inbox, My Issues, or issue detail.
- Typing updates the query.
- Search should use Linear issue text search, not only currently loaded board
  data.
- Search requests should be debounced, e.g. 300ms.
- `↑` / `↓` moves the result cursor.
- `Enter` opens the selected issue using the existing issue-tab flow.
- `Esc` closes search and restores focus to the previously selected tab.

## Dialog Concept

Search is a centered modal overlay. It is transient: users type, choose a result,
and leave. It should not be a tab.

## Empty State

```text
┌──────────────────────────────────────────────┐
│ Search Issues                                │
│ ──────────────────────────────────────────── │
│ ?                                            │
│ ──────────────────────────────────────────── │
│                                              │
│          Type to search across issues        │
│                                              │
│ ──────────────────────────────────────────── │
│ ↵ open  ↑↓ select  Esc close                 │
└──────────────────────────────────────────────┘
```

## Loading State

```text
┌──────────────────────────────────────────────┐
│ Search Issues                                │
│ ──────────────────────────────────────────── │
│ dark mode                                    │
│ ──────────────────────────────────────────── │
│                                              │
│              Searching…                      │
│                                              │
│ ──────────────────────────────────────────── │
│ ↵ open  ↑↓ select  Esc close                 │
└──────────────────────────────────────────────┘
```

## Results State

```text
┌──────────────────────────────────────────────┐
│ Search Issues                                │
│ ──────────────────────────────────────────── │
│ dark mode                                    │
│ ──────────────────────────────────────────── │
│ ▸ FE-142   Add dark mode toggle      In Pro… │
│   FE-98    Dark mode color tokens    Done    │
│   BE-211   Dark mode API prefs       Backlog │
│   FE-301   Fix dark mode flicker     Todo    │
│                                              │
│ ──────────────────────────────────────────── │
│ ↵ open  ↑↓ select  Esc close                 │
└──────────────────────────────────────────────┘
```

## No Results State

```text
┌──────────────────────────────────────────────┐
│ Search Issues                                │
│ ──────────────────────────────────────────── │
│ xyzzyplugh                                   │
│ ──────────────────────────────────────────── │
│                                              │
│              No issues found                 │
│                                              │
│ ──────────────────────────────────────────── │
│ ↵ open  ↑↓ select  Esc close                 │
└──────────────────────────────────────────────┘
```

## Error State

```text
┌──────────────────────────────────────────────┐
│ Search Issues                                │
│ ──────────────────────────────────────────── │
│ dark mode                                    │
│ ──────────────────────────────────────────── │
│                                              │
│        Unable to search — try again          │
│                                              │
│ ──────────────────────────────────────────── │
│ ↵ open  ↑↓ select  Esc close                 │
└──────────────────────────────────────────────┘
```

## Result Row Layout

Rows are single-line and selected rows use reverse style, consistent with Inbox
and My Issues.

```text
[cursor] [identifier]  [title]                  [state]
```

Example:

```text
▸ FE-142   Add dark mode toggle          In Pro…
```

Suggested columns:

- cursor marker: `▸` or space
- identifier: fixed width, e.g. 9-10 cells
- title: flexible, truncated with ellipsis
- state: fixed width, truncated with ellipsis

## Footer Hints

Search dialog footer:

```text
↵ open  ↑↓ select  Esc close
```

Global footer additions:

- Inbox: `? search`
- My Issues: `? search`
- Issue detail: `? search`

## Integration with Existing Tabs

Selecting a result should call the existing issue tab opening flow:

1. If an issue tab for that issue already exists, select it.
2. Otherwise add a new issue tab.
3. Close search.
4. Restore focus to the selected issue tab with `dispatch_after`.

## Proposed Implementation Shape

### `models/issues.ard`

Add:

```ard
fn search(api_key: Str, query: Str) [Issue]!Str
```

It should use Linear `searchIssues(term: ..., first: ...)`, request the same
issue fields needed by existing issue detail and board rendering, and decode
results into `issues::Issue`.

### `tui/search_view.ard`

Add a new stateful search widget with state similar to:

```ard
struct SearchUiState {
  query: Str,
  results: [issues::Issue],
  cursor: Int,
  loading: Bool,
  error: Str,
  searched: Bool,
}
```

Callbacks:

```ard
on_open(ctx, issue)
on_dismiss(ctx)
```

### `tui/logged_in_screen.ard`

Add `search_open: Bool` to `LoggedInUiState`.

Add global shortcut:

```ard
"?": "global.search"
```

Render search as a centered overlay:

```ard
ui::overlay(
  child: base,
  overlay: search,
  position: ui::OverlayPosition::center,
  trap_focus: true,
  auto_focus: true,
)
```

## Notes

- Prefer debounced search to avoid flooding the Linear API.
- Enter should open selected result, not submit the query.
- Search should be server-side; filtering only loaded board issues would miss
  most workspace issues.
