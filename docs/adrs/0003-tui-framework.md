# 0003: TUI Framework Layer

## Status

Accepted

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
  of "this view owns this rectangle; render within it."

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

(Terminology note: the rendered things are **views**, the trait is
**`View`**. Other frameworks call equivalents "widgets" or
"components"; this ADR uses "view" for our framework and reserves
"widget" for references to other systems' equivalents, e.g.
ratatui's `Widget` trait.)

## Decision

Introduce a small framework layer at `tui/core/` that the rest of
`tui/` builds on. The framework provides:

1. **`Rect{ x, y, w, h }`** — the universal unit of layout. Replaces
   the ad-hoc `tui/screen::Layout`.

2. **Layout engine** in `stack.ard` — two complementary models for
   composing multiple children:

   *Fit-content stacks*, for content-sized rows/columns where the
   parent's extent is the union of its children's drawn areas:
   - `VStack(children)` — children top-to-bottom
   - `HStack(children)` — children left-to-right
   - `Spacer::v(rows)` / `Spacer::h(cols)` — explicit gap

   *Constraint-based splits*, for screen-shell geometry where the
   parent owns a fixed area and partitions it among children:
   - `vsplit([Sized])` / `hsplit([Sized])` — take a list of
     `Sized{view, kind: SizeKind}` pairs.
   - `SizeKind` has six variants: `Fixed(n)`, `Percent(n)`,
     `Ratio(num, den)`, `Fill(weight)`, `AtLeast(n)`, `AtMost(n)`.
     Mirrors ratatui's `Length`, `Percentage`, `Ratio`, `Fill`,
     `Min`, `Max`.
   - The solver (`solve(available, [SizeKind]) [Int]`) is a pure
     function from total length + kinds to per-child sizes,
     unit-testable in isolation. Layout went in with twelve inline
     `test fn` cases for the solver covering all kinds and their
     interactions.

   The two stack families coexist deliberately: fit-content is the
   right model when children determine the parent's size (e.g. a
   modal that's as tall as its rows); constraint-based is the right
   model when the parent has a fixed area to divide (e.g. the app
   shell with a fixed tab bar, fixed status bar, and a body that
   takes the rest). There is intentionally **no** `FitContent`
   constraint that mixes the two — it requires a measure pass that
   doubles render cost and doesn't compose cleanly with the other
   kinds. Pick a stack flavor per axis instead.

3. **`Frame`** — the drawable surface a view renders into. Wraps a
   `vaxis::Vaxis` handle plus a pre-clipped `vaxis::Window`. Frame
   is **mutable** during render: as a view draws, the frame tracks
   the bounding rectangle of what was actually written. A parent
   reads `child_frame.dimensions()` after delegating to a child to
   discover how much space the child consumed, which is how
   fit-content stacks compute their layout without a separate
   measure pass.

   All drawing happens through Frame methods (`draw_text`,
   `draw_styled`, `clear`, `flush`). Composition happens through the
   free function `view::sub` (see #5).

4. **`trait View { fn render(mut frame: Frame) }`** — the universal
   component contract. Every component implements this trait. Real
   trait dispatch is used (not a closed union). The frame is `mut`
   because draws expand its tracked extent; the view itself never
   mutates its own state during render.

5. **`sub(frame: Frame, area: Rect) Frame`** — the framework's
   composition primitive, exposed as a free function in the `view`
   module. It carves a sub-rectangle of `frame.window` clipped to
   `area` and returns a new Frame whose own `(0, 0)` corresponds to
   `area`'s upper-left in the parent. A parent view composes by
   handing that sub-frame straight to its child's `render`:

   ```
   self.child.render(view::sub(frame, inner_area))
   ```

   This separates responsibilities cleanly: `sub` only produces
   frames; rendering only happens via `View::render`. Children always
   see a `(0, 0)..(w, h)` coordinate space scoped to their assigned
   region — the flexbox semantic where the parent decides layout and
   the child paints into the box it was given.

   `sub` is a free function rather than a `Frame` method because the
   Ard checker stack-overflows on impl methods that return their own
   struct type. Revert to a method when that compiler bug is fixed.

6. **A small set of primitive views**, factored by role rather than
   by use case:

   *Atoms* — produce visible content, no children:
   - `Text` — a single line of styled text with explicit fg / bg /
     bold / dim / italic / underline / reverse arguments.

   *Layout* — arrange children in space (see #2):
   - `VStack`, `HStack`, `Spacer::v`, `Spacer::h` (fit-content)
   - `vsplit`, `hsplit` (constraint-based, with `Sized`/`SizeKind`)

   *Container* — wrap a single child with decoration:
   - `Box(child, style: Style)` — the wrapping view.

   *Style* — the decoration *value* applied via `Box`, kept in its
   own module so it doesn't couple to any particular view (cut 1 of
   a future Lipgloss-flavored design where the same `Style` could
   apply to anything). `Style` holds:
   - `padding: Padding?`, `border: BorderStyle?`,
     `border_title: View?` (the title is itself a view, so it can
     be styled / composed), `background_char: Str?`
   - `width: Int?`, `height: Int?` (explicit sizing when needed,
     e.g. fullscreen overlays)
   - text attributes that propagate to the box's own painted cells
     (fg, bg, bold, dim, italic, underline, reverse)

   `Box::new(child, style)` is the single wrapping constructor;
   different decoration patterns are different `Style` values, not
   different constructors. `Padding::all(n)`, `Padding::x(n)`,
   `Padding::y(n)` are the padding factories.

   The set is deliberately primitive — there is no `Block`, no
   `Tabs`, no `List with selection`. Those are app-level
   compositions of the primitives above and live in tinear's `tui/`
   directory, not in `tui/core/` (see Conventions below). The first
   uses of the primitives will absorb the bespoke box-drawing in
   `modals.ard`, `board.ard`, `notifications.ard`, and the tab bar in
   `draw.ard`.

### Conventions

- **Storage uses concrete types.** State holds `Scrollview`, `Box`,
  etc. — not `View`. The trait coerces at call sites, which is enough
  for composition and dispatch. Heterogeneous view lists (`[View]`)
  are supported but rarely needed.
- **`render` is non-mutating on `self`.** Views own their state via
  `mut` fields, but state mutations happen only in input handlers.
  This preserves the "AppState is rebuilt each frame; draw paths are
  pure" property documented in `AGENTS.md`.
- **Frames are pre-clipped, not parameterized by area.** Views never
  receive their position. They receive a `Frame` already scoped to
  their region and address it with `(0,0)..(w,h)` coordinates. This
  matches flexbox: position is the parent's concern.
- **Composition uses `view::sub`.** A parent computes its child's
  `Rect`, calls `view::sub(frame, area)` to produce a clipped
  sub-frame, and forwards it directly to the child's `render`. Frame
  itself knows nothing about views.
- **App-specific compositions live in `tui/`, not `tui/core/`.**
  Tab bars, modal layouts, selectable lists, issue cards, the inbox
  detail panel — these are tinear-specific patterns built from `core`
  primitives. They stay in app code. A composition graduates into
  `tui/core/` only when it proves reusable across many call sites and
  generic enough to belong in the framework.
- **The framework lives in this repo for now.** Path: `tui/core/`.
  If it stabilizes and proves reusable, extracting to a standalone
  Ard module is a follow-up — not a goal for the first cut.

### Implementation status

Landed on branch `feat.tui-framework` (commits `591bd0c` through
`ddfe76f`):

- [x] `rect.ard`, `view.ard` (Frame + trait), `text.ard`, `box.ard`,
  `style.ard`, `stack.ard` (fit-content stacks + constraint solver +
  splits)
- [x] `commands/core_demo.ard` smoke test renders fullscreen mock
  inbox: tab bar + items + status bar pinned via `vsplit`
- [x] Layout solver has twelve inline `test fn` cases (`ard test`
  green)

Remaining sequencing for downstream work:

1. Port a modal as the first real surface (smallest and most-bordered;
   lowest risk for shaking out API gaps). Evaluate ergonomics.
2. Iterate on the framework API based on what the port reveals.
3. Port remaining screens incrementally (board, inbox, issue tabs).
4. Add `Scroll` once one of those ports needs it.
5. Per-view event handling — deferred to a separate ADR (see
   Consequences).

### Decisions worth flagging

A handful of choices came up during implementation that are worth
recording explicitly:

- **Style is its own module**, not a property bag baked into `Box`.
  Border / padding / title / fg / bg / attrs are widget-agnostic
  decoration values. Today only `Box` consumes a `Style`; future
  views can pull from the same module without depending on `Box`.
  A later cut may unify text attributes and spatial decoration into
  a single `styled(view, style)` operator that wraps any view
  (Lipgloss-flavored) — this module is the home that lands in.
- **Title is a `View?`**, not a `Str`. The title fragment composes
  like any other view (it can be styled, padded, swapped for a
  `Spacer`), and the `Box` doesn't have to know what's in it.
- **No `FitContent` constraint kind.** Mixing intrinsic content size
  with constraint solving requires a measure pass that doubles
  render work. If you need content-sized children, use the
  fit-content stacks; if you need to partition a known area, use
  the constraint splits.


## Consequences

### Positive

- The duplicated box-drawing, geometry math, and per-component draw
  signatures collapse into a small reusable surface.
- New screens compose existing views rather than reinvent rendering
  primitives.
- The `View` trait gives a clear extension point — a third party (or
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
  defines the rendering contract only. Per-view event handling
  (whether the `View` trait gains a second `handle_event` method,
  whether focus is tracked, how key events route to the active view)
  is a future concern, deferred until the rendering split clarifies
  what state actually needs to mutate. The existing input-dispatch
  loop in `tui/app.ard` is unaffected by this ADR.
- The `vaxis.ard` bindings stay as-is; the framework sits on top.

## Related

- [ADR 0002](0002-interactive-terminal-explorer.md) — established the
  TUI as a first-class feature; this ADR addresses the structural
  consequences of that growth.
- ratatui — the design influence for the immediate-mode + constraint-
  layout shape: https://ratatui.rs/
