# 0004: TUI App Loop and Event Dispatch

## Status

Proposed

## Context

[ADR 0003](0003-tui-framework.md) defined the rendering side of the TUI
framework: the `View` trait, `Frame` surface, `Style`, the layout
engine. It explicitly deferred event handling and the main loop:

> Event handling is out of scope for this ADR. The framework defines
> the rendering contract only. Per-view event handling (whether the
> `View` trait gains a second `handle_event` method, whether focus is
> tracked, how key events route to the active view) is a future
> concern...

The framework is now at a point where the demo only renders one frame
and exits on any keypress. To grow the rebuild — port the inbox,
modals, board, issue tabs — we need a loop that drives interaction
and a clear story for where event-handling logic lives.

There are three established models worth comparing:

| Model | Description |
|---|---|
| **Centralized app dispatch** | App owns all state, all event handling lives in one big `match`. Views are pure render. This is what `tui/app.ard` does today. |
| **Per-view dispatch through a trait** | Views have a `handle_event` method. Containers route events to children (typically the focused child). The framework owns routing. |
| **Focus-based / browser-shaped** | App tracks a focused view; events route via capture/bubble through the render tree to the focus and back. |

None of these are the Elm Architecture (Bubble Tea). Pure-functional
`Update(model, msg) -> (model, Cmd)` requires cheap immutability and
sum-typed message composition that Ard does not give us today.
Mutable receivers (`fn mut handle(...)`) are the language's grain;
fighting that grain to get TEA's purity benefits would cost more in
boilerplate than the benefits would return. Bubble Tea-style is off
the table.

The choice between centralized, per-view, and focus-based hinges on:

- **Focus in a terminal is less obvious than in a GUI.** There is no
  cursor pointer, no click-to-focus. Which view "has focus" is
  always an explicit app-level decision (the current mode, the
  active modal, the selected tab). The framework cannot infer it.
- **The project's existing grain is centralized.** `AppState` is
  rebuilt each frame in `run_loop` from local mutable variables.
  State mutations live exclusively in the input handler. Stateful
  components (`Scrollview`, `CommentList`) are concrete-typed fields
  in `AppState` that the input handler reaches into directly.

A focus-based model would require formalizing something the program
already manages by hand, with no clear benefit. A per-view trait
method (`fn handle_event` on `View`) would force every view —
including atoms like `Text` and structurals like `VStack` — to
declare a method most of them don't care about, without resolving
the "which child gets the event?" question for multi-child
containers (which the framework can't answer without a focus
concept).

## Decision

The framework adds two pieces — a small `App` trait and a runner.
The `View` trait stays render-only. Event-handling logic lives on
concrete component types, and the app routes events explicitly
through its own state.

### `App` trait and runner (`tui/core/app.ard`)

```ard
trait App {
  fn render(mut frame: view::Frame)
  fn mut handle_event(event: vaxis::Event)
}

fn run(mut app: App, vx: vaxis::Vaxis) {
  loop {
    let frame = view::root(vx)
    frame.clear()
    app.render(frame)
    frame.flush()

    let event = vaxis::next_event(vx)
    match event {
      vaxis::QuitEvent => break,
      _ => app.handle_event(event),
    }
  }
}
```

The two trait methods correspond to the two halves of every frame:

- `render` builds and renders the view tree from the app's current
  state.
- `handle_event` mutates the app's state in response to an event. Both
  methods return `Void`; the app does not communicate control flow to
  the runner through return values.

The runner owns the loop's lifecycle. It exits when it observes a
`QuitEvent` in the event stream — vaxis's natural shutdown signal,
produced when the terminal closes, when a signal interrupts
`next_event`, or when something else has explicitly posted one. The
app is not involved in this decision; it never sees `QuitEvent`. Apps
do not currently have a way to terminate the loop programmatically
(see Neutral consequences); the only exit path in slice 3a is the
vaxis-produced `QuitEvent`.

### Dispatch convention: top-down via app state, no framework routing

The `View` trait stays single-purpose:

```ard
trait View {
  fn render(mut frame: view::Frame)
}
```

Stateful, behavior-bearing views (e.g. a selectable list, a text
input, a scrollable region) implement `View` like any other, *and*
expose their own concrete-type methods with whatever signature suits
them — `handle_key(k: KeyEvent) Bool`, `scroll_up()`, `submit()`. The
framework does not constrain these method signatures.

The app holds those stateful views as **concrete-typed fields** in
its own state struct. Its `handle_event` is the single dispatch
site: a plain `match` over the app's state that decides which
concrete component gets the event.

