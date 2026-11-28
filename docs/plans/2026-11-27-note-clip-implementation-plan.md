# Web Clip Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** `POST /api/v1/notes/clip` turns an article URL into a note with readable TipTap content, a `clipping` tag, and typed properties.

**Architecture:** Go owns transport (URL validation, SSRF guards, fetch, caps); Node owns extraction (bundled Defuddle script, exec'd per clip, HTML on stdin, JSON on stdout). A pure Go converter turns the extracted markdown into TipTap blocks. The service creates the note, then writes properties via the existing endpoint path. Stacked on `feat/note-views`.

**Tech Stack:** Go 1.26 (Chi, sqlx), Node ≥ 20 (defuddle + linkedom + esbuild), React 19 frontend, alpine Docker image gains `nodejs`.

**Design doc:** `docs/plans/2026-11-27-note-clip-design.md`

---

### Task 1: Extraction script and bundle

**Files:**
- Create: `tools/clip/package.json`
- Create: `tools/clip/clip.mjs`
- Create: `tools/clip/build.mjs`
- Create: `backend/internal/infrastructure/external/testdata/clip_article.html`, `clip_minimal.html`, `clip_empty.html`
- Modify: `.gitignore`, `Makefile`

**Requirements:**

1. `tools/clip/package.json`: dependencies `defuddle` (pinned ^0.19) and `linkedom`; devDependency `esbuild`. Add `scripts.test` that runs `node --test clip.test.mjs` (see 5).
2. `tools/clip/clip.mjs`:
   - Reads the full HTML from stdin (binary-safe until EOF).
   - Takes the page URL as `process.argv[2]`.
   - Shims `DOMParser`/`document` with linkedom **before** importing Defuddle (the obsidian-clipper CLI does the same; Turndown probes `window.DOMParser` at module init).
   - Uses Defuddle to parse the HTML and produce markdown. Verify the exact exports against the installed package (`node_modules/defuddle/dist/*.d.ts`): `parse()` for metadata and the `defuddle/full` markdown content helper that `obsidian-clipper` uses (`createMarkdownContent`). If the helper's signature differs, adapt and note it in the report.
   - Prints exactly one JSON object to stdout (contract below) and exits 0. On any failure prints `{"error":"<bounded message>"}` to stdout, exits 1. Nothing else may write to stdout.
3. Output contract (field names fixed; the Go side parses these):
   ```json
   {"title":"...","content_markdown":"...","byline":null,"published":null,"site_name":null,"domain":"example.com","word_count":123}
   ```
   `published` is `YYYY-MM-DD` or null. `content_markdown` is the article body. `domain` from the URL host.
4. `tools/clip/build.mjs`: esbuild bundles `clip.mjs` → `dist/clip.cjs` (platform node, format cjs, target node20, `linkedom` NOT external — bundle it; defuddle bundles too).
5. Golden fixtures under `backend/internal/infrastructure/external/testdata/`:
   - `clip_article.html`: realistic article — `<title>`, `<article>` with two headings, a paragraph containing a bold span, a link, an inline code element, one bullet list (one nested sub-list), a blockquote; `<meta name="author">`; `<time datetime="2024-03-11">`.
   - `clip_minimal.html`: title + one paragraph, no metadata.
   - `clip_empty.html`: JS-only shell with no extractable article content.
   Keep them hand-written (no network). Write an expected-JSON for `clip_article.html` (`clip_article.expected.json`) capturing title/byline/published/site_name/word_count — NOT the full markdown (too brittle); assert `content_markdown` contains the bold/link/code/list markers.
6. `tools/clip/clip.test.mjs` using `node:test`: reads each fixture from `../../../backend/internal/infrastructure/external/testdata/`, execs the built `dist/clip.cjs` via `child_process`, and asserts the contract above (`clip_empty.html` must exit non-zero with an `error` object).
7. `.gitignore`: `tools/clip/node_modules/`, `tools/clip/dist/`.
8. `Makefile` target `build-clip`: `cd tools/clip && npm install && node build.mjs`.

**Verify:**

```bash
cd tools/clip && npm install && node build.mjs && npm test
```

All fixtures pass. Report the Defuddle API shapes you actually found.

**Commit:** `feat(clip): add defuddle extraction script`

---

### Task 2: Markdown → TipTap blocks converter (pure Go, TDD)

**Files:**
- Create: `backend/internal/application/note/markdown.go`
- Test: `backend/internal/application/note/markdown_test.go`

**Contract:**

```go
// MarkdownToBlocks converts article markdown into a TipTap document JSON
// string ({"type":"doc","content":[...]}). Inline syntax becomes marks;
// images stay literal. Empty input yields a single paragraph.
func MarkdownToBlocks(md string) (string, error)
```

**Block-level rules:**

- Normalize `\r\n` → `\n`. Lines are grouped; blank lines separate blocks.
- `#`–`######` → `{"type":"heading","attrs":{"level":N},"content":[inline]}`.
- `- ` or `* ` → `bulletList`; `N. ` → `orderedList`; items are
  `{"type":"listItem","content":[{"type":"paragraph","content":[inline]}, <nested list>?]}`.
  Nesting by indentation: 2 spaces per step, tolerates 4, and a TAB counts as one step (Defuddle indents continuation lines with tabs). One or more levels.
- `> ` → `{"type":"blockquote","content":[paragraphs]}` (consecutive lines merge into one quote with paragraphs).
- `---`/`***` alone → `{"type":"horizontalRule"}`.
- Fenced code ``` ```lang ```` → `{"type":"codeBlock","attrs":{"language":"lang"}}` with one text child, content verbatim (no inline parsing inside).
- Everything else → `{"type":"paragraph","content":[inline]}`. A paragraph with empty inline content is skipped.
- Empty markdown → one paragraph containing the raw input as text.

**Inline rules (order matters):**

1. Backslash escapes: backslash escapes only ASCII punctuation (CommonMark); otherwise the backslash stays literal (no mark).
2. Inline code: `` `code` `` → text node with mark `{"type":"code"}`; no nesting inside.
3. Images `![alt](src)` → literal text (no image node).
4. Links `[text](href)` → text node with mark `{"type":"link","attrs":{"href":href}}`; nested marks inside link text are allowed.
5. Bold `**x**` / `__x__` → mark `{"type":"bold"}`; `***x***` / `___x___` opens bold AND italic (Turndown emits this for `<strong><em>` in both nesting orders; test the triple delimiter before the double).
6. Italic `*x*` / `_x_` → mark `{"type":"italic"}` (single underscore only when not intra-word).
7. Strikethrough `~~x~~` → mark `{"type":"strike"}`.
8. Unmatched markers stay literal.

Text nodes are `{"type":"text","text":"...","marks":[...]}` with marks sorted by open order. Adjacent identical-mark runs may merge.

**Tests (table-driven):** each block type; nested lists; blockquote with two paragraphs; code fence with backticks inside; every inline mark alone and combined (bold inside link, code containing asterisks); image literal; escapes; unmatched markers; empty input; `\r\n` input; and a round-trip assertion that `blocksToText` on the output recovers the plain text (walkNode joins text nodes — verify the FTS path works).

**Verify:**

```bash
cd backend && go test ./internal/application/note/ -count=1 -v && go vet ./... && gofmt -l ./internal
```

**Commit:** `feat(notes): convert markdown to tipTap blocks`

---

### Task 3: Fetch guards and extractor (infrastructure)

**Files:**
- Create: `backend/internal/infrastructure/external/clipper.go`
- Test: `backend/internal/infrastructure/external/clipper_test.go`

**Public surface:**

```go
type ExtractedPage struct {
    TitleMarkdown      string  // article markdown
    Title, Byline      string
    Published           string // YYYY-MM-DD or ""
    SiteName, Domain   string
    WordCount           int
}

type DefuddleClipper struct{ /* script path, timeout */ }
func NewDefuddleClipper(scriptPath string) *DefuddleClipper
func (c *DefuddleClipper) Extract(ctx context.Context, rawURL string) (*ExtractedPage, error)
```

**`Extract` does everything, in order:**

1. `validateClipURL(rawURL)`: http/https only, no userinfo, host present. Pure, exported for tests.
2. `fetchPage(ctx, rawURL)`: resolve DNS; reject loopback/private/link-local/multicast/unspecified for every A and AAAA address; `CLIP_ALLOW_PRIVATE=1` (read per call, `os.LookupEnv`) bypasses only the IP check — scheme rules stay. `http.Client` with 10s timeout, 5MB cap enforced via a limited reader while streaming, a browser-ish `User-Agent`, and `CheckRedirect` that re-validates each hop (max 5, manual re-resolution per hop). Non-2xx → fetch error.
3. Run the script: `exec.CommandContext(ctx, "node", scriptPath, url)` with the HTML on stdin, 15s timeout, stdout/stderr captured; non-zero exit or bad JSON → extraction error.

**Error taxonomy (exported sentinels in this package):**

```go
var ErrClipInvalidURL  = errors.New("invalid clip url")
var ErrClipFetch       = errors.New("fetch failed")
var ErrClipExtraction  = errors.New("could not extract content")
```

Wrap causes with `%w`; messages bounded (no full HTML in errors).

**Tests:**

- `TestValidateClipURL`: table — valid https, http, missing scheme, ftp, userinfo, empty host.
- `TestBlockedTargets`: table of hosts/IPs (127.0.0.1, ::1, 10.0.0.1, 192.168.1.1, 169.254.169.254, ::ffff:127.0.0.1, multicast) all rejected; a public IP and a public hostname pass the pure IP check (`validateClipHost` split so the table needs no DNS).
- `TestFetchPage_AllowPrivate`: with `t.Setenv("CLIP_ALLOW_PRIVATE","1")`, fetch a loopback `httptest.Server` serving `clip_article.html` → 200 + body.
- `TestDefuddleClipper_Extract`: golden run against `testdata/clip_article.html` (needs `node` and `CLIP_SCRIPT` pointing at a built bundle — build it in the test via `exec` of `npm`? No: skip if the bundle is absent, with `t.Skipf` naming the `make build-clip` prerequisite; run it for real when present). Assert title, byline, published, domain, word_count, and that `content_markdown` contains the bold/link/code markers.

**Verify:**

```bash
cd backend && make -C .. build-clip 2>/dev/null || (cd ../tools/clip && npm install && node build.mjs)
cd backend && go test ./internal/infrastructure/external/ -count=1 -v
```

**Commit:** `feat(clip): fetch pages safely and extract with defuddle`

---

### Task 4: Repository lookup by property

**Files:**
- Modify: `backend/internal/domain/note/repository.go`
- Modify: `backend/internal/infrastructure/sqlite/note_repo.go`
- Test: `backend/internal/infrastructure/sqlite/note_repo_test.go`

Add to the `Repository` interface:

```go
// FindNoteIDByProperty returns the id of the first note holding the exact
// property value, or "" when none does. value is the JSON-encoded stored form.
FindNoteIDByProperty(ctx context.Context, key, value string) (string, error)
```

Implementation uses the existing index (`WHERE key = ? AND value = ? LIMIT 1`, deterministic by `note_id`). Adding it to the interface breaks the `mockRepo` in `application/note/service_test.go` — add the minimal stub there (a `findByProperty` field defaulting to `""`), as done for `SetProperties`.

Test: two notes with properties; lookup by existing key/value hits the right note; a value stored under a different key misses; JSON quoting matters (`"https://x"` as stored, not the raw URL); missing key → `""`, no error.

**Verify:** `cd backend && go test ./... -count=1`

**Commit:** `feat(notes): look up notes by property value`

---

### Task 5: Clip service (TDD)

**Files:**
- Create: `backend/internal/application/note/clip.go`
- Test: `backend/internal/application/note/clip_test.go`
- Modify: `backend/internal/application/note/service_test.go` (mock fields only)

**Contract:**

```go
type ExtractedPage struct { /* same fields as the infrastructure struct, redeclared here */ }

// PageExtractor turns a URL into an extracted page. The real one fetches and
// shells out to Defuddle; tests provide fakes. Defined here, where it is
// consumed (the Embedder pattern from rag.go).
type PageExtractor interface {
    Extract(ctx context.Context, rawURL string) (*ExtractedPage, error)
}

type ClipInput struct { URL string; FolderID *string }

type ClipService struct{ repo note.Repository; extractor PageExtractor }
func NewClipService(repo note.Repository, extractor PageExtractor) *ClipService
func (s *ClipService) Clip(ctx context.Context, in ClipInput) (*note.Note, error)
```

**Error taxonomy (in this package):**

```go
var ErrClipInvalidURL = errors.New("invalid clip url")
type ErrDuplicateClip struct{ NoteID string }   // Error(): "already clipped"
type ErrClipExtraction struct{ Err error }     // wraps; Error(): "could not extract content"
type ErrClipFetch struct{ Err error }          // wraps; Error(): "fetch failed"
```

**Flow:**

1. Basic URL validation at this layer (scheme + host) so `400` is testable with a fake extractor — full SSRF stays in the real one.
2. Dedupe: `FindNoteIDByProperty(ctx, "source", jsonEncode(in.URL))` → hit returns `ErrDuplicateClip{NoteID}`.
3. `extractor.Extract` — map its failure to `ErrClipExtraction`/`ErrClipFetch` (the infrastructure sentinels are matched with `errors.Is` against this package's own copies via shared message? No: the infrastructure returns its own sentinels; the service wraps whatever non-sentinel error it gets into the right category by checking `errors.Is(err, infra.ErrClipFetch)` would import infrastructure from application — layering violation. Instead: the service wraps ALL extractor errors into `ErrClipExtraction` with the cause, EXCEPT it defines the categories and the HTTP handler maps `ErrClipFetch` vs `ErrClipExtraction`… resolution: keep two service-level types and have the HANDLER receive the infrastructure error through the service's wrap (cause chain). The handler checks `errors.As` on the cause chain for the infrastructure sentinels — no: keep it simple — the service returns `ErrClipExtraction` wrapping the cause; the handler maps 422 for `ErrClipExtraction` and inspects `errors.Is(err, infra.ErrClipFetch)` on the CAUSE via a helper `UnwrapClipKind(err)` exported by the application layer. Decide in implementation; the tests must pin: fetch errors → 502-class, extraction errors → 422-class, and the mapping helper is unit-tested.)
4. Build the note: `Title` trimmed to 200 runes (fallback `domain`), `Blocks` via `MarkdownToBlocks` (fallback: single paragraph with the raw markdown on error), `Tags: []string{"clipping"}`, `FolderID` from input.
5. `repo.Create` (properties stay empty — the create guard rejects payloads).
6. `repo.SetProperties` with `source` (text), `author` (text), `published` (date), `domain` (text) — only non-empty values, keys already normalized.
7. Populate `n.Properties` with the written set and return the note.

**Tests:** duplicate path (repo fake hit); happy path (fake extractor with full metadata; assert note fields, tags, blocks contain a heading node — parse the JSON; properties exactly the four); sparse metadata (only title) → properties only `source`+`domain`; extraction error → `ErrClipExtraction`, no note created (assert `repo.created == nil`); invalid URL → `ErrClipInvalidURL` before the extractor is called (assert extractor not invoked); fetch-class error maps to the fetch category; `MarkdownToBlocks` failure falls back gracefully (inject by… converter is pure; simulate by empty markdown → still a note).

**Verify:** `cd backend && go test ./... -count=1 && go vet ./... && gofmt -l ./internal`

**Commit:** `feat(notes): clip urls into notes`

---

### Task 6: HTTP endpoint

**Files:**
- Modify: `backend/internal/infrastructure/http/handler/note.go`
- Modify: `backend/internal/infrastructure/http/handler/note_test.go`
- Modify: `backend/internal/infrastructure/http/server.go`, `backend/cmd/api/main.go`

**Route:** `r.Post("/clip", h.clip)` inside the `/notes` router.

**Request:** `{"url":"...","folder_id":"optional"}` — `url` required.

**Mapping:**

- `400` invalid/missing URL (service `ErrClipInvalidURL`, the extractor's `ErrClipInvalidURL` in the chain, property validation, or decode failure)
- `502` blocked target or upstream fetch failure (`ErrClipFetch`: SSRF rejection, timeout, non-2xx, size cap)
- `409` `{"error":"already clipped","existing_id":"<id>"}` from `ErrDuplicateClip`
- `422` `{"error":"could not extract content"}` from `ErrClipExtraction` (fetch-class cause → `502` `{"error":"fetch failed"}`)
- `201` the created note JSON (properties included — the service populated them)

**Wiring:** `NewNoteHandler(noteSvc, ragSvc, clipSvc *appnote.ClipService)` — nil-safe like `rag` (nil → 503 `"clip not configured"`). `server.go` and `main.go` construct `NewDefuddleClipper(os.Getenv("CLIP_SCRIPT") with default "tools/clip/dist/clip.cjs")` and the `ClipService`. The `architecture` package has layer tests — keep the handler depending only on the application service.

**Handler tests (fake extractor defined in the test file):**

- 201 happy path: note created (GET it back: tags contain `clipping`, properties present, blocks parse as a TipTap doc).
- 400 missing url / bad scheme.
- 409 with `existing_id` (pre-clip once, then again).
- 422 extraction failure (no note created).
- 502 fetch-class failure.
- 503 when clipSvc is nil.

**Verify:** `cd backend && go test ./... -count=1 && go vet ./... && gofmt -l ./internal`

**Commit:** `feat(notes): expose the clip endpoint`

---

### Task 7: Frontend clip UI

**Files:**
- Modify: `frontend/src/hooks/useNotes.ts` (add `useClipNote`)
- Modify: `frontend/src/pages/Notes/index.tsx`
- Create: `frontend/src/components/organisms/ClipUrlModal.tsx`
- Modify: `frontend/src/types/index.ts` (clip payload types)

**Behavior:**

- "Clip" button in the Notes page header (near "New note") opens `ClipUrlModal` (the existing `atoms/Modal`): a URL input (required, autofocus), a hint showing the target folder (current selection), Submit/Cancel.
- `useClipNote()` posts to `/notes/clip` with `{url, folder_id}`; on `201` the created note opens (`setSelectedNoteId(note.id)`) and the notes list refetches.
- On `409` with `existing_id`: show an inline option in the modal — "Already clipped. Open it?" — that selects the existing note.
- Network/4xx/5xx failures surface via `role="alert"` in the modal (same `text-xs text-red-400` pattern); nothing else changes.
- The modal must not collide with the view selector state; closing it never touches `activeViewId`.

**Verify:**

```bash
cd frontend && pnpm run build && pnpm run lint && pnpm vitest run
```

**Commit:** `feat(notes): clip urls from the notes page`

---

### Task 8: Docker and Makefile wiring

**Files:**
- Modify: `backend/Dockerfile`
- Modify: `Makefile` (done in Task 1 — only touch what remains)

**Dockerfile:**

- New stage `clipper`: `FROM node:22-alpine`, `WORKDIR /clip`, copy `tools/clip/package*.json`, `npm ci`, copy the source, run the esbuild bundle → `/clip/dist/clip.cjs`.
  (The backend build context is `./backend`; the compose file builds with that context. Either move the clip build into the frontend/root context or change the api build context to the repo root and adjust `COPY backend/` paths. Inspect `docker-compose.yml` and pick the smaller diff; report the choice.)
- Final stage: `apk add --no-cache nodejs ca-certificates tzdata`, `COPY --from=clipper /clip/dist/clip.cjs /app/clip/clip.cjs`, `ENV CLIP_SCRIPT=/app/clip/clip.cjs`.

**Verify:** `docker build` the api image locally if Docker is running (report skip otherwise); `docker run --rm <image> node --version` proves Node exists.

**Commit:** `feat(clip): bundle the extractor into the api image`

---

### Task 9: Verification

1. Full suites: backend (`go test ./... -count=1`, vet, gofmt) and frontend (build, lint, vitest).
2. End-to-end against a DB copy with `CLIP_ALLOW_PRIVATE=1`: serve `testdata/clip_article.html` on a loopback port (`python3 -m http.server`), clip it via curl, and assert the 201 response: title, `clipping` tag, blocks containing a heading and a bold mark, and the four properties; clip again → 409 with the same id; clip `clip_empty.html` → 422 with no note created.
3. Real-URL smoke (best effort, network permitting): clip one real article and eyeball the note content in the response JSON.
4. Walk the five acceptance criteria from the design doc.

---

## Out of Scope

Browser extension; image handling beyond literal text; scheduled clipping; full-text dedupe; inline marks beyond bold/italic/code/strike/link.
