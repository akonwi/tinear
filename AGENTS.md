# tinear

## Agent rules

- **Do not run `git commit` without explicit direction from the user.**

CLI for [Linear](https://linear.app/) built with [Ard](https://ard.run). One
binary, one entrypoint, runs the interactive TUI; auth is handled inside the
TUI via the welcome screen rather than a separate `login` subcommand.

The UI is built **directly against the Go
[vaxis/ui](https://git.sr.ht/~rockorager/vaxis) widget framework** via Ard's
direct Go interop (`use go:go.rockorager.dev/vaxis/ui`). There is no binding
layer: widgets are Go struct literals, theme fields are Go fields, and the
only Go code tinear owns is the small shim package under `ffi/`.

## Build, Run, Test

```bash
ard check main.ard               # type-check the entrypoint and everything it transitively touches
ard build main.ard --out tinear  # build the binary (gitignored)
ard run main.ard                 # build + run in one step
ard test                         # run all unit tests
```

## Project Layout

| Path | Purpose |
|------|---------|
| `main.ard` | App entrypoint: AppState (`api_key`, toasts, `current_modal`), dispatch-safe `show_modal`/`close_modal`/`notify`, always-mounted overlay. Routes to `welcome_screen` if no API key, else `logged_in_screen`. Ctrl+C quits; Super+c (Cmd+C, where the terminal forwards it) invokes the copy-selection intent so the detail-pane SelectionArea copies the active selection. |
| `config.ard` | Read/write `~/.tinear/config.json` and `$LINEAR_API_KEY` via `go:os` + `go:encoding/json`. |
| `decode.ard` | Composable JSON decoding over opaque `Any`: `from_json`, `run`, primitives, `nullable`/`list`/`field`/`path` combinators with path-carrying errors. |
| `linear/client.ard` | Shared GraphQL client (`graphql(api_key, query)`) over `go:net/http` with a 30s client timeout and GraphQL-error extraction. |
| `os.ard` | `open_url` wrapper over `ffi::OpenURL`. |
| `ffi/*.go` | Tinear's Go shim package, imported as `use go:tinear/ffi`. See *The ffi shim*. |
| `models/*.ard` | Stateless fetch/decode/mutation modules for `inbox`, `issues`, and the persisted `cache`. |
| `tui/*.ard` | TUI implementation, split by concern. See *Active TUI Architecture*. |

## The ffi shim

`ffi/` holds only what Ard cannot express, plus stopgaps for tracked
compiler gaps. Keep it minimal; prefer direct `use go:` calls.

| File | Contents |
|------|----------|
| `ffi/stateful.go` | `Stateful` widget (owns the `ui.StateBase` struct embedding Ard can't do), `StateCtx`, generic `Update[T]` (mutate live state + rebuild), `StateRef[T]`/`StateValue[T]`, `MarkDirty`, `Dispatch` (posts onto the UI thread), optional `OnTick`, and the per-widget `Animate`/`AnimateMs` animation controller. |
| `ffi/intents.go` | `Intent(t)` — wraps a string as a `ui.Intent` (Ard can't implement a Go interface method on a string newtype). |
| `ffi/host.go` | `OpenURL` (platform browser launch), plus stopgaps: `StrFromBytes` (ard#283) and `Millis` (ard#284). Delete each when its upstream issue lands. |

## Active TUI Architecture

Direct vaxis/ui composition plus thin Ard modules. Do not reintroduce a
binding/framework layer; if something is genuinely inexpressible, add the
smallest possible shim to `ffi/`.

| File | Purpose |
|------|---------|
| `tui/welcome_screen.ard` | Auth screen for unauthenticated users. Validates the API key against Linear's `viewer` query and persists it via `config::write`. |
| `tui/logged_in_screen.ard` | Tab shell. Owns the `Model` (tab_index, issue_tabs), tab bar, `IndexedStack` body, hints bar, and the `?` global-search modal. Hydrates from `models/cache` on first build; persists via `select_tab`/`replace_model` (`ffi::Update` + async persist). |
| `tui/inbox_view.ard` | Inbox list + detail pane split. Optimistic delete, detail fetch keyed by cursor with stale-result discard. |
| `tui/my_issues_view.ard` | Kanban board. Per-column `CustomScrollView` + `SliverListBuilder` + `SliverListController.RevealIndex` for auto-scroll. Horizontal outer `ScrollView`, metric-driven overflow chevrons. 5-minute silent refresh preserving cursor by issue id. |
| `tui/issue_detail_view.ard` | Issue detail tab. Sticky header (id, title, field grid), Description / Comments section tabs, scrollable body, compose modal entry (`n`), picker entry points (`s`/`y`), background refresh of both issue and comments. |
| `tui/search_view.ard` | Debounced (300ms, generation-guarded) global search modal body. Up/Down/Enter intercepted at the modal level so the TextField doesn't swallow them. |
| `tui/compose_view.ard` | Comment-compose modal body. `TextArea` + capture-phase shortcuts for `Ctrl+Enter`/`Ctrl+m`/`Ctrl+j`. |
| `tui/picker.ard` | Reusable single-select picker (state/cycle pickers, etc.). |
| `tui/issue_pickers.ard` | State/cycle picker builders for an issue. Take a caller-supplied `on_change(new_id)` so each surface can choose what to refresh after the mutation. |
| `tui/modal.ard` | `show_modal(theme, title, child, footer, width, on_dismiss)` frame. Does **not** wrap `child` in a focusable; the body must bring its own so events bubble through the body's `Shortcuts`/`Actions` before hitting the frame. |
| `tui/refresh.ard` | `schedule_periodic(interval_ms, action)` — spawns one fiber per caller. Process-lived, no cancellation yet. |
| `tui/toast.ard` | Animated toast overlay. Variants (info / success / error). Data owned by `AppState`; each toast owns an `AnimationController` over its TTL. |
| `tui/hints_bar.ard` | Footer hints. `new` adds a top divider; `row` is the bare row for embedding under a divider the caller already supplies (e.g. modal frames). |
| `tui/state.ard` | Workflow / priority helpers (rank, glyph, label). |
| `tui/text.ard` | Byte- and display-width string helpers. |

## vaxis/ui Best Practices

- Compose Go widgets directly: `ui::Flex{...}`, `ui::Text{...}`,
  `ui::ListTile{...}`, `ui::ScrollView{...}`, `ui::Actions{...}`,
  `ui::Shortcuts{...}`, etc. Constructor helpers (`ui::Center`,
  `ui::Padding`, `ui::Expanded`, `ui::DecoratedBox`, `ui::BorderAll`) are
  plain Go functions.
- **`ui::Widget` is Go `any`**, so a widget-list literal needs an explicit
  annotation to box element-wise: `mut children: [ui::Widget] = [...]`.
- **`ui::Symmetric(horizontal, vertical)`** — horizontal first. Getting this
  backwards silently breaks layouts (it did during the port).
- **TextField/TextArea size themselves from `MinWidth`** — wrapping in an
  outer `SizedBox` does not grow the editable region.
- Scroll/list controllers are Go pointer types held as `mut ui::X` fields:
  create with `mut ctrl = ui::ScrollController{}` then store `mut ctrl`
  (direct `mut ui::X{}` literals are blocked on ard#285).
- Separate keybindings from behavior:
  - `ui::Shortcuts{Bindings: ["j": ffi::Intent("board.row_down")], ...}`
    maps keys to intents (Str literals widen into `ui::IntentType` keys).
  - `ui::Actions{Bindings: ["board.row_down": fn(ev, intent) {...}], ...}`
    maps intents to handlers.
  - Name intents after the *behaviour* (`board.row_down`,
    `issue_detail.compose`), not the key.
- **Inner `Shortcuts` can't rebind `Tab` / `Shift+Tab` / `Escape`.** Install
  handlers for the upstream intents (`vaxis.next-focus`,
  `vaxis.previous-focus`, `vaxis.dismiss`) in your `Actions` instead.
- **Conditional `Actions` handlers can't "forward" with `Ignored`** to outer
  Actions — dispatch stops at the first binding. But returning `Ignored`
  from a capture-phase `Shortcuts` action DOES let the event continue to the
  focused element's target phase; that's how `1`/`2`/`?` type into modal
  text fields when `modal_open` is true.
- Focus management with multiple mounted tab views:
  - `IndexedStack` keeps every child mounted; hidden tabs must not pull
    focus. **Plumb an `active: Bool` from the shell and pass it as both
    `AutoFocus` and `ReclaimFocus`** on the tab's `FocusScope`. This is the
    pattern in `inbox_view`, `my_issues_view`, `issue_detail_view`.
- Modals via the `tui/modal::show_modal` frame:
  - Caller's `child` must include a focusable (a `ui::Focus(mut node, ...)`
    wrapper, a `ListTile` with `OnPressed`, a text input, ...). The frame
    deliberately does not add one.
  - For keys a text input would otherwise swallow (Up/Down/Enter in search,
    Ctrl+Enter in compose), bind them via the modal body's `Shortcuts` —
    capture phase intercepts before the focused element's target phase.
- Widget reconciliation pitfall: switching the outer widget type at a build
  slot between rebuilds (e.g. `ui::Center(...)` while loading and a Flex
  once loaded) can drop the new tree. Wrap the conditional in a stable outer
  widget type; let the conditional content live one level deeper.

## State and Async Patterns

Widget-local UI data lives in a plain Ard struct stored in the shim's
`StateCtx`. Stateless model modules perform fetch/decode/mutation; views
store the results in state.

```ard
fn new(api_key: Str) ui::Widget {
  ffi::Stateful{
    Init: fn() Any {
      View{started: false, items: [], loading: true, error: ""}
    },
    Build: fn(c: mut ffi::StateCtx, ctx: ui::BuildContext) ui::Widget {
      // First-build kick: StateCtx is created lazily, so Init cannot
      // reach it. `started` makes this once-only across rebuilds.
      let s = ffi::StateRef<View>(c)
      if not s.started {
        s.started = true
        load(c, api_key)
      }
      let theme = ui::MustDepend<ui::Theme>(ctx)
      let model = ffi::StateValue<View>(c)
      ...
    },
  }
}
```

Rules:

- **Mutate state with `ffi::Update<State>(c, fn(s: mut State) { ... })`.** The
  callback receives the **live `mut State`** (a pointer into persistent
  storage, not a copy), mutates fields in place, and the rebuild is scheduled
  for you — you can't forget `MarkDirty`. Spell the callback param
  `fn(s: mut State)` (mut in the type), never `fn(mut s: State)`.
- `ffi::StateValue<T>(c)` reads a snapshot for rendering or for a guard read
  before an `Update`.
- The raw `ffi::StateRef<T>(c)` (live `mut T`) + `ffi::MarkDirty(c)` are still
  available, but `Update` is the API for anything that changes what renders.
  The one place `StateRef` is used directly is the first-build `started`
  bookkeeping flag, which deliberately does **not** rebuild.
- **Initial fetches happen on first build behind a `started` guard** (the
  shim's StateCtx doesn't exist during `Init`).
- Background work must re-enter the UI thread via `ffi::Dispatch` before
  touching state; `Update` goes inside the dispatch:

  ```ard
  async::start(fn() {
    let result = load()
    ffi::Dispatch(c, fn() {
      ffi::Update<View>(c, fn(s: mut View) {
        match result {
          ok(items) => { s.items = items, s.error = "" },
          err(e) => { s.error = e },
        }
        s.loading = false
      })
    })
  })
  ```

- Work that must happen *after* the mutation (a follow-up fetch, a scroll
  reveal read from `StateValue`, `persist`) goes after the `Update` call, not
  inside the callback.
- `show_modal` / `close_modal` / `notify` from `main.ard` are dispatch-safe:
  callable from event handlers and background fibers alike.

### Periodic background work

`tui/refresh::schedule_periodic(interval_ms, action)` spawns a fiber that
sleeps then runs `action` forever. Used by my-issues and each issue detail
tab for the 5-minute auto-refresh. Fibers are process-lived and not
cancellable yet — `ffi::Dispatch` against a disposed widget is a safe no-op,
so closed tabs continue polling until app exit (wasteful but harmless).

## Testing

Currently no widget tests — testing is pure unit coverage:

| Test file | Covers |
|-----------|--------|
| `decode_test.ard` | JSON decode combinators: scalars, lists, nullable, path walking, error paths. |
| `models/cache_test.ard` | Versioned cache round-trip + invalid-input handling. |
| `models/issues_test.ard` | Mojibake normalization in issue/comment decode. |
| `tui/logged_in_screen_test.ard` | Tab math: cycling, open/close issue tabs, labels. |
| `tui/state_test.ard` | Priority + workflow-state-type ordering helpers. |
| `tui/text_test.ard` | Display-width truncation. |

Run with `ard test`.

## Code Style

- Name functions for UI intent (`issue_card`, `refresh_board`, `open_state_picker`) over mechanics.
- Use named arguments when calling a function with three or more arguments (Go struct literals use their field names instead).
- Model modules own fetch / decode / mutation. TUI modules own presentation and state wiring.
- Keep errors user-facing: surface failures via toasts (`notify`) or inline error panes; `main()` panics with a clear message only for startup failure.
- Linear auth header is `Authorization: <raw key>` — no `Bearer` prefix.
- Use `tinear/decode`'s `nullable`/`path` combinators (and the `models/*` helpers built on them) for nullable nested JSON fields; GraphQL inline fragments often omit them entirely.
- Use display-width-aware text helpers (`tui/text`) when fitting strings into fixed terminal cells. Byte-based helpers are ASCII-only.
- Auto-refresh handlers should be **silent** (no `loading = true` flash) and **preserve cursor by id** (look up the previously-cursored entity's id after the swap and snap to its new index). See `refresh_board` in `tui/my_issues_view.ard`.

## Auth and CLI surface

Single entrypoint (`main.ard`) launches the TUI. Auth happens in the welcome
screen, which validates the key against Linear's `viewer` query and writes it
to `~/.tinear/config.json`. To bootstrap without the TUI, set
`$LINEAR_API_KEY` or hand-edit the config file.

## Ard Language Notes

- **No `return` keyword** — last expression is the return value. Use `try` for Result propagation.
- **`and` / `not`** instead of `&&` / `!`.
- **`Void!Str`** uses `Result::ok(())` for the Ok variant.
- **String interpolation** with `{var}`; literal braces need `\{` / `\}`.
- **Mutating methods** with `fn mut name() { self.field = v }`; callable only on mutable receivers.
- **`mut <expr>` creates a mutable reference** (ADR 0045): `mut ctrl` aliases a mut binding; reads of a `mut T` variable deref (copy). Two current parser limits: `mut pkg::Type{}` doesn't parse (ard#285 — bind first, then `mut binding`) and `mut x` can't be a block's final expression (parses as a declaration).
- **Match arms must unify types.** Arms returning different foreign widget types need an explicit annotation on the binding (`let w: ui::Widget = match ...`).
- **`.at(i)` returns `T?`** on lists — `.expect("bounds checked")` after an explicit bounds check. Map access is `.get(key)` returning `V?`; removal is `.delete(key)`.
- **JSON numbers arrive as `Float64`** through `Any`; `decode::int` does the integrality check.
- **Go slice params/results are `mut [Byte]`-shaped**: pass a `mut` binding (`mut bytes = text.bytes()`), and bind results with `mut`.
- **Avoid `try expr -> e { Result::err(...) }`** until ard#282 lands — catch blocks producing `Result` values miscompile under the `(T, error)` ABI. Use `.map_err(named_fn)` + plain `try`, or an explicit `match`.
- **Cross-module generic helpers must be public** (second half of ard#282): private generics instantiated from another module emit empty bodies.
- **`async::start` returns `Void`** — no discard needed. Fibers must re-enter the UI thread via `ffi::Dispatch` before touching widget state.
- **Direct Go variadics** accept at most one trailing argument (the variadic element); calls needing multiple variadic values go through a shim.

## Upstream references

- The vaxis/ui Go source is the authoritative widget API reference (fields,
  constants, controller methods). Read it before guessing a widget's shape.
- Compiler issues tracked from this port: ard#282 (try-catch Result ABI +
  private cross-module generics), ard#283 (`Str::from_bytes`), ard#284
  (Int → sized-scalar conversions), ard#285 (`mut pkg::Type{}` parse).
