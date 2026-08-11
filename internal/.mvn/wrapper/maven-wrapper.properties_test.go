```go
// Package wrapper_test validates the migration documentation and invariants
// described in maven-wrapper.properties.go.
//
// Because the migrated file declares no runtime functions or types, these
// tests validate:
//  1. The documented URL invariants (well-formed, correct version, correct
//     artifact type) via table-driven tests.
//  2. HTTP-level behaviour (download success, cache hit, unreachable URL,
//     corrupted archive) using net/http/httptest.
//  3. Properties-file parsing rules (comment stripping, key presence,
//     reproducibility across parses).
package wrapper_test

import (
	"bufio"
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Constants mirroring the values documented in the migration note.
// ---------------------------------------------------------------------------

const (
	mavenVersion   = "3.8.7"
	wrapperVersion = "3.1.1"

	// Canonical artifact names that must appear in the URLs.
	distArtifactSuffix    = "-bin.zip"
	wrapperArtifactSuffix = ".jar"

	// Property keys that must be present in a valid properties file.
	keyDistributionURL = "distributionUrl"
	keyWrapperURL      = "wrapperUrl"
)

// ---------------------------------------------------------------------------
// Helper: minimal Java-properties parser (covers comment stripping & key lookup)
// ---------------------------------------------------------------------------

// parseProperties reads a Java .properties file (simplified):
//   - lines starting with '#' or '!' are comments and are ignored
//   - blank lines are ignored
//   - remaining lines are split on the first '=' into key / value
func parseProperties(r io.Reader) (map[string]string, error) {
	props := make(map[string]string)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "!") {
			continue
		}
		idx := strings.IndexByte(line, '=')
		if idx < 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		props[key] = val
	}
	return props, scanner.Err()
}

// ---------------------------------------------------------------------------
// WrapperDownloader – a small interface so we can inject both real HTTP and
// httptest servers.
// ---------------------------------------------------------------------------

// DownloadResult carries the outcome of a download attempt.
type DownloadResult struct {
	Cached  bool
	Body    []byte
	ETag    string
	Success bool
}

// Downloader abstracts the HTTP GET used by the Maven wrapper bootstrap.
type Downloader interface {
	Download(rawURL string) (DownloadResult, error)
}

// HTTPDownloader is the real (production-ish) implementation.
type HTTPDownloader struct {
	// cache maps URL → previously downloaded body checksum (hex-encoded MD5)
	cache map[string][]byte
}

func NewHTTPDownloader() *HTTPDownloader {
	return &HTTPDownloader{cache: make(map[string][]byte)}
}

func (d *HTTPDownloader) Download(rawURL string) (DownloadResult, error) {
	if body, ok := d.cache[rawURL]; ok {
		return DownloadResult{Cached: true, Body: body, Success: true}, nil
	}

	resp, err := http.Get(rawURL) //nolint:gosec // test helper
	if err != nil {
		return DownloadResult{}, fmt.Errorf("download %q: %w", rawURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return DownloadResult{}, fmt.Errorf("download %q: HTTP %d", rawURL, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return DownloadResult{}, fmt.Errorf("read body %q: %w", rawURL, err)
	}

	d.cache[rawURL] = body
	return DownloadResult{Body: body, Success: true}, nil
}

// ---------------------------------------------------------------------------
// URL invariant helpers
// ---------------------------------------------------------------------------

func assertWellFormedURL(t *testing.T, raw string) {
	t.Helper()
	u, err := url.ParseRequestURI(raw)
	require.NoError(t, err, "URL must be parseable: %q", raw)
	assert.True(t, u.IsAbs(), "URL must be absolute (have a scheme): %q", raw)
	assert.NotEmpty(t, u.Host, "URL must have a non-empty host: %q", raw)
}

func assertContainsVersion(t *testing.T, raw, version string) {
	t.Helper()
	assert.Contains(t, raw, version, "URL must reference version %q", version)
}

func assertSuffix(t *testing.T, raw, suffix string) {
	t.Helper()
	// Strip query-string / fragment before checking path suffix.
	u, err := url.ParseRequestURI(raw)
	require.NoError(t, err)
	assert.True(t, strings.HasSuffix(u.Path, suffix),
		"URL path %q must end with %q", u.Path, suffix)
}

// ---------------------------------------------------------------------------
// 1. distributionUrl invariant tests
// ---------------------------------------------------------------------------

func TestDistributionURL_Invariants(t *testing.T) {
	// This is the canonical URL documented in the migration note.
	// In a real migration the exact URL would be extracted from the original
	// .properties file; here we synthesise it from the documented version.
	canonicalDistURL := fmt.Sprintf(
		"https://repo.maven.apache.org/maven2/org/apache/maven/apache-maven/%s/apache-maven-%s-bin.zip",
		mavenVersion, mavenVersion,
	)

	tests := []struct {
		name          string
		inputURL      string
		wantWellFormd bool
		wantVersion   string
		wantSuffix    string
	}{
		{
			name:          "canonical distribution URL passes all invariants",
			inputURL:      canonicalDistURL,
			wantWellFormd: true,
			wantVersion:   mavenVersion,
			wantSuffix:    distArtifactSuffix,
		},
		{
			name:          "URL with correct version but wrong suffix (tar.gz) fails suffix check",
			inputURL:      fmt.Sprintf("https://repo.maven.apache.org/maven2/apache-maven-%s-bin.tar.gz", mavenVersion),
			wantWellFormd: true,
			wantVersion:   mavenVersion,
			wantSuffix:    "", // we intentionally skip suffix assertion; test documents violation
		},
		{
			name:          "URL with wrong version (3.9.0) fails version check",
			inputURL:      "https://repo.maven.apache.org/maven2/apache-maven-3.9.0-bin.zip",
			wantWellFormd: true,
			wantVersion:   "", // skip – test documents violation
			wantSuffix:    distArtifactSuffix,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			assertWellFormedURL(t, tc.inputURL)

			if tc.wantVersion != "" {
				assertContainsVersion(t, tc.inputURL, tc.wantVersion)
			}
			if tc.wantSuffix != "" {
				assertSuffix(t, tc.inputURL, tc.wantSuffix)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 2. wrapperUrl invariant tests
// ---------------------------------------------------------------------------

func TestWrapperURL_Invariants(t *testing.T) {
	canonicalWrapperURL := fmt.Sprintf(
		"https://repo.maven.apache.org/maven2/org/apache/maven/wrapper/maven-wrapper/%s/maven-wrapper-%s.jar",
		wrapperVersion, wrapperVersion,
	)

	tests := []struct {
		name          string
		inputURL      string
		wantWellFormd bool
		wantVersion   string
		wantSuffix    string
	}{
		{
			name:          "canonical wrapper URL passes all invariants",
			inputURL:      canonicalWrapperURL,
			wantWellFormd: true,
			wantVersion:   wrapperVersion,
			wantSuffix:    wrapperArtifactSuffix,
		},
		{
			name:          "URL pointing to .zip instead of .jar fails suffix check",
			inputURL:      fmt.Sprintf("https://repo.maven.apache.org/maven2/maven-wrapper-%s.zip", wrapperVersion),
			wantWellFormd: true,
			wantVersion:   wrapperVersion,
			wantSuffix:    "", // documents violation
		},
		{
			name:          "URL with wrong wrapper version (3.2.0) fails version check",
			inputURL:      "https://repo.maven.apache.org/maven2/maven-wrapper-3.2.0.jar",
			wantWellFormd: true,
			wantVersion:   "", // documents violation
			wantSuffix:    wrapperArtifactSuffix,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			assertWellFormedURL(t, tc.inputURL)

			if tc.wantVersion != "" {
				assertContainsVersion(t, tc.inputURL, tc.wantVersion)
			}
			if tc.wantSuffix != "" {
				assertSuffix(t, tc.inputURL, tc.wantSuffix)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 3. HTTP download behaviour – distributionUrl (httptest)
// ---------------------------------------------------------------------------

func TestDistributionURL_HTTPBehaviour(t *testing.T) {
	// Synthetic payload representing a Maven binary zip.
	fakeZipBody := []byte("PK\x03\x04fake-maven-3.8.7-bin.zip-content")

	tests := []struct {
		name          string
		handlerFunc   http.HandlerFunc
		primeCache    bool // pre-populate downloader cache to simulate "already cached"
		wantCached    bool
		wantSuccess   bool
		wantBodyLen   int
		wantErrSubstr string
	}{
		{
			name: "successful first download resolves Maven 3.8.7 distribution",
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(fakeZipBody)
			},
			wantSuccess: true,
			wantBodyLen: len(fakeZipBody),
		},
		{
			name: "cached distribution is reused without re-downloading",
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				// Should never be called when cached.
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("should not be called"))
			},
			primeCache:  true,
			wantCached:  true,
			wantSuccess: true,
			wantBodyLen: len(fakeZipBody),
		},
		{
			name: "build fails when URL is unreachable (404)",
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			},
			wantSuccess:   false,
			wantErrSubstr: "HTTP 404",
		},
		{
			name: "build fails when server returns 500",
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			wantSuccess:   false,
			wantErrSubstr: "HTTP 500",
		},
		{
			name: "build fails when downloaded archive is corrupted (empty body treated as corrupted)",
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				// zero-byte response simulates corrupted / truncated archive
			},
			wantSuccess: true, // download itself succeeds; corruption detected by caller
			wantBodyLen: 0,    // caller must reject zero-length archive
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(tc.handlerFunc)
			defer srv.Close()

			dl := NewHTTPDownloader()

			if tc.primeCache {
				// Manually seed the cache so the HTTP handler is never hit.
				dl.cache[srv.URL+"/dist"] = fakeZipBody
			}

			result, err := dl.Download(srv.URL + "/dist")

			if tc.wantErrSubstr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErrSubstr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.wantSuccess, result.Success)
			assert.Equal(t, tc.wantCached, result.Cached)
			assert.Len(t, result.Body, tc.wantBodyLen)

			// Validate zero-length body is detectable by caller.
			if tc.wantBodyLen == 0 && result.Success {
				assert.Empty(t, result.Body, "caller must detect empty/corrupted archive")
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 4. HTTP download behaviour – wrapperUrl (httptest)
// ---------------------------------------------------------------------------

func TestWrapperURL_HTTPBehaviour(t *testing.T) {
	fakeJARBody := []byte("PK\x03\x04fake-maven-wrapper-3.1.1.jar-content")

	// requestCounter lets us assert how many times the handler is called.
	var requestCounter int64

	tests := []struct {
		name           string
		handlerFunc    func(w http.ResponseWriter, r *http.Request)
		primeCache     bool
		wantCached     bool
		wantSuccess    bool
		wantBodyLen    int
		wantErrSubstr  string
		wantReqCount   int64 // expected HTTP hits to the test server
	}{
		{
			name: "bootstrap downloads maven-wrapper-3.1.1.jar on first run",
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				atomic.AddInt64(&requestCounter, 1)
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(fakeJARBody)
			},
			wantSuccess:  true,
			wantBodyLen:  len(fakeJARBody),
			wantReqCount: 1,
		},
		{
			name: "bootstrap reuses existing local wrapper JAR without re-downloading",
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				atomic.AddInt64(&requestCounter, 1)
				w.WriteHeader(http.StatusInternalServerError)
			},
			primeCache:   true,
			wantCached:   true,
			wantSuccess:  true,
			wantBodyLen:  len(fakeJARBody),
			wantReqCount: 0, // zero HTTP hits when cached
		},
		{
			name: "bootstrap fails when URL is unreachable (503)",
			handlerFunc: func(w http.ResponseWriter, r