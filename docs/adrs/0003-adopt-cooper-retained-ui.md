# 0003: Adopt Cooper Retained UI

## Status

Accepted

## Context

Tinear's current implementation on `main` is built directly on vaxis/ui's
Flutter-inspired widget and reconciliation model. It provides the product
behavior we want to preserve, but application state, lifecycle, focus, and
asynchronous work are coupled to widget rebuilds and a Go shim that Ard needs
to express stateful widgets and intents.

Cooper now provides an Ard-native imperative retained-mode framework with
persistent controls, direct mutation, explicit focus, selection, scrolling,
Input and TextArea controls, application dispatch, and deterministic headless
testing. Tinear is also a useful real-application proving ground for Cooper.

ADR 0002's durable product decisions remain useful, but its rendering-layer
constraint and phased feature list no longer describe the active direction.

## Decision

Rewrite Tinear's UI using Cooper's public retained controls directly.

- Each screen is a long-lived controller with one stable root control and
  concrete references to the controls it mutates.
- The initial retained tree is built before the application starts. Later tree
  and control mutations happen only in Cooper callbacks or through
  `Runtime.dispatch` through `application.context`.
- The application owns one mode-aware command router, modal and toast layers,
  tab focus policy, and per-controller lifecycle scopes.
- Inactive permanent tabs remain mounted and are hidden rather than rebuilt.
  Dynamic tabs are explicitly disposed when closed.
- Background work uses application cancellation, per-controller cancellation,
  and request generations so stale or late completions never touch destroyed
  controls.
- Linear, config, cache, and decoding modules remain independent of Cooper.
- The application preserves terminal-default foreground/background and owns a
  small semantic palette for muted, accent, border, surface, danger, warning,
  and success roles.
- Tinear uses only Cooper's supported public API and does not import
  `cooper/core/*` or introduce a widget/binding compatibility layer.
- Cooper `TestApp` tests cover deterministic UI behavior; PTY tests are reserved
  for terminal integration.

The reachable behavior and tests on `main` are the parity reference. The living
checklist is maintained in `docs/cooper-rewrite-roadmap.md`.

## Consequences

- The TUI is a clean rewrite rather than a source-compatible port.
- Existing model and parser behavior can be preserved, while vaxis/ui-specific
  stateful and intent shims are removed.
- Modal focus isolation, command precedence, and view-level async disposal are
  explicit application responsibilities.
- Cooper gaps discovered by the rewrite should be addressed through supported
  Cooper APIs or documented parity waivers, not by reaching into internals.
- Significant changes to this architecture require a superseding ADR.

## Related

- [0001: Record Architecture Decisions](0001-record-architecture-decisions.md)
- [0002: Interactive Terminal Explorer](0002-interactive-terminal-explorer.md)
- [Cooper rewrite parity roadmap](../cooper-rewrite-roadmap.md)
