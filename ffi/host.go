package ffi

import (
	"fmt"
	"os/exec"
	"reflect"
	goruntime "runtime"
	"strings"
	"time"
	"unicode/utf8"

	"git.sr.ht/~rockorager/vaxis"
	"git.sr.ht/~rockorager/vaxis/ui"
	"git.sr.ht/~rockorager/vaxis/ui/uitest"
	ardruntime "github.com/akonwi/ard/runtime"
)

// OpenURL launches the system's default browser/handler for the given URL.
func OpenURL(url string) error {
	var cmd *exec.Cmd
	switch goruntime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}

func TestVaxisNil() *vaxis.Vaxis {
	return nil
}

var linearTextNormalizer = strings.NewReplacer(
	// Correct common mojibake forms first.
	"â\u0080\u0094", "-",
	"â\u0080\u0093", "-",
	"â\u0080\u0098", "'",
	"â\u0080\u0099", "'",
	"â\u0080\u009c", "\"",
	"â\u0080\u009d", "\"",
	"â\u0080¦", "...",
	"â\u0080¢", "*",
	"â€”", "-",
	"â€“", "-",
	"â€˜", "'",
	"â€™", "'",
	"â€œ", "\"",
	"â€�", "\"",
	"â€¦", "...",
	"â€¢", "*",
	"‚Äî", "-",
	"‚Äì", "-",
	"‚Äò", "'",
	"‚Äô", "'",
	"‚Äú", "\"",
	"‚Äù", "\"",
	"‚Ä¶", "...",
	"‚Ä¢", "*",
	"\u00a0", " ",
)

func NormalizeLinearText(value string) string {
	normalized := linearTextNormalizer.Replace(value)
	// Degraded mojibake can arrive/render as variable runs of Â around ¢. Treat
	// the padding as noise and the cent marker as a dash fallback.
	normalized = strings.ReplaceAll(normalized, "Â", "")
	normalized = strings.ReplaceAll(normalized, "¢", "-")
	normalized = strings.ReplaceAll(normalized, "â", "-")
	normalized = strings.ReplaceAll(normalized, "�", "")
	return normalized
}

type UiStringIntent string

func (i UiStringIntent) IntentType() ui.IntentType { return ui.IntentType(i) }

func UiRun(root ui.Widget) error {
	shortcuts := ui.DefaultShortcuts()
	delete(shortcuts, "Tab")
	delete(shortcuts, "Shift+Tab")
	return ui.Run(root, ui.WithShortcuts(shortcuts))
}

type UiTestApp struct{ app *uitest.App }

func UiTestNew(root ui.Widget) *UiTestApp {
	return &UiTestApp{app: uitest.New(root)}
}

func UiTestPump(app *UiTestApp, width, height int) {
	if app == nil || app.app == nil {
		return
	}
	app.app.Pump(width, height)
}

func UiTestKey(app *UiTestApp, text string) {
	if app == nil || app.app == nil {
		return
	}
	app.app.Key(text)
}

func UiTestEnter(app *UiTestApp) {
	if app == nil || app.app == nil {
		return
	}
	app.app.Enter()
}

func UiTestCtrlJ(app *UiTestApp) {
	if app == nil || app.app == nil {
		return
	}
	app.app.Send(vaxis.Key{Text: "j", Keycode: 'j', Modifiers: vaxis.ModCtrl})
}

func UiTestEscape(app *UiTestApp) {
	if app == nil || app.app == nil {
		return
	}
	app.app.Send(vaxis.Key{Keycode: vaxis.KeyEsc})
}

func UiTestUp(app *UiTestApp) {
	if app == nil || app.app == nil {
		return
	}
	app.app.Send(vaxis.Key{Keycode: vaxis.KeyUp})
}

func UiTestDown(app *UiTestApp) {
	if app == nil || app.app == nil {
		return
	}
	app.app.Send(vaxis.Key{Keycode: vaxis.KeyDown})
}

func UiTestLeft(app *UiTestApp) {
	if app == nil || app.app == nil {
		return
	}
	app.app.Send(vaxis.Key{Keycode: vaxis.KeyLeft})
}

func UiTestRight(app *UiTestApp) {
	if app == nil || app.app == nil {
		return
	}
	app.app.Send(vaxis.Key{Keycode: vaxis.KeyRight})
}

func UiTestTab(app *UiTestApp) {
	if app == nil || app.app == nil {
		return
	}
	app.app.Tab()
}

func UiTestShiftTab(app *UiTestApp) {
	if app == nil || app.app == nil {
		return
	}
	app.app.ShiftTab()
}

func UiTestClick(app *UiTestApp, x, y int) {
	if app == nil || app.app == nil {
		return
	}
	app.app.Click(x, y)
}

func UiTestContains(app *UiTestApp, text string) bool {
	if app == nil || app.app == nil {
		return false
	}
	return app.app.Contains(text)
}

func UiTestText(app *UiTestApp) string {
	if app == nil || app.app == nil {
		return ""
	}
	return app.app.Text()
}

func UiTestCellGrapheme(app *UiTestApp, x, y int) string {
	if app == nil || app.app == nil {
		return ""
	}
	return app.app.Cell(x, y).Grapheme
}

func UiTestCellReverse(app *UiTestApp, x, y int) bool {
	if app == nil || app.app == nil {
		return false
	}
	return app.app.Cell(x, y).Attribute&vaxis.AttrReverse != 0
}

func UiTestShouldQuit(app *UiTestApp) bool {
	if app == nil || app.app == nil {
		return false
	}
	return app.app.ShouldQuit()
}

type UiEventContext struct{ handle ui.EventContext }

