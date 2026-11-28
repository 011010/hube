package note

import (
	"encoding/json"
	"strings"
	"unicode"
	"unicode/utf8"
)

// MarkdownToBlocks converts article markdown into a TipTap document JSON
// string ({"type":"doc","content":[...]}). Inline syntax becomes marks;
// images stay literal. Empty input yields a single paragraph.
//
// This is a purpose-built converter for Defuddle's markdown output, not a
// CommonMark implementation: one heading style, tight lists, one-level
// blockquotes, fenced code, and the inline marks TipTap renders. Anything
// unmatched stays literal text.
func MarkdownToBlocks(md string) (string, error) {
	blocks := parseBlocks(md)
	if len(blocks) == 0 {
		// Empty or whitespace-only input still yields one paragraph holding
		// the raw text, so a stored doc is never contentless.
		blocks = []docNode{{Type: "paragraph", Content: []docNode{{Type: "text", Text: &md}}}}
	}
	encoded, err := json.Marshal(docNode{Type: "doc", Content: blocks})
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

// markNode is a TipTap mark (bold, italic, code, strike, link).
type markNode struct {
	Type  string         `json:"type"`
	Attrs map[string]any `json:"attrs,omitempty"`
}

// docNode is a TipTap node: text nodes set Text and Marks, containers set
// Content, headings and code blocks set Attrs. Text is a pointer so an
// empty text node still marshals its "text" key.
type docNode struct {
	Type    string         `json:"type"`
	Attrs   map[string]any `json:"attrs,omitempty"`
	Content []docNode      `json:"content,omitempty"`
	Text    *string        `json:"text,omitempty"`
	Marks   []markNode     `json:"marks,omitempty"`
}

// parseBlocks splits normalized markdown into block nodes; blank lines
// separate blocks and consecutive plain lines merge into one paragraph.
func parseBlocks(md string) []docNode {
	lines := strings.Split(strings.ReplaceAll(md, "\r\n", "\n"), "\n")
	var blocks []docNode
	i := 0
	for i < len(lines) {
		line := lines[i]
		trimmed := strings.TrimSpace(line)
		level := headingLevel(trimmed)
		switch {
		case trimmed == "":
			i++
		case isFence(line):
			lang := strings.TrimSpace(trimmed[3:])
			i++
			var body []string
			for i < len(lines) && !isFence(lines[i]) {
				body = append(body, lines[i])
				i++
			}
			if i < len(lines) {
				i++ // step past the closing fence
			}
			code := strings.Join(body, "\n")
			blocks = append(blocks, docNode{
				Type:    "codeBlock",
				Attrs:   map[string]any{"language": lang},
				Content: []docNode{{Type: "text", Text: &code}},
			})
		case level > 0:
			blocks = append(blocks, docNode{
				Type:    "heading",
				Attrs:   map[string]any{"level": level},
				Content: parseInline(strings.TrimSpace(trimmed[level:]), nil),
			})
			i++
		case trimmed == "---" || trimmed == "***":
			blocks = append(blocks, docNode{Type: "horizontalRule"})
			i++
		case strings.HasPrefix(trimmed, ">"):
			var inner []string
			for i < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[i]), ">") {
				body := strings.TrimPrefix(strings.TrimSpace(lines[i]), ">")
				inner = append(inner, strings.TrimPrefix(body, " "))
				i++
			}
			blocks = append(blocks, docNode{Type: "blockquote", Content: quoteParagraphs(inner)})
		case isItemLine(line):
			var items []itemLine
			for i < len(lines) {
				it, ok := parseItemLine(lines[i])
				if !ok {
					break
				}
				items = append(items, it)
				i++
			}
			blocks = append(blocks, buildLists(items)...)
		default:
			var para []string
			for i < len(lines) && isPlainLine(lines[i]) {
				para = append(para, strings.TrimSpace(lines[i]))
				i++
			}
			content := parseInline(strings.Join(para, " "), nil)
			if len(content) > 0 {
				blocks = append(blocks, docNode{Type: "paragraph", Content: content})
			}
		}
	}
	return blocks
}

