# 0002: Interactive Terminal Explorer (TUI)

## Status

Proposed

## Context

The existing `tui` command shows a static read-only view of the user's
open issues (My Board). We want to expand it into a full explorer that
lets users navigate Linear data interactively — inbox, projects, issue
details — without leaving the terminal.

Key requirements from discussion:

- Tab-based navigation where each tab is a screen (inbox, projects,
  issue detail, filtered list)
- Inbox (notifications) as the default, permanently-open tab
- A command palette overlay inspired by Kit's UX pattern: triggered by
  `:`, shows a filterable list of commands
- A status bar at the bottom showing keybinding hints in Kit's format:
  `↑↓ select · r refresh · : command · Esc close · q quit`
- The `vaxis` terminal library as the rendering layer (already vendored)

Non-requirements (for now):

- Mouse support
- Full color theming
- Persistent session state across restarts

## Decision

The TUI will be built incrementally in these phases:

### Phase 1 — Inbox screen + tab scaffolding
- Replace the current My Board view with an Inbox (notifications) screen
- Introduce a tab bar with a single permanent Inbox tab
- Status bar with keybinding hints
- Separates concerns: tab management, screen rendering, input dispatch

### Phase 2 — Projects screen + tab switching
- Second permanent tab: Projects
- Tab switching (`Tab` / `Shift+Tab`)
- Selecting a project opens a new tab with its filtered issues

### Phase 3 — Issue detail tab
- Selecting an issue opens a new tab with its full detail view
- Tabs are closeable (`Esc` closes non-root tabs)

### Phase 4 — Command palette
- Overlay dialog with filterable command list
- Commands: switch tabs, open issue/project, filter, help
- Triggered by `:`

### Key constraint
- Use the existing `vaxis` FFI bindings — no new external dependencies.

## Consequences

- Building in phases keeps each change reviewable and testable.
- The tab/screen abstraction will need to be designed upfront to avoid
  major refactors between phases.
- Inbox is a new GraphQL query (`notifications`) — adds one more API
  surface to maintain.
- The current `draw_board` / `fetch_board` / `run_loop` structure will
  be replaced by a more general screen-rendering pattern.
- The command palette overlay requires rendering multiple layers
  (background dim, palette dialog) on top of the active screen.

## Related

- [0001: Record Architecture Decisions](0001-record-architecture-decisions.md)
- Linear API — `notifications` GraphQL query
- Kit's command palette pattern (`../kit/app/src/shell/CommandPalette.tsx`)
