#!/usr/bin/env node
// hube clip extractor.
//
// Reads a full HTML document from stdin and the page URL from argv[2],
// extracts the article (metadata + body) with Defuddle, converts the body
// to Markdown with the helper bundled in defuddle/full, and prints exactly
// one JSON object on stdout:
//
//   {"title":"...","content_markdown":"...","byline":null,"published":null,
//    "site_name":null,"domain":"example.com","word_count":123}
//
// `published` is "YYYY-MM-DD" or null. On any failure it prints
// {"error":"..."} and exits 1.
//
// Usage: node clip.mjs <url> < page.html (source; build.mjs bundles it
// into dist/clip.cjs, which is what the Go backend execs).
// Caller must exec under a deadline (e.g. 15s) and cap input size;
// pathological pages can spin CPU indefinitely.

// stdout carries only the final JSON object. Defuddle and its bundled
// turndown report conversion failures via console.log; route those to
// stderr so the stdout contract cannot be violated.
console.log = (...args) => console.error(...args);

// --- DOM shim: must be in place before defuddle loads ----------------------
// Turndown (bundled inside defuddle/full) reads the global `document` when
// it converts HTML at runtime, and defuddle's extractors probe for a global
// DOMParser. Node provides neither, so install linkedom-backed globals
// BEFORE importing defuddle. ESM static imports are hoisted, so defuddle
// must be imported dynamically after this shim runs.
import { parseHTML } from 'linkedom';

function LinkedomDOMParser() {}
LinkedomDOMParser.prototype.parseFromString = function (html) {
  return parseHTML(html).document;
};

if (typeof globalThis.window === 'undefined') globalThis.window = globalThis;
if (!globalThis.DOMParser) globalThis.DOMParser = LinkedomDOMParser;
if (!globalThis.window.DOMParser) globalThis.window.DOMParser = LinkedomDOMParser;
if (typeof globalThis.document === 'undefined') {
  globalThis.document = parseHTML(
    '<!DOCTYPE html><html><head></head><body></body></html>'
  ).document;
}

// defuddle/full is a UMD bundle: its markdown helper lives on the module's
// default export (module.exports), not as a statically detectable named
// export in Node's ESM-CJS interop. The imports are dynamic and sit inside
// main() so the shim above always runs first (esbuild CJS output forbids
// top-level await, and defuddle must load after the globals exist anyway).
async function loadDefuddle() {
  const { default: Defuddle } = await import('defuddle');
  const defuddleFull = await import('defuddle/full');
  return { Defuddle, createMarkdownContent: defuddleFull.default.createMarkdownContent };
}

// --- helpers ----------------------------------------------------------------

const MAX_ERROR_MESSAGE = 300;

/** Read all of stdin until EOF, binary-safe, and decode as UTF-8. */
async function readStdin() {
  const chunks = [];
  for await (const chunk of process.stdin) chunks.push(chunk);
  return Buffer.concat(chunks).toString('utf8');
}

/** Collapse and truncate an error message so the output stays bounded. */
function boundedMessage(error) {
  const raw = error instanceof Error ? error.message : String(error);
  return raw.replace(/\s+/g, ' ').trim().slice(0, MAX_ERROR_MESSAGE);
}

/**
 * Normalize Defuddle's raw published string to "YYYY-MM-DD" or null.
 * Date-only strings are returned verbatim (no timezone round-trip);
 * anything else must parse as a Date and is formatted from calendar
 * components: strings carrying an explicit offset (Z or +HH:MM) use UTC
 * components so the date is deterministic regardless of the machine's
 * timezone, while offset-free strings parse as local time and use local
 * components, matching how humans write dates like "March 11, 2024".
 */
function normalizePublished(raw) {
  if (!raw) return null;
  const value = String(raw).trim();
  const dateOnly = value.match(/^(\d{4}-\d{2}-\d{2})(?:[T ]|$)/);
  if (dateOnly) return dateOnly[1];
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return null;
  // An explicit trailing offset (Z, +05:30, -0500) pins the string to one
  // UTC instant whose calendar date shifts with the local timezone, so
  // those are formatted from UTC components; offset-free strings stay on
  // the local path.
  const hasOffset = /(?:Z|[+-]\d{2}:?\d{2})$/i.test(value);
  const year = hasOffset ? date.getUTCFullYear() : date.getFullYear();
  const month = hasOffset ? date.getUTCMonth() : date.getMonth();
  const day = hasOffset ? date.getUTCDate() : date.getDate();
  return `${year}-${String(month + 1).padStart(2, '0')}-${String(day).padStart(2, '0')}`;
}

/**
 * Defuddle falls back to the domain when no real site name exists. The
 * output contract already carries the domain, so treat a site name equal to
 * the URL hostname (defuddle's fallback value) as "no site name".
 */
function normalizeSiteName(site, hostname) {
  if (!site) return null;
  const normalize = (value) => value.trim().toLowerCase().replace(/^www\./, '');
  if (normalize(site) === normalize(hostname)) return null;
  return site;
}

/** Whitespace-separated token count of the markdown text. */
function estimateWordCount(markdown) {
  return markdown.split(/\s+/).filter(Boolean).length;
}

/** Write one JSON object to stdout and record the exit code. */
function writeJson(object, exitCode) {
  process.stdout.write(JSON.stringify(object) + '\n');
  // exitCode (not process.exit) so pending stdout writes always flush.
  process.exitCode = exitCode;
}

// --- main --------------------------------------------------------------------

async function main() {
  const url = process.argv[2];
  if (!url) {
    throw new Error('usage: clip.cjs <url> (HTML on stdin)');
  }
  const pageUrl = new URL(url); // throws on invalid URL: error JSON, exit 1
  // URL userinfo must never reach the note: defuddle resolves relative
  // links against this URL, so credentials would be baked into every
  // absolutized link in the markdown. Clear them before any use.
  pageUrl.username = '';
  pageUrl.password = '';

  const html = await readStdin();
  if (!html.trim()) {
    throw new Error('empty html input');
  }

  const { Defuddle, createMarkdownContent } = await loadDefuddle();
  const doc = parseHTML(html).document;
  const result = new Defuddle(doc, { url: pageUrl.href }).parse();
  const markdown = createMarkdownContent(result.content, pageUrl.href);

  // defuddle 0.19.13 (pinned in package-lock.json) reports a failed
  // turndown pass by returning a string that starts with this marker
  // instead of throwing. Reject it: the note would be a raw-HTML dump.
  // The lockfile pin keeps the prefix stable; if a defuddle bump renames
  // the marker, the deep-nesting test flips to exit 0 and catches it.
  if (markdown.startsWith('Partial conversion completed with errors')) {
    throw new Error('markdown conversion failed: partial conversion output');
  }

  if (!markdown.trim()) {
    throw new Error('no extractable article content');
  }

  writeJson(
    {
      title: result.title,
      content_markdown: markdown,
      byline: result.author || null,
      published: normalizePublished(result.published),
      site_name: normalizeSiteName(result.site, pageUrl.hostname),
      domain: pageUrl.hostname,
      word_count:
        result.wordCount > 0 ? result.wordCount : estimateWordCount(markdown),
    },
    0
  );
}

main().catch((error) => {
  writeJson({ error: boundedMessage(error) }, 1);
});