type UiActionBinding struct {
	Name    string
	Handler func(ui.EventContext, string) ui.EventResult
}

func UiAction[T ~int](name string, handler func(*UiEventContext, string) T) UiActionBinding {
	return UiActionBinding{
		Name: name,
		Handler: func(ctx ui.EventContext, intent string) ui.EventResult {
			if handler == nil {
				return ui.EventIgnored
			}
			return ui.EventResult(handler(&UiEventContext{handle: ctx}, intent))
		},
	}
}

func UiActions(child ui.Widget, bindings []UiActionBinding) ui.Widget {
	mapped := map[ui.IntentType]ui.ActionFunc{}
	for _, binding := range bindings {
		intentName := binding.Name
		action := binding.Handler
		mapped[ui.IntentType(intentName)] = func(ctx ui.EventContext, intent ui.Intent) ui.EventResult {
			if action == nil {
				return ui.EventIgnored
			}
			return action(ctx, string(intent.IntentType()))
		}
	}
	return ui.Actions{Bindings: mapped, Child: child}
}

type uiShortcutsCompat struct {
	Bindings map[string]ui.Intent
	Child    ui.Widget
	Capture  bool
}

type renderShortcutsCompat struct {
	ui.SingleChildRenderObject
	Bindings map[string]ui.Intent
	Capture  bool
}

func UiShortcuts(child ui.Widget, bindings map[string]string) ui.Widget {
	mapped := map[string]ui.Intent{}
	for binding, intentName := range bindings {
		mapped[binding] = UiStringIntent(intentName)
	}
	return uiShortcutsCompat{Bindings: mapped, Child: child}
}

func UiShortcutsCapture(child ui.Widget, bindings map[string]string) ui.Widget {
	mapped := map[string]ui.Intent{}
	for binding, intentName := range bindings {
		mapped[binding] = UiStringIntent(intentName)
	}
	return uiShortcutsCompat{Bindings: mapped, Child: child, Capture: true}
}

func (w uiShortcutsCompat) WidgetChild() ui.Widget {
	return w.Child
}

func (w uiShortcutsCompat) CreateRenderObject(ctx ui.BuildContext) ui.RenderObject {
	return &renderShortcutsCompat{Bindings: w.Bindings, Capture: w.Capture}
}

func (w uiShortcutsCompat) UpdateRenderObject(ctx ui.BuildContext, ro ui.RenderObject) {
	ro.(*renderShortcutsCompat).Bindings = w.Bindings
	ro.(*renderShortcutsCompat).Capture = w.Capture
}

func (r *renderShortcutsCompat) Layout(ctx ui.LayoutContext, c ui.Constraints) {
	if child := r.Child(); child != nil {
		child.Layout(ctx, c)
		r.SetSize(child.Base().Size())
		return
	}
	r.SetSize(c.Constrain(ui.Size{}))
}

func (r *renderShortcutsCompat) DryLayout(ctx ui.LayoutContext, c ui.Constraints) ui.Size {
	if child := r.Child(); child != nil {
		return ui.DryLayout(ctx, child, c)
	}
	return c.Constrain(ui.Size{})
}

func (r *renderShortcutsCompat) Paint(p *ui.Painter, off ui.Offset) {
	if child := r.Child(); child != nil {
		child.Paint(p, off)
	}
}

func (r *renderShortcutsCompat) HitTest(*ui.HitTestResult, ui.Point) bool {
	return false
}

func (r *renderShortcutsCompat) HandleEvent(ctx ui.EventContext, ev ui.Event) ui.EventResult {
	if r.Capture {
		if ctx.Phase() != ui.CapturePhase {
			return ui.EventIgnored
		}
	} else if ctx.Phase() != ui.TargetPhase && ctx.Phase() != ui.BubblePhase {
		return ui.EventIgnored
	}
	key, ok := ev.(ui.Key)
	if !ok || key.EventType == ui.EventRelease {
		return ui.EventIgnored
	}
	for binding, intent := range r.Bindings {
		if shortcutKeyMatches(key, binding) && ctx.Invoke(intent) == ui.EventHandled {
			return ui.EventHandled
		}
	}
	return ui.EventIgnored
}

func shortcutKeyMatches(key ui.Key, binding string) bool {
	if key.MatchString(binding) {
		return true
	}
	// vaxis 0.16 has no MatchString spelling for keypad Enter. Treat it as
	// Enter so app shortcuts don't have to know about that backend wart.
	return strings.EqualFold(binding, "Enter") && key.Keycode == vaxis.KeyKeyPadEnter
}

var uiFocusNodes = map[string]*ui.FocusNode{}

func uiFocusNode(key string) *ui.FocusNode {
	if node, ok := uiFocusNodes[key]; ok {
		return node
	}
	node := &ui.FocusNode{}
	uiFocusNodes[key] = node
	return node
}

func UiFocusable(child ui.Widget) ui.Widget {
	return ui.Focus(nil, child)
}

func UiFocusableKey(key string, child ui.Widget) ui.Widget {
	return ui.Focus(uiFocusNode(key), child)
}

func UiFocusScope(child ui.Widget, trap bool, autoFocus bool) ui.Widget {
	return ui.FocusScope{Trap: trap, AutoFocus: autoFocus, Child: child}
}

func UiRequestFocus(key string) {
	uiFocusNode(key).RequestFocus()
}

type uiKeyDebug struct{ Child ui.Widget }

type toastVariant int

const (
	toastInfo toastVariant = iota
	toastError
	toastDebug
)

const defaultToastTTL = 5 * time.Second

