package ffi

import (
	"os/exec"
	"runtime"
	"time"
	"unicode"
	"unicode/utf8"

	"git.sr.ht/~rockorager/vaxis"
)

// OpenURL launches the system's default browser/handler for the given URL.
func OpenURL(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
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

func Clear(term *vaxis.Vaxis) {
	if term == nil {
		return
	}
	term.Window().Clear()
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

func DrawText(term *vaxis.Vaxis, x int, y int, text string) {
	if term == nil {
		return
	}
	WindowDrawText(term.Window(), x, y, text)
}

// DrawLink renders text wrapped in an OSC 8 hyperlink. Terminals that
// support OSC 8 (iTerm2, Kitty, WezTerm, modern macOS Terminal, ...)
// make the text clickable. The text is also underlined so users can
// see it's a link.
func DrawLink(term *vaxis.Vaxis, x int, y int, text string, url string) {
	if term == nil {
		return
	}
	win := term.Window()
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

func DrawTextStyle(term *vaxis.Vaxis, x int, y int, text string, fg int, bg int, bold bool, dim bool, italic bool, underline bool, reverse bool) {
	if term == nil {
		return
	}
	WindowDrawTextStyle(term.Window(), x, y, text, fg, bg, bold, dim, italic, underline, reverse)
}

func Render(term *vaxis.Vaxis) error {
	if term == nil {
		return nil
	}
	term.Render()
	return nil
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

func WindowClear(win vaxis.Window) {
	win.Clear()
}

func WindowDrawText(win vaxis.Window, x int, y int, text string) {
	width, height := win.Size()
	if x < 0 || y < 0 || x >= width || y >= height {
		return
	}
	win.New(x, y, width-x, 1).Print(vaxis.Segment{Text: text})
}

func WindowDrawTextStyle(win vaxis.Window, x int, y int, text string, fg int, bg int, bold bool, dim bool, italic bool, underline bool, reverse bool) {
	width, height := win.Size()
	if x < 0 || y < 0 || x >= width || y >= height {
		return
	}
	win.New(x, y, width-x, 1).Print(vaxis.Segment{Text: text, Style: makeStyle(fg, bg, bold, dim, italic, underline, reverse)})
}

func WindowFill(win vaxis.Window, text string, fg int, bg int, bold bool, dim bool, italic bool, underline bool, reverse bool) {
	ch := " "
	if text != "" {
		ch = firstCharacter(text)
	}
	win.Fill(vaxis.Cell{Character: vaxis.Character{Grapheme: ch, Width: 1}, Style: makeStyle(fg, bg, bold, dim, italic, underline, reverse)})
}

func WindowShowCursor(win vaxis.Window, x int, y int) {
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
	return k.String()
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

// RefreshEvent is the "data is stale, refetch now" signal posted by
// the Ard-side refresh fiber via PostRefreshEvent. Distinct from
// vaxis.Redraw (terminal-level repaint) and from any per-frame
// animation tick; the only thing it carries is intent. vaxis's event
// queue is the bridge because it is the only thread-safe multiplexer
// that can unblock a ReadEvent call.
type RefreshEvent struct{}

// PostRefreshEvent enqueues a RefreshEvent into the vaxis event loop.
// Called from the Ard refresh fiber after each sleep interval.
// PostEvent is non-blocking and drops silently if the queue is full,
// which is the behaviour we want for a periodic refresh signal.
func PostRefreshEvent(term *vaxis.Vaxis) {
	if term == nil {
		return
	}
	term.PostEvent(RefreshEvent{})
}

// ReadEvent blocks until the next event from the terminal and returns
// it. Non-press key events are filtered out at this layer.
func ReadEvent(term *vaxis.Vaxis) (vaxis.Event, error) {
	if term == nil {
		return vaxis.QuitEvent{}, nil
	}
	for ev := range term.Events() {
		if k, ok := ev.(vaxis.Key); ok && k.EventType != vaxis.EventPress {
			continue
		}
		// vaxis needs to be told the new size before subsequent draws
		// pick it up; otherwise it keeps drawing at the old dimensions.
		if r, ok := ev.(vaxis.Resize); ok {
			term.Resize(r)
		}
		return ev, nil
	}
	return vaxis.QuitEvent{}, nil
}

// EventKind returns a string discriminator the Ard side uses to dispatch
// to the right Event variant constructor.
func EventKind(e vaxis.Event) string {
	switch e.(type) {
	case vaxis.Key:
		return "key"
	case vaxis.Mouse:
		return "mouse"
	case vaxis.Resize:
		return "resize"
	case vaxis.Redraw:
		return "redraw"
	case vaxis.FocusIn, vaxis.FocusOut:
		return "focus"
	case vaxis.PasteStartEvent, vaxis.PasteEndEvent:
		return "paste"
	case vaxis.QuitEvent:
		return "quit"
	case RefreshEvent:
		return "refresh"
	default:
		return "unknown"
	}
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

func EventFocusFocused(e vaxis.Event) bool {
	_, isIn := e.(vaxis.FocusIn)
	return isIn
}
func EventPasteStarted(e vaxis.Event) bool {
	_, isStart := e.(vaxis.PasteStartEvent)
	return isStart
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