```ard
struct InboxApp {
  active_tab: Int,
  list: SelectableList,
  modal: Modal?,
  // ...
}

impl core::App for InboxApp {
  fn render(mut frame: view::Frame) {
    let body = stack::vsplit([
      stack::fixed(tab_bar(self.active_tab), 1),
      stack::fill(self.list, 1),
      stack::fixed(hint_bar(), 1),
    ])
    body.render(frame)
  }

  fn mut handle_event(event: vaxis::Event) {
    // Modals win when open.
    match self.modal {
      m => {
        if m.handle_event(event) { return }
      },
      _ => (),
    }

    // Active mode's primary component.
    match event {
      vaxis::KeyEvent(k) => {
        if self.list.handle_key(k) { return }
      },
      _ => (),
    }

    // App-level fallbacks: tab switch, mode toggles.
    match event {
      vaxis::KeyEvent(k) => match k.name {
        "Tab" => { self.active_tab = (self.active_tab + 1) % 3 },
        _ => (),
      },
      _ => (),
    }
  }
}
```

This routing is **top-down through the app's state tree**, not
through the render tree. Events never enter the render tree at all.
The render tree is a per-frame artifact; the state tree is the
persistent source of truth that receives events.

### Conventions

- **One `App` instance per program.** The `run` runner takes
  ownership of it for the program's lifetime. Initial state is
  constructed before `run` is called.
- **`render` builds the view tree fresh each frame** from the app's
  current state. No memoization, no diffing at the framework level
  — vaxis's cell-buffer diffing handles redraw efficiency.
- **`handle_event` is the only mutation site for app state.** It is
  called exactly once per event, sees the event as a `vaxis::Event`
  union value, and matches on whatever variants it cares about.
- **Stateful views own their state.** `SelectableList`, `TextInput`,
  etc. carry their cursor/scroll/value as `mut` fields and expose
  concrete-type methods for input. They implement `View` for
  rendering; they are not framework-required to implement any
  particular event-handling interface.
- **The app routes events explicitly.** When the app needs to
  forward an event to a stateful child, it calls a concrete method
  on that child (`self.list.handle_key(k)`). The framework does
  not auto-route.
- **Single-child structural views do not forward events.** A `Box`
  wrapping a `SelectableList` doesn't relay events to the list —
  the app calls the list's `handle_key` directly through its
  concrete-typed field. The render tree's job is to draw; routing
  is the state tree's job.

## Consequences

### Positive

- **Small framework surface.** Two trait methods and one runner.
  Nothing about focus, routing, control flow, or events lives in the
  framework beyond the loop itself.
- **Exit is a natural consequence of the event stream.** The runner
  watches for `QuitEvent` and breaks; the app stays out of it. No
  return value, no flag, no enum, no extra method on the trait.
- **No design risk pushed into the framework prematurely.** Focus,
  bubbling, capture, dispatch order — all of these are app-level
  concerns we can grow into if/when needed, without disturbing
  view authors who don't need them.
- **View trait stays orthogonal to event handling.** Rendering
  polymorphism and event handling are independent concerns. A
  third party (or a future contributor) can write a view that's
  purely visual without learning anything about event flow.
- **Concrete-typed access to stateful components is the only honest
  model in the absence of focus.** Without focus tracking, a
  trait-typed dispatch through `[View]` has no defensible "which
  child wins" answer. Naming the components and routing to them by
  name is what the program already does, just made explicit.

### Negative / open

- **Stateful views aren't drop-in interchangeable across apps
  through trait dispatch alone.** A view that wants both rendering
  and event handling has an app-private contract beyond `View`.
  This is fine for an internal framework; if/when the framework is
  extracted as a library, the contract may want to be formalized.
- **`handle_event`'s big-match risk.** As the app grows, its
  `handle_event` will accumulate cases. The current `tui/app.ard`
  is the proof point — it's manageable but gets long. If/when it
  becomes unwieldy, the answer is per-mode dispatch functions, not
  framework-level routing.
- **Multi-tab / multi-screen flows must explicitly forward** to the
  active tab's component. No automatic "the visible view gets the
  event."

### Neutral / deferred

- **App-initiated quit is not yet supported.** Today the loop exits
  only when vaxis emits a `QuitEvent` (terminal close, signal). An
  app cannot say "quit now" in response to a `q` key without a way
  to inject a `QuitEvent` into the event queue. This is deferred
  intentionally — it's the same fundamental problem as views
  wanting to wake the loop for background work, and both will get
  a single coherent answer in a later slice.
- **Re-rendering on non-input signals (background fetches, timers)
  is similarly deferred.** The mechanism is the same as above: a
  way for code outside the synchronous event-handling path to push
  events into the queue. vaxis already exposes
  `post`; whether and how to surface that to apps and
  views is a future ADR.

## Related

- [ADR 0003](0003-tui-framework.md) — established the rendering
  framework and deferred event handling. This ADR fills in the
  rendering driver and event dispatch story; programmatic exit and
  external event injection remain deferred to a future ADR.
