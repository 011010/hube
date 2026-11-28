package external

import (
	"bytes"
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// clipToolsDir points at the Node extraction tool. go test always runs a
// package's tests with the package source directory as the working
// directory, so this relative path holds no matter where the module is
// checked out.
var clipToolsDir = filepath.Join("..", "..", "..", "..", "tools", "clip")

// clipBundlePath returns the bundled extractor path, building the bundle
// first if it is missing. Tests that need to exec node call this and are
// skipped when the toolchain or bundle is unavailable, so `go test ./...`
// stays green on machines without the clip toolchain.
func clipBundlePath(t *testing.T) string {
	t.Helper()

	node, err := exec.LookPath("node")
	if err != nil {
		t.Skipf("node not found in PATH (run make build-clip to test extraction): %v", err)
	}

	bundle := filepath.Join(clipToolsDir, "dist", "clip.cjs")
	if _, err := os.Stat(bundle); errors.Is(err, os.ErrNotExist) {
		build := exec.Command(node, "build.mjs")
		build.Dir = clipToolsDir
		if out, err := build.CombinedOutput(); err != nil {
			t.Skipf("clip bundle missing and build failed (run `make build-clip`): %v\n%s", err, out)
		}
	}
	if _, err := os.Stat(bundle); err != nil {
		t.Skipf("clip bundle not found at %s (run `make build-clip`)", bundle)
	}
	return bundle
}

func TestValidateClipURL(t *testing.T) {
	tests := []struct {
		name    string
		rawURL  string
		wantErr bool
	}{
		{"valid https", "https://example.com/posts/wal", false},
		{"valid http", "http://example.com/", false},
		{"valid with port and query", "https://example.com:8443/posts/wal?utm=x", false},
		{"missing scheme", "example.com/posts/wal", true},
		{"ftp scheme", "ftp://example.com/pub", true},
		{"userinfo credentials", "https://user:pass@example.com/", true},
		{"fragment", "https://example.com/posts/wal#section", true},
		{"bare hash carries no fragment", "https://example.com/posts/wal#", false},
		{"empty host", "http:///posts/wal", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateClipURL(tt.rawURL)
			if tt.wantErr && !errors.Is(err, ErrClipInvalidURL) {
				t.Fatalf("validateClipURL(%q) = %v; want ErrClipInvalidURL", tt.rawURL, err)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("validateClipURL(%q) = %v; want nil", tt.rawURL, err)
			}
		})
	}
}

func TestValidateClipHost(t *testing.T) {
	tests := []struct {
		host    string
		wantErr bool
	}{
		// Rejected: loopback, private, link-local (incl. cloud metadata),
		// IPv4-mapped loopback, IPv6 unique-local, multicast, unspecified.
		{"127.0.0.1", true},
		{"::1", true},
		{"10.0.0.1", true},
		{"192.168.1.1", true},
		{"172.16.0.1", true},
		{"169.254.169.254", true},
		{"::ffff:127.0.0.1", true},
		{"fc00::1", true},
		{"fd12:3456::1", true},
		{"224.0.0.1", true},
		{"0.0.0.0", true},
		{"fe80::1", true},
		{"ff02::1", true},
		{"::", true},
		// Public destinations pass.
		{"1.1.1.1", false},
		{"93.184.216.34", false},
		{"2606:4700::1111", false},
	}
	for _, tt := range tests {
		err := validateClipHost(tt.host)
		if tt.wantErr && err == nil {
			t.Errorf("validateClipHost(%q) = nil; want rejection", tt.host)
		}
		if !tt.wantErr && err != nil {
			t.Errorf("validateClipHost(%q) = %v; want nil", tt.host, err)
		}
	}
}

