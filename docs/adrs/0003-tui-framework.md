# 0003: TUI Framework Layer

## Status

Proposed

## Context

The interactive TUI (introduced in [ADR 0002](0002-interactive-terminal-explorer.md))
has grown into a substantial codebase under `tui/`. After improving the
`vaxis` bindings to map more directly onto the Go library, several
recurring patterns became visible:

- **Geometry is ad hoc.** Every screen invents its own way to compute
  rectangles. `tui/screen.ard` has a one-off `Layout` struct; `tui/board.ard`
  computes columns by hand; `tui/modals.ard` hard-codes box widths.
- **Boxes are hand-drawn everywhere.** `modals.ard`, `board.ard`,
  `notifications.ard`, `issue_tab.ard` all reach for `┌─┐│└┘` characters
  individually. Hundreds of lines of essentially-duplicated bordering.
- **No common component contract.** Existing components (`Scrollview`,
  `CommentList`) each have a bespoke `draw` signature. New components
  invent another one. There's no abstraction that says "I am drawable."
- **No clipping / sub-region abstraction.** Draws go straight to the
  terminal cell buffer at absolute coordinates. There's no equivalent
  of "this widget owns this rectangle; render within it."

Three established TUI frameworks were surveyed to choose a direction:

| Framework | Model | Layout | Trade-off |
|---|---|---|---|
| Bubble Tea (Go) | Elm Architecture: `Init/Update(Msg)/View() string` | Mostly hand-rolled (Lipgloss) | Pure-functional message dispatch is hard without ergonomic sum-type matching at every level. |
| ratatui (Rust) | Immediate-mode: `terminal.draw(\|frame\| {...})`, `Widget` trait | Constraint-based, declarative split | App owns state directly; widgets are essentially "given an area, paint it." |
| OpenTUI (Zig+TS) | Retained scene-graph + reactive bindings | Flexbox | Tree diffing buys little for a fully-redraws-each-frame app and adds significant indirection. |

The tinear TUI is **already shaped like ratatui** without anyone planning
it that way: `AppState` is rebuilt each frame in `app::run_loop` and
passed (immutably) to `draw_screen`, components hold persistent state
(`scroll`, `cursor`, `top`) while receiving per-frame layout from the
parent, and state mutations live exclusively in the input handler. The
missing pieces are a `Rect` / `Layout` / `Frame` system and a unified
component contract.