type toastEntry struct {
	title      []ui.Character
	desc       []ui.Character
	titleWidth int
	descWidth  int
	variant    toastVariant
	seq        int
}

type renderKeyDebug struct {
	ui.SingleChildRenderObject
	items []toastEntry
	seq   int
}

var activeToastHost *renderKeyDebug

func UiKeyDebug(child ui.Widget) ui.Widget {
	return uiKeyDebug{Child: child}
}

func (w uiKeyDebug) WidgetChild() ui.Widget {
	return w.Child
}

func (w uiKeyDebug) CreateRenderObject(ctx ui.BuildContext) ui.RenderObject {
	r := &renderKeyDebug{}
	activeToastHost = r
	return r
}

func (w uiKeyDebug) UpdateRenderObject(ctx ui.BuildContext, ro ui.RenderObject) {
	if r, ok := ro.(*renderKeyDebug); ok {
		activeToastHost = r
	}
}

func (r *renderKeyDebug) Layout(ctx ui.LayoutContext, c ui.Constraints) {
	if child := r.Child(); child != nil {
		child.Layout(ctx, c)
		r.SetSize(child.Base().Size())
		return
	}
	r.SetSize(c.Constrain(ui.Size{}))
}

func (r *renderKeyDebug) DryLayout(ctx ui.LayoutContext, c ui.Constraints) ui.Size {
	if child := r.Child(); child != nil {
		return ui.DryLayout(ctx, child, c)
	}
	return c.Constrain(ui.Size{})
}

func (r *renderKeyDebug) HitTest(*ui.HitTestResult, ui.Point) bool {
	return false
}

func toastTTL(ttlMs *int) time.Duration {
	if ttlMs != nil && *ttlMs > 0 {
		return time.Duration(*ttlMs) * time.Millisecond
	}
	return defaultToastTTL
}

func toastCharacters(value string, maxWidth int) ([]ui.Character, int) {
	chars := vaxis.Characters(value)
	width := 0
	for _, ch := range chars {
		width += ch.Width
	}
	if maxWidth <= 0 || width <= maxWidth {
		return chars, width
	}
	if maxWidth == 1 {
		return []ui.Character{{Grapheme: "…", Width: 1}}, 1
	}
	out := []ui.Character{}
	used := 0
	for _, ch := range chars {
		if used+ch.Width > maxWidth-1 {
			break
		}
		out = append(out, ch)
		used += ch.Width
	}
	out = append(out, ui.Character{Grapheme: "…", Width: 1})
	return out, used + 1
}

func (e toastEntry) height() int {
	if len(e.desc) > 0 {
		return 2
	}
	return 1
}

func (e toastEntry) width() int {
	return 3 + max(e.titleWidth, e.descWidth)
}

func (r *renderKeyDebug) show(runtime ui.Runtime, title string, desc *string, variant toastVariant, ttl time.Duration) {
	size := r.Size()
	maxWidth := min(48, max(7, size.Width/2))
	contentWidth := max(4, maxWidth-3)
	titleChars, titleWidth := toastCharacters(title, contentWidth)
	descChars := []ui.Character(nil)
	descWidth := 0
	if desc != nil && *desc != "" {
		descChars, descWidth = toastCharacters(*desc, contentWidth)
	}

	r.seq++
	seq := r.seq
	r.items = append(r.items, toastEntry{title: titleChars, desc: descChars, titleWidth: titleWidth, descWidth: descWidth, variant: variant, seq: seq})
	if len(r.items) > 3 {
		r.items = r.items[len(r.items)-3:]
	}
	r.MarkNeedsPaint()
	time.AfterFunc(ttl, func() {
		runtime.Dispatch(func() {
			for i, item := range r.items {
				if item.seq == seq {
					r.items = append(r.items[:i], r.items[i+1:]...)
					r.MarkNeedsPaint()
					break
				}
			}
		})
	})
}

func toastBaseStyle() ui.Style {
	return ui.Style{}
}

func toastAccentStyle(variant toastVariant) ui.Style {
	style := toastBaseStyle()
	switch variant {
	case toastError:
		style.Foreground = vaxis.IndexColor(9)
		style.Attribute |= ui.AttrBold
	case toastDebug:
		style.Foreground = vaxis.IndexColor(11)
	default:
		style.Foreground = vaxis.IndexColor(14)
	}
	return style
}

func toastTitleStyle(variant toastVariant) ui.Style {
	return toastBaseStyle()
}

func toastDescStyle(variant toastVariant) ui.Style {
	style := toastBaseStyle()
	style.Attribute |= ui.AttrDim
	return style
}

func (r *renderKeyDebug) paintEntry(p *ui.Painter, e toastEntry, off ui.Offset, right int, topY int) {
	w := e.width()
	x := right - w
	base := toastBaseStyle()
	blank := ui.Cell{Character: ui.Character{Grapheme: " ", Width: 1}, Style: base}
	for row := 0; row < e.height(); row++ {
		for i := 0; i < w; i++ {
			p.DrawCell(ui.Point{X: x + i, Y: topY + row}, blank)
		}
	}

	p.DrawCell(ui.Point{X: x, Y: topY}, ui.Cell{Character: ui.Character{Grapheme: "╿", Width: 1}, Style: toastAccentStyle(e.variant)})
	cx := x + 2
	for _, ch := range e.title {
		p.DrawCell(ui.Point{X: cx, Y: topY}, ui.Cell{Character: ch, Style: toastTitleStyle(e.variant)})
		cx += ch.Width
	}

	if len(e.desc) > 0 {
		cx = x + 2
		for _, ch := range e.desc {
			p.DrawCell(ui.Point{X: cx, Y: topY + 1}, ui.Cell{Character: ch, Style: toastDescStyle(e.variant)})
			cx += ch.Width
		}
	}
}

