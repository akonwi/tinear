package ffi

import (
	"time"
	"unicode"
	"unicode/utf8"

	"git.sr.ht/~rockorager/vaxis"
)

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

// rawEvent holds the most-recently read vaxis event so the Ard side
// can query its fields via the nullary accessors below.
//
// Ideally ReadEvent would return a *RawEvent handle that gets passed
// back into each accessor (optionally with Ard-side struct wrappers
// and methods for clean field access), but blocked by
// https://github.com/akonwi/ard/issues/153 — the compiler doesn't add
// the projectffi import to main.go when emitting code that references
// the extern type. The Ard wrapper in vaxis.ard reads the kind then
// queries each accessor
// before the next ReadEvent call, so the package-level state is safe
// for the current single-threaded TUI loop.
type rawEvent struct {
	kind string

	// Key
	keyName  string
	keyCtrl  bool
	keyShift bool
	keyAlt   bool
	keyMeta  bool

	// Mouse
	mouseCol    int
	mouseRow    int
	mouseButton string
	mouseKind   string
	mouseCtrl   bool
	mouseShift  bool
	mouseAlt    bool

	// Resize
	resizeCols int
	resizeRows int

	// Focus
	focusFocused bool

	// Paste
	pasteStarted bool
}

var lastEvent rawEvent

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

// ReadEvent blocks until the next event and stores its details in
// lastEvent. Returns the event kind as a string.
func ReadEvent(term *vaxis.Vaxis) (string, error) {
	if term == nil {
		lastEvent = rawEvent{kind: "quit"}
		return lastEvent.kind, nil
	}
	for ev := range term.Events() {
		switch ev := ev.(type) {
		case vaxis.Key:
			if ev.EventType != vaxis.EventPress {
				continue
			}
			lastEvent = rawEvent{
				kind:     "key",
				keyName:  keyName(ev),
				keyCtrl:  ev.Modifiers&vaxis.ModCtrl != 0,
				keyShift: ev.Modifiers&vaxis.ModShift != 0,
				keyAlt:   ev.Modifiers&vaxis.ModAlt != 0,
				keyMeta:  ev.Modifiers&vaxis.ModMeta != 0,
			}
			return lastEvent.kind, nil
		case vaxis.Mouse:
			lastEvent = rawEvent{
				kind:        "mouse",
				mouseCol:    ev.Col,
				mouseRow:    ev.Row,
				mouseButton: mouseButtonName(ev.Button),
				mouseKind:   mouseKindName(ev.EventType),
				mouseCtrl:   ev.Modifiers&vaxis.ModCtrl != 0,
				mouseShift:  ev.Modifiers&vaxis.ModShift != 0,
				mouseAlt:    ev.Modifiers&vaxis.ModAlt != 0,
			}
			return lastEvent.kind, nil
		case vaxis.Resize:
			lastEvent = rawEvent{
				kind:       "resize",
				resizeCols: ev.Cols,
				resizeRows: ev.Rows,
			}
			return lastEvent.kind, nil
		case vaxis.Redraw:
			lastEvent = rawEvent{kind: "redraw"}
			return lastEvent.kind, nil
		case vaxis.FocusIn:
			lastEvent = rawEvent{kind: "focus", focusFocused: true}
			return lastEvent.kind, nil
		case vaxis.FocusOut:
			lastEvent = rawEvent{kind: "focus", focusFocused: false}
			return lastEvent.kind, nil
		case vaxis.PasteStartEvent:
			lastEvent = rawEvent{kind: "paste", pasteStarted: true}
			return lastEvent.kind, nil
		case vaxis.PasteEndEvent:
			lastEvent = rawEvent{kind: "paste", pasteStarted: false}
			return lastEvent.kind, nil
		case vaxis.QuitEvent:
			lastEvent = rawEvent{kind: "quit"}
			return lastEvent.kind, nil
		}
	}
	lastEvent = rawEvent{kind: "quit"}
	return lastEvent.kind, nil
}

// Accessors. Each returns a field of the most-recently-read event.
func EventKeyName() string  { return lastEvent.keyName }
func EventKeyCtrl() bool    { return lastEvent.keyCtrl }
func EventKeyShift() bool   { return lastEvent.keyShift }
func EventKeyAlt() bool     { return lastEvent.keyAlt }
func EventKeyMeta() bool    { return lastEvent.keyMeta }

func EventMouseCol() int       { return lastEvent.mouseCol }
func EventMouseRow() int       { return lastEvent.mouseRow }
func EventMouseButton() string { return lastEvent.mouseButton }
func EventMouseKind() string   { return lastEvent.mouseKind }
func EventMouseCtrl() bool     { return lastEvent.mouseCtrl }
func EventMouseShift() bool    { return lastEvent.mouseShift }
func EventMouseAlt() bool      { return lastEvent.mouseAlt }

func EventResizeCols() int { return lastEvent.resizeCols }
func EventResizeRows() int { return lastEvent.resizeRows }

func EventFocusFocused() bool { return lastEvent.focusFocused }
func EventPasteStarted() bool { return lastEvent.pasteStarted }

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


