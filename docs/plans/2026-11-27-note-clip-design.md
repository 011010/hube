# Web Clip Design

Date: 2026-11-27
Status: Approved

## Overview

`POST /api/v1/notes/clip` turns a web article into a note: fetch the page,
extract the article with Defuddle, and create a note with readable content, a
`clipping` tag, and typed properties (`source`, `author`, `published`,
`domain`). This is the capture half of the knowledge pipeline: properties
(from #9) give the metadata a home, and views (from #10) can later filter the
inbox.

## Architecture

Two runtimes, one request, no extra service:

- **Go owns transport.** URL validation, SSRF guards, fetching, timeouts, and
  size caps live in the Go backend, where they are table-testable.
- **Node owns extraction.** Defuddle needs a DOM; it runs in a small bundled
  Node script that the Go process execs per clip. HTML goes in on stdin, JSON
  comes out on stdout. No second container, no HTTP hop, no long-lived
  process. Clip actions are user-initiated and rare, so spawn latency
  (~200ms) is irrelevant.

This mirrors how `obsidian-clipper` itself uses Defuddle in Node (linkedom as
the DOM shim), but keeps the fetching hardened in Go.

## Runtime and Packaging

- `tools/clip/` holds `package.json` (defuddle, linkedom, esbuild), the
  source `clip.mjs`, and a bundle step producing `dist/clip.cjs` (gitignored).
- The extractor finds the script via `CLIP_SCRIPT` (default
  `tools/clip/dist/clip.cjs`). `make build-clip` builds it for development.
- The backend Dockerfile gains a stage that installs the bundle, and the final
  image installs `nodejs` via apk (alpine ships a supported LTS). Image grows
  ~80MB; acceptable for a self-hosted single-binary-plus-node deployment.
- Extraction failure (paywalls, JS-only pages) returns `422` and creates no
  note.

## SSRF and Fetch Guards

Only the Go side fetches. Guards, in order:

1. Scheme must be `http` or `https`; no userinfo in the URL; no fragments.
2. Resolve DNS; reject loopback, private, link-local, multicast, and
   unspecified ranges for **every** resolved address (v4 and v6).
3. Follow redirects manually, re-running validation per hop (max 5).
4. `GET` with a browser-ish User-Agent, 10s timeout, 5MB response cap
   (enforced while reading, not after).
5. `CLIP_ALLOW_PRIVATE=1` (default off, documented as dev/test only) permits
   loopback targets so local fixture servers can be used in verification.

## Extraction Contract

The script reads HTML on stdin plus the page URL as an argument (Defuddle
uses it for metadata) and prints one JSON object:

```json
{
  "title": "string",
  "content_markdown": "string",
  "byline": "string|null",
  "published": "YYYY-MM-DD|null",
  "site_name": "string|null",
  "domain": "string",
  "word_count": 123
}
```

A non-zero exit or invalid JSON is a `422`.

## Note Construction

- `title`: page title, trimmed to 200 runes, fallback `domain` + date.
- `content`: the article markdown converted to TipTap blocks by a Go converter
  with **block-level and inline support**: headings (`#`–`######`), bullet and
  ordered lists (nested by indent), blockquotes, horizontal rules, fenced
  code blocks, and paragraphs; inline `**bold**`, `*italic*`, `` `code` ``,
  `~~strikethrough~~`, `[text](href)` links (as link marks), and backslash
  escapes. Images stay as literal markdown text (image handling is a
  follow-up). Empty input yields a single paragraph with the raw text.
- `tags`: `["clipping"]`.
- `folder_id`: from the request (the current folder selection in the UI);
  omitted means no folder.
- `properties`: `source` (text, the URL), `author` (text, from byline),
  `published` (date), `domain` (text) — each only when the extractor produced
  a value. Written via `SetProperties` after `Create` (the create path
  rejects property payloads by design).
- The response returns the note with the clipped properties populated.

## Dedupe

Clipping the same URL twice is a likely mistake. The repository gains a
lookup by exact property (`key` + JSON-encoded `value`), backed by the
existing `idx_note_properties_key`. A duplicate returns `409` with
`{"error": ..., "existing_id": "<note id>"}`; no note is created.

## API

```
POST /api/v1/notes/clip
{"url": "https://example.com/post", "folder_id": "optional"}
```

- `201` → the created note (with properties)
- `400` → invalid URL, blocked target, or oversized body
- `409` → already clipped, with `existing_id`
- `422` → extraction failed
- `502` → upstream fetch error (timeout, status ≥ 400, size cap)

## Frontend

A "Clip" button in the Notes page header opens a small modal: paste a URL
(the current folder is used, shown as a hint). Submitting calls the endpoint;
on success the created note opens in the editor. A `409` offers to open the
existing note instead. Failures surface via the existing `role="alert"`
pattern.

## Testing

- Script: golden HTML fixtures under `tools/clip/testdata/`, executed through
  the Go extractor test (`exec` + parse) so the integration itself is covered
  by `go test ./...`.
- SSRF: pure table tests for the URL/IP validator; no network.
- Service: mock extractor interface (the `Embedder` pattern from
  `rag.go` — interface defined where it is consumed).
- Handler: mock extractor; `400/409/422/502` paths.
- End-to-end: a local fixture page served on loopback with
  `CLIP_ALLOW_PRIVATE=1`, exercising fetch → extract → note → properties.

## Non-Goals

- Browser extension and highlighter.
- Image downloading, rewriting, or inline image nodes (markdown image syntax
  stays as literal text).
- Scheduled/bulk clipping.
- Full-text dedupe beyond the exact source URL.

## Acceptance Criteria

1. Clip a real article: the note opens with a readable title, block-converted
   content with real bold/italic/code/link marks, the `clipping` tag, and
   `source`/`author`/`published`/`domain` properties in the requested folder.
2. Clip the same URL again → `409` with the existing note's id.
3. `http://localhost/x` and a private-IP target → `400`, nothing fetched.
4. A page with no extractable article → `422`, no note created.
5. The Clip modal creates and opens the note; a duplicate offers to open the
   existing one.

## Follow-Ups

1. Saved view preset for the clipping inbox (e.g. "Unread clippings" via a
   `read` boolean property).
2. Images: download or inline-proxy.
3. A browser-extension or share-target front end for the endpoint.
4. A markdown lint pass on Defuddle output (stray escapes, nested-list
   edge cases).
