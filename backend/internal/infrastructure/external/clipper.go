package external

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Clip error taxonomy. Every error returned by DefuddleClipper wraps one
// of these; use errors.Is to branch. Messages are bounded and never
// include page HTML.
var (
	// ErrClipInvalidURL marks a URL that fails shape validation: scheme,
	// embedded credentials, or missing host.
	ErrClipInvalidURL = errors.New("invalid clip url")
	// ErrClipFetch marks transport failures: SSRF guard rejections, DNS,
	// dialing, non-2xx statuses, and the response size cap.
	ErrClipFetch = errors.New("fetch failed")
	// ErrClipExtraction marks extractor failures: the Node script exited
	// non-zero, timed out, or produced output outside the contract.
	ErrClipExtraction = errors.New("could not extract content")
)

const (
	// maxClipBodyBytes caps how much HTML the clipper reads. The cap is
	// enforced while streaming: the read stops one byte past the limit,
	// so an oversized page never fully transits.
	maxClipBodyBytes = 5 << 20 // 5 MB

	// maxClipRedirects bounds the requests a redirect chain may make: the
	// original plus the redirects followed. The hop that would exceed it
	// is refused inside CheckRedirect.
	maxClipRedirects = 5

	// maxClipDetailRunes bounds upstream detail (extractor and transport
	// error text) embedded in our messages so a hostile page cannot use
	// error strings as an amplification channel.
	maxClipDetailRunes = 300

	// clipUserAgent mirrors a desktop browser: many sites serve the real
	// article markup only to browser-like clients.
	clipUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"
)

// ExtractedPage is one clipped web page: the article body as Markdown plus
// the metadata Defuddle could find. Missing metadata is empty strings.
type ExtractedPage struct {
	Title           string
	ContentMarkdown string `json:"content_markdown"`
	Byline          string
	Published       string // "YYYY-MM-DD" or ""
	SiteName        string
	Domain          string
	WordCount       int
}

// clipScriptResult mirrors the one-line JSON contract of tools/clip's
// dist/clip.cjs. Nulls decode to zero values, so a single struct covers
// both the success and the {"error":...} failure shape.
type clipScriptResult struct {
	Title           string `json:"title"`
	ContentMarkdown string `json:"content_markdown"`
	Byline          string `json:"byline"`
	Published       string `json:"published"`
	SiteName        string `json:"site_name"`
	Domain          string `json:"domain"`
	WordCount       int    `json:"word_count"`
	Error           string `json:"error"`
}

// DefuddleClipper clips web pages. Go owns the transport — URL validation,
// SSRF guards, redirect re-validation, size caps, timeouts — while Node
// (the bundled tools/clip/dist/clip.cjs script) owns extraction.
type DefuddleClipper struct {
	scriptPath   string
	fetchTimeout time.Duration
	execTimeout  time.Duration
	// transport is the guarded dial path shared by every fetch. It is
	// built once per clipper: http.Transport is safe for concurrent use,
	// keep-alives are disabled so it pools nothing, and the guard reads
	// CLIP_ALLOW_PRIVATE at dial time — no per-request rebuild needed.
	transport *http.Transport
}

// NewDefuddleClipper builds a clipper around the bundled extraction script.
func NewDefuddleClipper(scriptPath string) *DefuddleClipper {
	return &DefuddleClipper{
		scriptPath:   scriptPath,
		fetchTimeout: 10 * time.Second,
		execTimeout:  15 * time.Second,
		transport:    newClipTransport(),
	}
}

// newClipTransport builds the fetch transport. The default transport is
// unusable here: it pools connections (a pooled connection skips the
// guarded dial on reuse), it honors environment proxies (which resolve
// targets themselves, out of reach of every guard), and it resolves hosts
// without validating the address actually connected to.
func newClipTransport() *http.Transport {
	return &http.Transport{
		DialContext: dialClipGuarded,
		// A pooled connection to the same host would skip DialContext —
		// and with it the address guard — on every reuse.
		DisableKeepAlives: true,
		// Environment proxies resolve and connect on their own, beyond
		// the guards. Clip fetches always dial directly.
		Proxy: nil,
	}
}