func (r *renderKeyDebug) Paint(p *ui.Painter, off ui.Offset) {
	if child := r.Child(); child != nil {
		child.Paint(p, off)
	}
	if len(r.items) == 0 {
		return
	}
	size := r.Size()
	visible := len(r.items)
	if size.Height < 12 && visible > 2 {
		visible = 2
	}
	right := off.X + max(0, size.Width-1)
	cursorY := off.Y + max(0, size.Height-2)
	start := len(r.items) - visible
	for i := len(r.items) - 1; i >= start; i-- {
		entry := r.items[i]
		h := entry.height()
		topY := cursorY - h + 1
		if topY < off.Y {
			break
		}
		r.paintEntry(p, entry, off, right, topY)
		cursorY = topY - 2
	}
}

func (r *renderKeyDebug) HandleEvent(ctx ui.EventContext, ev ui.Event) ui.EventResult {
	key, ok := ev.(ui.Key)
	if !ok || ctx.Phase() != ui.CapturePhase {
		return ui.EventIgnored
	}
	desc := fmt.Sprintf("code=%d text=%q mods=%d", key.Keycode, key.Text, key.Modifiers)
	r.show(ctx.Runtime(), key.String(), &desc, toastDebug, defaultToastTTL)
	return ui.EventIgnored
}

func normalizeToastVariant(variant any) toastVariant {
	idx := 0
	value := reflect.ValueOf(variant)
	if value.IsValid() {
		switch value.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			idx = int(value.Int())
		}
	}
	v := toastVariant(idx)
	if v < toastInfo || v > toastDebug {
		v = toastInfo
	}
	return v
}

func maybeStringPtr(value ardruntime.Maybe[string]) *string {
	if value.IsSome() {
		v := value.Value()
		return &v
	}
	return nil
}

func maybeIntPtr(value ardruntime.Maybe[int]) *int {
	if value.IsSome() {
		v := value.Value()
		return &v
	}
	return nil
}

func UiShowToast(ctx *UiEventContext, title string, description ardruntime.Maybe[string], variant any, ttlMs ardruntime.Maybe[int]) {
	if ctx == nil || activeToastHost == nil {
		return
	}
	activeToastHost.show(ctx.handle.Runtime(), title, maybeStringPtr(description), normalizeToastVariant(variant), toastTTL(maybeIntPtr(ttlMs)))
}

func UiShowToastRuntime(runtime ui.Runtime, title string, description ardruntime.Maybe[string], variant any, ttlMs ardruntime.Maybe[int]) {
	if activeToastHost == nil {
		return
	}
	runtime.Dispatch(func() {
		if activeToastHost != nil {
			activeToastHost.show(runtime, title, maybeStringPtr(description), normalizeToastVariant(variant), toastTTL(maybeIntPtr(ttlMs)))
		}
	})
}

func UiRequestFrameWrapped(ctx *UiEventContext) {
	if ctx == nil {
		return
	}
	ctx.handle.Runtime().Dispatch(func() {})
}

func UiDispatchAfterWrapped(ctx *UiEventContext, delayMs int, callback func()) {
	if ctx == nil {
		return
	}
	runtime := ctx.handle.Runtime()
	time.AfterFunc(time.Duration(delayMs)*time.Millisecond, func() {
		runtime.Dispatch(callback)
	})
}

func UiQuitWrapped(ctx *UiEventContext) {
	if ctx == nil {
		return
	}
	ctx.handle.Quit()
}

type UiStateContext struct{ state *ardState }

type ardStateful struct {
	key   string
	init  func(ui.BuildContext, *UiStateContext)
	build func(ui.BuildContext, *UiStateContext) ui.Widget
}

type ardState struct {
	ui.StateBase
	value        any
	initializing bool
	disposed     bool
	init         func(ui.BuildContext, *UiStateContext)
	build        func(ui.BuildContext, *UiStateContext) ui.Widget
}

func UiStatefulValueInit[T any](init func(ui.BuildContext, *UiStateContext) T, build func(ui.BuildContext, *UiStateContext) ui.Widget) ui.Widget {
	return ardStateful{
		init: func(ctx ui.BuildContext, state *UiStateContext) {
			state.state.value = init(ctx, state)
		},
		build: build,
	}
}

func UiStatefulValueInitKey[T any](key string, init func(ui.BuildContext, *UiStateContext) T, build func(ui.BuildContext, *UiStateContext) ui.Widget) ui.Widget {
	return ardStateful{
		key: key,
		init: func(ctx ui.BuildContext, state *UiStateContext) {
			state.state.value = init(ctx, state)
		},
		build: build,
	}
}

func (w ardStateful) WidgetKey() ui.KeyValue {
	return ui.KeyValue(w.key)
}

func (w ardStateful) CreateState() ui.State {
	return &ardState{init: w.init, build: w.build}
}

func (s *ardState) InitState() {
	if s.init != nil {
		s.initializing = true
		s.init(s.Context(), &UiStateContext{state: s})
		s.initializing = false
	}
}

func (s *ardState) DidUpdateWidget(old ui.Widget) {
	w := s.Widget().(ardStateful)
	s.init = w.init
	s.build = w.build
}

func (s *ardState) Dispose() {
	s.disposed = true
}

func (s *ardState) setValue(fn func()) {
	if s.initializing {
		fn()
		return
	}
	if s.disposed {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			if fmt.Sprint(r) != "ui: MarkNeedsBuild called after Dispose" {
				panic(r)
			}
		}
	}()
	s.SetState(fn)
}

