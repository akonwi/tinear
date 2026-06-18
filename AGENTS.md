# tinear

## Agent rules

- **Do not run `git commit` without explicit direction from the user.**

CLI for [Linear](https://linear.app/) built with [Ard](https://ard.run). One
binary, one entrypoint, runs the interactive TUI; auth is handled inside the
TUI via the welcome screen rather than a separate `login` subcommand.

## Build, Run, Test

```bash
ard check main.ard               # type-check the entrypoint and everything it transitively touches
ard build main.ard --out tinear  # build the binary (gitignored)
ard run main.ard                 # build + run in one step
ard test                         # run unit tests under models/ and tui/
```

## Project Layout

| Path | Purpose |
|------|---------|
| `main.ard` | App entrypoint: AppState (`api_key`, toasts, `current_modal`), `show_modal`/`close_modal` callbacks, overlay slot wiring. Routes to `welcome_screen` if no API key, else `logged_in_screen`. |
| `config.ard` | Read/write `~/.tinear/config` and `$LINEAR_API_KEY`; resolve `~/.tinear` as the config directory. |
| `linear/client.ard` | Shared GraphQL client (`graphql(api_key, query)`). |
| `commands/login.ard` | API-key validation + save. Called by the welcome screen, not as a CLI verb. |
| `os.ard` + `ffi/host.go` | Tinear-local FFI for `open_url` (macOS/Linux/Windows). The only Go FFI tinear owns. |
| `models/*.ard` | Stateless fetch/decode/mutation modules for `inbox`, `issues`, and the persisted `cache`. |
| `tui/*.ard` | TUI implementation, split by concern. See *Active TUI Architecture*. |
| `ard.toml` | Points the `vaxis` dependency at `../vaxis-ard` (sister repo). |

vaxis/ui primitives are imported from external **[vaxis-ard](https://github.com/akonwi/vaxis-ard)** (Ard binding around `git.sr.ht/~rockorager/vaxis/ui`). Tinear used to keep a local copy under `tui/ui.ard` and a Go FFI under `ffi/`; both are gone. Tricky lessons live in **`vaxis-ard/docs/`** (events/focus, widget reconciliation) and should be the first read when something keyboard- or focus-related misbehaves.

## Active TUI Architecture

vaxis/ui primitives plus thin Ard composition. Do not reintroduce app-local
framework code; prefer adding to vaxis-ard if a primitive is missing.

| File | Purpose |
|------|---------|
| `tui/welcome_screen.ard` | Auth screen for unauthenticated users. Calls `commands/login::login` to validate and save the key. |
| `tui/logged_in_screen.ard` | Tab shell. Owns the `Model` (tab_index, issue_tabs), tab bar, `ui::indexed_stack` body, hints bar. Hydrates from `models/cache` on init; persists via a `set_and_persist` helper around tab/issue-tab mutations. |
| `tui/inbox_view.ard` | Inbox list + detail pane. Owns its own 5-minute background refresh with cursor-by-id preservation. |
| `tui/my_issues_view.ard` | Kanban board. Per-column `custom_scroll_view` + `sliver_list_builder` + `SliverListController.reveal_index` for auto-scroll. Horizontal outer scroll via `scroll_view(axis: horizontal)`. Same background-refresh pattern. |
| `tui/issue_detail_view.ard` | Issue detail tab. Sticky header (id, title, field grid), Description / Comments section tabs, scrollable body, compose modal entry, picker entry points, background refresh of both issue and comments. |
| `tui/search_view.ard` | Debounced global search modal body. Up/Down/Enter intercepted at the modal level so the text_field doesn't swallow them. |
| `tui/compose_view.ard` | Comment-compose modal body. `text_area` + capture-phase shortcuts for `Ctrl+Enter`. |
| `tui/picker.ard` | Reusable single-select picker (state/cycle pickers, etc.). |
| `tui/issue_pickers.ard` | State/cycle picker builders for an issue. Take a caller-supplied `on_change(new_id)` so each surface can choose what to refresh after the mutation. |
| `tui/modal.ard` | `show_modal(theme, title, child, footer, width, on_dismiss)` frame. Does **not** wrap `child` in `ui::focus`; the body must bring its own focusable so events bubble through the body's `shortcuts`/`actions` before hitting the frame. |
| `tui/refresh.ard` | `schedule_periodic(interval_ms, action)` — spawns one fiber per caller. Process-lived, no cancellation yet. |
| `tui/toast.ard` | Animated toast overlay. Variants (info / warning / error / success). Owned by `AppState`. |
| `tui/hints_bar.ard` | Footer hints. `new` adds a top divider; `row` is the bare row for embedding under a divider the caller already supplies (e.g. modal frames). |
| `tui/state.ard` | Workflow / priority helpers (rank, glyph, label). |
| `tui/text.ard` | Byte- and display-width string helpers. |

## vaxis/ui Best Practices

- Prefer native vaxis/ui composition (`ui::row`, `ui::column`, `ui::scroll_view`, `ui::custom_scroll_view`, `ui::list_tile`, `ui::actions`, `ui::shortcuts`, etc.) over app-specific rendering abstractions. If a primitive is missing, extend vaxis-ard.
- Keep Ard bindings (in vaxis-ard) thin: translate Ard nullable/default arguments into the FFI shape, expose typed handles (`ui::BuildContext`, `ui::Runtime`, `ui::State`, `ui::EventContext`).
- Separate keybindings from behavior:
  - `ui::shortcuts` maps keys to intent names.
  - `ui::actions` maps intent names to handlers with `ui::action(...)`.
  - The intent name shows up in profiles, error messages, and tests — name them after the *behaviour* (`board.row_down`, `issue_detail.compose`), not the key.
- `vaxis-ard/docs/events-and-focus.md` is the source of truth for event flow.
  Two pitfalls that bit tinear and are documented there:
  - **Inner `Shortcuts` can't rebind `Tab` / `Shift+Tab` / `Escape`.** Install handlers for the upstream intents (`vaxis.next-focus`, `vaxis.previous-focus`, `vaxis.dismiss`) in your `Actions` instead.
  - **Conditional `Actions` handlers can't "forward" with `Ignored`.** Once an `Actions` widget has a binding for an intent, dispatch stops there — `Ignored` does not bubble to outer Actions. If you want the outer handler to run sometimes, *omit the binding* (build the bindings list dynamically with `mut bindings`).
- Focus management with multiple mounted tab views:
  - `indexed_stack` keeps every child mounted. Their `focus_scope`s all rebuild
    on every cascade.
  - If you give the hidden tabs `auto_focus: true` / `reclaim_focus: true`, they
    will fight for focus on every rebuild and the last-rebuilt sibling wins.
  - **Plumb an `active: Bool` from the parent and pass it as both `auto_focus`
    and `reclaim_focus`.** Only the visible tab participates in focus
    management. This is the pattern in `inbox_view`, `my_issues_view`,
    `issue_detail_view`.
- Modals via the `tui/modal::show_modal` frame:
  - Caller's `child` must include a focusable. The frame deliberately does
    not wrap it in `ui::focus` because that would put the focusable above the
    body's `shortcuts`/`actions` and events would bubble past them.
  - For keys that the text input would otherwise swallow (Up/Down/Enter on the
    search modal, Ctrl+Enter on compose), bind them via the modal-level
    `ui::shortcuts` — Shortcuts dispatch in the capture phase, so it intercepts
    before the focused element's target phase.
- When a modal is open, the app-level shortcuts at `logged_in_screen` gate their actions on a `modal_open: Bool` plumbed from `main.ard` (`AppState.current_modal.is_some()`). Returning `Ignored` from those actions lets the keystroke continue through to the modal's text input. See `vaxis-ard/docs/events-and-focus.md §4` for why this works.
- Widget reconciliation pitfall (`vaxis-ard/docs/widget-reconciliation.md`):
  switching the outer widget type at a build slot between rebuilds (e.g.
  `ui::center(...)` while loading and `ui::column(...)` once loaded) silently
  drops the new tree. Wrap the conditional in a stable outer widget type so
  the slot's type never changes; let the conditional content live one level
  deeper.

## State and Async Patterns

Widget-local UI data lives in typed `ui::State` values. Stateless model modules perform fetch/decode/mutation; views store the results in state.

Use `ui::stateful(init: ..., build: ..., key: ...)`:

```ard
ui::stateful(
  key: "my-widget",
  init: fn(ctx: ui::BuildContext) MyState {
    let rt = ctx.runtime<MyState>()
    load(api_key, rt)
    MyState{ loading: true, items: [], error: "" }
  },
  build: fn(ctx: ui::BuildContext, state: ui::State<MyState>) ui::Widget {
    let model = state.value()
    ...
  },
)
```

Rules:

- Initial fetches go in `init`. There was a prior cargo-cult workaround that
  moved them into `build`; the real issue was the widget-reconciliation
  pitfall, not init-time dispatch. See the `init` comment in
  `tui/inbox_view.ard`.
- Background work must re-enter the UI runtime before mutating UI state:

  ```ard
  let _ = async::start(fn() {
    let result = load()
    rt.dispatch(fn(s: ui::State<MyState>) {
      s.set(fn(mut next: MyState) {
        match result {
          ok(items) => { next.items = items; next.error = "" },
          err(e)    => { next.error = e },
        }
        next.loading = false
      })
    })
  })
  ```

- Prefer typed handles (`ui::State<MyState>`, `ui::Runtime<MyState>`) and read
  with `state.value()` / `s.value()` for immutable rendering snapshots.
- Use stable `key:` values for statefuls that must survive reordering or tab
  switching.
- `state.set(mut)` is synchronous (updates the cached value before returning),
  so `state.value()` immediately after returns the new value. Useful for
  post-mutation persistence helpers like `set_and_persist` in
  `tui/logged_in_screen.ard`.

### Periodic background work

`tui/refresh::schedule_periodic(interval_ms, action)` spawns a fiber that
sleeps then runs `action` forever. Used by inbox, my-issues, and each issue
detail tab for the 5-minute auto-refresh. Fibers are process-lived and not
cancellable yet — `rt.dispatch` against a disposed state is a safe no-op per
the FFI, so closed tabs continue polling until app exit (wasteful but harmless).

## Testing

Currently no widget tests — testing is pure unit coverage:

| Test file | Covers |
|-----------|--------|
| `models/cache_test.ard` | Versioned cache round-trip + invalid-input handling. |
| `models/issues_test.ard` | Mojibake normalization in issue/comment decode. |
| `tui/state_test.ard` | Priority + workflow-state-type ordering helpers. |
| `tui/text_test.ard` | Display-width truncation. |

Run with `ard test`. If we add widget tests later, wire them through
`vaxis-ard`'s uitest helpers; we no longer ship a tinear-local test binding.

## Code Style

- Name functions for UI intent (`issue_card`, `refresh_inbox`, `open_state_picker`) over mechanics.
- Use named arguments when calling a function with three or more arguments.
- Model modules own fetch / decode / mutation. TUI modules own presentation and state wiring.
- Keep errors user-facing. `main()` returns `Void`; surface failures via toasts or panics with a clear message.
- Linear auth header is `Authorization: <raw key>` — no `Bearer` prefix.
- Use `models/inbox::optional_field` (and `models/issues` helpers) for nullable nested JSON fields; GraphQL inline fragments often omit them entirely.
- Use display-width-aware text helpers (`tui/text`) when fitting strings into fixed terminal cells. Byte-based helpers are ASCII-only.
- Auto-refresh handlers should be **silent** (no `loading = true` flash) and **preserve cursor by id** (look up the previously-cursored entity's id after the swap and snap to its new index). See `refresh_inbox` / `refresh_board` for the pattern.

## Auth and CLI surface

Single entrypoint (`main.ard`) launches the TUI. There is no `login` subcommand any more; auth happens in the welcome screen, which delegates to `commands/login::login`. To bootstrap without the TUI, set `$LINEAR_API_KEY` or hand-edit `~/.tinear/config`.

## Ard Language Notes

- **No `return` keyword** — last expression is the return value. Use `try` for Result propagation.
- **Functions must be defined before use** within a file. Cross-file order doesn't matter.
- **`and` / `not`** instead of `&&` / `!`.
- **No `!=`**. Use `not (a == b)`.
- **`Void!Str`** uses `Result::ok(())` for the Ok variant.
- **`{ ... }` in `match`** must be followed by a newline — no inline `match x { a => b, _ => c }`.
- **String interpolation** with `{var}`; literal braces need `\{` / `\}`.
- **`use ard/list as List`** to access `List::drop()`.
- **Mutating methods** with `fn mut name() { self.field = v }`. Methods can mutate fields when the receiver is `mut`.
- **`mut` doesn't propagate through match/for bindings** — `mut x = ...; for item in xs { item.field = v }` doesn't compile. Hoist a copy:
  `mut t = xs.at(i); t.field = v; xs.set(i, t)`.
- **`while true { ... }`** is the idiomatic infinite loop (used by `schedule_periodic`).
- **Nullable callback gotcha**: `fn(X) Void?` is **not** a nullable function — the `?` binds to the return type. For a nullable callback, omit the return type: `on_pressed: fn(EventContext)?`.
- **Generic struct definitions** (e.g. `struct Box<$T> { ... }`) are tricky in the current parser. If a stateful needs to vary by a type parameter from the parent, keep the state struct concrete and have the parent close over its own typed runtime in a callback (see how `tui/compose_view` exposes `refresh_comments: fn()` instead of being generic over Detail's state type).
- **`async::start` returns a `Fiber`**. Discard it with `let _ = async::start(...)` when the surrounding expression's expected return type is `Void`.

## vaxis-ard companion docs

Read these before debugging anything keyboard- or focus-related:

- `vaxis-ard/docs/events-and-focus.md` — event dispatch (capture / target / bubble), Shortcuts/Actions semantics, focus_scope with `auto_focus` / `reclaim_focus`, conditional-handler / Ignored pitfall.
- `vaxis-ard/docs/widget-reconciliation.md` — the "state updates but the screen doesn't" trap, plus the sibling FFI bug where `uiStatefulState` cached the widget at `CreateState` and never refreshed parent-supplied params (fixed; explained as a misdiagnosis-prone class of bugs).
