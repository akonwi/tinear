# Filesystem Data Cache Plan

## Goal

Use `~/.tinear/data.json` as a local backup of user-visible TUI state so the
app can reopen quickly into the same useful state it had before closing, even
before the network refresh completes.

The cache is not the source of truth. Linear remains the source of truth. The
cache is a best-effort startup snapshot that is refreshed in the background.

## Current State

The typed model boundary is already in a good place for caching:

- `issues::Issue` no longer stores the original raw Linear `Dynamic` response.
- Picker construction uses typed fields through `tui/issue_picker.ard`.
- `models/issues.ard` remains the boundary that converts Linear GraphQL
  `Dynamic` responses into typed app data.

This means cached issues can be explicit typed structs instead of opaque raw
GraphQL payloads.

The UI state currently has three main cacheable owners:

- `LoggedInUiState` in `tui/logged_in_screen.ard`
- `InboxUiState` in `tui/inbox_view.ard`
- `MyIssuesUiState` in `tui/my_issues_view.ard`

Issue-detail comments live in each issue detail widget's local state. Cache those
later unless startup restoration feels incomplete without them.

## Cache File

Path:

```text
~/.tinear/data.json
```

Recommended module:

```text
models/cache.ard
```

It should use `config::get_config_dir()` or a new `config::get_data_path()` helper.

## Cache Data Shape

MVP cache shape:

```ard
struct DataCache {
  version: Int,
  logged_in: LoggedInCache,
  inbox: InboxCache,
  my_issues: MyIssuesCache,
}

struct LoggedInCache {
  tab: Str,
  issue_tabs: [issues::Issue],
}

struct InboxCache {
  items: [inbox::Item],
  cursor: Int,
}

struct MyIssuesCache {
  columns: [issues::BoardColumn],
  col: Int,
  row: Int,
  scroll: Int,
}
```

Start with `version = 1`. If parsing fails or the version is unknown, ignore the
cache and start empty.

## What to Persist

Persist durable, user-visible state:

- selected tab
- open issue tabs
- inbox items
- inbox cursor
- My Issues columns
- My Issues selected column/row
- My Issues horizontal scroll

Do not persist transient state:

- `loading` / `loaded` flags
- error strings
- `active_picker`
- `search_open`
- search query/results
- comment composer text
- compose sending state
- viewport width
- scroll controller handles

`active_picker` must not be cached because `picker::Props` contains callbacks.
Those callbacks are runtime closures, not data.

## Cache Module API

Use section-level helpers so each UI owner can save its slice without knowing the
whole app state.

Suggested public API:

```ard
fn load() DataCache?
fn save(cache: DataCache) Void!Str

fn load_logged_in() LoggedInCache?
fn load_inbox() InboxCache?
fn load_my_issues() MyIssuesCache?

fn save_logged_in(cache: LoggedInCache) Void!Str
fn save_inbox(cache: InboxCache) Void!Str
fn save_my_issues(cache: MyIssuesCache) Void!Str
```

The `save_*` helpers should do read-modify-write:

1. load existing `DataCache`, or create an empty default cache
2. replace only the target section
3. write the full `DataCache` back

This avoids `InboxView` overwriting `MyIssuesView` cache data and vice versa.

Writes are best-effort. A failed cache write should not disrupt the TUI.

## Startup Behavior

On app startup after auth:

1. Load cached values best-effort.
2. Use cached values as initial UI state if available.
3. Render immediately from cached data.
4. Start normal network refreshes.
5. Replace cached data with fresh Linear data when refresh succeeds.
6. Save fresh successful data back to cache.

If the cache is missing, corrupt, or outdated, ignore it and keep current network
loading behavior.

## Save Behavior

Save cache after meaningful state changes.

### Logged-in screen

Save after:

- selected tab changes
- issue tab opens
- issue tab closes
- issue tab is replaced after status/cycle update

Cache only:

```ard
LoggedInCache{tab: selected_tab, issue_tabs: issue_tabs}
```

Do not save `active_picker` or `search_open`.

