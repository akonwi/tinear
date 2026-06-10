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
| `config.ard` | load config and API key from env var or `~/.tinear/config` |
| `linear/client.ard` | Shared GraphQL client (one `graphql()` function) |
| `commands/login.ard` | Interactive `login` command (saves key to config) |
| `commands/tui.ard` | Thin TUI entrypoint — delegates to `tui/app::run_loop` |
| `tui/*.ard` | TUI implementation, split by concern (see below) |
| `vaxis.ard` + `ffi/` | Terminal UI bindings (Ard extern declarations + Go FFI) |

## TUI Layout

`commands/tui.ard` is just an entrypoint. The actual implementation lives in `tui/`:

| File | Purpose |
|------|---------|
| `tui/app.ard` | `run_loop` (event loop + input dispatch) and `draw_screen` |
| `tui/state.ard` | All data structs (AppState, IssueTab, etc.) + workflow / priority helpers |
| `tui/components.ard` | `Scrollview`, `CommentList`, `draw_comment` — view components with impls |
| `tui/api.ard` | Every GraphQL fetch/mutation + `open_state_picker` |
| `tui/text.ard` | Byte- and display-width string helpers (truncate, pad, wrap) |
| `tui/screen.ard` | `Layout` struct + terminal geometry helpers |
| `tui/decode.ard` | `optional_field` — missing-or-null aware field decoder |
| `tui/draw.ard` | Shared draw primitives (tab bar, label rows) |
| `tui/notifications.ard` | Notification helpers + inbox view (list, detail panels) |
| `tui/board.ard` | Board geometry + card/column rendering |
| `tui/issue_tab.ard` | Open-issue tab rendering |
| `tui/modals.ard` | Status picker + comment composer modal rendering |

### Architectural conventions

- **Persistent state vs ephemeral layout**: view components store only the bits
  that survive across frames (`scroll`, `cursor`, `top`). The per-frame layout
  rectangle (`screen::Layout`) is computed in the parent and passed into the
  component's `draw` method. Use this pattern for any new component.
- **`AppState` is rebuilt each frame** in `app::run_loop` from local mutable
  variables and passed (immutably) into `draw_screen`. State mutations live in
  the input handler, not in draw paths.
- **Dependency direction**: text/screen/decode → components → state → api →
  draw/board/notifications/modals/issue_tab → app. Don't introduce upward
  dependencies.

## Vaxis FFI

`vaxis.ard` lives at the project root with `ffi/host.go` holding the Go bindings.
The Go module includes both.

### `vaxis/ui` binding conventions

- Primitive widget factory functions should accept a nullable `style: Style?`
  argument when the underlying widget/rendering can be styled.
- Prefer a consistent, discoverable API like `ui::text(value, style: ...)` and
  `ui::row(children, style: ...)` over use-case-specific helpers such as
  `text_reverse`, `background`, or other single-purpose styling wrappers.
- Keep FFI-only details private/internal when possible. Public Ard functions
  should translate ergonomic nullable/default arguments into the required Go FFI
  shape.
- Prefer optional arguments for widget customization/lifecycle hooks instead of
  separate public functions for the same widget. For example, use
  `ui::stateful(build, init: ...)` rather than a separate `stateful_with_init`.

Key bits of the event API:

- `vaxis::next_event(vx) Event` — blocking; returns the typed `Event` union
  (`KeyEvent | MouseEvent | ResizeEvent | FocusEvent | PasteEvent | RedrawEvent | CustomEvent | ColorThemeEvent | QuitEvent`).
- `KeyEvent.name` is the plain key name (no modifier prefixes): `"Escape"`,
  `"Enter"`, `"Up"`, `"Tab"`, `"a"`, etc.
- `KeyEvent.text` is the rendered text (handles shift/layout/IME). Empty for
  non-textual keys. Prefer `.text` for text inputs.

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
- _(no command)_ — launches the interactive TUI (Inbox + My Issues tabs,
  per-issue detail tabs, status picker, comment composer)

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