func TestFetchPage_AllowPrivate(t *testing.T) {
	html, err := os.ReadFile(filepath.Join("testdata", "clip_article.html"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	// The handler runs on the server goroutine; report the user agent
	// through a channel so the test never races on shared state.
	uaCh := make(chan string, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uaCh <- r.Header.Get("User-Agent")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(html)
	}))
	defer srv.Close()

	t.Setenv("CLIP_ALLOW_PRIVATE", "1")

	c := NewDefuddleClipper("unused") // fetchPage never execs the script
	body, err := c.fetchPage(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("fetchPage() = %v; want nil", err)
	}
	if !bytes.Equal(body, html) {
		t.Fatalf("fetched body (%d bytes) does not match the fixture (%d bytes)", len(body), len(html))
	}
	if ua := <-uaCh; !strings.Contains(ua, "Mozilla") {
		t.Errorf("request User-Agent = %q; want a browser-like value", ua)
	}
}

func TestFetchPage_SizeCap(t *testing.T) {
	// The handler streams forever; only the client-side size cap can end
	// this exchange. A read-everything-then-check implementation would
	// hang until the fetch timeout instead of stopping at 5 MB.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		chunk := make([]byte, 128<<10)
		for {
			if _, err := w.Write(chunk); err != nil {
				return // client hung up: expected once the cap trips
			}
		}
	}))
	defer srv.Close()

	t.Setenv("CLIP_ALLOW_PRIVATE", "1")

	c := NewDefuddleClipper("")
	start := time.Now()
	body, err := c.fetchPage(context.Background(), srv.URL)
	elapsed := time.Since(start)
	if !errors.Is(err, ErrClipFetch) {
		t.Fatalf("fetchPage() = %v; want ErrClipFetch", err)
	}
	if body != nil {
		t.Errorf("fetchPage() returned %d bytes; want nil when the cap trips", len(body))
	}
	if elapsed > 5*time.Second {
		t.Fatalf("size cap did not stop the read early: took %s", elapsed)
	}
}

func TestFetchPage_RedirectsRejected(t *testing.T) {
	t.Run("private redirect target", func(t *testing.T) {
		t.Setenv("CLIP_ALLOW_PRIVATE", "0") // force unset semantics, ambient-proof

		dialed := make(chan struct{}, 1)
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			dialed <- struct{}{}
			http.Redirect(w, r, "http://127.0.0.1:1/x", http.StatusFound)
		}))
		defer srv.Close()

		c := NewDefuddleClipper("")
		_, err := c.fetchPage(context.Background(), srv.URL)
		if !errors.Is(err, ErrClipFetch) {
			t.Fatalf("fetchPage() = %v; want ErrClipFetch", err)
		}
		select {
		case <-dialed:
			t.Fatal("the private-address guard must reject before any dial")
		default:
		}
	})

	t.Run("scheme revalidated under allow-private", func(t *testing.T) {
		// The bypass reaches the loopback origin, but scheme rules always
		// apply: the ftp hop must die inside CheckRedirect.
		t.Setenv("CLIP_ALLOW_PRIVATE", "1")

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "ftp://127.0.0.1/pub/x", http.StatusFound)
		}))
		defer srv.Close()

		c := NewDefuddleClipper("")
		_, err := c.fetchPage(context.Background(), srv.URL)
		if !errors.Is(err, ErrClipFetch) {
			t.Fatalf("fetchPage() = %v; want ErrClipFetch", err)
		}
		if !errors.Is(err, ErrClipInvalidURL) {
			t.Fatalf("fetchPage() = %v; want the redirect re-validation to surface ErrClipInvalidURL", err)
		}
	})

	t.Run("guard rejects private redirect hop", func(t *testing.T) {
		t.Setenv("CLIP_ALLOW_PRIVATE", "0")

		req := &http.Request{URL: &url.URL{Scheme: "http", Host: "127.0.0.1:1", Path: "/x"}}
		if err := checkClipRedirect(context.Background(), req, nil); !errors.Is(err, ErrClipFetch) {
			t.Fatalf("checkClipRedirect() = %v; want ErrClipFetch", err)
		}
	})

	t.Run("guard rejects redirect chains over the cap", func(t *testing.T) {
		t.Setenv("CLIP_ALLOW_PRIVATE", "1")

		req := &http.Request{URL: &url.URL{Scheme: "http", Host: "1.1.1.1", Path: "/x"}}
		via := make([]*http.Request, maxClipRedirects)
		if err := checkClipRedirect(context.Background(), req, via); !errors.Is(err, ErrClipFetch) {
			t.Fatalf("checkClipRedirect() = %v; want ErrClipFetch", err)
		}
	})

	t.Run("allow-private lets the guard pass private hops", func(t *testing.T) {
		t.Setenv("CLIP_ALLOW_PRIVATE", "1")

		req := &http.Request{URL: &url.URL{Scheme: "http", Host: "127.0.0.1:1", Path: "/x"}}
		if err := checkClipRedirect(context.Background(), req, nil); err != nil {
			t.Fatalf("checkClipRedirect() = %v; want nil under allow-private", err)
		}
	})
}

