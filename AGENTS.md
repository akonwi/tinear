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
  only prop and renders it. The controller doubles as the handle the shell
  needs for key routing and focus.
- That split is a deliberate choice, not a framework limit. Keys can be
  consumed declaratively: view-level `on_key` bubbles from the focused node to
  the root and stops at `stop_propagation()`, and `ctx.on_key` subscriptions
  see the same mutable event. Cooper's `examples/cui_tinear.ard` routes a
  similar app that way. Tinear instead routes every key through
  `application_shell` with explicit precedence (modal, shell globals, then the
  active screen), which does not depend on focus being correct. Focus has been
  the fragile part of this codebase, so keep that independence unless you are
  deliberately redesigning routing.
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
  have none. A ref is only populated by a commit, so anything imperative that
  targets a freshly described view must run behind `ctx.dispatch` rather than
  immediately after the mutation that describes it.
- A declarative view becomes visible only at the next commit, and a control
  inside a `display: none` subtree cannot take focus. The shell therefore
  defers its focus claim through `focus_later` after changing the active tab.
- Loading, error, and loaded screens keep the same keyed scroll target mounted.
  Do not retry screen focus from data completions: a newer modal or user focus
  choice supersedes the original request.
- `tui/focus` isolates the pinned Cooper `core/node` dependency needed to
  inspect focus within a retained subtree; CUI has no equivalent query.
- Modal presentations use a generation-keyed subtree so replacing a body resets
  its component state even when close/open coalesce into one commit.
- `cui::box` is focusable only when it describes `on_key` or `focused`. Attach
  a no-op `on_key` to make a panel focusable; describing `focused` instead
  re-asserts focus on every commit.
- Every shell screen is declarative: build one with
  `shell_controller::declarative_screen`. There is no imperative screen shape.
- `tui/cui_bridge` hosts an imperative subtree inside a component. It is needed
  where CUI lacks a feature — today the markdown renderer, which requires text
  link callbacks, and imperative modal bodies. Pass `destroy_on_unmount: false`
  for controls that outlive the bridge.
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