func (s *ardState) Build(ctx ui.BuildContext) ui.Widget {
	return s.build(ctx, &UiStateContext{state: s})
}

func UiBuildContextRuntime(ctx ui.BuildContext) ui.Runtime {
	return ctx.Runtime()
}

func UiRuntimeDispatch(rt ui.Runtime, state ardruntime.Maybe[*UiStateContext], callback func(*UiStateContext)) {
	if rt == nil || callback == nil {
		return
	}
	if state.IsNone() || state.Value() == nil {
		panic("ui runtime dispatch requires state")
	}
	stateCtx := state.Value()
	rt.Dispatch(func() {
		if stateCtx.state == nil || stateCtx.state.disposed {
			return
		}
		callback(stateCtx)
	})
}

func UiStateValue[T any](ctx *UiStateContext) T {
	var zero T
	if ctx == nil || ctx.state == nil {
		return zero
	}
	if value, ok := ctx.state.value.(T); ok {
		return value
	}
	return zero
}

func UiStateSetValue[T any](ctx *UiStateContext, value T) {
	if ctx == nil || ctx.state == nil {
		return
	}
	ctx.state.setValue(func() { ctx.state.value = value })
}

type UiStyle struct {
	Bold       bool
	Reverse    bool
	Foreground string
}

func UiStyleNew(bold bool, reverse bool, foreground string) UiStyle {
	return UiStyle{Bold: bold, Reverse: reverse, Foreground: foreground}
}

func (s UiStyle) vaxisStyle() ui.Style {
	return s.vaxisStyleWithTheme(ui.Theme{})
}

func (s UiStyle) vaxisStyleWithTheme(theme ui.Theme) ui.Style {
	style := ui.Style{}
	if s.Bold {
		style.Attribute |= ui.AttrBold
	}
	if s.Reverse {
		style.Attribute |= ui.AttrReverse
	}
	switch s.Foreground {
	case "border":
		style.Foreground = theme.Border
	}
	return style
}

type uiThemedText struct {
	Value string
	Style ardruntime.Maybe[UiStyle]
}

func (w uiThemedText) CreateState() ui.State { return &uiThemedTextState{} }

type uiThemedTextState struct{ ui.StateBase }

func (s *uiThemedTextState) Build(ctx ui.BuildContext) ui.Widget {
	w := s.Widget().(uiThemedText)
	widget := ui.Text{Value: w.Value}
	if w.Style.IsSome() {
		theme := ui.MustDepend[ui.Theme](ctx)
		widget.Style = w.Style.Value().vaxisStyleWithTheme(theme)
	}
	return widget
}

func UiText(value string, style ardruntime.Maybe[UiStyle]) ui.Widget {
	return uiThemedText{Value: value, Style: style}
}

type uiBackground struct {
	Child ui.Widget
	Style UiStyle
}

type renderBackground struct {
	ui.SingleChildRenderObject
	Style ui.Style
}

func uiStyledBackground(child ui.Widget, style UiStyle) ui.Widget {
	return uiBackground{Child: child, Style: style}
}

func (w uiBackground) WidgetChild() ui.Widget {
	return w.Child
}

func (w uiBackground) CreateRenderObject(ctx ui.BuildContext) ui.RenderObject {
	return &renderBackground{Style: w.Style.vaxisStyle()}
}

func (w uiBackground) UpdateRenderObject(ctx ui.BuildContext, ro ui.RenderObject) {
	r := ro.(*renderBackground)
	r.Style = w.Style.vaxisStyle()
	r.MarkNeedsLayout()
}

func (r *renderBackground) Layout(ctx ui.LayoutContext, c ui.Constraints) {
	width := 0
	height := 0
	if child := r.Child(); child != nil {
		childConstraints := c
		if c.HasBoundedWidth() {
			childConstraints.MinWidth = c.MaxWidth
			childConstraints.MaxWidth = c.MaxWidth
		}
		child.Layout(ctx, childConstraints)
		size := child.Base().Size()
		width = size.Width
		height = size.Height
	}
	if c.HasBoundedWidth() {
		width = c.MaxWidth
	}
	r.SetSize(c.Constrain(ui.Size{Width: width, Height: height}))
}

func (r *renderBackground) DryLayout(ctx ui.LayoutContext, c ui.Constraints) ui.Size {
	width := 0
	height := 0
	if child := r.Child(); child != nil {
		childConstraints := c
		if c.HasBoundedWidth() {
			childConstraints.MinWidth = c.MaxWidth
			childConstraints.MaxWidth = c.MaxWidth
		}
		size := ui.DryLayout(ctx, child, childConstraints)
		width = size.Width
		height = size.Height
	}
	if c.HasBoundedWidth() {
		width = c.MaxWidth
	}
	return c.Constrain(ui.Size{Width: width, Height: height})
}

func (r *renderBackground) Paint(p *ui.Painter, off ui.Offset) {
	size := r.Size()
	blank := ui.Cell{Character: ui.Character{Grapheme: " ", Width: 1}, Style: r.Style}
	for y := 0; y < size.Height; y++ {
		for x := 0; x < size.Width; x++ {
			p.DrawCell(ui.Point{X: off.X + x, Y: off.Y + y}, blank)
		}
	}
	if child := r.Child(); child != nil {
		child.Paint(p, off)
	}
}

func (r *renderBackground) HitTest(*ui.HitTestResult, ui.Point) bool {
	return false
}

func UiRow(children []ui.Widget, style ardruntime.Maybe[UiStyle]) ui.Widget {
	row := ui.Row(children...)
	if style.IsSome() {
		return uiStyledBackground(row, style.Value())
	}
	return row
}

