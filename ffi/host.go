package ffi

import (
	"fmt"
	"os/exec"
	goruntime "runtime"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"git.sr.ht/~rockorager/vaxis"
	"git.sr.ht/~rockorager/vaxis/ui"
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

type UiStringIntent string

func (i UiStringIntent) IntentType() ui.IntentType { return ui.IntentType(i) }

func UiRun(root ui.Widget) error {
	shortcuts := ui.DefaultShortcuts()
	delete(shortcuts, "Tab")
	delete(shortcuts, "Shift+Tab")
	return ui.Run(root, ui.WithShortcuts(shortcuts))
}

func UiActions[T ~int](child ui.Widget, bindings map[string]func(ui.EventContext, string) T) ui.Widget {
	mapped := map[ui.IntentType]ui.ActionFunc{}
	for name, handler := range bindings {
		intentName := name
		action := handler
		mapped[ui.IntentType(intentName)] = func(ctx ui.EventContext, intent ui.Intent) ui.EventResult {
			return ui.EventResult(action(ctx, string(intent.IntentType())))
		}
	}
	return ui.Actions{Bindings: mapped, Child: child}
}

type uiShortcutsCompat struct {
	Bindings map[string]ui.Intent
	Child    ui.Widget
}

type renderShortcutsCompat struct {
	ui.SingleChildRenderObject
	Bindings map[string]ui.Intent
}

func UiShortcuts(child ui.Widget, bindings map[string]string) ui.Widget {
	mapped := map[string]ui.Intent{}
	for binding, intentName := range bindings {
		mapped[binding] = UiStringIntent(intentName)
	}
	return uiShortcutsCompat{Bindings: mapped, Child: child}
}

func (w uiShortcutsCompat) WidgetChild() ui.Widget {
	return w.Child
}

func (w uiShortcutsCompat) CreateRenderObject(ctx ui.BuildContext) ui.RenderObject {
	return &renderShortcutsCompat{Bindings: w.Bindings}
}

func (w uiShortcutsCompat) UpdateRenderObject(ctx ui.BuildContext, ro ui.RenderObject) {
	ro.(*renderShortcutsCompat).Bindings = w.Bindings
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

func UiFocusable(child ui.Widget) ui.Widget {
	return ui.Focus(nil, child)
}

type uiKeyDebug struct{ Child ui.Widget }

type renderKeyDebug struct {
	ui.SingleChildRenderObject
	last  string
	chars []ui.Character
	seq   int
}

func UiKeyDebug(child ui.Widget) ui.Widget {
	return uiKeyDebug{Child: child}
}

func (w uiKeyDebug) WidgetChild() ui.Widget {
	return w.Child
}

func (w uiKeyDebug) CreateRenderObject(ctx ui.BuildContext) ui.RenderObject {
	return &renderKeyDebug{}
}

func (w uiKeyDebug) UpdateRenderObject(ctx ui.BuildContext, ro ui.RenderObject) {}

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

func (r *renderKeyDebug) Paint(p *ui.Painter, off ui.Offset) {
	if child := r.Child(); child != nil {
		child.Paint(p, off)
	}
	if r.last == "" {
		return
	}
	style := ui.Style{Attribute: ui.AttrReverse}
	width := 0
	for _, ch := range r.chars {
		width += ch.Width
	}
	width += 2
	size := r.Size()
	x := off.X + max(0, size.Width-width)
	y := off.Y
	blank := ui.Cell{Character: ui.Character{Grapheme: " ", Width: 1}, Style: style}
	for i := 0; i < width; i++ {
		p.DrawCell(ui.Point{X: x + i, Y: y}, blank)
	}
	cx := x + 1
	for _, ch := range r.chars {
		p.DrawCell(ui.Point{X: cx, Y: y}, ui.Cell{Character: ch, Style: style})
		cx += ch.Width
	}
}

func (r *renderKeyDebug) HandleEvent(ctx ui.EventContext, ev ui.Event) ui.EventResult {
	key, ok := ev.(ui.Key)
	if !ok || ctx.Phase() != ui.CapturePhase {
		return ui.EventIgnored
	}
	r.seq++
	seq := r.seq
	r.last = fmt.Sprintf("key %q code=%d text=%q mods=%d", key.String(), key.Keycode, key.Text, key.Modifiers)
	r.chars = vaxis.Characters(r.last)
	r.MarkNeedsPaint()
	runtime := ctx.Runtime()
	time.AfterFunc(3*time.Second, func() {
		runtime.Dispatch(func() {
			if r.seq == seq {
				r.last = ""
				r.chars = nil
				r.MarkNeedsPaint()
			}
		})
	})
	return ui.EventIgnored
}

func UiRequestFrame(ctx ui.EventContext) {
	ctx.Runtime().Dispatch(func() {})
}

func UiDispatchAfter(ctx ui.EventContext, delayMs int, callback func()) {
	runtime := ctx.Runtime()
	time.AfterFunc(time.Duration(delayMs)*time.Millisecond, func() {
		runtime.Dispatch(callback)
	})
}

func UiQuit(ctx ui.EventContext) {
	ctx.Quit()
}

type UiStateContext struct{ state *ardState }

type ardStateful struct {
	build func(*UiStateContext) ui.Widget
}

type ardState struct {
	ui.StateBase
	values map[string]any
	build  func(*UiStateContext) ui.Widget
}

func UiStateful(build func(*UiStateContext) ui.Widget) ui.Widget {
	return ardStateful{build: build}
}

func (w ardStateful) CreateState() ui.State {
	return &ardState{values: map[string]any{}, build: w.build}
}

func (s *ardState) DidUpdateWidget(old ui.Widget) {
	s.build = s.Widget().(ardStateful).build
}

func (s *ardState) Build(ctx ui.BuildContext) ui.Widget {
	return s.build(&UiStateContext{state: s})
}

func UiStateString(ctx *UiStateContext, key string) string {
	if ctx == nil || ctx.state == nil {
		return ""
	}
	if value, ok := ctx.state.values[key].(string); ok {
		return value
	}
	return ""
}

func UiSetStateString(ctx *UiStateContext, key string, value string) {
	if ctx == nil || ctx.state == nil {
		return
	}
	ctx.state.SetState(func() { ctx.state.values[key] = value })
}

func UiStateBool(ctx *UiStateContext, key string) bool {
	if ctx == nil || ctx.state == nil {
		return false
	}
	if value, ok := ctx.state.values[key].(bool); ok {
		return value
	}
	return false
}

func UiSetStateBool(ctx *UiStateContext, key string, value bool) {
	if ctx == nil || ctx.state == nil {
		return
	}
	ctx.state.SetState(func() { ctx.state.values[key] = value })
}

func UiStateInt(ctx *UiStateContext, key string) int {
	if ctx == nil || ctx.state == nil {
		return 0
	}
	if value, ok := ctx.state.values[key].(int); ok {
		return value
	}
	return 0
}

func UiSetStateInt(ctx *UiStateContext, key string, value int) {
	if ctx == nil || ctx.state == nil {
		return
	}
	ctx.state.SetState(func() { ctx.state.values[key] = value })
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

func UiOverlayModal(child ui.Widget, modal ui.Widget) ui.Widget {
	return ui.Overlay{
		Child: child,
		Entries: []ui.OverlayEntry{{Modal: true, Child: modal}},
	}
}

type UiBoxBorder struct {
	Top   bool
	Right bool
	Bottom bool
	Left  bool
	Style ardruntime.Maybe[UiStyle]
}

func UiBorderNew(top bool, right bool, bottom bool, left bool, style ardruntime.Maybe[UiStyle]) UiBoxBorder {
	return UiBoxBorder{Top: top, Right: right, Bottom: bottom, Left: left, Style: style}
}

type uiBox struct {
	Child  ui.Widget
	Border ardruntime.Maybe[UiBoxBorder]
	Style  ardruntime.Maybe[UiStyle]
}

func (w uiBox) CreateState() ui.State { return &uiBoxState{} }

type uiBoxState struct{ ui.StateBase }

func (s *uiBoxState) Build(ctx ui.BuildContext) ui.Widget {
	w := s.Widget().(uiBox)
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
	return ui.DecoratedBox(decoration, w.Child)
}

func UiBox(child ui.Widget, border ardruntime.Maybe[UiBoxBorder], style ardruntime.Maybe[UiStyle]) ui.Widget {
	return uiBox{Child: child, Border: border, Style: style}
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

func UiButton(label string, onPressed func(ui.EventContext)) ui.Widget {
	return ui.Button{Label: label, OnPressed: func(ctx ui.EventContext) {
		if onPressed != nil {
			onPressed(ctx)
		}
	}}
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

func New(title string) (*vaxis.Vaxis, error) {
	vx, err := vaxis.New(vaxis.Options{DisableKittyKeyboard: true})
	if err != nil {
		return nil, err
	}
	if title != "" {
		vx.SetTitle(title)
	}
	vx.HideCursor()
	drainStartupEvents(vx)
	return vx, nil
}

func Close(term *vaxis.Vaxis) error {
	if term == nil {
		return nil
	}
	term.Close()
	return nil
}

func Refresh(term *vaxis.Vaxis) {
	if term != nil {
		term.Refresh()
	}
}

func Bell(term *vaxis.Vaxis) {
	if term != nil {
		term.Bell()
	}
}

func SetTitle(term *vaxis.Vaxis, title string) {
	if term != nil {
		term.SetTitle(title)
	}
}

func HideCursor(term *vaxis.Vaxis) {
	if term != nil {
		term.HideCursor()
	}
}

func ShowCursor(term *vaxis.Vaxis, x int, y int) {
	if term != nil {
		term.ShowCursor(x, y, vaxis.CursorBlock)
	}
}

func Suspend(term *vaxis.Vaxis) error {
	if term == nil {
		return nil
	}
	return term.Suspend()
}

func Resume(term *vaxis.Vaxis) error {
	if term == nil {
		return nil
	}
	return term.Resume()
}

func TerminalID(term *vaxis.Vaxis) string {
	if term == nil {
		return ""
	}
	return term.TerminalID()
}

func RenderedWidth(term *vaxis.Vaxis, text string) int {
	if term == nil {
		return len([]rune(text))
	}
	return term.RenderedWidth(text)
}

func CanRGB(term *vaxis.Vaxis) bool {
	return term != nil && term.CanRGB()
}

func CanSixel(term *vaxis.Vaxis) bool {
	return term != nil && term.CanSixel()
}

func CanKittyGraphics(term *vaxis.Vaxis) bool {
	return term != nil && term.CanKittyGraphics()
}

func CanUnicodeCore(term *vaxis.Vaxis) bool {
	return term != nil && term.CanUnicodeCore()
}

func CanDisplayGraphics(term *vaxis.Vaxis) bool {
	return term != nil && term.CanDisplayGraphics()
}

func CanReportColor(term *vaxis.Vaxis) bool {
	return term != nil && term.CanReportColor()
}

func CanReportForegroundColor(term *vaxis.Vaxis) bool {
	return term != nil && term.CanReportForegroundColor()
}

func CanReportBackgroundColor(term *vaxis.Vaxis) bool {
	return term != nil && term.CanReportBackgroundColor()
}

func CanExplicitWidth(term *vaxis.Vaxis) bool {
	return term != nil && term.CanExplicitWidth()
}

func CanInBandResize(term *vaxis.Vaxis) bool {
	return term != nil && term.CanInBandResize()
}

func CanSetAppID(term *vaxis.Vaxis) bool {
	return term != nil && term.CanSetAppID()
}

func NotifyWorkingDirectory(term *vaxis.Vaxis, path string) {
	if term != nil {
		term.NotifyWorkingDirectory(path)
	}
}

func SetAppID(term *vaxis.Vaxis, id string) {
	if term != nil {
		term.SetAppID(id)
	}
}

// SetMouseShape sets the terminal mouse cursor shape. shape is the raw CSS
// cursor name used by vaxis: "default", "text", "pointer", "help", "wait",
// "progress", "ew-resize", "ns-resize", "cell".
func SetMouseShape(term *vaxis.Vaxis, shape string) {
	if term != nil {
		term.SetMouseShape(vaxis.MouseShape(shape))
	}
}

func Notify(term *vaxis.Vaxis, title string, body string) {
	if term != nil {
		term.Notify(title, body)
	}
}

func ClipboardPush(term *vaxis.Vaxis, text string) {
	if term != nil {
		term.ClipboardPush(text)
	}
}

// WindowDrawLink renders text as an OSC 8 clickable hyperlink. Terminals that
// support OSC 8 (iTerm2, Kitty, WezTerm, modern macOS Terminal, ...)
// make the text clickable. The text is also underlined so users can
// see it's a link.
func WindowDrawLink(win vaxis.Window, x int, y int, text string, url string) {
	width, height := win.Size()
	if x < 0 || y < 0 || x >= width || y >= height {
		return
	}
	style := vaxis.Style{
		Hyperlink:      url,
		UnderlineStyle: vaxis.UnderlineSingle,
	}
	win.New(x, y, width-x, 1).Print(vaxis.Segment{Text: text, Style: style})
}

func Render(term *vaxis.Vaxis) {
	if term == nil {
		return
	}
	term.Render()
}

func Root(term *vaxis.Vaxis) vaxis.Window {
	if term == nil {
		return vaxis.Window{}
	}
	return term.Window()
}

func RootWindow(term *vaxis.Vaxis, x int, y int, width int, height int) vaxis.Window {
	return Subwindow(Root(term), x, y, width, height)
}

func Subwindow(win vaxis.Window, x int, y int, width int, height int) vaxis.Window {
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}
	return win.New(x, y, width, height)
}

func WindowWidth(win vaxis.Window) int {
	width, _ := win.Size()
	return width
}

func WindowHeight(win vaxis.Window) int {
	_, height := win.Size()
	return height
}

func WindowOriginCol(win vaxis.Window) int {
	col, _ := win.Origin()
	return col
}

func WindowOriginRow(win vaxis.Window) int {
	_, row := win.Origin()
	return row
}

func WindowClear(win vaxis.Window) {
	if win.Vx == nil {
		writeTestFill(win, testCell{Grapheme: " "})
		return
	}
	win.Clear()
}

func WindowDrawText(win vaxis.Window, x int, y int, text string) {
	width, height := win.Size()
	if x < 0 || y < 0 || x >= width || y >= height {
		return
	}
	if win.Vx == nil {
		writeTestText(win, x, y, text, testCell{})
		return
	}
	win.New(x, y, width-x, 1).Print(vaxis.Segment{Text: text})
}

func WindowDrawTextStyle(win vaxis.Window, x int, y int, text string, fg int, bg int, bold bool, dim bool, italic bool, underline bool, reverse bool) {
	width, height := win.Size()
	if x < 0 || y < 0 || x >= width || y >= height {
		return
	}
	if win.Vx == nil {
		writeTestText(win, x, y, text, testCell{Fg: fg, Bg: bg, Bold: bold, Dim: dim, Italic: italic, Underline: underline, Reverse: reverse})
		return
	}
	win.New(x, y, width-x, 1).Print(vaxis.Segment{Text: text, Style: makeStyle(fg, bg, bold, dim, italic, underline, reverse)})
}

func WindowFill(win vaxis.Window, text string, fg int, bg int, bold bool, dim bool, italic bool, underline bool, reverse bool) {
	ch := " "
	if text != "" {
		ch = firstCharacter(text)
	}
	if win.Vx == nil {
		writeTestFill(win, testCell{Grapheme: ch, Fg: fg, Bg: bg, Bold: bold, Dim: dim, Italic: italic, Underline: underline, Reverse: reverse})
		return
	}
	win.Fill(vaxis.Cell{Character: vaxis.Character{Grapheme: ch, Width: 1}, Style: makeStyle(fg, bg, bold, dim, italic, underline, reverse)})
}

func WindowShowCursor(win vaxis.Window, x int, y int) {
	if win.Vx == nil {
		return
	}
	win.ShowCursor(x, y, vaxis.CursorBlock)
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

func firstCharacter(text string) string {
	for _, r := range text {
		if r == 0 || !unicode.IsPrint(r) {
			break
		}
		return string(r)
	}
	return " "
}

// Event accessors take a vaxis.Event (interface{}) and type-assert to
// the concrete variant. No wrapper struct is needed — we pass the
// vaxis event itself across the FFI boundary.

// keyName returns the plain key name without modifier prefixes.
func keyName(k vaxis.Key) string {
	k.Modifiers = 0
	name := k.String()
	for {
		trimmed := name
		for _, prefix := range []string{"Meta+", "Hyper+", "Super+", "Ctrl+", "Alt+", "Shift+"} {
			trimmed = strings.TrimPrefix(trimmed, prefix)
		}
		if trimmed == name {
			return name
		}
		name = trimmed
	}
}

func mouseButtonName(b vaxis.MouseButton) string {
	switch b {
	case vaxis.MouseLeftButton:
		return "left"
	case vaxis.MouseMiddleButton:
		return "middle"
	case vaxis.MouseRightButton:
		return "right"
	case vaxis.MouseNoButton:
		return "none"
	case vaxis.MouseWheelUp:
		return "wheel_up"
	case vaxis.MouseWheelDown:
		return "wheel_down"
	case vaxis.MouseWheelLeft:
		return "wheel_left"
	case vaxis.MouseWheelRight:
		return "wheel_right"
	default:
		return "unknown"
	}
}

func mouseKindName(t vaxis.EventType) string {
	switch t {
	case vaxis.EventPress:
		return "press"
	case vaxis.EventRelease:
		return "release"
	case vaxis.EventMotion:
		return "motion"
	default:
		return "unknown"
	}
}

func mouseButtonFromName(name string) vaxis.MouseButton {
	switch name {
	case "left":
		return vaxis.MouseLeftButton
	case "middle":
		return vaxis.MouseMiddleButton
	case "right":
		return vaxis.MouseRightButton
	case "wheel_up":
		return vaxis.MouseWheelUp
	case "wheel_down":
		return vaxis.MouseWheelDown
	case "wheel_left":
		return vaxis.MouseWheelLeft
	case "wheel_right":
		return vaxis.MouseWheelRight
	default:
		return vaxis.MouseNoButton
	}
}

func mouseKindFromName(name string) vaxis.EventType {
	switch name {
	case "release":
		return vaxis.EventRelease
	case "motion":
		return vaxis.EventMotion
	default:
		return vaxis.EventPress
	}
}

func modifiers(ctrl, shift, alt bool) vaxis.ModifierMask {
	var mods vaxis.ModifierMask
	if ctrl {
		mods |= vaxis.ModCtrl
	}
	if shift {
		mods |= vaxis.ModShift
	}
	if alt {
		mods |= vaxis.ModAlt
	}
	return mods
}

type PasteEvent struct {
	Content string
}

type CustomEvent struct {
	Name    string
	Payload ardruntime.Maybe[any]
}

func IsPasteEvent(e vaxis.Event) bool {
	_, ok := e.(PasteEvent)
	return ok
}

func PostKeyEvent(vx *vaxis.Vaxis, ev vaxis.Event) {
	vx.PostEvent(ev)
}

func PostMouseEvent(vx *vaxis.Vaxis, col, row int, button, kind string, ctrl, shift, alt bool) {
	vx.PostEvent(vaxis.Mouse{
		Col:       col,
		Row:       row,
		Button:    mouseButtonFromName(button),
		EventType: mouseKindFromName(kind),
		Modifiers: modifiers(ctrl, shift, alt),
	})
}

func PostResizeEvent(vx *vaxis.Vaxis, cols, rows, xPixel, yPixel int) {
	vx.PostEvent(vaxis.Resize{Cols: cols, Rows: rows, XPixel: xPixel, YPixel: yPixel})
}

func PostFocusEvent(vx *vaxis.Vaxis, focused bool) {
	if focused {
		vx.PostEvent(vaxis.FocusIn{})
	} else {
		vx.PostEvent(vaxis.FocusOut{})
	}
}

func PostPasteEvent(vx *vaxis.Vaxis, content string) {
	vx.PostEvent(PasteEvent{Content: content})
}

func PostRedrawEvent(vx *vaxis.Vaxis) {
	vx.PostEvent(vaxis.Redraw{})
}

func PostCustomEvent(vx *vaxis.Vaxis, name string, payload ardruntime.Maybe[any]) {
	vx.PostEvent(CustomEvent{Name: name, Payload: payload})
}

func PostColorThemeEvent(vx *vaxis.Vaxis, dark bool) {
	mode := vaxis.LightMode
	if dark {
		mode = vaxis.DarkMode
	}
	vx.PostEvent(vaxis.ColorThemeUpdate{Mode: mode})
}

func PostQuitEvent(vx *vaxis.Vaxis) {
	vx.PostEvent(vaxis.QuitEvent{})
}

// ReadEvent blocks until the next event from the terminal and returns
// it. Non-press key events are filtered out at this layer. Bracketed paste
// boundaries are collapsed into a single PasteEvent carrying the pasted text.
func ReadEvent(term *vaxis.Vaxis) vaxis.Event {
	for ev := range term.Events() {
		if _, ok := ev.(vaxis.PasteStartEvent); ok {
			var content strings.Builder
			for pasteEv := range term.Events() {
				switch pasteEv := pasteEv.(type) {
				case vaxis.PasteEndEvent:
					return PasteEvent{Content: content.String()}
				case vaxis.Key:
					if pasteEv.EventType == vaxis.EventPaste {
						content.WriteString(pasteEv.Text)
					}
				}
			}
			return PasteEvent{Content: content.String()}
		}

		if k, ok := ev.(vaxis.Key); ok && k.EventType != vaxis.EventPress {
			continue
		}
		// vaxis needs to be told the new size before subsequent draws
		// pick it up; otherwise it keeps drawing at the old dimensions.
		if r, ok := ev.(vaxis.Resize); ok {
			term.Resize(r)
		}
		return ev
	}
	return vaxis.QuitEvent{}
}

func IsKeyEvent(e vaxis.Event) bool {
	_, ok := e.(vaxis.Key)
	return ok
}

func IsMouseEvent(e vaxis.Event) bool {
	_, ok := e.(vaxis.Mouse)
	return ok
}

func IsFocusEvent(e vaxis.Event) bool {
	switch e.(type) {
	case vaxis.FocusIn, vaxis.FocusOut:
		return true
	default:
		return false
	}
}

func IsRedrawEvent(e vaxis.Event) bool {
	_, ok := e.(vaxis.Redraw)
	return ok
}

func IsColorThemeEvent(e vaxis.Event) bool {
	_, ok := e.(vaxis.ColorThemeUpdate)
	return ok
}

func IsQuitEvent(e vaxis.Event) bool {
	_, ok := e.(vaxis.QuitEvent)
	return ok
}

func EventKeyName(e vaxis.Event) string {
	if k, ok := e.(vaxis.Key); ok {
		return keyName(k)
	}
	return ""
}
func EventKeyCtrl(e vaxis.Event) bool {
	if k, ok := e.(vaxis.Key); ok {
		return k.Modifiers&vaxis.ModCtrl != 0
	}
	return false
}
func EventKeyShift(e vaxis.Event) bool {
	if k, ok := e.(vaxis.Key); ok {
		return k.Modifiers&vaxis.ModShift != 0
	}
	return false
}
func EventKeyAlt(e vaxis.Event) bool {
	if k, ok := e.(vaxis.Key); ok {
		return k.Modifiers&vaxis.ModAlt != 0
	}
	return false
}
func EventKeyMeta(e vaxis.Event) bool {
	if k, ok := e.(vaxis.Key); ok {
		return k.Modifiers&vaxis.ModMeta != 0
	}
	return false
}

// EventKeyText returns the rendered text the key would insert into a
// text buffer, accounting for shift, keyboard layout, and IME. Empty
// for non-textual keys (Enter, arrows, function keys, control combos
// like Ctrl+C, etc.). Use this when implementing text inputs so you
// don't have to manually shift/layout characters.
func EventKeyText(e vaxis.Event) string {
	if k, ok := e.(vaxis.Key); ok {
		return k.Text
	}
	return ""
}

func EventKeyPretty(e vaxis.Event) string {
	if k, ok := e.(vaxis.Key); ok {
		return k.String()
	}
	return ""
}

func EventMouseCol(e vaxis.Event) int {
	if m, ok := e.(vaxis.Mouse); ok {
		return m.Col
	}
	return 0
}
func EventMouseRow(e vaxis.Event) int {
	if m, ok := e.(vaxis.Mouse); ok {
		return m.Row
	}
	return 0
}
func EventMouseButton(e vaxis.Event) string {
	if m, ok := e.(vaxis.Mouse); ok {
		return mouseButtonName(m.Button)
	}
	return ""
}
func EventMouseKind(e vaxis.Event) string {
	if m, ok := e.(vaxis.Mouse); ok {
		return mouseKindName(m.EventType)
	}
	return ""
}
func EventMouseCtrl(e vaxis.Event) bool {
	if m, ok := e.(vaxis.Mouse); ok {
		return m.Modifiers&vaxis.ModCtrl != 0
	}
	return false
}
func EventMouseShift(e vaxis.Event) bool {
	if m, ok := e.(vaxis.Mouse); ok {
		return m.Modifiers&vaxis.ModShift != 0
	}
	return false
}
func EventMouseAlt(e vaxis.Event) bool {
	if m, ok := e.(vaxis.Mouse); ok {
		return m.Modifiers&vaxis.ModAlt != 0
	}
	return false
}

func EventResizeCols(e vaxis.Event) int {
	if r, ok := e.(vaxis.Resize); ok {
		return r.Cols
	}
	return 0
}
func EventResizeRows(e vaxis.Event) int {
	if r, ok := e.(vaxis.Resize); ok {
		return r.Rows
	}
	return 0
}

func EventResizeXPixel(e vaxis.Event) int {
	if r, ok := e.(vaxis.Resize); ok {
		return r.XPixel
	}
	return 0
}

func EventResizeYPixel(e vaxis.Event) int {
	if r, ok := e.(vaxis.Resize); ok {
		return r.YPixel
	}
	return 0
}

func IsResizeEvent(e vaxis.Event) bool {
	_, ok := e.(vaxis.Resize)
	return ok
}

func EventColorThemeDark(e vaxis.Event) bool {
	if u, ok := e.(vaxis.ColorThemeUpdate); ok {
		return u.Mode == vaxis.DarkMode
	}
	return false
}

func EventFocusFocused(e vaxis.Event) bool {
	_, isIn := e.(vaxis.FocusIn)
	return isIn
}
func EventPasteContent(e vaxis.Event) string {
	if p, ok := e.(PasteEvent); ok {
		return p.Content
	}
	return ""
}

func EventCustomName(e vaxis.Event) string {
	if c, ok := e.(CustomEvent); ok {
		return c.Name
	}
	return fmt.Sprintf("%T", e)
}

func EventCustomPayload(e vaxis.Event) ardruntime.Maybe[any] {
	if c, ok := e.(CustomEvent); ok {
		return c.Payload
	}
	return ardruntime.None[any]()
}

func drainStartupEvents(vx *vaxis.Vaxis) {
	quiet := time.NewTimer(100 * time.Millisecond)
	defer quiet.Stop()
	for {
		select {
		case <-vx.Events():
			if !quiet.Stop() {
				<-quiet.C
			}
			quiet.Reset(100 * time.Millisecond)
		case <-quiet.C:
			return
		}
	}
}