// isFence reports whether a line opens or closes a fenced code block.
func isFence(line string) bool { return strings.HasPrefix(strings.TrimSpace(line), "```") }

// headingLevel reports the level of a "#"-prefixed line, or 0 when the
// line is not a heading. More than six hashes stays plain text.
func headingLevel(trimmed string) int {
	n := 0
	for n < len(trimmed) && trimmed[n] == '#' {
		n++
	}
	if n >= 1 && n <= 6 && (n == len(trimmed) || trimmed[n] == ' ') {
		return n
	}
	return 0
}

// itemLine is one raw list-item line with its indent and marker kind.
type itemLine struct {
	indent  int
	ordered bool
	text    string
}

func isItemLine(line string) bool {
	_, ok := parseItemLine(line)
	return ok
}

// parseItemLine recognizes "- text", "* text", and "N. text" with any
// leading indent of spaces and tabs: a tab is one step (2) and a space
// half a step (1), so Defuddle's tab-indented continuations nest the
// same way a two-space indent would.
func parseItemLine(line string) (itemLine, bool) {
	indent, i := 0, 0
	for i < len(line) && (line[i] == ' ' || line[i] == '\t') {
		if line[i] == '\t' {
			indent += 2
		} else {
			indent++
		}
		i++
	}
	rest := line[i:]
	if strings.HasPrefix(rest, "- ") || strings.HasPrefix(rest, "* ") {
		return itemLine{indent: indent, text: strings.TrimSpace(rest[2:])}, true
	}
	n := 0
	for n < len(rest) && rest[n] >= '0' && rest[n] <= '9' {
		n++
	}
	if n > 0 && n+1 < len(rest) && rest[n] == '.' && rest[n+1] == ' ' {
		return itemLine{indent: indent, ordered: true, text: strings.TrimSpace(rest[n+2:])}, true
	}
	return itemLine{}, false
}

// buildLists turns a run of item lines into list nodes, splitting runs by
// root indent and marker type.
func buildLists(items []itemLine) []docNode {
	var out []docNode
	for len(items) > 0 {
		node, consumed := buildOneList(items)
		out = append(out, node)
		items = items[consumed:]
	}
	return out
}

// buildOneList builds a single list from items sharing one indent and
// marker type, nesting deeper lines under the closest preceding item. It
// always consumes at least one item, so buildLists always makes progress.
func buildOneList(items []itemLine) (docNode, int) {
	base, ordered := items[0].indent, items[0].ordered
	node := docNode{Type: "bulletList"}
	if ordered {
		node.Type = "orderedList"
	}
	i := 0
	for i < len(items) && items[i].indent == base && items[i].ordered == ordered {
		item := docNode{Type: "listItem", Content: []docNode{{
			Type:    "paragraph",
			Content: parseInline(items[i].text, nil),
		}}}
		i++
		j := i
		for j < len(items) && items[j].indent > base {
			j++
		}
		if j > i {
			item.Content = append(item.Content, buildLists(items[i:j])...)
			i = j
		}
		node.Content = append(node.Content, item)
	}
	return node, i
}

// quoteParagraphs groups blockquote inner lines into paragraphs: blank
// inner lines separate paragraphs, consecutive lines merge with spaces.
func quoteParagraphs(inner []string) []docNode {
	var paras []docNode
	var cur []string
	flush := func() {
		if len(cur) == 0 {
			return
		}
		content := parseInline(strings.Join(cur, " "), nil)
		if len(content) > 0 {
			paras = append(paras, docNode{Type: "paragraph", Content: content})
		}
		cur = nil
	}
	for _, line := range inner {
		if strings.TrimSpace(line) == "" {
			flush()
			continue
		}
		cur = append(cur, strings.TrimSpace(line))
	}
	flush()
	return paras
}