func UiColumn(children []ui.Widget) ui.Widget {
	return ui.Column(children...)
}

func UiColumnStretch(children []ui.Widget) ui.Widget {
	return ui.Flex{
		Axis:               ui.Vertical,
		CrossAxisAlignment: ui.CrossAxisStretch,
		Children:           children,
	}
}

func UiIndexedStack(index int, children []ui.Widget) ui.Widget {
	return ui.IndexedStack{Index: index, Children: children}
}

func UiScrollView(child ui.Widget) ui.Widget {
	return ui.ScrollView{Child: child}
}

func UiScrollPaneController() *ui.ScrollPaneController {
	return &ui.ScrollPaneController{}
}

func UiScrollPane(controller *ui.ScrollPaneController, child ui.Widget) ui.Widget {
	return ui.ScrollPane{Controller: controller, Child: child}
}

func UiScrollPaneScrollBy(controller *ui.ScrollPaneController, cols, rows int) bool {
	if controller == nil {
		return false
	}
	return controller.ScrollBy(cols, rows)
}

func UiColumnMin(children []ui.Widget) ui.Widget {
	return ui.Flex{
		Axis:               ui.Vertical,
		MainAxisSize:       ui.MainAxisSizeMin,
		CrossAxisAlignment: ui.CrossAxisStretch,
		Children:           children,
	}
}

func UiCenter(child ui.Widget) ui.Widget {
	return ui.Center(child)
}

func UiPaddingAll(all int, child ui.Widget) ui.Widget {
	return ui.Padding(ui.All(all), child)
}

func UiPaddingHorizontal(horizontal int, child ui.Widget) ui.Widget {
	return ui.Padding(ui.Symmetric(horizontal, 0), child)
}

func UiSizedBox(width int, height int) ui.Widget {
	return ui.SizedBox{Width: width, Height: height}
}

func UiSizedBoxChild(width int, height int, child ui.Widget) ui.Widget {
	return ui.SizedBox{Width: width, Height: height, Child: child}
}

func UiConstrainedWidth(width int, child ui.Widget) ui.Widget {
	return ui.ConstrainedBox{Constraints: ui.Constraints{MinWidth: width, MaxWidth: width}, Child: child}
}

func UiExpanded(child ui.Widget) ui.Widget {
	return ui.Expanded(child)
}

type uiWidthObserver struct {
	Child   ui.Widget
	OnWidth func(int)
}

func (w uiWidthObserver) WidgetChild() ui.Widget {
	return w.Child
}

func (w uiWidthObserver) CreateRenderObject(ctx ui.BuildContext) ui.RenderObject {
	return &renderWidthObserver{OnWidth: w.OnWidth, lastWidth: -1}
}

func (w uiWidthObserver) UpdateRenderObject(ctx ui.BuildContext, ro ui.RenderObject) {
	r := ro.(*renderWidthObserver)
	r.OnWidth = w.OnWidth
	r.MarkNeedsLayout()
}

type renderWidthObserver struct {
	ui.SingleChildRenderObject
	OnWidth   func(int)
	lastWidth int
}

func (r *renderWidthObserver) notify(width int) {
	if width == r.lastWidth {
		return
	}
	r.lastWidth = width
	if r.OnWidth != nil {
		r.OnWidth(width)
	}
}

func (r *renderWidthObserver) Layout(ctx ui.LayoutContext, c ui.Constraints) {
	if child := r.Child(); child != nil {
		child.Layout(ctx, c)
		size := child.Base().Size()
		if c.HasBoundedWidth() {
			size.Width = c.MaxWidth
		}
		r.SetSize(c.Constrain(size))
	} else {
		r.SetSize(c.Constrain(ui.Size{}))
	}
	r.notify(r.Size().Width)
}

func (r *renderWidthObserver) DryLayout(ctx ui.LayoutContext, c ui.Constraints) ui.Size {
	if child := r.Child(); child != nil {
		size := ui.DryLayout(ctx, child, c)
		if c.HasBoundedWidth() {
			size.Width = c.MaxWidth
		}
		return c.Constrain(size)
	}
	return c.Constrain(ui.Size{})
}

func (r *renderWidthObserver) Paint(p *ui.Painter, off ui.Offset) {
	if child := r.Child(); child != nil {
		child.Paint(p, off)
	}
}

func (r *renderWidthObserver) HitTest(result *ui.HitTestResult, point ui.Point) bool {
	if child := r.Child(); child != nil {
		return child.HitTest(result, point)
	}
	return false
}

func UiOnWidth(child ui.Widget, onWidth func(int)) ui.Widget {
	return uiWidthObserver{Child: child, OnWidth: onWidth}
}

type uiThemedDivider struct {
	Style    ardruntime.Maybe[UiStyle]
	Vertical bool
}

func (w uiThemedDivider) CreateState() ui.State { return &uiThemedDividerState{} }

type uiThemedDividerState struct{ ui.StateBase }

func (s *uiThemedDividerState) Build(ctx ui.BuildContext) ui.Widget {
	w := s.Widget().(uiThemedDivider)
	widget := ui.Divider{}
	if w.Vertical {
		widget.Axis = ui.Vertical
	}
	if w.Style.IsSome() {
		theme := ui.MustDepend[ui.Theme](ctx)
		widget.Style = w.Style.Value().vaxisStyleWithTheme(theme)
	}
	return widget
}

func UiDivider(style ardruntime.Maybe[UiStyle], vertical bool) ui.Widget {
	return uiThemedDivider{Style: style, Vertical: vertical}
}

