package note

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/husari/hube/internal/domain/note"
)

// fakeExtractor counts calls so tests can prove the service never reaches
// extraction (invalid URLs, duplicates) and that the URL passes through.
type fakeExtractor struct {
	calls  int
	gotURL string
	page   *ExtractedPage
	err    error
}

func (f *fakeExtractor) Extract(_ context.Context, rawURL string) (*ExtractedPage, error) {
	f.calls++
	f.gotURL = rawURL
	if f.err != nil {
		return nil, f.err
	}
	return f.page, nil
}

// errExtractorBoom stands in for an infrastructure sentinel. The service
// must keep the wrap chain intact so the HTTP layer can classify the cause.
var errExtractorBoom = errors.New("extractor exploded")

// fullPage is the richest page an extractor can produce.
func fullPage() *ExtractedPage {
	return &ExtractedPage{
		ContentMarkdown: "# How to Clip\n\nSome **bold** text.\n",
		Title:           "How to Clip Anything",
		Byline:          "Ada Author",
		Published:       "2024-03-11",
		SiteName:        "Example Weekly",
		Domain:          "example.com",
		WordCount:       42,
	}
}

// unmarshalBlocks decodes a TipTap doc with the package's own node type.
func unmarshalBlocks(t *testing.T, blocks string) docNode {
	t.Helper()
	var doc docNode
	if err := json.Unmarshal([]byte(blocks), &doc); err != nil {
		t.Fatalf("blocks are not a TipTap doc: %v\nblocks: %s", err, blocks)
	}
	return doc
}

// propertySet indexes props by key so tests assert set equality, not order.
func propertySet(t *testing.T, props []note.Property) map[string]note.Property {
	t.Helper()
	set := make(map[string]note.Property, len(props))
	for _, p := range props {
		if _, dup := set[p.Key]; dup {
			t.Fatalf("duplicate property key %q in %+v", p.Key, props)
		}
		set[p.Key] = p
	}
	return set
}

func TestClipService_Clip_InvalidURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{"ftp scheme", "ftp://example.com/file"},
		{"missing scheme", "example.com/post"},
		{"empty url", ""},
		{"empty host", "http://"},
		{"non-http scheme", "mailto:user@example.com"},
		{"malformed", "http://[::1"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockRepo{}
			extractor := &fakeExtractor{page: fullPage()}
			svc := NewClipService(repo, extractor)

			_, err := svc.Clip(context.Background(), ClipInput{URL: tc.url})
			if !errors.Is(err, ErrClipInvalidURL) {
				t.Fatalf("error = %v, want ErrClipInvalidURL", err)
			}
			if extractor.calls != 0 {
				t.Errorf("extractor was called %d times, want 0 for an invalid URL", extractor.calls)
			}
			if repo.created != nil {
				t.Error("no note should be created for an invalid URL")
			}
		})
	}
}

func TestClipService_Clip_Duplicate(t *testing.T) {
	// The ampersand pins the stored form: the dedupe lookup must pass the
	// JSON encoding SetProperties writes (json.Marshal HTML-escapes & as
	// \u0026), not the raw URL and not strconv.Quote.
	rawURL := "https://example.com/post?utm_source=rss&ref=feed"
	stored, err := json.Marshal(rawURL)
	if err != nil {
		t.Fatalf("marshal url: %v", err)
	}

	repo := &mockRepo{findByProperty: "note-7"}
	extractor := &fakeExtractor{}
	svc := NewClipService(repo, extractor)

	_, err = svc.Clip(context.Background(), ClipInput{URL: rawURL})
	if err == nil {
		t.Fatal("Clip expected a duplicate error")
	}
	var dup *ErrDuplicateClip
	if !errors.As(err, &dup) {
		t.Fatalf("error type = %T, want *ErrDuplicateClip", err)
	}
	if dup.NoteID != "note-7" {
		t.Errorf("ErrDuplicateClip.NoteID = %q, want %q", dup.NoteID, "note-7")
	}
	if repo.findByPropertyKey != "source" {
		t.Errorf("lookup key = %q, want %q", repo.findByPropertyKey, "source")
	}
	if repo.findByPropertyValue != string(stored) {
		t.Errorf("lookup value = %q, want the JSON-encoded stored form %q", repo.findByPropertyValue, string(stored))
	}
	if extractor.calls != 0 {
		t.Errorf("extractor was called %d times, want 0 on a duplicate", extractor.calls)
	}
	if repo.created != nil {
		t.Error("Clip should not create a note for a duplicate URL")
	}
}

