package note

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/husari/hube/internal/domain/note"
)

// maxClipTitleRunes caps a clipped note's title. Page titles can run long;
// 200 runes keeps the note list readable without touching the content.
const maxClipTitleRunes = 200

// ExtractedPage is what a page extractor produces: the article body as
// markdown plus the metadata the extractor could find. Declared here so the
// service and its tests never depend on infrastructure; the concrete
// extractor converts into this shape at the wiring edge.
type ExtractedPage struct {
	ContentMarkdown string
	Title           string
	Byline          string
	Published       string // YYYY-MM-DD or ""
	SiteName        string
	Domain          string
	WordCount       int
}

// PageExtractor turns a URL into an extracted page. The real implementation
// fetches under SSRF guards and shells out to the Defuddle bundle; tests
// provide fakes. Declared where it is consumed, like Embedder in rag.go.
type PageExtractor interface {
	Extract(ctx context.Context, rawURL string) (*ExtractedPage, error)
}

type ClipInput struct {
	URL      string
	FolderID *string
}

// ErrDuplicateClip reports that this URL was already clipped. NoteID points
// at the existing note so the caller can offer to open it instead.
type ErrDuplicateClip struct{ NoteID string }

func (e *ErrDuplicateClip) Error() string { return "already clipped" }

// ErrClipInvalidURL marks a URL the clip flow refuses before any network
// work. The full guard set (credentials, fragments, private targets) lives
// in the real extractor, whose own sentinel carries the same message; the
// HTTP layer classifies both.
var ErrClipInvalidURL = errors.New("invalid clip url")

// ClipService clips web pages into notes.
type ClipService struct {
	repo      note.Repository
	extractor PageExtractor
	// markdownToBlocks is the converter seam. It defaults to MarkdownToBlocks;
	// tests swap it to exercise the fallback without depending on the stdlib's
	// version-specific depth limits (the reason this seam exists — see the
	// converter-fallback test).
	markdownToBlocks func(string) (string, error)
}

func NewClipService(repo note.Repository, extractor PageExtractor) *ClipService {
	return &ClipService{repo: repo, extractor: extractor, markdownToBlocks: MarkdownToBlocks}
}

// Clip turns in.URL into a note: dedupe against previously clipped sources,
// extract the page, convert the article to TipTap blocks, validate the note
// and its metadata, create the note, then attach the extracted metadata as
// properties.
//
// Extractor errors are wrapped only with a "clip:" prefix and their chain
// is left intact, so the HTTP layer can classify them via the
// infrastructure sentinels (ErrClipFetch, ErrClipExtraction, ...). This
// package deliberately does not import infrastructure to branch on them.
// Property validation failures get the same wrap and keep their
// *note.ValidationError in the chain, so the HTTP layer can classify them
// as a 400-class request failure.
func (s *ClipService) Clip(ctx context.Context, in ClipInput) (*note.Note, error) {
	if err := validateClipURL(in.URL); err != nil {
		return nil, err
	}

	// Dedupe on the stored form of the source property: the JSON encoding
	// SetProperties writes, so ampersands and quotes in the URL match
	// byte-for-byte what a previous clip stored.
	stored, err := json.Marshal(in.URL)
	if err != nil {
		return nil, fmt.Errorf("clip: encode source url: %w", err)
	}
	existing, err := s.repo.FindNoteIDByProperty(ctx, "source", string(stored))
	if err != nil {
		return nil, fmt.Errorf("clip: find existing clip: %w", err)
	}
	if existing != "" {
		return nil, &ErrDuplicateClip{NoteID: existing}
	}

	page, err := s.extractor.Extract(ctx, in.URL)
	if err != nil {
		return nil, fmt.Errorf("clip: %w", err)
	}
	if page == nil {
		return nil, fmt.Errorf("clip: extractor returned no page")
	}

	blocks, err := s.clipBlocks(page.ContentMarkdown, clipLogLabel(page, in.URL))
	if err != nil {
		return nil, err
	}
	n := &note.Note{
		Title:    clipTitle(page),
		Blocks:   blocks,
		Tags:     []string{"clipping"},
		FolderID: in.FolderID,
	}
	n.Normalize()
	if err := n.Validate(); err != nil {
		return nil, err
	}

	// Metadata is validated before anything is persisted: a hostile page
	// (non-ISO published date, oversized text) must fail the clip outright
	// rather than leave a bare note row behind — invisible to dedupe, so a
	// retry would create a second note.
	props := clipProperties(in.URL, page)
	if err := note.ValidateProperties(props); err != nil {
		return nil, fmt.Errorf("clip: validate properties: %w", err)
	}

	// The create path rejects property payloads by design, so the note is
	// created bare and the metadata is written right after.
	if err := s.repo.Create(ctx, n); err != nil {
		return nil, fmt.Errorf("clip: create note: %w", err)
	}

	if err := s.repo.SetProperties(ctx, n.ID, props); err != nil {
		return nil, fmt.Errorf("clip: set properties: %w", err)
	}
	n.Properties = props
	return n, nil
}

