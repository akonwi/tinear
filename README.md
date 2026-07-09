# tinear

A terminal UI for [Linear](https://linear.app/) built with [Ard](https://ard.run).

<img width="1302" height="1013" alt="Ghostty000486" src="https://github.com/user-attachments/assets/0cb3b3a2-2beb-49e7-af53-a67ca33ac294" />

## Setup

1. **Get an API key** from Linear Settings → API → Personal API keys.

2. **Provide the key** via environment variable or config file:

   ```bash
   export LINEAR_API_KEY=lin_api_xxx
   ```

   Or run `tinear login` to save it interactively to `~/.tinear/config`.

## Usage

Run without arguments to launch the interactive TUI:

```bash
tinear
```

The TUI shows your **Inbox** and **My Issues** tabs, with per-issue detail views,
a status picker, and a comment composer.

### Commands

| Command | Description |
|---------|-------------|
| `tinear` (no command) | Launch the interactive TUI |
| `tinear login` | Save an API key interactively |
| `tinear help` | Show usage |

## Build

```bash
ard check main.ard              # type-check
ard build main.ard --out ard-out/tinear
```

The binary at `ard-out/tinear` is standalone and gitignored.

## License

MIT
