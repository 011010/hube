package note

import (
	"encoding/json"
	"strings"
	"testing"
)

// The builders below mirror the TipTap node shapes MarkdownToBlocks must
// emit. Expected documents are built from the same structs the converter
// marshals, then compared as exact JSON bytes.

func textNode(s string, ms ...markNode) docNode {
	n := docNode{Type: "text", Text: &s}
	if len(ms) > 0 {
		n.Marks = ms
	}
	return n
}

func para(ns ...docNode) docNode {
	return docNode{Type: "paragraph", Content: ns}
}

func heading(level int, ns ...docNode) docNode {
	return docNode{Type: "heading", Attrs: map[string]any{"level": level}, Content: ns}
}

func bulletList(ns ...docNode) docNode {
	return docNode{Type: "bulletList", Content: ns}
}

func orderedList(ns ...docNode) docNode {
	return docNode{Type: "orderedList", Content: ns}
}

func listItem(ns ...docNode) docNode {
	return docNode{Type: "listItem", Content: ns}
}

func blockquote(ns ...docNode) docNode {
	return docNode{Type: "blockquote", Content: ns}
}

func horizontalRule() docNode {
	return docNode{Type: "horizontalRule"}
}

func codeBlock(lang, body string) docNode {
	return docNode{
		Type:    "codeBlock",
		Attrs:   map[string]any{"language": lang},
		Content: []docNode{textNode(body)},
	}
}

