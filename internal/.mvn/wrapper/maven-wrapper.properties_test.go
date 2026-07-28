```go
// Package wrapper contains tests documenting the deliberate migration decision
// to discard the Maven Wrapper configuration. Since the migrated file contains
// no runnable code, these tests validate the package-level documentation
// invariants and the structural decisions recorded in the migration notes.
//
// Because the original behavioral specs describe Maven Wrapper behaviour (a JVM
// build-tool concept with no Go counterpart), the tests below:
//   - Confirm the package compiles and is importable.
//   - Validate every invariant that can be expressed in Go (URL well-formedness,
//     version strings, scheme, host, archive extension).
//   - Use table-driven patterns and testify/assert throughout.
//   - Include an httptest-backed suite that simulates the download scenarios
//     described in the specs (reachable / unreachable / corrupt archive) so
//     that the intent of every behavioural scenario is exercised in Go terms.
package wrapper_test

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Constants that mirror the values from maven-wrapper.properties
// ---------------------------------------------------------------------------

const (
	distributionURL = "https://repo.maven.apache.org/maven2/org/apache/maven/apache-maven/3.8.7/apache-maven-3.8.7-bin.zip"
	wrapperURL      = "https://repo.maven.apache.org/maven2/org/apache/maven/wrapper/maven-wrapper/3.1.1/maven-wrapper-3.1.1.jar"

	mavenVersion   = "3.8.7"
	wrapperVersion = "3.1.1"
)

// ---------------------------------------------------------------------------
// Helpers / mock interfaces
// ---------------------------------------------------------------------------

// Downloader abstracts an HTTP GET so we can inject test servers.
type Downloader interface {
	Get(rawURL string) (*http.Response, error)
}

type httpDownloader struct{ client *http.Client }

func (d *httpDownloader) Get(rawURL string) (*http.Response, error) {
	return d.client.Get(rawURL)
}

// Cache simulates the local artefact cache used by the wrapper.
type Cache interface {
	Has(key string) bool
	Store(key string, data []byte)
	Load(key string) ([]byte, bool)
}

type inMemoryCache struct{ m map[string][]byte }

func newCache() *inMemoryCache { return &inMemoryCache{m: make(map[string][]byte)} }

func (c *inMemoryCache) Has(key string) bool        { _, ok := c.m[key]; return ok }
func (c *inMemoryCache) Store(key string, data []byte) { c.m[key] = data }
func (c *inMemoryCache) Load(key string) ([]byte, bool) {
	v, ok := c.m[key]
	return v, ok
}

// resolveArtefact downloads the artefact at rawURL via dl unless the cache
// already holds it, in which case the cached copy is returned.
func resolveArtefact(dl Downloader, cache Cache, rawURL string) ([]byte, error) {
	if cache.Has(rawURL) {
		data, _ := cache.Load(rawURL)
		return data, nil
	}
	resp, err := dl.Get(rawURL)
	if err != nil {
		return nil, fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading body: %w", err)
	}
	cache.Store(rawURL, data)
	return data, nil
}

// isValidZip returns true if data looks like a ZIP archive.
func isValidZip(data []byte) bool {
	_, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	return err == nil
}

// isValidJAR returns true if data looks like a ZIP (JAR is a ZIP).
func isValidJAR(data []byte) bool { return isValidZip(data) }

// makeZip produces a minimal valid ZIP archive.
func makeZip(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	f, err := w.Create("dummy.txt")
	require.NoError(t, err)
	_, err = f.Write([]byte("dummy"))
	require.NoError(t, err)
	require.NoError(t, w.Close())
	return buf.Bytes()
}

// ---------------------------------------------------------------------------
// 1. URL invariant tests (table-driven)
// ---------------------------------------------------------------------------

func TestDistributionURL_Invariants(t *testing.T) {
	t.Parallel()

	type invariantCase struct {
		name  string
		check func(t *testing.T, u *url.URL)
	}

	u, err := url.Parse(distributionURL)
	require.NoError(t, err, "distributionUrl must be parseable as a URL")

	cases := []invariantCase{
		{
			name: "must be well-formed",
			check: func(t *testing.T, u *url.URL) {
				assert.NotEmpty(t, u.Scheme, "scheme must not be empty")
				assert.NotEmpty(t, u.Host, "host must not be empty")
				assert.NotEmpty(t, u.Path, "path must not be empty")
			},
		},
		{
			name: "must use HTTPS scheme",
			check: func(t *testing.T, u *url.URL) {
				assert.Equal(t, "https", u.Scheme)
			},
		},
		{
			name: "must reference official Maven Central host",
			check: func(t *testing.T, u *url.URL) {
				assert.Equal(t, "repo.maven.apache.org", u.Host)
			},
		},
		{
			name: "must reference Maven version 3.8.7",
			check: func(t *testing.T, u *url.URL) {
				assert.Contains(t, u.Path, mavenVersion)
			},
		},
		{
			name: "must point to a binary distribution ZIP archive",
			check: func(t *testing.T, u *url.URL) {
				assert.True(t, strings.HasSuffix(u.Path, ".zip"),
					"path must end with .zip, got %s", u.Path)
				assert.Contains(t, u.Path, "-bin",
					"archive must be a binary distribution (-bin)")
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.check(t, u)
		})
	}
}

func TestWrapperURL_Invariants(t *testing.T) {
	t.Parallel()

	type invariantCase struct {
		name  string
		check func(t *testing.T, u *url.URL)
	}

	u, err := url.Parse(wrapperURL)
	require.NoError(t, err, "wrapperUrl must be parseable as a URL")

	cases := []invariantCase{
		{
			name: "must be well-formed",
			check: func(t *testing.T, u *url.URL) {
				assert.NotEmpty(t, u.Scheme)
				assert.NotEmpty(t, u.Host)
				assert.NotEmpty(t, u.Path)
			},
		},
		{
			name: "must use HTTPS scheme",
			check: func(t *testing.T, u *url.URL) {
				assert.Equal(t, "https", u.Scheme)
			},
		},
		{
			name: "must reference official Maven Central host",
			check: func(t *testing.T, u *url.URL) {
				assert.Equal(t, "repo.maven.apache.org", u.Host)
			},
		},
		{
			name: "must reference Maven Wrapper version 3.1.1",
			check: func(t *testing.T, u *url.URL) {
				assert.Contains(t, u.Path, wrapperVersion)
			},
		},
		{
			name: "must point to a JAR archive",
			check: func(t *testing.T, u *url.URL) {
				assert.True(t, strings.HasSuffix(u.Path, ".jar"),
					"path must end with .jar, got %s", u.Path)
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.check(t, u)
		})
	}
}

// ---------------------------------------------------------------------------
// 2. distributionUrl behavioural specs (httptest-backed, table-driven)
// ---------------------------------------------------------------------------

func TestDistributionURL_DownloadBehaviours(t *testing.T) {
	t.Parallel()
	validZip := makeZip(t)

	type tc struct {
		name            string
		handler         http.HandlerFunc
		primeCache      bool // put something in the cache before the call
		wantErr         bool
		wantErrContains string
		wantCached      bool
		validateData    func(t *testing.T, data []byte)
	}

	cases := []tc{
		{
			name: "downloads distribution when not cached and server is reachable",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(validZip)
			},
			primeCache:   false,
			wantErr:      false,
			wantCached:   true,
			validateData: func(t *testing.T, data []byte) { assert.True(t, isValidZip(data)) },
		},
		{
			name: "uses cached distribution without re-downloading",
			handler: func(w http.ResponseWriter, r *http.Request) {
				// If this is called the test should fail – but we track it.
				t.Error("server was called even though cache was primed")
				w.WriteHeader(http.StatusInternalServerError)
			},
			primeCache:   true,
			wantErr:      false,
			wantCached:   true,
			validateData: func(t *testing.T, data []byte) { assert.True(t, isValidZip(data)) },
		},
		{
			name: "fails when server is unreachable",
			handler: func(w http.ResponseWriter, r *http.Request) {
				// Server closes connection immediately.
				hj, ok := w.(http.Hijacker)
				if !ok {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				conn, _, _ := hj.Hijack()
				conn.Close()
			},
			primeCache:      false,
			wantErr:         true,
			wantErrContains: "download failed",
			wantCached:      false,
		},
		{
			name: "fails when server returns non-200",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte("not found"))
			},
			primeCache:      false,
			wantErr:         true,
			wantErrContains: "unexpected status",
			wantCached:      false,
		},
		{
			name: "fails when downloaded archive is corrupt",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("this is not a zip"))
			},
			primeCache: false,
			wantErr:    false, // download itself succeeds; validation is caller's job
			wantCached: true,
			validateData: func(t *testing.T, data []byte) {
				assert.False(t, isValidZip(data), "corrupt archive must not be a valid zip")
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(tc.handler)
			defer srv.Close()

			cache := newCache()
			targetURL := srv.URL + "/apache-maven-3.8.7-bin.zip"

			if tc.primeCache {
				cache.Store(targetURL, validZip)
			}

			dl := &httpDownloader{client: srv.Client()}
			data, err := resolveArtefact(dl, cache, targetURL)

			if tc.wantErr {
				require.Error(t, err)
				if tc.wantErrContains != "" {
					assert.Contains(t, err.Error(), tc.wantErrContains)
				}
				if !tc.wantCached {
					assert.False(t, cache.Has(targetURL))
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, data)

			if tc.wantCached {
				assert.True(t, cache.Has(targetURL))
			}
			if tc.validateData != nil {
				tc.validateData(t, data)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 3. wrapperUrl behavioural specs (httptest-backed, table-driven)
// ---------------------------------------------------------------------------

func TestWrapperURL_DownloadBehaviours(t *testing.T) {
	t.Parallel()
	validJAR := makeZip(t) // JARs are ZIPs

	type tc struct {
		name            string
		handler         http.HandlerFunc
		primeCache      bool
		wantErr         bool
		wantErrContains string
		wantCached      bool
		validateData    func(t *testing.T, data []byte)
	}

	cases := []tc{
		{
			name: "downloads wrapper JAR when not cached and server is reachable",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(validJAR)
			},
			primeCache:   false,
			wantErr:      false,
			wantCached:   true,
			validateData: func(t *testing.T, data []byte) { assert.True(t, isValidJAR(data)) },
		},
		{
			name: "uses cached wrapper JAR without re-downloading",
			handler: func(w http.ResponseWriter, r *http.Request) {
				t.Error("server was called even though cache was primed")
				w.WriteHeader(http.StatusInternalServerError)
			},
			primeCache:   true,
			wantErr:      false,
			wantCached:   true,
			validateData: func(t *testing.T, data []byte) { assert.True(t, isValidJAR(data)) },
		},
		{
			name: "bootstrap fails when server is unreachable",
			handler: func(w http.ResponseWriter, r *http.Request) {
				hj, ok := w.(http.Hijacker)
				if !ok {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				conn, _, _ := hj.Hijack()
				conn.Close()
			},
			primeCache:      false,
			wantErr:         true,
			wantErrContains: "download failed",
			wantCached:      false,
		},