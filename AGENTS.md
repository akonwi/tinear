# tinear

## Agent rules

- **Do not run `git commit` without explicit direction from the user.**

Tinear is an Ard application using
[Cooper](https://github.com/akonwi/cooper), an imperative retained-mode TUI
framework. The previous vaxis/ui implementation remains available at commit
`84ff69a` and should be consulted with `git show 84ff69a:<path>` rather than
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
- Application and custom-view modules should normally import only `cooper` for
  App, Runtime, and events plus `cooper/ui` for controls and view values. Keep
  focused imports such as `cooper/animation`, `cooper/testing`, or
  `cooper/event` only for specialized APIs and test event constructors.
- Construct retained controls once and mutate them through setters.
- Build the initial tree before starting the application.
- After startup, mutate controls only from Cooper callbacks or through
  `Runtime.dispatch` (`application.context.dispatch`).
- Keep Linear fetching and decoding separate from retained UI controllers.
- Give every view explicit lifecycle/cancellation state before starting
  background work.
- Because `Runtime.dispatch` is app-scoped, every async completion must check
  its controller's disposal and request-generation state before mutating controls.
- Prefer deterministic Cooper `TestApp` coverage for UI behavior and PTY tests
  only for terminal integration.
- Cross-view sync happens over the shell-scoped event bus (`events.ard`): mutation sites publish, data-owning controllers
  subscribe with a silent refresh, and every subscription's unsubscribe fn must
  run in the owner's dispose path. Publish only from the dispatch context;
  subscribers run synchronously. Do not use channels for broadcast — Ard
  channels are point-to-point.
