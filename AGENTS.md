# tinear

## Agent rules

- **Do not run `git commit` without explicit direction from the user.**

Tinear is an Ard application using
[Cooper](https://github.com/akonwi/cooper). The UI is built with `cooper/cui`,
Cooper's declarative component layer; the imperative retained API remains
available and is still used where CUI has no equivalent. The previous vaxis/ui
implementation remains available at commit `84ff69a` and should be consulted
with `git show 84ff69a:<path>` rather than restored wholesale.

## Build, run, and test

```sh
ard check main.ard
ard build main.ard --out tinear
ard run main.ard
ard test
```

The Ard dependency is pinned to a remote Cooper commit; `../cooper` remains the local framework source checkout for joint development and reference.

## Architecture guidance

- Describe UI with `cooper/cui` components. Import `cooper` for Runtime and
  event types, `cooper/ui` for styles and view values, and `cooper/cui` for
  components. Keep `cooper/animation`, `cooper/testing`, and `cooper/event` to
  specialized APIs and test event constructors.
- Screens split a controller from a component: the controller owns data,
  requests, and cancellation; the component is passed the controller as its
  only prop and renders it. The split is not optional — `ctx.on_key` handlers
  return nothing, so key consumption and precedence (modal, then shell
  globals, then the active screen) cannot be expressed declaratively, and the
  shell needs an imperative handle per screen. CUI has no component refs, so
  the controller is that handle.
- Controllers request renders through `self.invalidate`, bound to
  `ctx.invalidate_root` by the component's `mounted` hook and reset to a no-op
  on `unmounting`. Never take a render callback as a constructor argument.
- `render` must be deterministic and effect-free. Mutate state in event
  callbacks and lifecycle hooks only.
- Describe focus with state (see the issue creation form's `FormFocus`).
  Reconciliation re-asserts a described `focused: true` on every commit, so a
  one-shot focus must either be cleared after its first commit or performed
  imperatively through a ref.
- Use refs for geometry, scrolling, and imperative capabilities the declarative
  API does not cover; guard `ref.current` since offscreen and unmounted views
  have none.
- `tui/cui_bridge` hosts an imperative subtree inside a component. It is needed
  where CUI lacks a feature — today the markdown renderer, which requires text
  link callbacks. Pass `destroy_on_unmount: false` for controls that outlive
  the bridge.
- Keep Linear fetching and decoding separate from UI code.
- Every async completion must check its controller's disposal and
  request-generation state before mutating state. Prefer routing service
  completions (which already land on the UI thread) directly, and reserve
  component `ctx.dispatch` for work that must be suppressed on unmount.
- Modal bodies are declarative: `modal_host::present_view`. Reserve
  `present` for imperative bodies.
- Prefer deterministic Cooper `TestApp` coverage for UI behavior and PTY tests
  only for terminal integration. Tests drive a renderer with
  `cui/renderer::mount` and must `flush()` before `render()` to commit queued
  renders. Mounting is all the wiring a controller needs.
- Cross-view sync happens over the shell-scoped event bus (`events.ard`): mutation sites publish, data-owning controllers
  subscribe with a silent refresh, and every subscription's unsubscribe fn must
  run in the owner's dispose path. Publish only from the dispatch context;
  subscribers run synchronously. Do not use channels for broadcast — Ard
  channels are point-to-point.
