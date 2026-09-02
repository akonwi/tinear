# tinear

A retained terminal UI for [Linear](https://linear.app/) built with [Ard](https://ard.run) and [Cooper](https://github.com/akonwi/cooper).

<img width="1302" height="1013" alt="Ghostty000486" src="https://github.com/user-attachments/assets/0cb3b3a2-2beb-49e7-af53-a67ca33ac294" />

## Setup

1. **Get an API key** from Linear Settings → API → Personal API keys.

2. **Provide the key** via environment variable or config file:

   ```bash
   export LINEAR_API_KEY=lin_api_xxx
   ```

   Or run `tinear` and paste the key in the UI, which will be saved to `~/.tinear/config.json`.

## Usage

Run without arguments to launch the interactive TUI:

```bash
tinear
```

The TUI provides:

- A notification Inbox with preview, archive, and browser routing
- A retained My Issues board with state, cycle, priority, assignee, and project editing
- Persistent issue and document tabs with Markdown detail views
- Threaded comments, replies, and issue-description editing
- Global issue/document search and issue creation

Use `?` to search, `c` to create an issue from Inbox or My Issues, and
`Tab`/`Shift+Tab` to cycle persistent tabs. Contextual shortcuts are shown in
the footer.

### Commands

| Command | Description |
|---------|-------------|
| `tinear` (no command) | Launch the interactive TUI |

## Build and test

Building requires Ard v0.40.0 or newer. The Cooper dependency is pinned to an exact remote revision in `ard.toml` and `ard.lock`.

```bash
ard check main.ard
ard build main.ard --out ard-out/tinear
ard test
go test ./...
python3 test/pty/test_tinear.py  # live-terminal smoke test
```

## License

MIT
