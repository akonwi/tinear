# Filesystem Data Cache Plan

## Goal

Use `~/.tinear/data.json` as a local backup of user-visible TUI state so the
app can reopen quickly into the same useful state it had before closing, even
before the network refresh completes.

The cache is not the source of truth. Linear remains the source of truth. The
cache is a best-effort startup snapshot that is refreshed in the background.

## Typed Issue Boundary

`issues::Issue` should remain a fully typed app model. It should not carry the
original raw Linear GraphQL `Dynamic` response.

Picker construction now uses typed issue fields instead of re-decoding raw JSON:

- `issue.id`
- `issue.identifier`
- `issue.state_id`
- `issue.team_id`
- `issue.team_name`
- `issue.cycle_id`

`models/issues.ard` is the boundary that converts Linear `Dynamic` responses into
typed app data. After decoding, the rest of the app should use typed fields.

This keeps `data.json` cleaner and easier to version because cached issues are
explicit typed structs, not partially opaque raw GraphQL payloads.

## Cache File

Path:

```text
~/.tinear/data.json
```

Add path helpers near the existing config helpers, likely in `config.ard`:

```ard
fn get_data_path() Str!Str
```

Or introduce a dedicated cache module that uses `config::get_config_dir()`.

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

- loading flags
- error strings
- open picker modals
- comment composer text
- compose sending state
- viewport width
- scroll controller handles

## Startup Behavior

On app startup after auth:

1. Load `~/.tinear/data.json` best-effort.
2. Use cached values as initial UI state if available.
3. Render immediately from cached data.
4. Start normal network refreshes.
5. Replace cached data with fresh Linear data when refresh succeeds.

If the cache is missing, corrupt, or outdated, ignore it and keep current network
loading behavior.

## Save Behavior

Save cache after meaningful state changes:

- issue tab opened/closed/replaced
- selected tab changes
- inbox refresh succeeds
- inbox cursor changes
- My Issues refresh succeeds
- My Issues cursor/scroll changes

Writes should be best-effort. A failed cache write should not disrupt the TUI.

Prefer atomic writes:

1. write `data.json.tmp`
2. rename to `data.json`

If Ard filesystem APIs do not expose rename yet, use direct write for MVP and
add atomic writes later.

## Privacy and Account Safety

`data.json` may contain Linear issue titles, descriptions, inbox items, and open
issue tabs. Treat it as private local user data.

Rules:

- Do not store the API token in `data.json`.
- Store only app data and UI state.
- Future improvement: include `viewer_id` and ignore cache if a different user
  is authenticated.

## Implementation Order

1. Add cache structs and load/save helpers.
2. Add tests for missing/corrupt/versioned cache parsing.
3. Hydrate My Issues from cache.
4. Persist My Issues after refresh/navigation.
5. Hydrate inbox from cache.
6. Persist inbox after refresh/navigation.
7. Hydrate logged-in screen tab/open issue tabs from cache.
8. Persist logged-in screen after tab/open/close changes.
9. Optional follow-up: cache issue detail comments.

## Open Questions

- Should cache be per-account immediately via `viewer_id`, or deferred?
- Should cache writes be debounced, or is direct write on meaningful state change
  acceptable for now?
- Should issue detail comments be part of MVP, or a follow-up?
