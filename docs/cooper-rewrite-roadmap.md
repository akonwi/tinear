# Cooper Rewrite Parity Roadmap

> Living implementation checklist for the Cooper rewrite on `rewrite/cooper`.
>
> Behavioral parity reference: `main` at `84ff69a`.
>
> Update this document as phases land, scope changes, or parity waivers are accepted.

This roadmap tracks behavior that is reachable on `main`. Working principles,
architectural guidance, and session-specific reasoning stay in the active Kit
scratchpad; durable architectural decisions should be recorded as ADRs.

## Status snapshot

- **Current implementation focus:** Phase 9 parity hardening and the remaining explicit Phase 0/1 framework and transport gates.
- **Feature slices complete:** authentication, shell/cache, Inbox, My Issues, field picker/mutation flows, Markdown, issue details/comments/editing, document details, global search, and issue creation.
- **Open framework/foundation gates:** remotely pin Cooper, terminal-title parity, optional terminal-browser app mode, transport injection/limits, fallible decoder cleanup, and true HTTP cancellation.
- **Latest validation:** `ard check main.ard`, `ard build main.ard`, 121 Ard tests, and Go Markdown parser tests pass.

## Parity ledger

### Application, authentication, and platform behavior

- [x] One no-argument TUI entrypoint.
- [x] `$LINEAR_API_KEY` takes precedence over `~/.tinear/config.json`.
- [x] Missing credentials show the welcome/login form.
- [x] Login validates the key with Linear's `viewer` query, persists it, and transitions in place.
- [x] `Ctrl+C` cleanly destroys the application.
- [x] `Super+C` copies the current Cooper selection through OSC 52.
- [x] URLs prefer `terminal-browser open --split right --size 0.55`, then fall back to the system opener.
- [ ] Restore optional terminal-browser app-mode probing where requested.
- [x] Semantic foreground/background, muted, accent, danger, warning, and success presentation is available.
- [x] Toasts support info/success/error variants and deterministic configurable expiry.
- [x] At most one modal is active; it blocks background input and restores focus on close.
- [ ] Inbox count updates the terminal title, or receives an explicit temporary parity waiver until Cooper exposes a title API.

### Logged-in shell and persistence

- [x] Permanent Inbox and My Issues tabs.
- [x] Dynamic issue and document tabs, deduplicated by `(kind, id)`.
- [x] Inactive tab controllers stay mounted and preserve local state.
- [x] `Tab` / `Shift+Tab` cycle tabs.
- [x] `1` and `2` jump to Inbox and My Issues.
- [x] Mouse activation selects tabs.
- [x] `Escape` closes the active dynamic tab and returns to My Issues.
- [x] Document tab titles are cached, truncated for display, and updated after fetch.
- [x] `~/.tinear/data.json` schema v2 persists active tab index plus issue/document `TabRef`s.
- [x] Corrupt, unknown-version, unknown-kind, or out-of-range cache data safely falls back.
- [x] Closing a dynamic tab disposes controls, listeners, requests, and periodic refresh work.

### Inbox

- [x] Load Linear issue and pull-request notifications.
- [x] Render a navigable list pane and scrollable detail pane.
- [x] Preserve issue/PR-specific detail metadata, comments, history, and triggering-comment presentation used by `main`.
- [x] `j`/Down and `k`/Up move the cursor and reveal it.
- [x] `h`/Left, `l`/Right, Space, and Shift+Space scroll/page the detail pane.
- [x] Enter opens an issue tab or the external target for non-issue notifications.
- [x] `o` opens the selected item in the browser.
- [x] Backspace optimistically archives the notification and reports failure via toast.
- [x] `r` manually refreshes.
- [x] Detail requests reject stale responses after cursor changes.
- [x] Five-minute silent refresh preserves the selected notification by ID.
- [x] Empty, loading, and error states match current behavior.

### My Issues board

- [x] Load the viewer and assigned issues, excluding canceled work.
- [x] Group and order columns by workflow rank/position.
- [x] Hide the Duplicate state and cap completed columns as `main` does.
- [x] Render horizontal columns with per-column vertical scrolling.
- [x] Render priority glyph, identifier, wrapped title, and cycle/project metadata.
- [x] `j`/Down, `k`/Up, `h`/Left, and `l`/Right navigate cards and skip empty columns.
- [x] Cursor movement reveals the active card and horizontally reveals its column.
- [x] Enter/mouse opens an issue tab.
- [x] `r` manually refreshes.
- [x] Five-minute silent refresh preserves the selected issue by ID.
- [x] `s`, `y`, `p`, `a`, and Shift+`p` edit state, cycle, priority, assignee, and project.
- [x] Successful mutations refresh the board without losing cursor identity.