// isPlainLine reports whether a line continues the current paragraph: it
// has visible content and starts no other block kind.
func isPlainLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || isFence(line) || headingLevel(trimmed) > 0 ||
		trimmed == "---" || trimmed == "***" || strings.HasPrefix(trimmed, ">") {
		return false
	}
	_, ok := parseItemLine(line)
	return !ok
}

// parseInline converts markdown text into TipTap text nodes. marks holds
// the marks opened by enclosing spans; every emitted node carries them in
// open order. Literal characters accumulate in a buffer and flush as runs;
// adjacent runs with identical mark sets merge at the end.
func parseInline(s string, marks []markNode) []docNode {
	var nodes []docNode
	var buf strings.Builder
	flush := func() {
		if buf.Len() == 0 {
			return
		}
		t := buf.String()
		nodes = append(nodes, docNode{Type: "text", Text: &t, Marks: marks})
		buf.Reset()
	}
	i := 0
	open := func(inner string, end int, ms ...markNode) {
		flush()
		next := marks
		for _, m := range ms {
			next = pushMark(next, m)
		}
		nodes = append(nodes, parseInline(inner, next)...)
		i = end
	}
	for i < len(s) {
		c := s[i]
		switch {
		case c == '\\':
			// CommonMark escapes: a backslash escapes only ASCII
			// punctuation; before anything else it stays literal, so
			// raw LaTeX ($e^{i \pi}$) and Windows paths round-trip.
			if i+1 < len(s) && isEscapablePunct(s[i+1]) {
				buf.WriteByte(s[i+1])
				i += 2
			} else {
				buf.WriteByte('\\')
				i++
			}
		case c == '`':
			if end := strings.IndexByte(s[i+1:], '`'); end > 0 {
				flush()
				code := s[i+1 : i+1+end]
				nodes = append(nodes, docNode{Type: "text", Text: &code, Marks: pushMark(marks, markNode{Type: "code"})})
				i += end + 2
			} else {
				buf.WriteByte('`')
				i++
			}
		case c == '!' && i+1 < len(s) && s[i+1] == '[':
			// Images stay literal text: keep the whole span verbatim.
			if end, ok := scanLinkSpan(s, i+1); ok {
				buf.WriteString(s[i:end])
				i = end
			} else {
				buf.WriteByte('!')
				i++
			}
		case c == '[':
			if inner, href, end, ok := scanLink(s, i); ok {
				open(inner, end, markNode{Type: "link", Attrs: map[string]any{"href": href}})
			} else {
				buf.WriteByte('[')
				i++
			}
		case c == '*' && i+2 < len(s) && s[i+1] == '*' && s[i+2] == '*':
			// Defuddle's Turndown renders <strong><em> — either
			// nesting order — as ***x***. Test the triple before the
			// double so both marks open instead of leaving a stray
			// asterisk that italicizes unrelated text.
			if inner, end, ok := scanSpan(s, i+3, "***"); ok {
				open(inner, end, markNode{Type: "bold"}, markNode{Type: "italic"})
			} else {
				buf.WriteByte('*')
				i++
			}
		case c == '*' && i+1 < len(s) && s[i+1] == '*':
			if inner, end, ok := scanSpan(s, i+2, "**"); ok {
				open(inner, end, markNode{Type: "bold"})
			} else {
				buf.WriteByte('*')
				i++
			}
		case c == '*':
			if inner, end, ok := scanSpan(s, i+1, "*"); ok {
				open(inner, end, markNode{Type: "italic"})
			} else {
				buf.WriteByte('*')
				i++
			}
		case c == '_' && i+2 < len(s) && s[i+1] == '_' && s[i+2] == '_':
			// ___x___ mirrors ***x***: bold and italic together.
			if inner, end, ok := scanSpan(s, i+3, "___"); ok {
				open(inner, end, markNode{Type: "bold"}, markNode{Type: "italic"})
			} else {
				buf.WriteByte('_')
				i++
			}
		case c == '_' && i+1 < len(s) && s[i+1] == '_':
			if inner, end, ok := scanSpan(s, i+2, "__"); ok {
				open(inner, end, markNode{Type: "bold"})
			} else {
				buf.WriteByte('_')
				i++
			}
		case c == '_' && !precededByAlnum(s, i):
			// Single underscore opens italic only when not intra-word.
			if inner, end, ok := scanSpan(s, i+1, "_"); ok {
				open(inner, end, markNode{Type: "italic"})
			} else {
				buf.WriteByte('_')
				i++
			}
		case c == '~' && i+1 < len(s) && s[i+1] == '~':
			if inner, end, ok := scanSpan(s, i+2, "~~"); ok {
				open(inner, end, markNode{Type: "strike"})
			} else {
				buf.WriteByte('~')
				i++
			}
		default:
			buf.WriteByte(c)
			i++
		}
	}
	flush()
	return mergeTextRuns(nodes)
}