func TestMarkdownToBlocks(t *testing.T) {
	bold := markNode{Type: "bold"}
	italic := markNode{Type: "italic"}
	code := markNode{Type: "code"}
	strike := markNode{Type: "strike"}
	link := func(href string) markNode {
		return markNode{Type: "link", Attrs: map[string]any{"href": href}}
	}

	cases := []struct {
		name string
		in   string
		want []docNode
	}{
		{
			"heading level one",
			"# Title",
			[]docNode{heading(1, textNode("Title"))},
		},
		{
			"heading level three with inline mark",
			"### Deep **thought**",
			[]docNode{heading(3, textNode("Deep "), textNode("thought", bold))},
		},
		{
			"heading level six",
			"###### Six",
			[]docNode{heading(6, textNode("Six"))},
		},
		{
			"seven hashes is a paragraph",
			"####### Seven",
			[]docNode{para(textNode("####### Seven"))},
		},
		{
			"simple paragraph",
			"Hello world.",
			[]docNode{para(textNode("Hello world."))},
		},
		{
			"consecutive lines merge into one paragraph",
			"first line\nsecond line",
			[]docNode{para(textNode("first line second line"))},
		},
		{
			"blank line separates paragraphs",
			"para one\n\npara two",
			[]docNode{para(textNode("para one")), para(textNode("para two"))},
		},
		{
			"bullet list with dashes",
			"- alpha\n- beta",
			[]docNode{bulletList(
				listItem(para(textNode("alpha"))),
				listItem(para(textNode("beta"))),
			)},
		},
		{
			"bullet list with stars",
			"* alpha\n* beta",
			[]docNode{bulletList(
				listItem(para(textNode("alpha"))),
				listItem(para(textNode("beta"))),
			)},
		},
		{
			"nested bullet list two levels",
			"- alpha\n  - alpha one\n  - alpha two\n- beta",
			[]docNode{bulletList(
				listItem(
					para(textNode("alpha")),
					bulletList(
						listItem(para(textNode("alpha one"))),
						listItem(para(textNode("alpha two"))),
					),
				),
				listItem(para(textNode("beta"))),
			)},
		},
		{
			"nested list tolerates four space indent",
			"- alpha\n    - deep\n- beta",
			[]docNode{bulletList(
				listItem(
					para(textNode("alpha")),
					bulletList(listItem(para(textNode("deep")))),
				),
				listItem(para(textNode("beta"))),
			)},
		},
		{
			"three levels of nesting",
			"- a\n  - b\n    - c",
			[]docNode{bulletList(
				listItem(
					para(textNode("a")),
					bulletList(listItem(
						para(textNode("b")),
						bulletList(listItem(para(textNode("c")))),
					)),
				),
			)},
		},
		{
			"ordered list",
			"1. first\n2. second",
			[]docNode{orderedList(
				listItem(para(textNode("first"))),
				listItem(para(textNode("second"))),
			)},
		},
		{
			"ordered list nested in bullet list",
			"- intro\n  1. one\n  2. two",
			[]docNode{bulletList(listItem(
				para(textNode("intro")),
				orderedList(
					listItem(para(textNode("one"))),
					listItem(para(textNode("two"))),
				),
			))},
		},
		{
			"marker type change splits sibling lists",
			"- a\n1. b",
			[]docNode{
				bulletList(listItem(para(textNode("a")))),
				orderedList(listItem(para(textNode("b")))),
			},
		},
		{
			"blockquote single line",
			"> quoted",
			[]docNode{blockquote(para(textNode("quoted")))},
		},
		{
			"blockquote consecutive lines merge",
			"> one\n> two",
			[]docNode{blockquote(para(textNode("one two")))},
		},
		{
			"blockquote with two paragraphs",
			"> first\n>\n> second",
			[]docNode{blockquote(
				para(textNode("first")),
				para(textNode("second")),
			)},
		},
		{
			"blockquote inline marks",
			"> **note**: careful",
			[]docNode{blockquote(para(
				textNode("note", bold),
				textNode(": careful"),
			))},
		},
		{
			"horizontal rule with dashes",
			"---",
			[]docNode{horizontalRule()},
		},
		{
			"horizontal rule with stars",
			"***",
			[]docNode{horizontalRule()},
		},
		{
			"code fence with language",
			"```go\nfmt.Println(\"hi\")\n```",
			[]docNode{codeBlock("go", "fmt.Println(\"hi\")")},
		},
		{
			"code fence without language",
			"```\nplain text\n```",
			[]docNode{codeBlock("", "plain text")},
		},
		{
			"code fence keeps markdown and backticks verbatim",
			"```md\n# not a heading\n**bold** and `code`\n```",
			[]docNode{codeBlock("md", "# not a heading\n**bold** and `code`")},
		},
		{
			"code fence preserves inner indentation",
			"```go\nfunc main() {\n\treturn\n}\n```",
			[]docNode{codeBlock("go", "func main() {\n\treturn\n}")},
		},
		{
			"unterminated fence consumes to end",
			"```go\nleftover",
			[]docNode{codeBlock("go", "leftover")},
		},
		{
			"bold with asterisks",
			"make **this** bold",
			[]docNode{para(
				textNode("make "),
				textNode("this", bold),
				textNode(" bold"),
			)},
		},
		{
			"bold with underscores",
			"make __this__ bold",
			[]docNode{para(
				textNode("make "),
				textNode("this", bold),
				textNode(" bold"),
			)},
		},
		{
			"italic with asterisk",
			"a *slanted* word",
			[]docNode{para(
				textNode("a "),
				textNode("slanted", italic),
				textNode(" word"),
			)},
		},
		{
			"italic with underscore",
			"a _slanted_ word",
			[]docNode{para(
				textNode("a "),
				textNode("slanted", italic),
				textNode(" word"),
			)},
		},
		{
			"intra word underscore stays literal",
			"snake_case_name",
			[]docNode{para(textNode("snake_case_name"))},
		},
		{
			"strikethrough",
			"gone ~~forever~~ now",
			[]docNode{para(
				textNode("gone "),
				textNode("forever", strike),
				textNode(" now"),
			)},
		},
		{
			"inline code",
			"run `make test` now",
			[]docNode{para(
				textNode("run "),
				textNode("make test", code),
				textNode(" now"),
			)},
		},
		{
			"code containing asterisks is not bold",
			"`**not bold**`",
			[]docNode{para(textNode("**not bold**", code))},
		},
		{
			"two code spans stay separate",
			"a `x` and `y`",
			[]docNode{para(
				textNode("a "),
				textNode("x", code),
				textNode(" and "),
				textNode("y", code),
			)},
		},
		{
			"link",
			"see [docs](https://example.com) here",
			[]docNode{para(
				textNode("see "),
				textNode("docs", link("https://example.com")),
				textNode(" here"),
			)},
		},
		{
			"bold inside link composes marks in open order",
			"[**strong** stuff](https://example.com)",
			[]docNode{para(
				textNode("strong", link("https://example.com"), bold),
				textNode(" stuff", link("https://example.com")),
			)},
		},
		{
			"image stays literal text",
			"before ![a photo](https://x/y.png) after",
			[]docNode{para(textNode("before ![a photo](https://x/y.png) after"))},
		},
		{
			"backslash escapes remove markers",
			"\\*not bold\\* and \\_this\\_",
			[]docNode{para(textNode("*not bold* and _this_"))},
		},
		{
			"escaped backslash collapses to one",
			"a\\\\b",
			[]docNode{para(textNode("a\\b"))},
		},
		{
			"unmatched single asterisk stays literal",
			"*bold",
			[]docNode{para(textNode("*bold"))},
		},
		{
			"unmatched double asterisk stays literal",
			"**bold",
			[]docNode{para(textNode("**bold"))},
		},
		{
			"unclosed bracket stays literal",
			"[text](",
			[]docNode{para(textNode("[text]("))},
		},
		{
			"link without closing paren stays literal",
			"[text](http",
			[]docNode{para(textNode("[text](http"))},
		},
		{
			"adjacent runs with identical marks merge",
			"[a](u)[b](u)",
			[]docNode{para(textNode("ab", link("u")))},
		},
		{
			"crlf input is normalized",
			"first\r\n\r\nsecond",
			[]docNode{para(textNode("first")), para(textNode("second"))},
		},
		{
			"empty input yields one paragraph with the raw text",
			"",
			[]docNode{para(textNode(""))},
		},
		{
			"whitespace only input yields one paragraph with the raw text",
			"   ",
			[]docNode{para(textNode("   "))},
		},
		{
			"whitespace lines only yield the raw text",
			"  \n  ",
			[]docNode{para(textNode("  \n  "))},
		},
		{
			"full document mixes every block type",
			"# Clip Title\n\n" +
				"Intro with **bold**, *italic*, and `code`.\n\n" +
				"- one\n- two\n  - nested\n\n" +
				"> quoted wisdom\n\n" +
				"---\n\n" +
				"## Section",
			[]docNode{
				heading(1, textNode("Clip Title")),
				para(
					textNode("Intro with "),
					textNode("bold", bold),
					textNode(", "),
					textNode("italic", italic),
					textNode(", and "),
					textNode("code", code),
					textNode("."),
				),
				bulletList(
					listItem(para(textNode("one"))),
					listItem(
						para(textNode("two")),
						bulletList(listItem(para(textNode("nested")))),
					),
				),
				blockquote(para(textNode("quoted wisdom"))),
				horizontalRule(),
				heading(2, textNode("Section")),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := MarkdownToBlocks(tc.in)
			if err != nil {
				t.Fatalf("MarkdownToBlocks(%q) returned error: %v", tc.in, err)
			}
			want, err := json.Marshal(docNode{Type: "doc", Content: tc.want})
			if err != nil {
				t.Fatalf("marshal expected doc: %v", err)
			}
			if got != string(want) {
				t.Errorf("MarkdownToBlocks(%q)\n\tgot:  %s\n\twant: %s", tc.in, got, want)
			}
		})
	}
}