### Issue and document details

- [x] Issue header shows identifier/title, state, priority, assignee, cycle, project, and team.
- [x] Description and Comments sections are switchable by click and `d`/`c`.
- [x] Issue body scrolls and remains selectable.
- [x] Comments are threaded one level deep, preserve bot/user attribution, and expose a cursor.
- [x] `j`/`k` navigate comments when Comments is active; arrows scroll.
- [x] `n` composes a top-level comment.
- [x] `r` replies to the selected comment's thread root.
- [x] `e` edits the description in a multiline editor; Ctrl+Enter/Ctrl+M/Ctrl+J save.
- [x] Description save guards against stale periodic refresh results.
- [x] `o` opens the issue/document in the browser.
- [x] Issue details expose all five field pickers.
- [x] Issue and comment loads have independent loading/error handling.
- [x] Document tabs render metadata and Markdown content, retitle themselves, and refresh silently every five minutes.
- [x] Closing an issue/document tab stops its refresh loop and suppresses late completions.

### Markdown

- [x] Retain the Goldmark CommonMark+GFM semantic parser and Go tests.
- [x] Map headings, paragraphs, emphasis, strike, inline code, links, images, lists, task lists, quotes, rules, fenced code, tables, and mermaid labels to retained Cooper controls.
- [x] Preserve clickable OSC-8 links while intercepting activation when terminal-browser routing is required.
- [x] Preserve selection across rendered Markdown text.
- [x] Keep syntax highlighting out of parity scope.

### Pickers, search, comments, and issue creation

- [x] Reusable modal single-select picker starts on the current value.
- [ ] Match `main`'s picker filter-visibility threshold for lists of five or fewer options.
- [ ] Confirm exact 12-result picker viewport parity; the retained picker is bounded and reveals keyboard selection, but its input/divider currently share the 12-row body.
- [x] Picker supports keyboard and mouse activation plus null/unassigned options where applicable.
- [x] `?` opens global issue/document search from logged-in views, including terminal-normalized Shift+`/` events.
- [x] Search debounces for 300 ms and rejects stale generations.
- [x] Issue results appear before document results.
- [x] One failed search source yields a warning with surviving results; both failing yields an error.
- [x] Search supports loading, empty-query, no-results, and error states.
- [x] Enter/mouse opens or reuses the correct dynamic tab and closes search.
- [x] `c` from Inbox/My Issues starts issue creation.
- [x] Creation loads teams: zero reports failure, one advances directly, many show team selection.
- [x] Creation requires a title, accepts an optional multiline Markdown description, and blocks duplicate/dismissed in-flight submission.
- [x] Successful creation shows a toast and opens the new issue tab.

## Delivery roadmap

**Current focus:** Phase 9 parity hardening, including the remaining explicit Phase 0/1 framework and transport gates.

### Phase 0 — Contract and Cooper integration gates

- [ ] Pin the exact Cooper revision used by the rewrite once it is remotely available; use the local path during active joint development.
- [x] Add an ADR superseding ADR 0002's vaxis/ui implementation decision with direct Cooper retained controls.
- [x] Decide app-local semantic palette behavior while retaining terminal-default foreground/background.
- [ ] Decide terminal-title parity: add a Cooper capability or document a temporary waiver.
- [x] Prove modal layering, mouse barrier, focus restoration, and selection copy in TestApp fixtures.
- [x] Prove retained hidden-tab roots and nested two-axis scrolling in shell/board fixtures.
- [x] Define common controller lifecycle and request-generation conventions in working guidance and `tui/lifecycle.ard`.

**Exit criteria:** architectural gaps have explicit decisions; no unsupported Cooper API is required.

### Phase 1 — Data and service foundation

Candidate modules:

```text
config.ard
linear/client.ard
models/tabs.ard
models/cache.ard
models/inbox.ard
models/issues.ard
models/documents.ard
platform/url.ard
ffi/markdown/*
```

- [x] Re-add `decode` as a direct Ard dependency.
- [x] Restore config path/load/write behavior.
- [x] Restore the Linear GraphQL client with timeout and GraphQL error extraction.
- [x] Restore typed tabs and shell cache v2.
- [x] Restore inbox, issue/comment/board/picker/create, and document models.
- [x] Move workflow/priority model helpers out of the old TUI dependency direction.
- [x] Restore the markdown parser as the only necessary Go-side app FFI.
- [x] Restore safe, non-blocking terminal-browser URL routing without the obsolete Go process shim.
- [ ] Restore optional terminal-browser app-mode probing without blocking the UI thread.
- [x] Introduce injectable service closures for controller tests.
- [x] Encode all GraphQL string arguments safely rather than interpolating IDs.
- [ ] Make network response decoders fallible instead of panicking on malformed or permission-dependent payloads.
- [ ] Inject/test the HTTP transport, status/error matrix, and response-size boundary.
- [ ] Thread per-controller cancellation into HTTP request contexts so closed views interrupt in-flight requests.
- [x] Make config/cache directories and files private and replace files atomically.
- [x] Serialize/coalesce shell cache writes so older snapshots cannot overwrite newer state.
- [x] Fix board/picker source defects while porting: fetch assignee/project IDs and exclude canceled projects correctly.

