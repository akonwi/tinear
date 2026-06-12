# tinear

## Agent rules

- **Do not run `git commit` without explicit direction from the user.**

CLI for [Linear](https://linear.app/) built with [Ard](https://ard.run).

## Build, Run, Test

```bash
ard check main.ard                          # type-check
ard build main.ard --out ard-out/tinear     # build into the gitignored ard-out/
ard run main.ard                            # run directly
ard test                                    # run unit/widget tests
```

## Project Layout

| Path | Purpose |
|------|---------|
| `main.ard` | Entrypoint, dispatches `help` / `login` / default-to-TUI |
| `config.ard` | Load config and API key from env var or `~/.tinear/config` |
| `linear/client.ard` | Shared GraphQL client (one `graphql()` function) |
| `commands/login.ard` | Interactive `login` command (saves key to config) |
| `commands/tui.ard` | vaxis/ui TUI entrypoint used by the default command |
| `models/*.ard` | Stateless fetch/decode model modules for inbox and issues |
| `tui/*.ard` | vaxis/ui implementation, split by concern |
| `tui/ui.ard` | Ard-facing vaxis/ui widget/state/runtime bindings |
| `tui/ui/test.ard` | Ard-facing `vaxis/ui/uitest` bindings for widget tests |
| `vaxis.ard` + `ffi/` | Terminal/vaxis/ui bindings (Ard extern declarations + Go FFI) |

## Active TUI Architecture

The active TUI is built on `git.sr.ht/~rockorager/vaxis/ui`. Do not add new
custom retained UI framework code; prefer vaxis/ui primitives plus thin Ard
wrappers.

| File | Purpose |
|------|---------|
| `tui/logged_in_screen.ard` | Logged-in tab shell; keeps tab bodies mounted with `ui::indexed_stack` |
| `tui/welcome_screen.ard` | Login/auth screen for unauthenticated users |
| `tui/inbox_view.ard` | Inbox list + notification detail modal |
| `tui/my_issues_view.ard` | My Issues board + picker modals |
| `tui/modal.ard` | Shared vaxis/ui modal/dialog wrapper |
| `tui/hints.ard` | Footer hint bar |
| `tui/api.ard` | GraphQL fetch/mutation helpers and picker construction |
| `tui/state.ard` | Shared Linear UI model structs + workflow/priority helpers |
| `tui/decode.ard` | `optional_field` — missing-or-null aware field decoder |
| `tui/text.ard` | Byte- and display-width string helpers (truncate, pad, wrap) |

## vaxis/ui Best Practices

- Prefer native vaxis/ui composition (`ui::row`, `ui::column`, `ui::text`,
  `ui::box`, `ui::actions`, `ui::shortcuts`, etc.) over app-specific rendering
  abstractions.
- Keep Ard bindings thin and ergonomic. Public functions should translate Ard
  nullable/default arguments into the exact Go FFI shape needed by vaxis/ui.
- Keep raw FFI handles private/internal where possible. Expose Ard wrapper
  structs like `BuildContext`, `Runtime`, `State`, and `EventContext`.
- Primitive widget factory functions should accept a nullable `style: Style?`
  argument when the underlying widget/rendering can be styled.
- Prefer a consistent, discoverable API like `ui::text(value, style: ...)` and
  `ui::row(children, style: ...)` over single-purpose helpers such as
  `text_reverse` or `background`.
- Prefer optional arguments for widget customization/lifecycle hooks instead of
  separate public functions for the same widget.
- Separate keybindings from behavior:
  - `ui::shortcuts` maps keys (`"j"`, `"Enter"`, `"Ctrl+c"`) to intent names.
  - `ui::actions` maps intent names to behavior with `ui::action(...)`.
- Keyboard events follow the focused widget path. If a state change mounts a new
  focus target (for example opening/selecting a dynamic tab), defer the focus
  request until after the rebuild with `ui::dispatch_after(ctx, 0, fn() { ... })`;
  requesting focus before the target is mounted silently leaves focus elsewhere.
- Modals should get key priority by being placed in `ui::overlay_modal`, whose
  FFI wrapper traps/autofocuses the modal subtree without drawing a shaded
  barrier. Avoid adding root-level Escape handlers that always return handled;
  return `ignored` when they do not apply so modal dismiss handlers can run.
- For scrollable detail/body regions, prefer native `ui::scroll_view` or
  controller-backed `ui::scroll_pane`. If focus remains on an outer wrapper,
  route `j/k` and arrow shortcuts to the scroll pane controller explicitly;
  mouse wheel scrolling can work even when keyboard focus is elsewhere.
- Keep inactive tab bodies mounted with `ui::indexed_stack` when their state
  should survive tab switches.

## State and Async Patterns

Widget-local UI data should live in typed `ui::State` values. Stateless model
modules perform API calls/decoding, then widgets store returned data in state.

Use `ui::stateful(init: ..., build: ..., key: ...)`:

```ard
ui::stateful(
  key: "my-widget",
  init: fn(ctx: ui::BuildContext) MyState {
    MyState{loading: false, items: []}
  },
  build: fn(ctx: ui::BuildContext, state: ui::State<MyState>) ui::Widget {
    let model = state.value()
    ...
  },
)
```

Rules:

- `init` creates initial state only. Do not perform blocking background work in
  `init` unless it is intentionally synchronous and cheap.
- Direct UI callbacks/build/event handlers may mutate state directly with
  `state.set(fn(mut next: MyState) { ... })`.
- Background work must re-enter the UI runtime before mutating UI state:

```ard
let rt = ctx.runtime<MyState>()
async::start(fn() {
  let result = load()
  rt.dispatch(fn(state: ui::State<MyState>) {
    state.set(fn(mut next: MyState) {
      match result {
        ok(items) => {
          next.items = items
          next.error = ""
        },
        err(e) => { next.error = e },
      }
      next.loading = false
    })
  })
})
```

- Prefer typed state handles (`ui::State<MyState>`, `ui::Runtime<MyState>`)
  with `state.value()` and `state.set(...)`.
- Prefer immutable snapshots for rendering (`let model = state.value()`) and
  focused mutation blocks for updates.
- Use stable `key:` values for stateful widgets whose identity must survive tree
  reordering or tab switching.

## Testing vaxis/ui Widgets

Use the Ard wrapper around upstream `vaxis/ui/uitest`:

```ard
use ard/testing
use tinear/tui/ui/test as uitest

let app = uitest::app(widget)
app.pump(80, 24)
try testing::assert(app.contains("Inbox"), "renders Inbox")
```

Available helpers include:

- `pump(width, height)` — rebuild/layout/paint at a fixed size
- `key(str)`, `enter()`, `escape()`, `tab()`, `shift_tab()`
- `up()`, `down()`, `left()`, `right()`
- `click(x, y)`
- `contains(str)`, `text()`
- `cell_grapheme(x, y)`, `cell_reverse(x, y)`
- `should_quit()`

Testing guidance:

- Prefer widget tests for rendering, focus/key behavior, actions, shortcuts,
  modal open/close, and state transitions.
- Pump after constructing the app and after input that should affect rendering.
- Assert on user-visible text or cell attributes; avoid coupling tests to
  incidental whitespace unless layout is the behavior under test.
- Keep pure helper tests in regular Ard unit tests when no widget tree is
  needed.

## Code Style

- Keep functions small and named for UI intent (`issue_card`, `detail_modal`,
  `load_inbox`) rather than implementation mechanics.
- Favor named arguments when calling a function with more than two arguments.
- Prefer model modules (`models/*.ard`) for fetch/decode logic and TUI modules
  for presentation/state wiring.
- Keep GraphQL decode edge cases close to API/model code, not inside render
  branches.
- Keep errors user-facing. `run()` entrypoints return `Void`, so they may panic
  or print a clear message on failure.
- Do not add a `Bearer` prefix to Linear auth. The header is
  `Authorization: <raw key>`.
- Use `tui/decode::optional_field(data, name)` for nullable nested JSON fields;
  GraphQL inline fragments can omit fields entirely.
- Use display-width-aware text helpers when fitting strings into fixed terminal
  cells. Byte-based helpers are only safe for ASCII.

## Commands

- `login` — prompts for API key, validates via `viewer` query, saves to
  `~/.tinear/config`
- `help` — prints usage
- _(no command)_ — launches the interactive TUI

## Ard Language Notes

- **No `return` keyword** — last expression is the return value. Use `try` for
  Result propagation.
- **Functions must be defined before use** within a file (cross-file is fine).
- **`and` / `not`** instead of `&&` / `!`.
- **`Void!Str`** uses `Result::ok(())` for the Ok variant.
- **`{ ... }` in `match`** must be followed by a newline — no inline
  `match x { a => b, _ => c }`.
- **String interpolation** with `{var}`, literal braces need `\{` / `\}`.
- **`use ard/list as List`** to access `List::drop()`.
- **Mutating methods** with `fn mut name() { self.field = v }`. Methods can
  mutate fields when the receiver is `mut`.
- **`mut` doesn't propagate through match/for bindings** — `mut x = ...; for
  item in xs { item.field = v }` doesn't compile. Hoist a copy:
  `mut t = xs.at(i); t.field = v; xs.set(i, t)`.