func TestClipService_Clip_ExtractionError(t *testing.T) {
	repo := &mockRepo{}
	extractor := &fakeExtractor{err: fmt.Errorf("defuddle: %w", errExtractorBoom)}
	svc := NewClipService(repo, extractor)

	_, err := svc.Clip(context.Background(), ClipInput{URL: "https://example.com/broken"})
	if err == nil {
		t.Fatal("Clip expected the extractor error to propagate")
	}
	if !strings.HasPrefix(err.Error(), "clip: ") {
		t.Errorf("error = %q, want a %q prefix", err.Error(), "clip: ")
	}
	if !errors.Is(err, errExtractorBoom) {
		t.Errorf("errors.Is dropped the wrap chain: %v", err)
	}
	if repo.created != nil {
		t.Error("no note should be created when extraction fails")
	}
}

func TestClipService_Clip_NilPageFromExtractor(t *testing.T) {
	// A broken extractor returning (nil, nil) violates its contract; the
	// service reports it instead of panicking on the nil page.
	repo := &mockRepo{}
	extractor := &fakeExtractor{}
	svc := NewClipService(repo, extractor)

	_, err := svc.Clip(context.Background(), ClipInput{URL: "https://example.com/ghost"})
	if err == nil {
		t.Fatal("Clip expected an error when the extractor returns no page")
	}
	if !strings.Contains(err.Error(), "no page") {
		t.Errorf("error = %q, want it to mention %q", err.Error(), "no page")
	}
	if repo.created != nil {
		t.Error("no note should be created without an extracted page")
	}
}

func TestClipService_Clip_InvalidExtractedMetadataRejectedBeforeCreate(t *testing.T) {
	// A hostile page can feed the extractor garbage metadata. Property
	// validation must fire before Create: failing after would leave a bare
	// note row behind — invisible to dedupe, so a retry would create a
	// second note.
	repo := &mockRepo{}
	extractor := &fakeExtractor{page: &ExtractedPage{
		ContentMarkdown: "Body.\n",
		Title:           "Hostile Page",
		Published:       "March 11, 2024",
		Domain:          "example.com",
	}}
	svc := NewClipService(repo, extractor)

	_, err := svc.Clip(context.Background(), ClipInput{URL: "https://example.com/hostile"})
	if err == nil {
		t.Fatal("Clip expected a validation error for a non-ISO published date")
	}
	if !strings.HasPrefix(err.Error(), "clip: ") {
		t.Errorf("error = %q, want a %q prefix", err.Error(), "clip: ")
	}
	if !strings.Contains(err.Error(), "published") {
		t.Errorf("error = %q, want it to mention the published property", err.Error())
	}
	var valErr *note.ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("error type = %T, want *note.ValidationError in the chain", err)
	}
	if valErr.Field != "properties.published" {
		t.Errorf("ValidationError.Field = %q, want %q", valErr.Field, "properties.published")
	}
	if extractor.calls != 1 {
		t.Errorf("extractor was called %d times, want 1", extractor.calls)
	}
	if repo.created != nil {
		t.Error("no note should be created when property validation fails")
	}
	if repo.savedProperties != nil {
		t.Error("SetProperties should not run when property validation fails")
	}
}

