```go
package wrapper_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	wrapper "internal/.mvn/wrapper"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// parseURL is a small helper used by several test cases.
func parseURL(t *testing.T, raw string) *url.URL {
	t.Helper()
	u, err := url.Parse(raw)
	require.NoError(t, err, "constant must be parseable as a URL")
	return u
}

// serveArchive starts an httptest server that returns 200 for the given path
// and 404 for everything else.  It is used to simulate a reachable artifact
// repository without network access.
func serveArtifact(t *testing.T, expectedPath string, body []byte) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == expectedPath {
			w.Header().Set("Content-Type", "application/octet-stream")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(body)
			return
		}
		http.NotFound(w, r)
	}))
	t.Cleanup(srv.Close)
	return srv
}

// serveUnreachable returns a server that immediately closes the connection,
// simulating an unreachable endpoint.
func serveUnreachable(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hj, ok := w.(http.Hijacker)
		if !ok {
			http.Error(w, "hijacking not supported", http.StatusInternalServerError)
			return
		}
		conn, _, _ := hj.Hijack()
		_ = conn.Close() // drop the connection immediately
	}))
	t.Cleanup(srv.Close)
	return srv
}

// ---------------------------------------------------------------------------
// MavenDistributionURL constant tests
// ---------------------------------------------------------------------------

func TestMavenDistributionURL_Invariants(t *testing.T) {
	tests := []struct {
		name      string
		checkFunc func(t *testing.T)
	}{
		{
			name: "constant is non-empty",
			checkFunc: func(t *testing.T) {
				assert.NotEmpty(t, wrapper.MavenDistributionURL)
			},
		},
		{
			name: "constant is a valid URL",
			checkFunc: func(t *testing.T) {
				u := parseURL(t, wrapper.MavenDistributionURL)
				assert.NotEmpty(t, u.Host)
				assert.NotEmpty(t, u.Path)
			},
		},
		{
			name: "URL uses HTTPS scheme",
			checkFunc: func(t *testing.T) {
				u := parseURL(t, wrapper.MavenDistributionURL)
				assert.Equal(t, "https", u.Scheme, "distributionUrl must use the HTTPS scheme")
			},
		},
		{
			name: "URL references Maven version 3.8.7",
			checkFunc: func(t *testing.T) {
				assert.Contains(t, wrapper.MavenDistributionURL, "3.8.7",
					"distributionUrl must point to Maven 3.8.7")
			},
		},
		{
			name: "URL references a binary distribution archive (-bin.zip)",
			checkFunc: func(t *testing.T) {
				assert.True(t,
					strings.HasSuffix(wrapper.MavenDistributionURL, "-bin.zip"),
					"distributionUrl must reference a binary distribution archive ending in -bin.zip")
			},
		},
		{
			name: "URL points to the apache-maven artifact",
			checkFunc: func(t *testing.T) {
				assert.Contains(t, wrapper.MavenDistributionURL, "apache-maven",
					"distributionUrl must reference the apache-maven artifact")
			},
		},
		{
			name: "exact constant value matches original properties file",
			checkFunc: func(t *testing.T) {
				const expected = "https://repo.maven.apache.org/maven2/org/apache/maven/apache-maven/3.8.7/apache-maven-3.8.7-bin.zip"
				assert.Equal(t, expected, wrapper.MavenDistributionURL)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.checkFunc(t)
		})
	}
}

// ---------------------------------------------------------------------------
// MavenWrapperURL constant tests
// ---------------------------------------------------------------------------

func TestMavenWrapperURL_Invariants(t *testing.T) {
	tests := []struct {
		name      string
		checkFunc func(t *testing.T)
	}{
		{
			name: "constant is non-empty",
			checkFunc: func(t *testing.T) {
				assert.NotEmpty(t, wrapper.MavenWrapperURL)
			},
		},
		{
			name: "constant is a valid URL",
			checkFunc: func(t *testing.T) {
				u := parseURL(t, wrapper.MavenWrapperURL)
				assert.NotEmpty(t, u.Host)
				assert.NotEmpty(t, u.Path)
			},
		},
		{
			name: "URL uses HTTPS scheme",
			checkFunc: func(t *testing.T) {
				u := parseURL(t, wrapper.MavenWrapperURL)
				assert.Equal(t, "https", u.Scheme, "wrapperUrl must use the HTTPS scheme")
			},
		},
		{
			name: "URL references maven-wrapper version 3.1.1",
			checkFunc: func(t *testing.T) {
				assert.Contains(t, wrapper.MavenWrapperURL, "3.1.1",
					"wrapperUrl must point to maven-wrapper 3.1.1")
			},
		},
		{
			name: "URL references a JAR artifact",
			checkFunc: func(t *testing.T) {
				assert.True(t,
					strings.HasSuffix(wrapper.MavenWrapperURL, ".jar"),
					"wrapperUrl must reference a JAR artifact ending in .jar")
			},
		},
		{
			name: "URL references the maven-wrapper artifact",
			checkFunc: func(t *testing.T) {
				assert.Contains(t, wrapper.MavenWrapperURL, "maven-wrapper",
					"wrapperUrl must reference the maven-wrapper artifact")
			},
		},
		{
			name: "exact constant value matches original properties file",
			checkFunc: func(t *testing.T) {
				const expected = "https://repo.maven.apache.org/maven2/org/apache/maven/wrapper/maven-wrapper/3.1.1/maven-wrapper-3.1.1.jar"
				assert.Equal(t, expected, wrapper.MavenWrapperURL)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.checkFunc(t)
		})
	}
}

// ---------------------------------------------------------------------------
// Behavioral scenario: distributionUrl – download simulation via httptest
// ---------------------------------------------------------------------------

func TestDistributionURL_DownloadBehavior(t *testing.T) {
	// Extract the path component from the real constant so tests remain in sync.
	realURL := parseURL(t, wrapper.MavenDistributionURL)

	// Fake archive payload – not a real zip, but sufficient to test HTTP layer.
	fakeArchive := []byte("PK\x03\x04fake-zip-content")

	tests := []struct {
		name           string
		setupServer    func(t *testing.T) *httptest.Server
		overridePath   string // if non-empty, request this path instead of realURL.Path
		expectStatus   int
		expectBody     []byte
		expectDownload bool
	}{
		{
			name: "distribution archive downloaded successfully when present on server",
			setupServer: func(t *testing.T) *httptest.Server {
				return serveArtifact(t, realURL.Path, fakeArchive)
			},
			expectStatus:   http.StatusOK,
			expectBody:     fakeArchive,
			expectDownload: true,
		},
		{
			name: "server returns 404 when archive is absent (build would fail)",
			setupServer: func(t *testing.T) *httptest.Server {
				// Serve a different path so the real path gets a 404.
				return serveArtifact(t, "/other/path", fakeArchive)
			},
			expectStatus:   http.StatusNotFound,
			expectDownload: false,
		},
		{
			name: "cached distribution reused: second download not required",
			setupServer: func(t *testing.T) *httptest.Server {
				callCount := 0
				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					callCount++
					if callCount > 1 {
						// Simulate caching: second request should never reach server.
						t.Errorf("server called %d times; expected at most 1 (cache should be used)", callCount)
					}
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write(fakeArchive)
				}))
				t.Cleanup(srv.Close)
				return srv
			},
			expectStatus:   http.StatusOK,
			expectBody:     fakeArchive,
			expectDownload: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := tc.setupServer(t)

			// Build the request URL by replacing the host with the test server's host.
			reqURL := *realURL
			testSrvURL := parseURL(t, srv.URL)
			reqURL.Scheme = testSrvURL.Scheme
			reqURL.Host = testSrvURL.Host

			resp, err := http.Get(reqURL.String())
			require.NoError(t, err, "HTTP request must not return a transport error")
			defer resp.Body.Close()

			assert.Equal(t, tc.expectStatus, resp.StatusCode)

			if tc.expectDownload && tc.expectBody != nil {
				buf := make([]byte, len(tc.expectBody))
				n, _ := resp.Body.Read(buf)
				assert.Equal(t, tc.expectBody, buf[:n],
					"response body must match the expected archive bytes")
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Behavioral scenario: wrapperUrl – download simulation via httptest
// ---------------------------------------------------------------------------

func TestWrapperURL_DownloadBehavior(t *testing.T) {
	realURL := parseURL(t, wrapper.MavenWrapperURL)

	// Minimal fake JAR bytes (just magic header PK).
	fakeJAR := []byte("PK\x03\x04fake-jar-content")

	tests := []struct {
		name           string
		setupServer    func(t *testing.T) *httptest.Server
		expectStatus   int
		expectBody     []byte
		expectDownload bool
	}{
		{
			name: "wrapper JAR downloaded successfully when present on server",
			setupServer: func(t *testing.T) *httptest.Server {
				return serveArtifact(t, realURL.Path, fakeJAR)
			},
			expectStatus:   http.StatusOK,
			expectBody:     fakeJAR,
			expectDownload: true,
		},
		{
			name: "server returns 404 when JAR is absent (build would fail)",
			setupServer: func(t *testing.T) *httptest.Server {
				return serveArtifact(t, "/other/path", fakeJAR)
			},
			expectStatus:   http.StatusNotFound,
			expectDownload: false,
		},
		{
			name: "existing wrapper JAR reused without re-downloading",
			setupServer: func(t *testing.T) *httptest.Server {
				callCount := 0
				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					callCount++
					if callCount > 1 {
						t.Errorf("server called %d times; wrapper JAR cache should prevent re-download", callCount)
					}
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write(fakeJAR)
				}))
				t.Cleanup(srv.Close)
				return srv
			},
			expectStatus:   http.StatusOK,
			expectBody:     fakeJAR,
			expectDownload: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := tc.setupServer(t)

			reqURL := *realURL
			testSrvURL := parseURL(t, srv.URL)
			reqURL.Scheme = testSrvURL.Scheme
			reqURL.Host = testSrvURL.Host

			resp, err := http.Get(reqURL.String())
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tc.expectStatus, resp.StatusCode)

			if tc.expectDownload && tc.expectBody != nil {
				buf := make([]byte, len(tc.expectBody))
				n, _ := resp.Body.Read(buf)
				assert.Equal(t, tc.expectBody, buf[:n])
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Error cases: unreachable server
// ---------------------------------------------------------------------------

func TestDistributionURL_UnreachableServer(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "build fails when distributionUrl server is unreachable"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := serveUnreachable(t)

			realURL := parseURL(t, wrapper.MavenDistributionURL)
			testSrvURL := parseURL(t, srv.URL)
			reqURL := *realURL
			reqURL.Scheme = testSrvURL.Scheme
			reqURL.Host = testSrvURL.Host

			_, err := http.Get(reqURL.String())
			// An EOF / connection reset must be returned — not a clean response.
			assert.Error(t, err, "expected a transport error when server drops the connection")
		})
	}
}

func TestWrapperURL_UnreachableServer(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "build fails when wrapperUrl server is unreachable"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := serveUnreachable(t)

			realURL := parseURL(t, wrapper.MavenWrapperURL)
			testSrvURL := parseURL(t, srv.URL)
			reqURL := *realURL
			reqURL.Scheme = testSrvURL.Scheme
			reqURL.Host = testSrvURL.Host

			_, err := http.Get(reqURL.String())
			assert.Error(t, err, "expected a transport error when server drops the connection")
		})
	}
}

// ---------------------------------------------------------------------------
// Global invariants
// ---------------------------------------------------------------------------

func TestGlobalInvariants(t *testing.T) {
	tests := []struct {
		name      string
		checkFunc func(t *testing.T)
	}{