// dialClipGuarded closes the DNS rebinding window: it resolves the host
// once, picks a single answer, validates that address with the same rules
// as the pre-flight guard, and connects to that exact address. DNS can
// change between the pre-flight check and this dial; only here is the
// address actually connected to guaranteed validated. It runs inside the
// client's timeout, so DNS burns the same 10s budget as the rest of the
// fetch.
func dialClipGuarded(ctx context.Context, network, addr string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, fmt.Errorf("%w: dial: %s", ErrClipFetch, boundedClipDetail(err.Error()))
	}
	var d net.Dialer
	if allowPrivateIPs() {
		// Dev/test bypass: the address guards are explicitly off.
		return d.DialContext(ctx, network, addr)
	}
	if net.ParseIP(host) != nil {
		// IP literal: validate, then connect to it as written. No
		// resolution window at all.
		if err := validateClipHost(host); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrClipFetch, err)
		}
		return d.DialContext(ctx, network, addr)
	}
	addrs, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, fmt.Errorf("%w: resolve: %s", ErrClipFetch, boundedClipDetail(err.Error()))
	}
	for _, a := range addrs {
		// Pick the first answer that passes the same IP rules and
		// connect to that address itself. If DNS flipped every answer
		// hostile since the pre-flight check, nothing is dialed.
		if err := validateClipHost(a.IP.String()); err != nil {
			continue
		}
		return d.DialContext(ctx, network, net.JoinHostPort(a.IP.String(), port))
	}
	return nil, fmt.Errorf("%w: %s resolves to no public address", ErrClipFetch, boundedClipDetail(host))
}

// Extract clips rawURL: validate the URL, fetch the page under the SSRF
// guards, then extract the article by execing the Node script.
func (c *DefuddleClipper) Extract(ctx context.Context, rawURL string) (*ExtractedPage, error) {
	if err := validateClipURL(rawURL); err != nil {
		return nil, err
	}
	html, err := c.fetchPage(ctx, rawURL)
	if err != nil {
		return nil, err
	}
	return c.runScript(ctx, rawURL, html)
}

// validateClipURL enforces the clip URL shape: http or https only, no
// embedded userinfo, no fragments, non-empty host. It is pure — no DNS,
// no environment.
func validateClipURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrClipInvalidURL, boundedClipDetail(err.Error()))
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("%w: scheme must be http or https", ErrClipInvalidURL)
	}
	if u.User != nil {
		return fmt.Errorf("%w: userinfo not allowed", ErrClipInvalidURL)
	}
	if u.Fragment != "" {
		return fmt.Errorf("%w: fragments not allowed", ErrClipInvalidURL)
	}
	if u.Hostname() == "" {
		return fmt.Errorf("%w: empty host", ErrClipInvalidURL)
	}
	return nil
}

// validateClipHost rejects IP literals that are not public destinations:
// loopback, private (RFC1918 and IPv6 unique-local), link-local (including
// the cloud metadata range), multicast, and unspecified addresses. It is
// pure — no DNS, no environment — so hostnames pass through and are
// checked per resolved address by the fetch path.
func validateClipHost(host string) error {
	ip := net.ParseIP(host)
	if ip == nil {
		return nil
	}
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() ||
		ip.IsMulticast() || ip.IsUnspecified() {
		return fmt.Errorf("ip %s is not a public destination", ip)
	}
	return nil
}

// allowPrivateIPs reports whether the dev/test bypass is active. It is
// read per call so tests can toggle it with t.Setenv. The bypass skips
// only the private-address checks; validateClipURL rules always apply.
func allowPrivateIPs() bool {
	v, ok := os.LookupEnv("CLIP_ALLOW_PRIVATE")
	return ok && v == "1"
}

// checkClipHost rejects hosts that are (or resolve to) non-public
// addresses. Hostnames are resolved and every A/AAAA address must pass
// validateClipHost.
func checkClipHost(ctx context.Context, host string) error {
	if allowPrivateIPs() {
		return nil
	}
	if net.ParseIP(host) != nil {
		return validateClipHost(host)
	}
	addrs, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return fmt.Errorf("resolve: %s", boundedClipDetail(err.Error()))
	}
	for _, a := range addrs {
		if err := validateClipHost(a.IP.String()); err != nil {
			return err
		}
	}
	return nil
}

// checkClipRedirect is the http.Client redirect policy: every hop re-runs
// the full URL and host validation, so a public origin cannot bounce the
// fetch into a private network or a non-HTTP scheme.
func checkClipRedirect(ctx context.Context, req *http.Request, via []*http.Request) error {
	if len(via) >= maxClipRedirects {
		// via holds the requests already made (the original plus the
		// redirects followed); the hop that would make request
		// maxClipRedirects+1 is refused here.
		return fmt.Errorf("%w: stopped after %d requests", ErrClipFetch, len(via))
	}
	if err := validateClipURL(req.URL.String()); err != nil {
		return fmt.Errorf("%w: redirect rejected: %w", ErrClipFetch, err)
	}
	if err := checkClipHost(ctx, req.URL.Hostname()); err != nil {
		return fmt.Errorf("%w: redirect rejected: %w", ErrClipFetch, err)
	}
	return nil
}

