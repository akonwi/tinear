# tinear

A CLI for [Linear](https://linear.app/) built with [Ard](https://ard.run).

## Setup

1. **Get an API key** from Linear Settings → API → Personal API keys.

2. **Provide the key** via environment variable or config file:

   ```bash
   export LINEAR_API_KEY=lin_api_xxx
   ```

   Or create `~/.tinear/config`:
   ```json
   { "api_key": "lin_api_xxx" }
   ```

   The env var takes precedence.

## Commands

| Command | Description |
|---------|-------------|
| `tinear me` | Show current user |
| `tinear teams [--json]` | List teams |
| `tinear issues [--team <key>] [--status <name>] [--json]` | List and filter issues |
| `tinear issue <id> [--json]` | View issue details |
| `tinear create-issue --team <key> --title <title> [--description <desc>]` | Create an issue |
| `tinear update-issue <id> --status <name>` | Update issue status |
| `tinear my-board [--json]` | List your open assigned issues |
| `tinear tui` | Interactive TUI showing you and your open assigned issues |
| `tinear help` | Show usage |

Use `--json` on any list/detail command for JSON output.

## Build

```bash
ard build main.ard
# binary: ./main
```
