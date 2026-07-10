// Package ffi bridges only what Ard cannot express: vaxis's stateful-widget
// pattern relies on embedding ui.StateBase (Go struct embedding). This generic
// shim owns the embedding and persists typed state in the framework's State
// object (surviving rebuilds), while all build/event logic stays in Ard. State
// is read/written with the generic StateValue/StateSet functions.
package ffi

import (
	"time"

	"go.rockorager.dev/vaxis/ui"
)

// StateCtx holds a stateful widget's persistent, caller-typed state plus its
// rebuild trigger. It is opaque to Ard, threaded through StateValue/StateSet.
type StateCtx struct {
	value          any
	markNeedsBuild func()
	anim           *ui.AnimationController
	runtime        ui.Runtime
	disposed       bool
}

// Dispatch posts fn onto the UI thread. Background fibers must route all
// state mutation through here: StateRef + MarkDirty are only safe on the
// UI thread. Dispatching after unmount is a silent no-op upstream.
func Dispatch(c *StateCtx, fn func()) {
	if c != nil && !c.disposed && c.runtime != nil {
		c.runtime.Dispatch(func() {
			if c.disposed {
				return
			}
			fn()
		})
	}
}

// Animation returns the StateBase-owned animation controller. It is nil unless
// the widget set Stateful.Animate. The controller's Value/Status/Forward/Stop/
// Reset are driven from Ard; only its creation (NewAnimation, Duration, Curve)
// is Go plumbing.
func Animation(c *StateCtx) *ui.AnimationController { return c.anim }

// StateValue returns a snapshot copy of the current state for rendering.
func StateValue[T any](c *StateCtx) T {
	if p, ok := c.value.(*T); ok {
		return *p
	}
	return c.value.(T)
}

// StateSet replaces the whole state value (copy-then-write model).
func StateSet[T any](c *StateCtx, v T) {
	if c.disposed {
		return
	}
	if p, ok := c.value.(*T); ok {
		*p = v
	} else {
		c.value = v
	}
	MarkDirty(c)
}

// StateRef returns a live *T handle to the persistent state so callers can
// mutate fields in place. On first use it promotes the stored value to a
// pointer; every later call returns that same pointer, so mutations persist
// across rebuilds. Ard binds *T to `mut T`, so this reads as `mut T`.
func StateRef[T any](c *StateCtx) *T {
	if p, ok := c.value.(*T); ok {
		return p
	}
	v := c.value.(T)
	p := &v
	c.value = p
	return p
}

// Update runs mutate against the live *T state pointer, then schedules a
// rebuild. React-like callback ergonomics with a live pointer (not a copy).
func Update[T any](c *StateCtx, mutate func(*T)) {
	if c.disposed {
		return
	}
	mutate(StateRef[T](c))
	MarkDirty(c)
}

// MarkDirty schedules a rebuild after in-place state mutation.
func MarkDirty(c *StateCtx) {
	if c.disposed {
		return
	}
	if c.markNeedsBuild != nil {
		c.markNeedsBuild()
	}
}

// Stateful is a StatefulWidget whose initial state and build are Ard closures.
// Build receives the persistent StateCtx and the live BuildContext so Ard can
// read inherited values (e.g. the theme) and drive the runtime.
type Stateful struct {
	Init  func() any
	Build func(c *StateCtx, ctx ui.BuildContext) ui.Widget
	// OnTick, when set, is invoked about once per second on the UI thread. It
	// lets Ard drive time-based state (a background goroutine posting through
	// Runtime.Dispatch is Go-only plumbing; the increment logic stays in Ard).
	OnTick func(c *StateCtx)
	// OnDispose, when set, is invoked exactly once when vaxis disposes this
	// stateful widget. Ard uses it to cancel widget-owned background work.
	OnDispose func(c *StateCtx)
	// Animate, when true, creates a StateBase-owned AnimationController the
	// runner ticks before each frame; reachable from Ard via ffi.Animation.
	Animate bool
	// AnimateMs overrides the animation duration (milliseconds) and switches
	// the curve to linear. Zero keeps the 1200ms EaseInOut default.
	AnimateMs int
}

func (w Stateful) CreateState() ui.State { return &statefulState{} }

type statefulState struct {
	ui.StateBase
	ctx  *StateCtx
	stop chan struct{}
	anim *ui.AnimationController
}

// InitState starts the once-per-second ticker when the widget wants one. The
// goroutine posts back through Runtime.Dispatch so the Ard OnTick handler runs
// on the UI thread, where mutating state and scheduling a rebuild is safe.
func (s *statefulState) InitState() {
	w := s.StateBase.Widget().(Stateful)
	if w.Animate {
		options := ui.AnimationOptions{Duration: 1200 * time.Millisecond, Curve: ui.EaseInOut}
		if w.AnimateMs > 0 {
			options = ui.AnimationOptions{Duration: time.Duration(w.AnimateMs) * time.Millisecond, Curve: ui.Linear}
		}
		s.anim = s.NewAnimation(options)
	}
	if w.OnTick == nil {
		return
	}
	rt := s.Context().Runtime()
	s.stop = make(chan struct{})
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				rt.Dispatch(func() {
					if s.ctx != nil {
						s.StateBase.Widget().(Stateful).OnTick(s.ctx)
					}
				})
			case <-s.stop:
				return
			}
		}
	}()
}

func (s *statefulState) Dispose() {
	if s.ctx != nil && !s.ctx.disposed {
		w := s.StateBase.Widget().(Stateful)
		if w.OnDispose != nil {
			w.OnDispose(s.ctx)
		}
		s.ctx.disposed = true
		s.ctx.runtime = nil
		s.ctx.markNeedsBuild = nil
	}
	if s.stop != nil {
		close(s.stop)
	}
}

func (s *statefulState) Build(bctx ui.BuildContext) ui.Widget {
	// Re-read the CURRENT widget config (a parent may have rebuilt us with a
	// new closure); state persists in s.ctx across rebuilds.
	w := s.StateBase.Widget().(Stateful)
	if s.ctx == nil {
		s.ctx = &StateCtx{markNeedsBuild: s.MarkNeedsBuild, value: w.Init(), runtime: s.Context().Runtime()}
	}
	s.ctx.anim = s.anim
	return w.Build(s.ctx, bctx)
}
