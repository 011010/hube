import { test } from 'node:test';
import assert from 'node:assert/strict';
import { execFile } from 'node:child_process';
import { readFile } from 'node:fs/promises';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const CLIP = path.join(__dirname, 'dist', 'clip.cjs');
const TESTDATA = path.resolve(
  __dirname,
  '../../backend/internal/infrastructure/external/testdata'
);

/**
 * Run the built clipper with the given HTML on stdin and fake page URL.
 * Asserts on the resolved child, so a crash surfaces as a test failure.
 */
function runClip(url, html) {
  return new Promise((resolve, reject) => {
    const child = execFile(process.execPath, [CLIP, url], {
      maxBuffer: 32 * 1024 * 1024,
    });
    let stdout = '';
    let stderr = '';
    child.stdout.setEncoding('utf8');
    child.stderr.setEncoding('utf8');
    child.stdout.on('data', (chunk) => (stdout += chunk));
    child.stderr.on('data', (chunk) => (stderr += chunk));
    // The error-path runs exit before reading all of stdin; ignore EPIPE so
    // feeding stdin never fails the run.
    child.stdin.on('error', () => {});
    child.stdin.end(html);
    child.on('error', reject);
    child.on('close', (code, signal) => resolve({ code, signal, stdout, stderr }));
  });
}

test('clip_article.html: metadata, markdown structure, and domain', async () => {
  const html = await readFile(path.join(TESTDATA, 'clip_article.html'), 'utf8');
  const expected = JSON.parse(
    await readFile(path.join(TESTDATA, 'clip_article.expected.json'), 'utf8')
  );

  const { code, stdout, stderr } = await runClip('https://example.com/article', html);
  assert.equal(code, 0, `clip failed (stderr: ${stderr})`);

  const parsed = JSON.parse(stdout);
  for (const [field, value] of Object.entries(expected)) {
    assert.deepEqual(parsed[field], value, `unexpected ${field}`);
  }

  assert.match(parsed.content_markdown, /^#{1,6} .+$/m, 'markdown heading');
  assert.ok(parsed.content_markdown.includes('**'), 'bold marker');
  assert.ok(parsed.content_markdown.includes(']('), 'link marker');
  assert.ok(parsed.content_markdown.includes('`'), 'code span');
  assert.equal(parsed.domain, 'example.com');
});

test('clip_minimal.html: title extracted, missing metadata is null', async () => {
  const html = await readFile(path.join(TESTDATA, 'clip_minimal.html'), 'utf8');

  const { code, stdout, stderr } = await runClip('https://example.com/minimal', html);
  assert.equal(code, 0, `clip failed (stderr: ${stderr})`);

  const parsed = JSON.parse(stdout);
  assert.equal(parsed.title, 'A minimal page');
  assert.equal(parsed.byline, null);
  assert.equal(parsed.published, null);
  assert.equal(parsed.site_name, null);
  assert.ok(parsed.content_markdown.trim().length > 0, 'non-empty markdown');
});

test('clip_empty.html: JS-only shell fails with a JSON error', async () => {
  const html = await readFile(path.join(TESTDATA, 'clip_empty.html'), 'utf8');

  const { code, stdout } = await runClip('https://example.com/shell', html);
  assert.notEqual(code, 0, 'expected a non-zero exit');

  const parsed = JSON.parse(stdout);
  assert.equal(typeof parsed.error, 'string');
  assert.ok(parsed.error.length > 0);
});