func normalizeOverlayPosition(position int) ui.Alignment {
	switch position {
	case 1:
		return ui.TopLeft
	case 2:
		return ui.TopRight
	case 3:
		return ui.BottomLeft
	case 4:
		return ui.BottomRight
	default:
		return ui.CenterAlign
	}
}

// using this generic with constraint for easier enum interop without reflecting
func UiOverlay[T ~int](child ui.Widget, overlay ui.Widget, position T, trapFocus bool, autoFocus bool) ui.Widget {
	entryChild := overlay
	if trapFocus || autoFocus {
		entryChild = ui.FocusScope{Trap: trapFocus, AutoFocus: autoFocus, Child: overlay}
	}
	return ui.Overlay{
		Child: child,
		Entries: []ui.OverlayEntry{{
			Child:     entryChild,
			Alignment: normalizeOverlayPosition(int(position)),
		}},
	}
}

type UiBoxBorder struct {
	Top    bool
	Right  bool
	Bottom bool
	Left   bool
	Style  ardruntime.Maybe[UiStyle]
}

func UiBorderNew(top bool, right bool, bottom bool, left bool, style ardruntime.Maybe[UiStyle]) UiBoxBorder {
	return UiBoxBorder{Top: top, Right: right, Bottom: bottom, Left: left, Style: style}
}

type uiBox struct {
	Child  ardruntime.Maybe[ui.Widget]
	Width  ardruntime.Maybe[int]
	Height ardruntime.Maybe[int]
	Border ardruntime.Maybe[UiBoxBorder]
	Style  ardruntime.Maybe[UiStyle]
}

func (w uiBox) CreateState() ui.State { return &uiBoxState{} }

type uiBoxState struct{ ui.StateBase }

func (s *uiBoxState) Build(ctx ui.BuildContext) ui.Widget {
	w := s.Widget().(uiBox)
	var child ui.Widget
	if w.Child.IsSome() {
		child = w.Child.Value()
	}

	hasDecoration := w.Style.IsSome() || w.Border.IsSome()
	if hasDecoration {
		theme := ui.MustDepend[ui.Theme](ctx)
		decoration := ui.Decoration{}
		if w.Style.IsSome() {
			decoration.Style = w.Style.Value().vaxisStyleWithTheme(theme)
		}
		if w.Border.IsSome() {
			b := w.Border.Value()
			borderStyle := ui.Style{Foreground: theme.Border}
			if b.Style.IsSome() {
				borderStyle = b.Style.Value().vaxisStyleWithTheme(theme)
			}
			decoration.Border = ui.Border{
				Top:    b.Top,
				Right:  b.Right,
				Bottom: b.Bottom,
				Left:   b.Left,
				Style:  borderStyle,
			}
		}
		child = ui.DecoratedBox(decoration, child)
	}

	if w.Width.IsSome() || w.Height.IsSome() {
		width := 0
		height := 0
		if w.Width.IsSome() {
			width = w.Width.Value()
		}
		if w.Height.IsSome() {
			height = w.Height.Value()
		}
		return ui.SizedBox{Width: width, Height: height, Child: child}
	}
	if child != nil {
		return child
	}
	return ui.SizedBox{}
}

func UiBox(child ardruntime.Maybe[ui.Widget], width ardruntime.Maybe[int], height ardruntime.Maybe[int], border ardruntime.Maybe[UiBoxBorder], style ardruntime.Maybe[UiStyle]) ui.Widget {
	return uiBox{Child: child, Width: width, Height: height, Border: border, Style: style}
}

func UiTextField(value string, placeholder string, minWidth int, obscure bool, onChanged func(ui.EventContext, string), onSubmitted func(ui.EventContext, string)) ui.Widget {
	return ui.TextField{
		Value:       value,
		Placeholder: placeholder,
		MinWidth:    minWidth,
		ObscureText: obscure,
		OnChanged: func(ctx ui.EventContext, next string) {
			if onChanged != nil {
				onChanged(ctx, next)
			}
		},
		OnSubmitted: func(ctx ui.EventContext, submitted string) {
			if onSubmitted != nil {
				onSubmitted(ctx, submitted)
			}
		},
	}
}

func UiTextFieldWrapped(value string, placeholder string, minWidth int, obscure bool, onChanged func(*UiEventContext, string), onSubmitted func(*UiEventContext, string)) ui.Widget {
	return UiTextField(
		value,
		placeholder,
		minWidth,
		obscure,
		func(ctx ui.EventContext, next string) {
			if onChanged != nil {
				onChanged(&UiEventContext{handle: ctx}, next)
			}
		},
		func(ctx ui.EventContext, submitted string) {
			if onSubmitted != nil {
				onSubmitted(&UiEventContext{handle: ctx}, submitted)
			}
		},
	)
}

func UiTextArea(value string, placeholder string, minWidth int, minHeight int, softWrap bool, onChanged func(ui.EventContext, string), onSubmitted func(ui.EventContext, string)) ui.Widget {
	area := ui.TextArea{
		Value:       value,
		Placeholder: placeholder,
		MinWidth:    minWidth,
		MinHeight:   minHeight,
		SoftWrap:    softWrap,
		OnChanged: func(ctx ui.EventContext, next string) {
			if onChanged != nil {
				onChanged(ctx, next)
			}
		},
	}
	const submitIntent = ui.IntentType("tinear.textarea.submit")
	return ui.Shortcuts{
		Bindings: map[string]ui.Intent{
			"Ctrl+Enter": UiStringIntent(submitIntent),
			"Ctrl+m":     UiStringIntent(submitIntent),
			"Ctrl+j":     UiStringIntent(submitIntent),
		},
		Child: ui.Actions{
			Bindings: map[ui.IntentType]ui.ActionFunc{
				submitIntent: func(ctx ui.EventContext, intent ui.Intent) ui.EventResult {
					if onSubmitted != nil {
						onSubmitted(ctx, value)
					}
					return ui.EventHandled
				},
			},
			Child: area,
		},
	}
}