**Exit criteria:** all model/cache/markdown tests from `main` are restored or replaced with equivalent coverage; model modules have no Cooper imports.

### Phase 2 — Root, lifecycle, modal, toast, and command infrastructure

Candidate modules:

```text
tui/app_controller.ard
tui/lifecycle.ard
tui/modal_host.ard
tui/toast_host.ard
tui/hints_bar.ard
tui/theme.ard
```

- [x] Build a permanent app root with content, modal, and toast planes.
- [x] Implement app-level key routing by active view and interaction mode.
- [x] Implement modal ticket/generation ownership, background mouse blocking, guarded dismiss, and focus restore.
- [x] Implement toast stacking, variants, TTL cancellation, and disposal.
- [x] Implement Super+C using `App.selection()` and `Context.clipboard.write()`.
- [x] Add TestApp coverage for modal isolation, toast expiry, selection copy, and teardown.

**Exit criteria:** infrastructure can host either welcome or logged-in content without reconstructing the App.

### Phase 3 — Welcome and authentication

- [x] Build retained logo, API-key Input, login action, loading state, and inline error.
- [x] Validate asynchronously without blocking the UI thread.
- [x] Persist a valid key and transition to the shell.
- [x] Ensure dismissal/quit during validation suppresses late completion.

**Exit criteria:** TestApp covers missing config, env/config precedence, invalid key, valid key, and clean shutdown.

### Phase 4 — Shell, tabs, focus, and shell cache

Candidate modules:

```text
tui/shell_controller.ard
tui/tab_bar.ard
```

- [x] Restore pure tab reducers/tests first.
- [x] Construct Inbox and My Issues controllers once.
- [x] Add dynamic issue/document controller creation and identity reuse.
- [x] Toggle inactive roots with `Display::none` and focus the active view explicitly.
- [x] Implement keyboard/mouse tab navigation and dynamic close behavior.
- [x] Hydrate/persist shell cache, including document retitling.
- [x] Serialize/coalesce persistence so older snapshots cannot overwrite newer state.

**Exit criteria:** TestApp covers open/reuse/switch/close/restore, hidden-state retention, focus restoration, malformed cache, and closed-tab disposal.

### Phase 5 — Inbox vertical slice

- [x] Build retained list rows and split detail layout.
- [x] Implement loading/error/empty states and initial load.
- [x] Implement cursor navigation/reveal and detail scrolling.
- [x] Implement stale-safe detail requests.
- [x] Implement issue/PR opening, terminal-browser routing, optimistic archive, and manual refresh.
- [x] Add five-minute silent refresh with cursor-by-ID preservation.
- [ ] Add terminal-title count behavior when the framework gate is resolved.

**Exit criteria:** Inbox is a complete network-backed daily-use slice with deterministic tests for stale requests, deletion, refresh preservation, routing, and disposal.

### Phase 6 — My Issues board and reusable pickers

- [x] Build an outer horizontal `ScrollBox` containing a retained row of columns.
- [x] Give each column its own vertical `ScrollBox` and retained card records.
- [x] Reconcile columns/cards by stable IDs and use `scroll_child_into_view`.
- [x] Restore card rendering, ordering, navigation, opening, manual refresh, and periodic refresh.
- [x] Build the app-local filterable picker controller.
- [x] Restore state, cycle, priority, assignee, and project pickers/mutations.
- [x] Benchmark the bounded 200-issue workload before considering app-local virtualization.
  - `ard run benchmarks/board_200.ard`: 200 cards / 10 columns measured ~6.5 ms construction, ~6.6 ms initial render, ~7.0 ms per rendered navigation, and ~10.2 ms narrow resize on the development machine. Keep eager retained cards; virtualization is not warranted yet.

**Exit criteria:** board navigation, nested scrolling, cursor preservation, every picker, and every mutation path pass TestApp coverage at narrow and wide sizes.

### Visual baseline checkpoint — Quiet Structure

The agreed visual system is recorded in `docs/design-language.md` and ADR 0004.
Complete this pass before resuming feature phases so new screens inherit a
coherent language instead of extending the bare parity styling.