// validateClipURL enforces the URL shape this layer can check without any
// network work: an http or https URL with a host. Obviously-bad input
// fails fast here; hostile-but-well-formed URLs are the real extractor's
// to catch before any fetch.
func validateClipURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrClipInvalidURL, boundedClipURLDetail(err))
	}
	if (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" {
		return ErrClipInvalidURL
	}
	return nil
}

// maxClipURLDetailRunes bounds the url.Parse error text wrapped into
// ErrClipInvalidURL: parse errors echo the raw URL, so a hostile input
// can inflate the message. 300 matches the infrastructure layer's cap
// for upstream detail.
const maxClipURLDetailRunes = 300

// boundedClipURLDetail collapses whitespace and truncates the url.Parse
// error text so a hostile URL cannot use the wrap as an amplification
// channel — the same discipline the infrastructure clipper applies to
// upstream errors, kept local because this package does not import it.
func boundedClipURLDetail(err error) string {
	detail := strings.Join(strings.Fields(err.Error()), " ")
	if r := []rune(detail); len(r) > maxClipURLDetailRunes {
		return string(r[:maxClipURLDetailRunes]) + "..."
	}
	return detail
}

// clipTitle picks the note title: the page title trimmed of surrounding
// space and capped at 200 runes, falling back to the domain — with the
// publish date appended when the extractor found one — when the page had
// no usable title.
func clipTitle(page *ExtractedPage) string {
	title := strings.TrimSpace(page.Title)
	if utf8.RuneCountInString(title) > maxClipTitleRunes {
		title = string([]rune(title)[:maxClipTitleRunes])
	}
	if title != "" {
		return title
	}
	if page.Domain == "" {
		return ""
	}
	if page.Published != "" {
		return page.Domain + " - " + page.Published
	}
	return page.Domain
}

// clipLogLabel picks what the clip logs identify a page by: its title,
// falling back to the URL, capped at maxClipTitleRunes so a hostile page
// cannot flood the log.
func clipLogLabel(page *ExtractedPage, rawURL string) string {
	label := strings.TrimSpace(page.Title)
	if label == "" {
		label = rawURL
	}
	if utf8.RuneCountInString(label) > maxClipTitleRunes {
		label = string([]rune(label)[:maxClipTitleRunes])
	}
	return label
}

// clipBlocks converts the extracted markdown into TipTap blocks, through
// the service's converter seam. When the converter fails — its json.Marshal
// step exhausts the encoder's depth limit somewhere around 2500
// list-nesting levels — the clip still keeps its content: the same
// single-paragraph doc MarkdownToBlocks builds for empty input, holding
// the raw markdown as one plain-text paragraph. The fallback logs label
// (page title or URL) so operators can see which pages lost their
// formatting.
func (s *ClipService) clipBlocks(md, label string) (string, error) {
	blocks, err := s.markdownToBlocks(md)
	if err == nil {
		return blocks, nil
	}
	log.Printf("clip: convert markdown %s: %v", label, err)
	fallback, err := json.Marshal(docNode{
		Type: "doc",
		Content: []docNode{{
			Type:    "paragraph",
			Content: []docNode{{Type: "text", Text: &md}},
		}},
	})
	if err != nil {
		// Unreachable in practice: a flat doc cannot exhaust the depth
		// limit and invalid UTF-8 encodes lossily rather than failing.
		return "", fmt.Errorf("clip: build fallback blocks: %w", err)
	}
	return string(fallback), nil
}

// clipProperties builds the property set for a clipped note from the values
// the extractor actually produced. source is always present — it is the
// dedupe anchor — and the rest are skipped when empty.
func clipProperties(rawURL string, page *ExtractedPage) []note.Property {
	props := []note.Property{
		{Key: "source", Type: note.PropertyText, Value: rawURL},
	}
	if page.Byline != "" {
		props = append(props, note.Property{Key: "author", Type: note.PropertyText, Value: page.Byline})
	}
	if page.Published != "" {
		props = append(props, note.Property{Key: "published", Type: note.PropertyDate, Value: page.Published})
	}
	if page.Domain != "" {
		props = append(props, note.Property{Key: "domain", Type: note.PropertyText, Value: page.Domain})
	}
	return props
}