// TestMarkdownToBlocks_RoundTripPlainText walks the produced JSON the same
// way the sqlite layer's blocksToText does (collect every text node, join
// with spaces) and asserts the plain text survives without markers. This is
// the path feeding the content column and FTS.
func TestMarkdownToBlocks_RoundTripPlainText(t *testing.T) {
	md := "# Real Talk\n\n" +
		"The **quick** brown fox *jumps* over `lazy` dogs and reads [the docs](https://example.com).\n\n" +
		"- alpha\n  - alpha one\n- beta\n\n" +
		"> quoted line one\n> quoted line two\n\n" +
		"1. first\n2. second\n"

	got, err := MarkdownToBlocks(md)
	if err != nil {
		t.Fatalf("MarkdownToBlocks returned error: %v", err)
	}

	var doc map[string]any
	if err := json.Unmarshal([]byte(got), &doc); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, got)
	}

	var parts []string
	collectText(doc, &parts)
	joined := strings.Join(parts, " ")

	want := strings.Join([]string{
		"Real Talk",
		"The ", "quick", " brown fox ", "jumps", " over ", "lazy", " dogs and reads ", "the docs", ".",
		"alpha", "alpha one", "beta",
		"quoted line one quoted line two",
		"first", "second",
	}, " ")

	if joined != want {
		t.Errorf("round trip mismatch\n\tgot:  %q\n\twant: %q", joined, want)
	}
}

// collectText mirrors walkNode in internal/infrastructure/sqlite: every text
// node contributes its text, children are walked in order.
func collectText(node map[string]any, out *[]string) {
	if t, ok := node["text"].(string); ok {
		*out = append(*out, t)
	}
	children, ok := node["content"].([]any)
	if !ok {
		return
	}
	for _, child := range children {
		if childMap, ok := child.(map[string]any); ok {
			collectText(childMap, out)
		}
	}
}

// TestMarkdownToBlocks_MalformedInputStaysValid guards the contract that any
// input — truncated markers, empty spans, stray delimiters — produces a
// valid document instead of panicking or erroring. The specific literal
// renderings are covered by the table above; this is the never-crash net
// for a parser fed external article content.
func TestMarkdownToBlocks_MalformedInputStaysValid(t *testing.T) {
	inputs := []string{
		"\\", "trailing backslash\\", "!", "![", "![alt](", "[", "[](", "[]()",
		"[a](u", "**", "****", "***x***", "*", "~~", "~~~~x", "`", "``", "`x",
		"- ", "1. ", "1.", ">", ">>", "#", "######", "#######", "```", "```go\n```",
		"- a\n    - b\n  - c\n- d", "C:\\Users\\path", "🎉 **bold** 🎉", "  \r\n\t\r\n  ",
	}
	for _, in := range inputs {
		got, err := MarkdownToBlocks(in)
		if err != nil {
			t.Errorf("MarkdownToBlocks(%q) error: %v", in, err)
			continue
		}
		var doc map[string]any
		if err := json.Unmarshal([]byte(got), &doc); err != nil {
			t.Errorf("MarkdownToBlocks(%q) produced invalid JSON: %v\n%s", in, err, got)
			continue
		}
		if doc["type"] != "doc" {
			t.Errorf("MarkdownToBlocks(%q) root type = %v, want doc", in, doc["type"])
		}
		if content, ok := doc["content"].([]any); !ok || len(content) == 0 {
			t.Errorf("MarkdownToBlocks(%q) has no blocks\n%s", in, got)
		}
	}
}