- [x] Establish a host-theme foundation using terminal-default foreground/background and indexed ANSI semantic colors.
- [x] Restyle shell tabs and footer with accent hierarchy and dim rules.
- [x] Restyle Inbox rows with hover/active surfaces and render detail as structured styled content.
- [x] Restyle board headers/cards with comfortable spacing, priority color, metadata, and hover/active surfaces.
- [x] Apply the shared list language to picker/search results.
- [x] Align modals and toasts with the terminal-default canvas.
- [ ] Add TestApp coverage at narrow, standard, and wide sizes for selected, focused, loading, empty, and error states.

**Exit criteria:** shell, Inbox, My Issues, picker, modal, and toast screenshots/frames follow the design guide and have no full-row reverse selection.

### Phase 7 — Markdown, issue detail, comments, and document detail

- [x] Map parsed markdown blocks/runs to persistent Cooper Box/Text controls.
- [x] Build issue header, section tabs, body scrolling, and selection.
- [x] Load issue/comments independently with lifecycle guards.
- [x] Restore comment threading/cursor, top-level compose, and replies.
- [x] Restore description editing/saving with stale-refresh protection.
- [x] Wire issue field pickers and browser opening.
- [x] Build document tabs with metadata and Markdown content, retitling, browser opening, and refresh.
- [x] Verify every per-tab worker stops on close.

**Exit criteria:** Markdown parser/render fixtures and complete issue/document interaction flows pass TestApp coverage.

### Phase 8 — Global search and issue creation

- [x] Build search modal Input/results viewport.
- [x] Restore debounce, generation checks, merged issue/document results, partial-failure warning, navigation, and open behavior.
- [x] Build issue creation as one modal-owned retained state machine.
- [x] Restore team selection, title/description form, submission chords, guarded dismissal, toasts, and tab opening.

**Exit criteria:** deterministic tests cover stale search, all search states, zero/one/many teams, validation, duplicate submission prevention, dismissal races, and successful creation.

### Phase 9 — Parity hardening and release gate

- [ ] Port or replace all pure tests from `main`.
- [ ] Add fake-Linear TestApp scenarios for every async state transition.
- [ ] Add resize coverage for narrow, standard, and wide terminals.
- [ ] Prove hidden/closed views receive no commands and apply no late results.
- [ ] Add PTY coverage for startup/restoration, key normalization, mouse/wheel, multiline paste, OSC 52 copy, and clean quit.
- [ ] Profile 50 notifications, 200 board issues, 50 comments, multiple dynamic tabs, and repeated refreshes.
- [x] Verify there are no `vaxis/ui`, `ffi/stateful`, or intent-shim imports.
- [ ] Verify every post-start control mutation occurs inside a Cooper callback or `Context.dispatch`.
- [ ] Run `ard check main.ard`, `ard build main.ard`, `ard test`, Go parser tests, and PTY smoke tests.
- [ ] Restore/update native multi-platform CI and release tooling for Cooper's CGO requirements.
- [ ] Update README/features docs and mark superseded implementation details accurately.

**Final parity gate:** every item in the parity ledger is automated, manually verified, or has an explicit documented waiver.

## Explicitly out of parity scope

These are documented ideas, not behavior that the rewrite must reproduce before cutover:

- [ ] Projects as a permanent tab
- [ ] `:` command palette
- [ ] Separate `login` CLI command / old `commands/*` architecture
- [ ] Full inbox/board/offline cache from `docs/data-cache-plan.md`
- [ ] Recent-documents list
- [ ] Labels display/editing
- [ ] Editing comments
- [ ] Filtered boards
- [ ] Syntax highlighting
- [ ] Mark-all-read and snooze
- [ ] Global `R` refresh
- [ ] Issue relations
- [ ] Document creation/editing

## Known framework/app risks to track

- [ ] Cooper has no public terminal-title API.
- [ ] Cooper has indexed host-palette colors but no holistic terminal-queried, contrast-derived semantic theme service.
- [ ] Cooper has no general app-facing modal/focus-trap control; Tinear must own modal policy.
- [ ] Cooper has no always-expanded ListBox/Combobox; Tinear needs a retained picker/search list.
- [ ] Cooper has no built-in virtualization; use bounded eager retained rows first and profile.
- [ ] Cooper has no animation/frame API; toast TTL parity comes before slide animation parity.
- [ ] App-scoped dispatch outlives individual controllers; every late completion needs controller disposal/generation checks.
- [ ] Automatic ScrollBox bars reserve cells and can change parity-sensitive wrapping/layout.
- [ ] Current Cooper is locally pinned; shared development requires a remotely reachable commit/tag.