func TestFetchPage_Non2xx(t *testing.T) {
	t.Setenv("CLIP_ALLOW_PRIVATE", "1")

	tests := []struct {
		name   string
		status int
	}{
		// 301 without a Location header is terminal: the client returns
		// the response instead of following anything.
		{"301 without location is terminal", http.StatusMovedPermanently},
		{"404 not found", http.StatusNotFound},
		{"500 server error", http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
			}))
			defer srv.Close()

			c := NewDefuddleClipper("")
			_, err := c.fetchPage(context.Background(), srv.URL)
			if !errors.Is(err, ErrClipFetch) {
				t.Fatalf("fetchPage() with status %d = %v; want ErrClipFetch", tt.status, err)
			}
		})
	}
}

func TestFetchPage_SizeBoundary(t *testing.T) {
	// Pins the exact cap: maxClipBodyBytes bytes succeed, one more fails.
	t.Setenv("CLIP_ALLOW_PRIVATE", "1")

	tests := []struct {
		name    string
		size    int
		wantErr bool
	}{
		{"exactly the cap succeeds", maxClipBodyBytes, false},
		{"one byte over the cap fails", maxClipBodyBytes + 1, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				chunk := make([]byte, 128<<10)
				for remaining := tt.size; remaining > 0; {
					n := min(len(chunk), remaining)
					if _, err := w.Write(chunk[:n]); err != nil {
						return // client hung up: cap tripped, expected
					}
					remaining -= n
				}
			}))
			defer srv.Close()

			c := NewDefuddleClipper("")
			body, err := c.fetchPage(context.Background(), srv.URL)
			if tt.wantErr {
				if !errors.Is(err, ErrClipFetch) {
					t.Fatalf("fetchPage(%d bytes) = %v; want ErrClipFetch", tt.size, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("fetchPage(%d bytes) = %v; want nil", tt.size, err)
			}
			if len(body) != maxClipBodyBytes {
				t.Fatalf("body length = %d; want exactly the %d byte cap", len(body), maxClipBodyBytes)
			}
		})
	}
}

func TestFetchPage_RedirectChainLimit(t *testing.T) {
	// End-to-end: an endless chain must die at the redirect cap. The cap
	// counts requests (original plus followed redirects), so a chain that
	// would need a 6th request must never get one.
	t.Setenv("CLIP_ALLOW_PRIVATE", "1")

	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		http.Redirect(w, r, "/hop/next", http.StatusFound)
	}))
	defer srv.Close()

	c := NewDefuddleClipper("")
	_, err := c.fetchPage(context.Background(), srv.URL)
	if !errors.Is(err, ErrClipFetch) {
		t.Fatalf("fetchPage() = %v; want ErrClipFetch", err)
	}
	if got := hits.Load(); got != maxClipRedirects {
		t.Fatalf("server saw %d requests; want exactly %d (original + %d redirects) before the client refused the next hop",
			got, maxClipRedirects, maxClipRedirects-1)
	}
}

