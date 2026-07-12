// Package markdown parses markdown into a flat, semantic block/run model
// that Ard can render with vaxis/ui widgets. Parsing is goldmark
// (CommonMark + GFM); no styling or widgets here — the Ard side maps
// semantics to theme styles.
package markdown

import (
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	extast "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/text"
)

// Run is one inline run of text with semantic flags.
type Run struct {
	Text   string
	Bold   bool
	Italic bool
	Code   bool
	Strike bool
	// Link is the destination URL when this run is a link.
	Link string
}

// Block kinds. Strings so Ard can match on them directly.
const (
	KindParagraph = "paragraph"
	KindHeading   = "heading"
	KindCode      = "code"
	KindQuote     = "quote"
	KindListItem  = "list_item"
	KindRule      = "rule"
	KindTableRow  = "table_row"
)

// Block is one renderable line-group.
type Block struct {
	Kind string
	// Level is the heading level (1-6) for headings, and 1 for a
	// table's header row.
	Level int
	// Lang is the fenced code block's info string ("go", "mermaid", ...).
	Lang string
	// Indent is the list nesting depth (0 for top-level items). Also set
	// for quote depth on quoted blocks.
	Indent int
	// Marker is the list item's rendered marker: "•", "3.", "☐", "☑".
	// Empty for continuation blocks inside an item.
	Marker string
	Runs   []Run
}

// Parse converts markdown source into the flat block model.
func Parse(source string) []Block {
	md := goldmark.New(goldmark.WithExtensions(extension.GFM))
	src := []byte(source)
	doc := md.Parser().Parse(text.NewReader(src))
	w := &walker{src: src}
	w.blocks = make([]Block, 0, 16)
	w.container(doc, 0, 0)
	return w.blocks
}

type walker struct {
	src    []byte
	blocks []Block
}

// container walks a node's block-level children. quote is the current
// blockquote depth; indent the current list nesting depth.
func (w *walker) container(n ast.Node, quote, indent int) {
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		w.block(c, quote, indent)
	}
}

func (w *walker) block(n ast.Node, quote, indent int) {
	switch t := n.(type) {
	case *ast.Heading:
		w.emit(Block{Kind: KindHeading, Level: t.Level, Indent: quote, Runs: w.inlines(n)})
	case *ast.Paragraph, *ast.TextBlock:
		kind := KindParagraph
		if quote > 0 {
			kind = KindQuote
		}
		w.emit(Block{Kind: kind, Indent: quoteOrIndent(quote, indent), Runs: w.inlines(n)})
	case *ast.FencedCodeBlock:
		lang := ""
		if t.Info != nil {
			lang = string(t.Info.Segment.Value(w.src))
		}
		w.emit(Block{Kind: KindCode, Lang: strings.TrimSpace(lang), Indent: quoteOrIndent(quote, indent), Runs: []Run{{Text: w.rawLines(n)}}})
	case *ast.CodeBlock:
		w.emit(Block{Kind: KindCode, Indent: quoteOrIndent(quote, indent), Runs: []Run{{Text: w.rawLines(n)}}})
	case *ast.Blockquote:
		w.container(n, quote+1, indent)
	case *ast.List:
		w.list(t, quote, indent)
	case *ast.ThematicBreak:
		w.emit(Block{Kind: KindRule})
	case *extast.Table:
		w.table(t)
	case *ast.HTMLBlock:
		// Render raw HTML as plain text; Linear rarely emits it.
		w.emit(Block{Kind: KindParagraph, Indent: quoteOrIndent(quote, indent), Runs: []Run{{Text: w.rawLines(n)}}})
	default:
		if n.Type() == ast.TypeBlock {
			w.container(n, quote, indent)
		}
	}
}

func (w *walker) list(l *ast.List, quote, indent int) {
	ordered := l.IsOrdered()
	number := l.Start
	if number == 0 {
		number = 1
	}
	for item := l.FirstChild(); item != nil; item = item.NextSibling() {
		marker := "•"
		if ordered {
			marker = itoa(number) + "."
			number++
		}
		// Task list items carry a TaskCheckBox as the first inline of
		// the item's first paragraph.
		if checked, ok := taskState(item); ok {
			if checked {
				marker = "☑"
			} else {
				marker = "☐"
			}
		}
		first := true
		for c := item.FirstChild(); c != nil; c = c.NextSibling() {
			switch c.(type) {
			case *ast.Paragraph, *ast.TextBlock:
				b := Block{Kind: KindListItem, Indent: indent, Runs: w.inlines(c)}
				if quote > 0 {
					b.Kind = KindQuote
					b.Indent = quote
				}
				if first {
					b.Marker = marker
				}
				w.emit(b)
			case *ast.List:
				w.list(c.(*ast.List), quote, indent+1)
			default:
				w.block(c, quote, indent+1)
			}
			first = false
		}
	}
}

