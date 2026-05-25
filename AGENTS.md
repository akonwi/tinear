# linear-cli

CLI for [Linear](https://linear.app/) built with [Ard](https://ard.run).

## Build & Run

```bash
ard check main.ard      # type-check
ard build main.ard       # build -> ard-out/go/main
ard run main.ard         # run directly
```

## Project Layout

| Path | Purpose |
|------|---------|
| `main.ard` | Entrypoint, command dispatch |
| `config.ard` | API key from env var `LINEAR_API_KEY` or `~/.linear-cli/config` |
| `util.ard` | Flag parsing, error helpers, usage printer |
| `linear/client.ard` | Shared GraphQL client (one `graphql()` function) |
| `commands/*.ard` | One file per subcommand, each exports `fn run(args: [Str])` |

## Key Patterns

- **Error handling**: `.expect("msg")` panics with a message. The `run()` entrypoints are `Void`, so they panic on failure. Keep error messages user-facing.
- **Nullable nested JSON**: Always decode parent objects one level at a time with `decode::nullable(decode::dynamic)`, then `match obj { v => ... maybe::some(name), _ => maybe::none() }`. Avoid `decode::path()` on nullable intermediates — it panics on null.
- **Auth header**: `Authorization: <raw key>` — no `Bearer` prefix. Linear rejects it.

## Commands

- `me` — viewer query, plain text = `Name <email>`
- `teams` — lists all teams
- `issues [--team <key>] [--status <name>] [--json]` — lists issues, filter args optional
- `issue <id> [--json]` — detail view with labels, description
- `create-issue --team <key> --title <title> [--description <desc>]` — resolves team key to ID
- `update-issue <id> --status <name>` — resolves issue team, looks up state by name
- `login` — prompts for API key, validates via `viewer` query, saves to `~/.linear-cli/config`

## Ard Language Notes

- **No `return` keyword** — last expression is the return value. Use `try` for Result propagation.
- **Functions must be defined before use** within a file.
- **No `private fn`** for regular functions — visibility isn't supported yet.
- **`and` / `not`** instead of `&&` / `!`.
- **`Void!Str`** uses `Result::ok(())` for the Ok variant.
- **`{ ... }` in `match`** must be followed by a newline — no inline `match x { a => b, _ => c }`.
- **String interpolation** with `{var}`, literal braces need `\{` / `\}`.
- **`use ard/list as List`** to access `List::drop()`.
