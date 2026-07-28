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

	wrapper "smartcontact/internal/.mvn/wrapper"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// serveZip stands up an httptest.Server that returns a minimal, non-empty
// response body so tests can simulate a reachable distribution archive.
func serveZip(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/zip")
		w.WriteHeader(http.StatusOK)
		// Minimal non-zero body simulates a valid archive download.
		_, _ = w.Write([]byte("PK\x03\x04")) // zip local-file magic bytes
	}))
	t.Cleanup(srv.Close)
	return srv
}

// serveJAR stands up an httptest.Server that returns a minimal JAR (which is
// itself a zip file) so tests can simulate a reachable wrapper JAR.
func serveJAR(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/java-archive")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("PK\x03\x04")) // JAR / zip magic bytes
	}))
	t.Cleanup(srv.Close)
	return srv
}

// serve404 returns a server that always responds 404 Not Found.
func serve404(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	t.Cleanup(srv.Close)
	return srv
}

// serveCorrupt returns a server whose body starts with garbage bytes (not a
// valid zip/JAR).
func serveCorrupt(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("CORRUPT_GARBAGE_DATA"))
	}))
	t.Cleanup(srv.Close)
	return srv
}

// isWellFormedURL returns true when rawURL can be parsed and has both a non-
// empty scheme and host component.
func isWellFormedURL(rawURL string) bool {
	u, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return false
	}
	return u.Scheme != "" && u.Host != ""
}

// endsWithExtension returns true when rawURL's path ends with the given suffix
// (case-insensitive comparison on the last segment).
func endsWithExtension(rawURL, ext string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	return strings.HasSuffix(strings.ToLower(u.Path), strings.ToLower(ext))
}

// fetchURL performs a real HTTP GET against a URL and returns the response.
// It is used to exercise the httptest servers in download-simulation tests.
func fetchURL(t *testing.T, rawURL string) *http.Response {
	t.Helper()
	resp, err := http.Get(rawURL) //nolint:noctx
	require.NoError(t, err, "unexpected error performing GET %s", rawURL)
	t.Cleanup(func() { _ = resp.Body.Close() })
	return resp
}

// ---------------------------------------------------------------------------
// DistributionURL tests
// ---------------------------------------------------------------------------

func TestDistributionURL_WellFormed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		rawURL  string
		wantOK  bool
		wantMsg string
	}{
		{
			name:    "canonical constant is a well-formed https URL",
			rawURL:  wrapper.DistributionURL,
			wantOK:  true,
			wantMsg: "DistributionURL must be a well-formed URL with scheme and host",
		},
		{
			name:    "empty string is not well-formed",
			rawURL:  "",
			wantOK:  false,
			wantMsg: "empty string should not be a well-formed URL",
		},
		{
			name:    "relative path is not well-formed",
			rawURL:  "/maven/apache-maven-3.8.7-bin.zip",
			wantOK:  false,
			wantMsg: "relative path should not be a well-formed URL",
		},
		{
			name:    "missing host is not well-formed",
			rawURL:  "https:///apache-maven-3.8.7-bin.zip",
			wantOK:  false,
			wantMsg: "URL without host should not be well-formed",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := isWellFormedURL(tc.rawURL)
			assert.Equal(t, tc.wantOK, got, tc.wantMsg)
		})
	}
}