// scanLink finds "[text](href)" starting at open. Empty link text is not
// a link; it stays literal.
func scanLink(s string, open int) (inner, href string, end int, ok bool) {
	j := strings.Index(s[open+1:], "](")
	if j <= 0 {
		return "", "", 0, false
	}
	hrefStart := open + 1 + j + 2
	k := strings.IndexByte(s[hrefStart:], ')')
	if k < 0 {
		return "", "", 0, false
	}
	return s[open+1 : open+1+j], s[hrefStart : hrefStart+k], hrefStart + k + 1, true
}

// scanLinkSpan reports the index just past a "[..](..)" span starting at
// open, without extracting its parts.
func scanLinkSpan(s string, open int) (int, bool) {
	j := strings.Index(s[open+1:], "](")
	if j < 0 {
		return 0, false
	}
	hrefStart := open + 1 + j + 2
	k := strings.IndexByte(s[hrefStart:], ')')
	if k < 0 {
		return 0, false
	}
	return hrefStart + k + 1, true
}

// scanSpan finds the closing delimiter for a mark span starting just after
// its opener. Empty spans stay literal.
func scanSpan(s string, from int, delim string) (string, int, bool) {
	k := strings.Index(s[from:], delim)
	if k <= 0 {
		return "", 0, false
	}
	return s[from : from+k], from + k + len(delim), true
}

// isEscapablePunct reports whether c is ASCII punctuation — the only
// characters a backslash escapes, per CommonMark. Turndown emits these
// escapes; before any other character the backslash stays literal.
func isEscapablePunct(c byte) bool {
	return strings.IndexByte("!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~", c) >= 0
}

// precededByAlnum reports whether offset i in s follows a letter or digit
// (decoded as a full rune).
func precededByAlnum(s string, i int) bool {
	if i == 0 {
		return false
	}
	r, _ := utf8.DecodeLastRuneInString(s[:i])
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}

// pushMark appends m to a private copy of marks so sibling spans never
// share a backing array.
func pushMark(marks []markNode, m markNode) []markNode {
	out := make([]markNode, 0, len(marks)+1)
	return append(append(out, marks...), m)
}

// mergeTextRuns joins adjacent text nodes carrying identical mark sets.
func mergeTextRuns(nodes []docNode) []docNode {
	if len(nodes) == 0 {
		return nil
	}
	out := make([]docNode, 0, len(nodes))
	for _, n := range nodes {
		if len(out) > 0 && n.Type == "text" {
			last := &out[len(out)-1]
			if last.Type == "text" && marksEqual(last.Marks, n.Marks) {
				merged := *last.Text + *n.Text
				last.Text = &merged
				continue
			}
		}
		out = append(out, n)
	}
	return out
}

func marksEqual(a, b []markNode) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Type != b[i].Type || !attrsEqual(a[i].Attrs, b[i].Attrs) {
			return false
		}
	}
	return true
}

func attrsEqual(a, b map[string]any) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if w, ok := b[k]; !ok || v != w {
			return false
		}
	}
	return true
}