A complication: Ard's trait system was, until very recently, partially
broken. v0.21.0 type-checked user traits but the Go backend rejected
dispatch ([akonwi/ard#172](https://github.com/akonwi/ard/issues/172),
fixed in v0.21.1). v0.21.1 dispatched at call sites but not at storage
positions ([akonwi/ard#175](https://github.com/akonwi/ard/issues/175),
fixed on the current dev branch). A separate codegen bug rejected
`Void`-returning trait methods (also fixed on the dev branch). With
those fixes in place, traits now support function parameters, struct
fields, list elements, and nested dispatch — enough to express a
genuine open-world `Widget` interface rather than a closed union with
central match dispatch.

## Decision

Introduce a small framework layer at `tui/core/` that the rest of
`tui/` builds on. The framework provides:

1. **`Rect{ x, y, w, h }`** — the universal unit of layout. Replaces
   the ad-hoc `tui/screen::Layout`.

2. **`Constraint` / `Direction` / `Layout`** — declarative,
   constraint-based subdivision of a `Rect`. A parent calls
   `Layout::vertical([…]).split(area)` and receives child rects. This
   is the ratatui constraint model: `Length(n)`, `Percent(n)`,
   `Min(n)`, `Max(n)`, `Fill(weight)`.

3. **`Frame`** — a value type carrying the `vaxis::Vaxis` handle plus a
   pre-clipped `vaxis::Window`. All drawing happens through Frame
   primitives. The Frame value itself is immutable; the underlying
   terminal buffer is mutated through FFI.

4. **`trait Widget { fn render(frame: Frame) }`** — the universal
   component contract. Every component implements this trait. Real
   trait dispatch is used (not a closed union).

5. **`Frame::render(self, area: Rect, w: Widget)`** — the
   framework's composition primitive, exposed as a method on Frame
   rather than a top-level helper. It creates a subwindow of
   `self.window` clipped to `area`, wraps it in a new Frame, and
   invokes `w.render(child_frame)`. Children always see a
   `(0,0)..(w,h)` coordinate space scoped to their assigned region.
   This is the flexbox semantic: the parent decides layout, the child
   paints into the box it was given.

   The trait method `Widget::render(self, frame: Frame)` and the Frame
   method `Frame::render(self, area: Rect, w: Widget)` share a name by
   design — the receiver type makes the intent unambiguous:

   ```
   widget.render(frame)         # "widget, paint yourself onto this frame"
   frame.render(area, widget)   # "frame, render this widget at this area"
   ```

6. **A small set of primitive widgets**, factored by role rather than
   by use case. Eight in total:

   *Atoms* — produce visible content, no children:
   - `Text(content, style, wrap)` — single-line or word-wrapped
     styled text. Truncates to frame width when `wrap` is false.
   - `Fill(char, style)` — fill the frame with one character.

   *Layout* — arrange children in space:
   - `VStack(children, constraints)` — stack top-to-bottom under
     `Layout`-style constraints.
   - `HStack(children, constraints)` — stack left-to-right.
   - `ZStack(children)` — overlay children at the same position;
     later children draw over earlier ones (for modals, popups).
   - `Spacer` — placeholder widget that consumes flex space when
     paired with `Constraint::Fill(weight)`.

   *Container* — wrap a single child with decoration:
   - `Box(child, border?, padding?, title?, background?)` — the
     only decoration primitive. Border, padding, title, and
     background fill are properties of `Box`, applied in a fixed
     order (background → border → title in top border → padding →
     child). Pure border is `Box{child, border: Single}`; pure
     padding is `Box{child, padding: All(1)}`; the classic
     "bordered titled panel" pattern is
     `Box{child, border: Single, title: " Inbox "}`. Decoration
     outside a border (margin) composes by nesting another `Box`.

   *Behavior* — wrap a single child with state:
   - `Scroll(child)` — child is logically larger than the frame;
     widget holds `scroll_offset` and clips to the visible window.

   The set is deliberately primitive — there is no `Block`, no
   `Tabs`, no `List with selection`. Those are app-level
   compositions of the primitives above and live in tinear's `tui/`
   directory, not in `tui/core/` (see Conventions below). The first
   uses of the primitives will absorb the bespoke box-drawing in
   `modals.ard`, `board.ard`, `notifications.ard`, and the tab bar in
   `draw.ard`.

### Conventions

- **Storage uses concrete types.** State holds `Scrollview`, `Box`,
  etc. — not `Widget`. The trait coerces at call sites, which is
  enough for composition and dispatch. Heterogeneous widget lists
  (`[Widget]`) are supported but rarely needed.
- **`render` is non-mutating on `self`.** Widgets own their state via
  `mut` fields, but state mutations happen only in input handlers.
  This preserves the "AppState is rebuilt each frame; draw paths are
  pure" property documented in `AGENTS.md`.
- **Frames are pre-clipped, not parameterized by area.** Widgets never
  receive their position. They receive a `Frame` already scoped to
  their region and address it with `(0,0)..(w,h)` coordinates. This
  matches flexbox: position is the parent's concern.
- **Composition flows through Frame methods.** There are no top-level
  framework helpers a widget needs to import. Everything a widget needs
  to draw or to embed children hangs off the `Frame` value it is
  handed.
- **App-specific compositions live in `tui/`, not `tui/core/`.**
  Tab bars, modal layouts, selectable lists, issue cards, the inbox
  detail panel — these are tinear-specific patterns built from `core`
  primitives. They stay in app code. A composition graduates into
  `tui/core/` only when it proves reusable across many call sites and
  generic enough to belong in the framework.
- **The framework lives in this repo for now.** Path: `tui/core/`.
  If it stabilizes and proves reusable, extracting to a standalone
  Ard module is a follow-up — not a goal for the first cut.

### Sequencing

1. Implement `rect.ard`, `layout.ard`, `frame.ard`, `widget.ard` (the
   trait).
2. Implement `Box` as the first widget; verify the composition
   chain end-to-end (Box exercises decoration, child rendering, and
   the Frame composition method in one shot).
3. Implement the layout widgets (`VStack`, `HStack`, `ZStack`,
   `Spacer`) and the remaining atoms (`Text`, `Fill`).
4. Add `Scroll` once a real screen needs it.
5. Port a modal as the first real surface (smallest and most-bordered;
   lowest risk for shaking out API gaps). Evaluate ergonomics.
6. Iterate on the framework API based on what the port reveals.
7. Port remaining screens incrementally.

## Consequences

### Positive

- The duplicated box-drawing, geometry math, and per-component draw
  signatures collapse into a small reusable surface.
- New screens compose existing widgets rather than reinvent rendering
  primitives.
- The `Widget` trait gives a clear extension point — a third party (or
  a future contributor) implements one method and the rest of the
  framework cooperates.
- Encourages a clean split between **state** (mutated by input
  handlers) and **render** (pure, side-effecting via Frame). This is
  already the project's convention; the framework enforces it.

### Negative / open

- A new abstraction layer to learn for contributors and future AI
  sessions. Mitigated by keeping the surface small and well-documented.
- The first port (a modal) is the riskiest moment — it may surface
  API gaps that force iteration on the framework before broader
  rollout.
- The framework is tinear-coupled until proven otherwise. Extraction
  is a deliberate later decision, not an unstated goal.

### Neutral

- **Event handling is out of scope for this ADR.** The framework
  defines the rendering contract only. Per-widget event handling
  (whether a `Widget` trait gains a second `handle_event` method,
  whether focus is tracked, how key events route to the active widget)
  is a future concern, deferred until the rendering split clarifies
  what state actually needs to mutate. The existing input-dispatch
  loop in `tui/app.ard` is unaffected by this ADR.
- The `vaxis.ard` bindings stay as-is; the framework sits on top.

## Related

- [ADR 0002](0002-interactive-terminal-explorer.md) — established the
  TUI as a first-class feature; this ADR addresses the structural
  consequences of that growth.
- [akonwi/ard#172](https://github.com/akonwi/ard/issues/172) — trait
  dispatch in the Go backend (fixed in v0.21.1).
- [akonwi/ard#175](https://github.com/akonwi/ard/issues/175) — trait
  coercion at storage positions (fixed on dev).
- ratatui — the design influence for the immediate-mode + constraint-
  layout shape: https://ratatui.rs/