### Inbox

Save after:

- inbox refresh succeeds
- inbox cursor changes
- archive reload succeeds

Cache only:

```ard
InboxCache{items: items, cursor: cursor}
```

Do not save detail modal state in the MVP. Reopening with the same inbox list and
cursor is enough for the first pass.

### My Issues

Save after:

- initial async load succeeds
- manual refresh succeeds
- background refresh succeeds
- selected row/column changes
- horizontal scroll changes
- status/cycle mutation reload succeeds

Cache only:

```ard
MyIssuesCache{columns: columns, col: col, row: row, scroll: scroll}
```

Do not save `active_picker`, `viewport_width`, loading flags, or errors.

## Hydration Details

### Logged-in screen

Current init starts as:

```ard
LoggedInUiState{tab: "Inbox", issue_tabs: [], active_picker: none, search_open: false}
```

Change it to read `cache::load_logged_in()` and initialize:

- `tab` from cache, defaulting to `"Inbox"`
- `issue_tabs` from cache, defaulting to `[]`
- `active_picker = none`
- `search_open = false`

Clamp/validate selected tab:

- `"Inbox"` and `"My Issues"` are always valid
- issue tab is valid only if its identifier exists in `issue_tabs`
- otherwise fall back to `"Inbox"`

### Inbox

Current init synchronously loads inbox from network.

Change initial state to:

- cached items/cursor if present
- `loaded = cached_items.size() > 0`
- `loading = false` initially if cached data exists
- then start a refresh after mount/background to get fresh Linear data

Avoid blocking startup on the network when cache exists.

### My Issues

Current init starts loading immediately and async-refreshes columns.

Change initial state to:

- cached columns/col/row/scroll if present
- `loaded = cached_columns.size() > 0`
- `loading = false` initially if cached data exists
- then keep the existing async load to refresh from Linear

Clamp `col`, `row`, and `scroll` after applying cached data.

## Atomicity

Prefer atomic writes:

1. write `data.json.tmp`
2. rename to `data.json`

If Ard filesystem APIs do not expose rename yet, use direct write for MVP and
add atomic writes later.

Because background refreshes can save while UI interactions also save, the MVP
should use read-modify-write section helpers. If races become visible, add a
small host-side file lock or route all cache writes through the UI event loop.

## Privacy and Account Safety

`data.json` may contain Linear issue titles, descriptions, inbox items, and open
issue tabs. Treat it as private local user data.

Rules:

- Do not store the API token in `data.json`.
- Store only app data and UI state.
- Future improvement: include `viewer_id` and ignore cache if a different user
  is authenticated.

For MVP, viewer validation can be deferred because the app currently stores one
local token. Add `viewer_id` before supporting easy account switching.

## Implementation Order

1. Add `models/cache.ard` structs and load/save helpers.
2. Add tests for missing/corrupt/versioned cache parsing.
3. Hydrate My Issues from cache.
4. Persist My Issues after successful refresh and navigation.
5. Hydrate inbox from cache.
6. Persist inbox after successful refresh and navigation.
7. Hydrate logged-in screen tab/open issue tabs from cache.
8. Persist logged-in screen after tab/open/close/replace changes.
9. Optional follow-up: cache issue detail comments.
10. Optional follow-up: add `viewer_id` ownership validation.
11. Optional follow-up: atomic writes or debounced writes if direct writes are noisy.

## Testing Plan

Add cache unit tests:

- missing `data.json` returns none/default
- corrupt JSON is ignored
- unknown version is ignored
- valid cache round-trips typed issues/inbox/my-issues data
- section save preserves other sections

Add UI tests where practical:

- My Issues initializes from cached columns without showing empty/loading state
- Inbox initializes from cached items and cursor
- Logged-in screen restores open issue tabs and selected tab

## Open Questions

- Should issue detail comments be included in MVP, or follow-up?
- Should cache writes be debounced immediately, or only if direct writes become noisy?
- Should `viewer_id` be included in v1 despite requiring an extra viewer fetch?
