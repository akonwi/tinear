# tinear

## Agent rules

- **Do not run `git commit` without explicit direction from the user.**

Tinear is being rewritten from scratch as an Ard application using
[Cooper](https://github.com/akonwi/cooper), an imperative retained-mode TUI
framework. The previous vaxis/ui implementation remains available on the
`main` branch and should be consulted with `git show main:<path>` rather than
restored wholesale.

## Build, run, and test

```sh
ard check main.ard
ard build main.ard --out tinear
ard run main.ard
ard test
```

The Ard dependency is pinned to a remote Cooper commit; `../cooper` remains the local framework source checkout for joint development and reference.

## Architecture guidance

- Use Cooper controls directly; do not add a widget or binding layer.
- Construct retained controls once and mutate them through setters.
- Build the initial tree before starting the application.
- After startup, mutate controls only from Cooper callbacks or through
  `Runtime.dispatch` (`application.context.dispatch`).
- Keep Linear fetching and decoding separate from retained UI controllers.
- Give every view explicit lifecycle/cancellation state before starting
  background work.
- Prefer deterministic Cooper `TestApp` coverage for UI behavior and PTY tests
  only for terminal integration.