func TestFetchPage_HostileRedirectLocation(t *testing.T) {
	// Regression: a 302 whose Location is too malformed to parse makes
	// the stdlib client echo the entire header in its error text (a 20k
	// rune Location measured at 40k runes end to end). The clipper must
	// bound whatever it wraps.
	t.Setenv("CLIP_ALLOW_PRIVATE", "1")

	hostile := "http://example.com/" + strings.Repeat("a", 20000) + "%zz"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Header set directly: a hostile origin writes whatever it wants.
		w.Header().Set("Location", hostile)
		w.WriteHeader(http.StatusFound)
	}))
	defer srv.Close()

	c := NewDefuddleClipper("")
	_, err := c.fetchPage(context.Background(), srv.URL)
	if !errors.Is(err, ErrClipFetch) {
		t.Fatalf("fetchPage() = %v; want ErrClipFetch", err)
	}
	if n := len([]rune(err.Error())); n > 400 {
		t.Fatalf("fetch error is %d runes; hostile redirect detail must stay bounded (< ~400)", n)
	}
}

func TestFetchPage_NoConnectionReuse(t *testing.T) {
	// A pooled connection to the same host would skip the guarded dial on
	// the second request. Every clip fetch must dial fresh so the dial
	// guard runs per request.
	t.Setenv("CLIP_ALLOW_PRIVATE", "1")

	var conns atomic.Int32
	srv := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	srv.Listener = &countingListener{Listener: ln, conns: &conns}
	srv.Start()
	defer srv.Close()

	c := NewDefuddleClipper("")
	for i := 0; i < 2; i++ {
		if _, err := c.fetchPage(context.Background(), srv.URL); err != nil {
			t.Fatalf("fetch %d: %v", i+1, err)
		}
	}
	if got := conns.Load(); got != 2 {
		t.Fatalf("server accepted %d connections for 2 fetches; want 2 (each request must dial through the guard)", got)
	}
}

func TestBoundedClipDetail(t *testing.T) {
	hostile := strings.Repeat("x", 20000)
	got := boundedClipDetail(hostile)
	if n := len([]rune(got)); n > maxClipDetailRunes+len("...") {
		t.Fatalf("boundedClipDetail() = %d runes; want at most %d plus an ellipsis", n, maxClipDetailRunes)
	}
	if !strings.HasSuffix(got, "...") {
		t.Fatalf("boundedClipDetail() = %q; want a trailing ellipsis on truncation", got)
	}
	if got := boundedClipDetail(" a\n\t b\r\n "); got != "a b" {
		t.Fatalf("boundedClipDetail() = %q; want whitespace collapsed", got)
	}
}

func TestRunScript_HostileErrorsBounded(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skipf("node not found in PATH (script failure tests need node): %v", err)
	}

	// Fake extractor scripts in a temp dir exercise the failure paths of
	// runScript without the real bundle: hostile {"error":...} stdout and
	// hostile stderr, alone and combined.
	tests := []struct {
		name            string
		script          string
		wantStderr      bool
		wantMaxErrRunes int
	}{
		{
			name:            "hostile extractor error",
			script:          `process.stdout.write(JSON.stringify({error: "E" + "x".repeat(20000)}))`,
			wantMaxErrRunes: 400,
		},
		{
			name:            "hostile stderr excerpt",
			script:          `process.stderr.write("S" + "y".repeat(20000)); process.exit(3)`,
			wantStderr:      true,
			wantMaxErrRunes: 400,
		},
		{
			name:            "hostile error and stderr together",
			script:          `process.stdout.write(JSON.stringify({error: "E" + "x".repeat(20000)})); process.stderr.write("S" + "y".repeat(20000)); process.exit(1)`,
			wantStderr:      true,
			wantMaxErrRunes: 700, // both excerpts are capped, but two caps add up
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "clip.cjs")
			if err := os.WriteFile(path, []byte(tt.script), 0o644); err != nil {
				t.Fatalf("write fake script: %v", err)
			}

			c := NewDefuddleClipper(path)
			_, err := c.runScript(context.Background(), "https://example.com/a", []byte("<html></html>"))
			if !errors.Is(err, ErrClipExtraction) {
				t.Fatalf("runScript() = %v; want ErrClipExtraction", err)
			}
			if n := len([]rune(err.Error())); n > tt.wantMaxErrRunes {
				t.Fatalf("error is %d runes; hostile extractor detail must stay bounded (< %d)", n, tt.wantMaxErrRunes)
			}
			if tt.wantStderr && !strings.Contains(err.Error(), "(stderr: ") {
				t.Fatalf("error = %q; want a bounded stderr excerpt", err)
			}
		})
	}
}