func TestClipService_Clip_HappyPath(t *testing.T) {
	repo := &mockRepo{}
	extractor := &fakeExtractor{page: fullPage()}
	svc := NewClipService(repo, extractor)

	folder := "folder-1"
	rawURL := "https://example.com/how-to-clip"
	n, err := svc.Clip(context.Background(), ClipInput{URL: rawURL, FolderID: &folder})
	if err != nil {
		t.Fatalf("Clip returned unexpected error: %v", err)
	}

	if extractor.calls != 1 || extractor.gotURL != rawURL {
		t.Errorf("extractor called %d times with url %q, want 1 with %q", extractor.calls, extractor.gotURL, rawURL)
	}
	if n.Title != "How to Clip Anything" {
		t.Errorf("Title = %q, want the extracted title", n.Title)
	}
	if n.Status != note.StatusDraft {
		t.Errorf("Status = %q, want %q", n.Status, note.StatusDraft)
	}
	if len(n.Tags) != 1 || n.Tags[0] != "clipping" {
		t.Errorf("Tags = %#v, want [clipping]", n.Tags)
	}
	if n.FolderID == nil || *n.FolderID != "folder-1" {
		t.Errorf("FolderID = %#v, want folder-1", n.FolderID)
	}

	// The create path rejects property payloads by design, so the note
	// must reach the repo bare and gain properties only via SetProperties.
	if repo.created != n {
		t.Fatal("Clip did not persist the built note")
	}
	if repo.createdPropertyCount != 0 {
		t.Errorf("Create received %d properties, want 0", repo.createdPropertyCount)
	}
	if repo.savedProperties == nil {
		t.Fatal("SetProperties received a nil slice")
	}

	set := propertySet(t, repo.savedProperties)
	want := map[string]struct {
		typ   string
		value any
	}{
		"source":    {note.PropertyText, rawURL},
		"author":    {note.PropertyText, "Ada Author"},
		"published": {note.PropertyDate, "2024-03-11"},
		"domain":    {note.PropertyText, "example.com"},
	}
	if len(set) != len(want) {
		t.Fatalf("SetProperties received %d properties (%+v), want %d", len(set), repo.savedProperties, len(want))
	}
	for key, w := range want {
		g, ok := set[key]
		if !ok {
			t.Errorf("missing property %q in %+v", key, repo.savedProperties)
			continue
		}
		if g.Type != w.typ || g.Value != w.value {
			t.Errorf("property %q = {%s %v}, want {%s %v}", key, g.Type, g.Value, w.typ, w.value)
		}
	}

	// The response carries exactly the written set.
	if len(n.Properties) != len(repo.savedProperties) {
		t.Fatalf("note carries %d properties, SetProperties wrote %d", len(n.Properties), len(repo.savedProperties))
	}
	for i := range repo.savedProperties {
		if n.Properties[i] != repo.savedProperties[i] {
			t.Errorf("note property %d = %+v, want %+v", i, n.Properties[i], repo.savedProperties[i])
		}
	}

	// The article converted to a TipTap doc with the heading on top.
	doc := unmarshalBlocks(t, n.Blocks)
	if doc.Type != "doc" || len(doc.Content) == 0 {
		t.Fatalf("blocks are not a doc with content: %s", n.Blocks)
	}
	if doc.Content[0].Type != "heading" {
		t.Errorf("first block = %q, want heading", doc.Content[0].Type)
	}
	if level, ok := doc.Content[0].Attrs["level"].(float64); !ok || level != 1 {
		t.Errorf("heading level = %#v, want 1", doc.Content[0].Attrs["level"])
	}
}

func TestClipService_Clip_SparseMetadata(t *testing.T) {
	repo := &mockRepo{}
	extractor := &fakeExtractor{page: &ExtractedPage{
		ContentMarkdown: "Just a paragraph.\n",
		Title:           "Sparse Page",
		Domain:          "example.org",
	}}
	svc := NewClipService(repo, extractor)

	n, err := svc.Clip(context.Background(), ClipInput{URL: "https://example.org/sparse"})
	if err != nil {
		t.Fatalf("Clip returned unexpected error: %v", err)
	}
	if n.Title != "Sparse Page" {
		t.Errorf("Title = %q, want the extracted title", n.Title)
	}
	if n.FolderID != nil {
		t.Errorf("FolderID = %v, want none when the input omits it", *n.FolderID)
	}
	set := propertySet(t, repo.savedProperties)
	if len(set) != 2 {
		t.Fatalf("SetProperties received %d properties (%+v), want exactly source and domain", len(set), repo.savedProperties)
	}
	if set["source"].Type != note.PropertyText || set["source"].Value != "https://example.org/sparse" {
		t.Errorf("source = %+v, want the clipped URL as text", set["source"])
	}
	if set["domain"].Type != note.PropertyText || set["domain"].Value != "example.org" {
		t.Errorf("domain = %+v, want the extracted domain as text", set["domain"])
	}
}

func TestClipService_Clip_TitleFallback(t *testing.T) {
	tests := []struct {
		name      string
		published string
		want      string
	}{
		{"domain and date", "2024-03-11", "example.com - 2024-03-11"},
		{"domain only", "", "example.com"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockRepo{}
			extractor := &fakeExtractor{page: &ExtractedPage{
				ContentMarkdown: "Body.\n",
				Domain:          "example.com",
				Published:       tc.published,
			}}
			svc := NewClipService(repo, extractor)

			n, err := svc.Clip(context.Background(), ClipInput{URL: "https://example.com/no-title"})
			if err != nil {
				t.Fatalf("Clip returned unexpected error: %v", err)
			}
			if n.Title != tc.want {
				t.Errorf("Title = %q, want %q", n.Title, tc.want)
			}
		})
	}
}

