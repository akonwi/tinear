# tinear

## Agent rules

- **Do not run `git commit` without explicit direction from the user.**

CLI for [Linear](https://linear.app/) built with [Ard](https://ard.run).

## Build & Run

```bash
ard check main.ard                          # type-check
ard build main.ard --out ard-out/tinear     # build into the gitignored ard-out/
ard run main.ard                            # run directly
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
| `tui/*.ard` | vaxis/ui implementation, split by concern (see below) |
| `vaxis.ard` + `ffi/` | Terminal/vaxis/ui bindings (Ard extern declarations + Go FFI) |

## TUI Layout

The active TUI is built on `git.sr.ht/~rockorager/vaxis/ui` via `tui/ui.ard`.

| File | Purpose |
|------|---------|
| `tui/ui.ard` | Ard-facing vaxis/ui widget/state/runtime bindings |
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

### Architectural conventions

- Prefer vaxis/ui primitives and thin Ard wrappers over custom retained UI code.
- Widget-local UI data lives in typed `ui::State` values. Stateless model modules
  perform API calls and decoding, then widgets store returned data in state.
- Use `ui::stateful(init: ..., build: ..., key: ...)`. `init` creates initial
  state only; background work should capture `ctx.runtime()` and dispatch state
  mutations back onto the UI runtime.
- Keep inactive tab bodies mounted with `ui::indexed_stack` when their state
  should survive tab switches.
- Keep FFI-only details private/internal when possible. Public Ard functions
  should translate ergonomic nullable/default arguments into the required Go FFI
  shape.

## vaxis/ui binding conventions

- Primitive widget factory functions should accept a nullable `style: Style?`
  argument when the underlying widget/rendering can be styled.
- Prefer a consistent, discoverable API like `ui::text(value, style: ...)` and
  `ui::row(children, style: ...)` over use-case-specific helpers such as
  `text_reverse`, `background`, or other single-purpose styling wrappers.
- Prefer optional arguments for widget customization/lifecycle hooks instead of
  separate public functions for the same widget.
- Re-enter the UI runtime explicitly from background work:

```ard
let rt = ctx.runtime()
async::start(fn() {
  let result = load()
  rt.dispatch(fn(state: ui::State) {
    state.set<MyState>(fn(mut s: MyState) { ... })
  })
})
```

## Key Patterns

- **Error handling**: `.expect("msg")` panics with a message. The `run()`
  entrypoints return `Void`, so they panic on failure. Keep error messages
  user-facing.
- **Nullable nested JSON**: use `tui/decode::optional_field(data, name)` to
  treat both missing-field and null-value cases as `None`. GraphQL inline
  fragments omit fields entirely when their parent type doesn't match.
- **Display-width-aware** text helpers (`text::pad_str_display`,
  `text::truncate_with_ellipsis_display`) require a `vx` argument and use
  `vaxis::rendered_width`. Use these whenever fitting into a fixed-cell slot;
  the byte-based versions are only safe for ASCII.
- **Auth header**: `Authorization: <raw key>` — no `Bearer` prefix. Linear
  rejects it.
- **Named args**: when calling a function with more than two arguments, use
  named arguments for readability.

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
