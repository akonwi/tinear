# linear-cli

A CLI for [Linear](https://linear.app/) built with [Ard](https://ard.run).

## Setup

1. **Get an API key** from Linear Settings → API → Personal API keys.

2. **Provide the key** via environment variable or config file:

   ```bash
   export LINEAR_API_KEY=lin_api_xxx
   ```

   Or create `~/.linear-cli/config`:
   ```json
   { "api_key": "lin_api_xxx" }
   ```

   The env var takes precedence.

## Commands

| Command | Description |
|---------|-------------|
| `linear-cli me` | Show current user |
| `linear-cli teams [--json]` | List teams |
| `linear-cli issues [--team <key>] [--status <name>] [--json]` | List and filter issues |
| `linear-cli issue <id> [--json]` | View issue details |
| `linear-cli create-issue --team <key> --title <title> [--description <desc>]` | Create an issue |
| `linear-cli update-issue <id> --status <name>` | Update issue status |
| `linear-cli my-board [--json]` | List your open assigned issues |
| `linear-cli tui` | Interactive TUI showing you and your open assigned issues |
| `linear-cli help` | Show usage |

Use `--json` on any list/detail command for JSON output.

## Build

```bash
ard build main.ard
# binary: ./main
```
