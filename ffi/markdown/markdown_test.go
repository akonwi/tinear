package markdown

import "testing"

func parseOne(t *testing.T, src string) []Block {
	t.Helper()
	return Parse(src)
}

func TestHeadingAndParagraph(t *testing.T) {
	blocks := parseOne(t, "# Title\n\nHello **bold** and *italic* and `code`.")
	if len(blocks) != 2 {
		t.Fatalf("want 2 blocks, got %d: %#v", len(blocks), blocks)
	}
	if blocks[0].Kind != KindHeading || blocks[0].Level != 1 || blocks[0].Runs[0].Text != "Title" {
		t.Fatalf("bad heading: %#v", blocks[0])
	}
	p := blocks[1]
	if p.Kind != KindParagraph {
		t.Fatalf("bad paragraph kind: %#v", p)
	}
	var bold, italic, code bool
	for _, r := range p.Runs {
		if r.Text == "bold" && r.Bold {
			bold = true
		}
		if r.Text == "italic" && r.Italic {
			italic = true
		}
		if r.Text == "code" && r.Code {
			code = true
		}
	}
	if !bold || !italic || !code {
		t.Fatalf("missing inline styles: %#v", p.Runs)
	}
}

func TestFencedCodeWithLang(t *testing.T) {
	blocks := parseOne(t, "```go\nfmt.Println(1)\nreturn\n```")
	if len(blocks) != 1 || blocks[0].Kind != KindCode || blocks[0].Lang != "go" {
		t.Fatalf("bad code block: %#v", blocks)
	}
	if blocks[0].Runs[0].Text != "fmt.Println(1)\nreturn" {
		t.Fatalf("bad code text: %q", blocks[0].Runs[0].Text)
	}
}

func TestMermaidLangSurvives(t *testing.T) {
	blocks := parseOne(t, "```mermaid\ngraph TD; A-->B\n```")
	if blocks[0].Lang != "mermaid" {
		t.Fatalf("mermaid info string lost: %#v", blocks[0])
	}
}

func TestLists(t *testing.T) {
	blocks := parseOne(t, "- one\n- two\n  - nested\n1. first\n2. second")
	if len(blocks) != 5 {
		t.Fatalf("want 5 items, got %d: %#v", len(blocks), blocks)
	}
	if blocks[0].Marker != "•" || blocks[0].Indent != 0 {
		t.Fatalf("bad bullet: %#v", blocks[0])
	}
	if blocks[2].Indent != 1 {
		t.Fatalf("nested item should indent: %#v", blocks[2])
	}
	if blocks[3].Marker != "1." || blocks[4].Marker != "2." {
		t.Fatalf("bad ordered markers: %#v %#v", blocks[3], blocks[4])
	}
}

func TestTaskList(t *testing.T) {
	blocks := parseOne(t, "- [x] done\n- [ ] todo")
	if blocks[0].Marker != "☑" || blocks[1].Marker != "☐" {
		t.Fatalf("bad task markers: %#v", blocks)
	}
}

func TestQuoteAndRule(t *testing.T) {
	blocks := parseOne(t, "> quoted text\n\n---")
	if blocks[0].Kind != KindQuote || blocks[0].Indent != 1 {
		t.Fatalf("bad quote: %#v", blocks[0])
	}
	if blocks[1].Kind != KindRule {
		t.Fatalf("bad rule: %#v", blocks[1])
	}
}

func TestLinksAndStrikethrough(t *testing.T) {
	blocks := parseOne(t, "See [docs](https://x.dev) and ~~gone~~ and https://auto.link")
	var link, strike, auto bool
	for _, r := range blocks[0].Runs {
		if r.Text == "docs" && r.Link == "https://x.dev" {
			link = true
		}
		if r.Text == "gone" && r.Strike {
			strike = true
		}
		if r.Link == "https://auto.link" {
			auto = true
		}
	}
	if !link || !strike || !auto {
		t.Fatalf("missing inline semantics: %#v", blocks[0].Runs)
	}
}

func TestTable(t *testing.T) {
	blocks := parseOne(t, "| a | b |\n|---|---|\n| 1 | 2 |")
	if len(blocks) != 2 || blocks[0].Kind != KindTableRow || blocks[0].Level != 1 {
		t.Fatalf("bad table: %#v", blocks)
	}
	if !blocks[0].Runs[0].Bold {
		t.Fatalf("header cells should be bold: %#v", blocks[0].Runs)
	}
	if blocks[1].Level != 0 {
		t.Fatalf("body row should not be header: %#v", blocks[1])
	}
}

func TestSoftBreakBecomesNewline(t *testing.T) {
	blocks := parseOne(t, "line one\nline two")
	joined := ""
	for _, r := range blocks[0].Runs {
		joined += r.Text
	}
	if joined != "line one\nline two" {
		t.Fatalf("soft break lost: %q", joined)
	}
}