func TestClipService_Clip_TitleCappedAt200Runes(t *testing.T) {
	// Multibyte runes prove the cap counts runes, not bytes: a byte cut
	// at 200 would slice through the middle of an é.
	long := strings.Repeat("é", 250) + strings.Repeat("Z", 50)
	repo := &mockRepo{}
	extractor := &fakeExtractor{page: &ExtractedPage{
		ContentMarkdown: "Body.\n",
		Title:           long,
		Domain:          "example.com",
	}}
	svc := NewClipService(repo, extractor)

	n, err := svc.Clip(context.Background(), ClipInput{URL: "https://example.com/long-title"})
	if err != nil {
		t.Fatalf("Clip returned unexpected error: %v", err)
	}
	if got := utf8.RuneCountInString(n.Title); got != 200 {
		t.Errorf("title holds %d runes, want exactly 200", got)
	}
	if want := string([]rune(long)[:200]); n.Title != want {
		t.Errorf("Title was not cut at rune 200")
	}
}

func TestClipService_Clip_ConverterFailureFallsBackToParagraph(t *testing.T) {
	// ~2500 nesting levels is where MarkdownToBlocks really fails: the
	// json.Marshal inside exhausts the encoder's depth limit (2400
	// converts fine; 2500 errors with "exceeded max depth"). Building the
	// pathological input — instead of injecting a converter error — pins
	// the fallback against the failure production can actually hit.
	var md strings.Builder
	for i := 0; i < 2500; i++ {
		md.WriteString(strings.Repeat(" ", 2*i))
		md.WriteString("- item\n")
	}
	raw := md.String()
	if _, err := MarkdownToBlocks(raw); err == nil {
		t.Fatal("fixture stopped triggering the converter depth failure; raise the nesting level")
	}

	repo := &mockRepo{}
	extractor := &fakeExtractor{page: &ExtractedPage{
		ContentMarkdown: raw,
		Title:           "Deeply Nested",
		Domain:          "example.com",
	}}
	svc := NewClipService(repo, extractor)

	n, err := svc.Clip(context.Background(), ClipInput{URL: "https://example.com/nested"})
	if err != nil {
		t.Fatalf("Clip returned unexpected error: %v", err)
	}
	doc := unmarshalBlocks(t, n.Blocks)
	if len(doc.Content) != 1 || doc.Content[0].Type != "paragraph" {
		t.Fatalf("fallback blocks hold %d blocks, want a single paragraph", len(doc.Content))
	}
	texts := doc.Content[0].Content
	if len(texts) != 1 || texts[0].Text == nil || *texts[0].Text != raw {
		t.Error("fallback paragraph does not hold the raw markdown as plain text")
	}
	// The clip still completes: properties were written.
	if repo.savedProperties == nil {
		t.Error("SetProperties was not called after the converter fallback")
	}
}

func TestClipService_Clip_EmptyMarkdownUsesConverterEmptyDoc(t *testing.T) {
	// The service passes the markdown through as-is, so empty markdown
	// lands in MarkdownToBlocks' own empty-input fallback: one paragraph
	// holding the empty raw text. Chosen behavior, pinned here.
	want, err := MarkdownToBlocks("")
	if err != nil {
		t.Fatalf("MarkdownToBlocks(empty) = %v", err)
	}
	repo := &mockRepo{}
	extractor := &fakeExtractor{page: &ExtractedPage{Title: "Empty Article", Domain: "example.com"}}
	svc := NewClipService(repo, extractor)

	n, err := svc.Clip(context.Background(), ClipInput{URL: "https://example.com/empty-article"})
	if err != nil {
		t.Fatalf("Clip returned unexpected error: %v", err)
	}
	if n.Blocks != want {
		t.Errorf("Blocks = %s, want the converter's empty-input doc %s", n.Blocks, want)
	}
}

func TestClipService_Clip_TitlelessDomainlessPageRejected(t *testing.T) {
	// The real extractor always fills Domain from the URL host, so this
	// page violates the extraction contract. The service still refuses to
	// store a titleless note: validation fires before the repo sees it.
	repo := &mockRepo{}
	extractor := &fakeExtractor{page: &ExtractedPage{
		ContentMarkdown: "Body.\n",
		Byline:          "Ada Author",
	}}
	svc := NewClipService(repo, extractor)

	_, err := svc.Clip(context.Background(), ClipInput{URL: "https://example.com/anonymous"})
	if err == nil {
		t.Fatal("Clip expected a validation error for a titleless, domainless page")
	}
	var valErr *note.ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("error type = %T, want *note.ValidationError", err)
	}
	if valErr.Field != "title" {
		t.Errorf("ValidationError.Field = %q, want %q", valErr.Field, "title")
	}
	if repo.created != nil {
		t.Error("no note should be created for a titleless page")
	}
}