// fetchPage downloads the page behind rawURL, which must already have
// passed validateClipURL (Extract enforces it). The host guard runs twice
// by design: checkClipHost rejects before any network work, and the
// guarded dial re-validates at connect time — the dial is the security
// boundary, the pre-flight is a fast-fail optimization. Redirects are
// re-validated per hop; the read stops one byte past maxClipBodyBytes
// instead of draining; non-2xx statuses are rejected.
func (c *DefuddleClipper) fetchPage(ctx context.Context, rawURL string) ([]byte, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrClipInvalidURL, boundedClipDetail(err.Error()))
	}
	if err := checkClipHost(ctx, u.Hostname()); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrClipFetch, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrClipFetch, boundedClipDetail(err.Error()))
	}
	req.Header.Set("User-Agent", clipUserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml")

	client := &http.Client{
		Timeout:   c.fetchTimeout, // covers DNS, dial, TLS, and body
		Transport: c.transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return checkClipRedirect(ctx, req, via)
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		var uerr *url.Error
		if errors.As(err, &uerr) {
			// Drop the wrapper: its message echoes the full request URL,
			// which a hostile redirect Location could inflate.
			err = uerr.Err
		}
		if errors.Is(err, ErrClipFetch) {
			// Redirect guard or guarded dial: already wrapped and bounded.
			// The passthrough also preserves chained sentinels, e.g.
			// ErrClipInvalidURL from redirect re-validation.
			return nil, err
		}
		// Bound the wrap: the raw transport error can echo attacker
		// text (a hostile 302 Location reaches this path measured at
		// 40k runes).
		return nil, fmt.Errorf("%w: %s", ErrClipFetch, boundedClipDetail(err.Error()))
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("%w: status %d %s", ErrClipFetch, resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	// io.LimitReader stops pulling from the socket one byte past the cap,
	// so an oversized page is cut mid-stream instead of being drained.
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxClipBodyBytes+1))
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrClipFetch, boundedClipDetail(err.Error()))
	}
	if len(body) > maxClipBodyBytes {
		return nil, fmt.Errorf("%w: page exceeds %d bytes", ErrClipFetch, maxClipBodyBytes)
	}
	return body, nil
}

// runScript execs the Node extractor with the page URL as argv[2] and the
// HTML on stdin, under a execTimeout ceiling (the caller's deadline may
// be tighter), then decodes the one-line JSON contract.
func (c *DefuddleClipper) runScript(ctx context.Context, pageURL string, html []byte) (*ExtractedPage, error) {
	execCtx, cancel := context.WithTimeout(ctx, c.execTimeout)
	defer cancel()

	cmd := exec.CommandContext(execCtx, "node", c.scriptPath, pageURL)
	cmd.Stdin = bytes.NewReader(html)

	out, err := cmd.Output()
	if err != nil {
		if execCtx.Err() != nil {
			if ctx.Err() != nil {
				// The caller's deadline was the tighter one.
				return nil, fmt.Errorf("%w: %w", ErrClipExtraction, ctx.Err())
			}
			return nil, fmt.Errorf("%w: node exceeded %s", ErrClipExtraction, c.execTimeout)
		}
		detail := boundedClipDetail(scriptDetail(out, err))
		// cmd.Output captures failed-run stderr on the ExitError; a
		// bounded excerpt beats a bare "exit status N" for debugging
		// without giving the script a channel to inflate our messages.
		var ee *exec.ExitError
		if errors.As(err, &ee) && len(ee.Stderr) > 0 {
			return nil, fmt.Errorf("%w: %s (stderr: %s)",
				ErrClipExtraction, detail, boundedClipDetail(string(ee.Stderr)))
		}
		return nil, fmt.Errorf("%w: %s", ErrClipExtraction, detail)
	}

	var res clipScriptResult
	if err := json.Unmarshal(out, &res); err != nil {
		return nil, fmt.Errorf("%w: invalid extractor output: %s", ErrClipExtraction, boundedClipDetail(err.Error()))
	}
	if res.Error != "" {
		return nil, fmt.Errorf("%w: %s", ErrClipExtraction, boundedClipDetail(res.Error))
	}
	if strings.TrimSpace(res.ContentMarkdown) == "" {
		return nil, fmt.Errorf("%w: empty article", ErrClipExtraction)
	}

	return &ExtractedPage{
		Title:           res.Title,
		ContentMarkdown: res.ContentMarkdown,
		Byline:          res.Byline,
		Published:       res.Published,
		SiteName:        res.SiteName,
		Domain:          res.Domain,
		WordCount:       res.WordCount,
	}, nil
}

// scriptDetail prefers the extractor's own {"error":...} message over the
// generic process error text.
func scriptDetail(stdout []byte, err error) string {
	var res struct {
		Error string `json:"error"`
	}
	if json.Unmarshal(stdout, &res) == nil && res.Error != "" {
		return res.Error
	}
	return err.Error()
}

// boundedClipDetail collapses whitespace and truncates upstream error
// detail so page content can neither leak into nor bloat our messages.
func boundedClipDetail(detail string) string {
	detail = strings.Join(strings.Fields(detail), " ")
	if r := []rune(detail); len(r) > maxClipDetailRunes {
		return string(r[:maxClipDetailRunes]) + "..."
	}
	return detail
}