func UiTextAreaWrapped(value string, placeholder string, minWidth int, minHeight int, softWrap bool, onChanged func(*UiEventContext, string), onSubmitted func(*UiEventContext, string)) ui.Widget {
	return UiTextArea(
		value,
		placeholder,
		minWidth,
		minHeight,
		softWrap,
		func(ctx ui.EventContext, next string) {
			if onChanged != nil {
				onChanged(&UiEventContext{handle: ctx}, next)
			}
		},
		func(ctx ui.EventContext, submitted string) {
			if onSubmitted != nil {
				onSubmitted(&UiEventContext{handle: ctx}, submitted)
			}
		},
	)
}

func UiButton(label string, onPressed func(ui.EventContext)) ui.Widget {
	return ui.Button{Label: label, OnPressed: func(ctx ui.EventContext) {
		if onPressed != nil {
			onPressed(ctx)
		}
	}}
}

func UiButtonWrapped(label string, onPressed func(*UiEventContext)) ui.Widget {
	return UiButton(label, func(ctx ui.EventContext) {
		if onPressed != nil {
			onPressed(&UiEventContext{handle: ctx})
		}
	})
}

var testRenderMarks = map[int]bool{}

type testCell struct {
	Grapheme  string
	Fg        int
	Bg        int
	Bold      bool
	Dim       bool
	Italic    bool
	Underline bool
	Reverse   bool
}

var testScreenWidth int
var testScreenHeight int
var testScreenCells []testCell

func TestWindowZero() vaxis.Window {
	return vaxis.Window{}
}

func TestResetScreen(width, height int) {
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}
	testScreenWidth = width
	testScreenHeight = height
	testScreenCells = make([]testCell, width*height)
}

func TestWindow(width, height int) vaxis.Window {
	TestResetScreen(width, height)
	return vaxis.Window{Width: width, Height: height}
}

func TestCellGrapheme(x, y int) string {
	if x < 0 || y < 0 || x >= testScreenWidth || y >= testScreenHeight {
		return ""
	}
	return testScreenCells[y*testScreenWidth+x].Grapheme
}

func TestCellReverse(x, y int) bool {
	if x < 0 || y < 0 || x >= testScreenWidth || y >= testScreenHeight {
		return false
	}
	return testScreenCells[y*testScreenWidth+x].Reverse
}

func TestLine(y int) string {
	if y < 0 || y >= testScreenHeight {
		return ""
	}
	var b strings.Builder
	for x := 0; x < testScreenWidth; x++ {
		g := TestCellGrapheme(x, y)
		if g == "" {
			g = " "
		}
		b.WriteString(g)
	}
	return b.String()
}

func TestResetRenderMarks() {
	testRenderMarks = map[int]bool{}
}

func TestMarkRendered(id int) {
	testRenderMarks[id] = true
}

func TestWasRendered(id int) bool {
	return testRenderMarks[id]
}

func TestMarkTabSelected(index int) {
	testRenderMarks[1000+index] = true
}

func TextBackspace(text string) string {
	if text == "" {
		return ""
	}
	_, size := utf8.DecodeLastRuneInString(text)
	if size <= 0 {
		return ""
	}
	return text[:len(text)-size]
}

func makeStyle(fg int, bg int, bold bool, dim bool, italic bool, underline bool, reverse bool) vaxis.Style {
	style := vaxis.Style{}
	if fg >= 0 {
		style.Foreground = colorFromInt(fg)
	}
	if bg >= 0 {
		style.Background = colorFromInt(bg)
	}
	if bold {
		style.Attribute |= vaxis.AttrBold
	}
	if dim {
		style.Attribute |= vaxis.AttrDim
	}
	if italic {
		style.Attribute |= vaxis.AttrItalic
	}
	if reverse {
		style.Attribute |= vaxis.AttrReverse
	}
	if underline {
		style.UnderlineStyle = vaxis.UnderlineSingle
	}
	return style
}

func testWindowOrigin(win vaxis.Window) (int, int) {
	col := 0
	row := 0
	w := win
	for {
		col += w.Column
		row += w.Row
		if w.Parent == nil {
			return col, row
		}
		w = *w.Parent
	}
}

func writeTestCell(win vaxis.Window, x, y int, cell testCell) {
	width, height := win.Size()
	if x < 0 || y < 0 || x >= width || y >= height {
		return
	}
	absX, absY := testWindowOrigin(win)
	absX += x
	absY += y
	if absX < 0 || absY < 0 || absX >= testScreenWidth || absY >= testScreenHeight {
		return
	}
	testScreenCells[absY*testScreenWidth+absX] = cell
}

func writeTestText(win vaxis.Window, x, y int, text string, style testCell) {
	col := x
	for _, r := range text {
		if r == '\n' {
			col = x
			y++
			continue
		}
		cell := style
		cell.Grapheme = string(r)
		writeTestCell(win, col, y, cell)
		col++
	}
}

func writeTestFill(win vaxis.Window, cell testCell) {
	width, height := win.Size()
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			writeTestCell(win, x, y, cell)
		}
	}
}

func colorFromInt(value int) vaxis.Color {
	if value >= 0 && value <= 255 {
		return vaxis.IndexColor(uint8(value))
	}
	return vaxis.HexColor(uint32(value))
}