func (w *walker) table(t *extast.Table) {
	header := true
	for row := t.FirstChild(); row != nil; row = row.NextSibling() {
		runs := make([]Run, 0, 8)
		for cell := row.FirstChild(); cell != nil; cell = cell.NextSibling() {
			if len(runs) > 0 {
				runs = append(runs, Run{Text: " │ "})
			}
			cellRuns := w.inlines(cell)
			if header {
				for i := range cellRuns {
					cellRuns[i].Bold = true
				}
			}
			runs = append(runs, cellRuns...)
		}
		level := 0
		if header {
			level = 1
		}
		w.emit(Block{Kind: KindTableRow, Level: level, Runs: runs})
		header = false
	}
}

// inlines flattens a block node's inline children into styled runs.
func (w *walker) inlines(n ast.Node) []Run {
	runs := make([]Run, 0, 4)
	w.inline(n, Run{}, &runs)
	return runs
}

func (w *walker) inline(n ast.Node, style Run, out *[]Run) {
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		switch t := c.(type) {
		case *ast.Text:
			r := style
			r.Text = string(t.Segment.Value(w.src))
			push(out, r)
			if t.HardLineBreak() || t.SoftLineBreak() {
				nl := style
				nl.Text = "\n"
				push(out, nl)
			}
		case *ast.String:
			r := style
			r.Text = string(t.Value)
			push(out, r)
		case *ast.CodeSpan:
			r := style
			r.Code = true
			r.Text = string(codeSpanText(t, w.src))
			push(out, r)
		case *ast.Emphasis:
			s := style
			if t.Level >= 2 {
				s.Bold = true
			} else {
				s.Italic = true
			}
			w.inline(c, s, out)
		case *extast.Strikethrough:
			s := style
			s.Strike = true
			w.inline(c, s, out)
		case *ast.Link:
			s := style
			s.Link = string(t.Destination)
			w.inline(c, s, out)
		case *ast.AutoLink:
			r := style
			url := string(t.URL(w.src))
			r.Text = url
			r.Link = url
			push(out, r)
		case *ast.Image:
			r := style
			r.Link = string(t.Destination)
			r.Text = "🖼 " + string(t.Text(w.src))
			push(out, r)
		case *extast.TaskCheckBox:
			// Consumed by the list walker for the marker; skip here.
		default:
			w.inline(c, style, out)
		}
	}
}

func (w *walker) rawLines(n ast.Node) string {
	var b strings.Builder
	lines := n.Lines()
	for i := 0; i < lines.Len(); i++ {
		seg := lines.At(i)
		b.Write(seg.Value(w.src))
	}
	return strings.TrimRight(b.String(), "\n")
}

func (w *walker) emit(b Block) {
	// Drop blocks that rendered to nothing (empty paragraphs).
	if b.Kind != KindRule && len(b.Runs) == 0 {
		return
	}
	w.blocks = append(w.blocks, b)
}

func push(out *[]Run, r Run) {
	if r.Text == "" {
		return
	}
	*out = append(*out, r)
}

func quoteOrIndent(quote, indent int) int {
	if quote > 0 {
		return quote
	}
	return indent
}

func taskState(item ast.Node) (checked bool, ok bool) {
	first := item.FirstChild()
	if first == nil {
		return false, false
	}
	for c := first.FirstChild(); c != nil; c = c.NextSibling() {
		if cb, isCb := c.(*extast.TaskCheckBox); isCb {
			return cb.IsChecked, true
		}
	}
	return false, false
}

func codeSpanText(n *ast.CodeSpan, src []byte) []byte {
	var b []byte
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		if t, ok := c.(*ast.Text); ok {
			b = append(b, t.Segment.Value(src)...)
		}
	}
	return b
}

func itoa(i int) string {
	if i < 0 {
		return "0"
	}
	if i == 0 {
		return "0"
	}
	var digits []byte
	for i > 0 {
		digits = append([]byte{byte('0' + i%10)}, digits...)
		i /= 10
	}
	return string(digits)
}