// countingListener wraps a listener to count accepted connections, used
// to prove each clip fetch dials fresh instead of reusing a pooled one.
type countingListener struct {
	net.Listener
	conns *atomic.Int32
}

func (l *countingListener) Accept() (net.Conn, error) {
	conn, err := l.Listener.Accept()
	if err == nil {
		l.conns.Add(1)
	}
	return conn, err
}

func TestDefuddleClipper_Extract(t *testing.T) {
	bundle := clipBundlePath(t)

	html, err := os.ReadFile(filepath.Join("testdata", "clip_article.html"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(html)
	}))
	defer srv.Close()

	t.Setenv("CLIP_ALLOW_PRIVATE", "1")

	srvURL, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatalf("parse server url: %v", err)
	}

	c := NewDefuddleClipper(bundle)
	page, err := c.Extract(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("Extract() = %v; want nil", err)
	}

	if page.Title != "Understanding SQLite WAL Mode" {
		t.Errorf("Title = %q", page.Title)
	}
	if page.Byline != "Jane Doe" {
		t.Errorf("Byline = %q", page.Byline)
	}
	if page.Published != "2024-03-11" {
		t.Errorf("Published = %q; want 2024-03-11", page.Published)
	}
	if page.SiteName != "Hube Engineering" {
		t.Errorf("SiteName = %q", page.SiteName)
	}
	if page.Domain != srvURL.Hostname() {
		t.Errorf("Domain = %q; want %q", page.Domain, srvURL.Hostname())
	}
	if page.WordCount != 182 {
		t.Errorf("WordCount = %d; want 182", page.WordCount)
	}
	for _, marker := range []string{
		"**wal file**", // bold
		"[official WAL documentation](https://www.sqlite.org/wal.html)", // link
		"`PRAGMA journal_mode=WAL`",                                     // inline code
		"\t- ",                                                          // TAB-indented nested list item
	} {
		if !strings.Contains(page.ContentMarkdown, marker) {
			t.Errorf("ContentMarkdown missing marker %q\nmarkdown:\n%s", marker, page.ContentMarkdown)
		}
	}
}

func TestDefuddleClipper_EmptyPage(t *testing.T) {
	bundle := clipBundlePath(t)

	html, err := os.ReadFile(filepath.Join("testdata", "clip_empty.html"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(html)
	}))
	defer srv.Close()

	t.Setenv("CLIP_ALLOW_PRIVATE", "1")

	c := NewDefuddleClipper(bundle)
	_, err = c.Extract(context.Background(), srv.URL)
	if !errors.Is(err, ErrClipExtraction) {
		t.Fatalf("Extract() = %v; want ErrClipExtraction", err)
	}
}

func TestDefuddleClipper_Extract_PrivateRejected(t *testing.T) {
	// The URL shape is valid, so guard order matters: validation passes,
	// then the private-address guard rejects the loopback target with a
	// fetch error before any dial.
	dialed := make(chan struct{}, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dialed <- struct{}{}
	}))
	defer srv.Close()

	t.Setenv("CLIP_ALLOW_PRIVATE", "0")

	c := NewDefuddleClipper("unused")
	_, err := c.Extract(context.Background(), srv.URL)
	if !errors.Is(err, ErrClipFetch) {
		t.Fatalf("Extract() = %v; want ErrClipFetch", err)
	}
	if errors.Is(err, ErrClipInvalidURL) {
		t.Fatalf("Extract() = %v; the URL is valid, so only the private-address guard may fire", err)
	}
	select {
	case <-dialed:
		t.Fatal("the guard must reject before any dial")
	default:
	}
}