func TestDistributionURL_ReferencesZipArchive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		rawURL string
		want   bool
	}{
		{
			name:   "constant references a .zip archive",
			rawURL: wrapper.DistributionURL,
			want:   true,
		},
		{
			name:   "JAR URL does not reference a .zip",
			rawURL: wrapper.WrapperURL,
			want:   false,
		},
		{
			name:   "arbitrary non-zip URL",
			rawURL: "https://example.com/file.tar.gz",
			want:   false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := endsWithExtension(tc.rawURL, ".zip")
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestDistributionURL_ContainsMavenVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		rawURL         string
		expectedSubstr string
		want           bool
	}{
		{
			name:           "URL encodes Maven 3.8.7",
			rawURL:         wrapper.DistributionURL,
			expectedSubstr: "3.8.7",
			want:           true,
		},
		{
			name:           "URL encodes apache-maven product name",
			rawURL:         wrapper.DistributionURL,
			expectedSubstr: "apache-maven",
			want:           true,
		},
		{
			name:           "URL contains -bin suffix (binary distribution)",
			rawURL:         wrapper.DistributionURL,
			expectedSubstr: "-bin",
			want:           true,
		},
		{
			name:           "URL does NOT contain a different version",
			rawURL:         wrapper.DistributionURL,
			expectedSubstr: "3.9.0",
			want:           false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := strings.Contains(tc.rawURL, tc.expectedSubstr)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestDistributionURL_Download_Reachable(t *testing.T) {
	t.Parallel()

	srv := serveZip(t)

	tests := []struct {
		name           string
		serverURL      string
		wantStatusCode int
		wantBodyMagic  []byte
	}{
		{
			name:           "reachable server returns 200 with zip magic bytes",
			serverURL:      srv.URL,
			wantStatusCode: http.StatusOK,
			wantBodyMagic:  []byte("PK"),
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			resp := fetchURL(t, tc.serverURL)
			assert.Equal(t, tc.wantStatusCode, resp.StatusCode)

			buf := make([]byte, 2)
			n, _ := resp.Body.Read(buf)
			assert.Equal(t, tc.wantBodyMagic, buf[:n],
				"expected zip magic bytes at start of downloaded archive")
		})
	}
}

func TestDistributionURL_Download_NotFound(t *testing.T) {
	t.Parallel()

	srv := serve404(t)

	tests := []struct {
		name           string
		serverURL      string
		wantStatusCode int
	}{
		{
			name:           "404 response simulates unreachable/missing archive",
			serverURL:      srv.URL,
			wantStatusCode: http.StatusNotFound,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			resp := fetchURL(t, tc.serverURL)
			assert.Equal(t, tc.wantStatusCode, resp.StatusCode,
				"build should fail when distribution archive is not found")
		})
	}
}

func TestDistributionURL_Download_Corrupt(t *testing.T) {
	t.Parallel()

	srv := serveCorrupt(t)

	tests := []struct {
		name          string
		serverURL     string
		wantZipMagic  bool
		wantBodyStart string
	}{
		{
			name:          "corrupt body does not start with zip magic bytes",
			serverURL:     srv.URL,
			wantZipMagic:  false,
			wantBodyStart: "CORRUPT",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			resp := fetchURL(t, tc.serverURL)
			require.Equal(t, http.StatusOK, resp.StatusCode)

			buf := make([]byte, 7)
			n, _ := resp.Body.Read(buf)
			body := string(buf[:n])

			isZip := strings.HasPrefix(body, "PK")
			assert.Equal(t, tc.wantZipMagic, isZip,
				"corrupt archive must not begin with zip magic bytes — build should fail")
			assert.True(t, strings.HasPrefix(body, tc.wantBodyStart))
		})
	}
}

// ---------------------------------------------------------------------------
// WrapperURL tests
// ---------------------------------------------------------------------------

func TestWrapperURL_WellFormed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		rawURL  string
		wantOK  bool
		wantMsg string
	}{
		{
			name:    "canonical constant is a well-formed https URL",
			rawURL:  wrapper.WrapperURL,
			wantOK:  true,
			wantMsg: "WrapperURL must be a well-formed URL with scheme and host",
		},
		{
			name:    "empty string is not well-formed",
			rawURL:  "",
			wantOK:  false,
			wantMsg: "empty string should not be well-formed",
		},
		{
			name:    "path-only is not well-formed",
			rawURL:  "maven-wrapper-3.1.1.jar",
			wantOK:  false,
			wantMsg: "bare filename should not be a well-formed URL",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := isWellFormedURL(tc.rawURL)
			assert.Equal(t, tc.wantOK, got, tc.wantMsg)
		})
	}
}

func TestWrapperURL_ReferencesJARFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		rawURL string
		want   bool
	}{
		{
			name:   "constant references a .jar file",
			rawURL: wrapper.WrapperURL,
			want:   true,
		},
		{
			name:   "distribution URL does not reference a .jar",
			rawURL: wrapper.DistributionURL,
			want:   false,
		},
		{
			name:   "arbitrary non-jar URL",
			rawURL: "https://example.com/tool.zip",
			want:   false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := endsWithExtension(tc.rawURL, ".jar")
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestWrapperURL_ContainsWrapperVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		rawURL         string
		expectedSubstr string
		want           bool
	}{
		{
			name:           "URL encodes version 3.1.1",
			rawURL:         wrapper.WrapperURL,
			expectedSubstr: "3.1.1",
			want:           true,
		},
		{
			name:           "URL contains maven-wrapper artifact name",
			rawURL:         wrapper.WrapperURL,
			expectedSubstr: "maven-wrapper",
			want:           true,
		},
		{
			name:           "URL does NOT contain a different version",
			rawURL:         wrapper.WrapperURL,
			expectedSubstr: "3.2.0",
			want:           false,
		},
		{
			name:           "URL is served from Apache Maven central repo",
			rawURL:         wrapper.WrapperURL,
			expectedSubstr: "repo.maven.apache.org",
			want:           true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := strings.Contains(tc.rawURL, tc.expectedSubstr)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestWrapperURL_Download_Reachable(t *testing.T) {
	t.Parallel()

	srv := serveJAR(t)

	tests := []struct {
		name           string
		serverURL      string
		wantStatusCode int
		wantBodyMagic  []byte
	}{
		{
			name:           "reachable server returns 200 with JAR (zip) magic bytes",
			serverURL:      srv.URL,
			wantStatusCode: http.StatusOK,
			wantBodyMagic:  []byte("PK"),
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			resp := fetchURL(t, tc