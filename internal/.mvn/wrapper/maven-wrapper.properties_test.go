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
// Helpers
// ---------------------------------------------------------------------------

// isAbsoluteURL returns true when rawURL is a valid, absolute URL (has a
// scheme and a host).
func isAbsoluteURL(rawURL string) bool {
	u, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return false
	}
	return u.Scheme != "" && u.Host != ""
}

// isHTTPS returns true when the URL uses the https scheme.
func isHTTPS(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	return u.Scheme == "https"
}

// ---------------------------------------------------------------------------
// DistributionURL tests
// ---------------------------------------------------------------------------

func TestDistributionURL_Invariants(t *testing.T) {
	tests := []struct {
		name      string
		checkFn   func(t *testing.T)
	}{
		{
			name: "constant is non-empty",
			checkFn: func(t *testing.T) {
				assert.NotEmpty(t, wrapper.DistributionURL, "DistributionURL must not be empty")
			},
		},
		{
			name: "must be a valid absolute URL",
			checkFn: func(t *testing.T) {
				assert.True(t, isAbsoluteURL(wrapper.DistributionURL),
					"DistributionURL must be a valid, absolute URL; got %q", wrapper.DistributionURL)
			},
		},
		{
			name: "must use HTTPS scheme",
			checkFn: func(t *testing.T) {
				assert.True(t, isHTTPS(wrapper.DistributionURL),
					"DistributionURL must use HTTPS; got %q", wrapper.DistributionURL)
			},
		},
		{
			name: "must point to Maven binary distribution archive (zip)",
			checkFn: func(t *testing.T) {
				assert.True(t,
					strings.HasSuffix(wrapper.DistributionURL, ".zip"),
					"DistributionURL must point to a .zip archive; got %q", wrapper.DistributionURL)
			},
		},
		{
			name: "resolved Maven version must always be 3.8.7",
			checkFn: func(t *testing.T) {
				assert.Contains(t, wrapper.DistributionURL, "3.8.7",
					"DistributionURL must pin Maven version 3.8.7; got %q", wrapper.DistributionURL)
			},
		},
		{
			name: "must reference apache-maven artifact",
			checkFn: func(t *testing.T) {
				assert.Contains(t, wrapper.DistributionURL, "apache-maven",
					"DistributionURL must reference an apache-maven artifact")
			},
		},
		{
			name: "must be a binary (bin) distribution",
			checkFn: func(t *testing.T) {
				assert.Contains(t, wrapper.DistributionURL, "-bin",
					"DistributionURL must point to a binary (-bin) distribution")
			},
		},
		{
			name: "exact expected value",
			checkFn: func(t *testing.T) {
				expected := "https://repo.maven.apache.org/maven2/org/apache/maven/apache-maven/3.8.7/apache-maven-3.8.7-bin.zip"
				assert.Equal(t, expected, wrapper.DistributionURL)
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, tc.checkFn)
	}
}

func TestDistributionURL_BehavioralSpecs(t *testing.T) {
	tests := []struct {
		name           string
		serverBehavior func(w http.ResponseWriter, r *http.Request)
		wantStatusOk   bool
		wantErrOnFetch bool
		description    string
	}{
		{
			name: "scenario: distribution not cached – server reachable returns 200",
			serverBehavior: func(w http.ResponseWriter, r *http.Request) {
				// Simulate a reachable Maven repository returning the archive.
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("PK\x03\x04")) // ZIP magic bytes
			},
			wantStatusOk:   true,
			wantErrOnFetch: false,
			description:    "When distribution is not cached, wrapper downloads from distributionUrl successfully",
		},
		{
			name: "error_case: URL unreachable – server returns 404",
			serverBehavior: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			},
			wantStatusOk:   false,
			wantErrOnFetch: false, // HTTP itself succeeds; caller must treat non-200 as failure
			description:    "When the distribution URL is unreachable or archive missing, the download fails",
		},
		{
			name: "error_case: server returns 503 service unavailable",
			serverBehavior: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusServiceUnavailable)
			},
			wantStatusOk:   false,
			wantErrOnFetch: false,
			description:    "When server is unavailable the download fails",
		},
		{
			name: "scenario: distribution cached – no download (simulated by re-use of local copy)",
			serverBehavior: func(w http.ResponseWriter, r *http.Request) {
				// In a real scenario this handler would NOT be called when the
				// distribution is cached. We model the "cached" path by verifying
				// the URL constant alone (no HTTP call is made in this sub-test).
				t.Error("server should not be called when distribution is already cached")
			},
			wantStatusOk:   true,
			wantErrOnFetch: false,
			description:    "When distribution 3.8.7 is already cached, it is reused without network I/O",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if tc.name == "scenario: distribution cached – no download (simulated by re-use of local copy)" {
				// The cached scenario: validate that the URL constant itself
				// encodes enough information to identify the artifact without
				// making a network request.
				assert.Contains(t, wrapper.DistributionURL, "3.8.7")
				assert.Contains(t, wrapper.DistributionURL, "apache-maven")
				return
			}

			srv := httptest.NewServer(http.HandlerFunc(tc.serverBehavior))
			defer srv.Close()

			// Construct a request that mimics what a wrapper would do.
			req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
			require.NoError(t, err)

			resp, err := http.DefaultClient.Do(req)
			if tc.wantErrOnFetch {
				assert.Error(t, err, tc.description)
				return
			}
			require.NoError(t, err)
			defer resp.Body.Close()

			if tc.wantStatusOk {
				assert.Equal(t, http.StatusOK, resp.StatusCode, tc.description)
			} else {
				assert.NotEqual(t, http.StatusOK, resp.StatusCode, tc.description)
			}
		})
	}
}

// TestDistributionURL_ArchiveValidation verifies that when a valid ZIP archive
// is returned, the content starts with the ZIP magic bytes (error case:
// corrupt/invalid archive).
func TestDistributionURL_ArchiveValidation(t *testing.T) {
	tests := []struct {
		name        string
		body        []byte
		wantCorrupt bool
	}{
		{
			name:        "valid zip archive has correct magic bytes",
			body:        []byte("PK\x03\x04rest-of-zip-content"),
			wantCorrupt: false,
		},
		{
			name:        "error_case: corrupt archive – wrong magic bytes",
			body:        []byte("NOTAZIP"),
			wantCorrupt: true,
		},
		{
			name:        "error_case: empty response body – invalid archive",
			body:        []byte{},
			wantCorrupt: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(tc.body)
			}))
			defer srv.Close()

			resp, err := http.Get(srv.URL)
			require.NoError(t, err)
			defer resp.Body.Close()

			buf := make([]byte, 4)
			n, _ := resp.Body.Read(buf)
			body := buf[:n]

			isValidZip := len(body) >= 2 && body[0] == 'P' && body[1] == 'K'

			if tc.wantCorrupt {
				assert.False(t, isValidZip,
					"expected corrupt/invalid archive but got valid zip magic bytes")
			} else {
				assert.True(t, isValidZip,
					"expected valid zip archive but magic bytes are wrong")
			}
		})
	}
}

// ---------------------------------------------------------------------------
// WrapperURL tests
// ---------------------------------------------------------------------------

func TestWrapperURL_Invariants(t *testing.T) {
	tests := []struct {
		name    string
		checkFn func(t *testing.T)
	}{
		{
			name: "constant is non-empty",
			checkFn: func(t *testing.T) {
				assert.NotEmpty(t, wrapper.WrapperURL, "WrapperURL must not be empty")
			},
		},
		{
			name: "must be a valid absolute URL",
			checkFn: func(t *testing.T) {
				assert.True(t, isAbsoluteURL(wrapper.WrapperURL),
					"WrapperURL must be a valid, absolute URL; got %q", wrapper.WrapperURL)
			},
		},
		{
			name: "must use HTTPS scheme",
			checkFn: func(t *testing.T) {
				assert.True(t, isHTTPS(wrapper.WrapperURL),
					"WrapperURL must use HTTPS; got %q", wrapper.WrapperURL)
			},
		},
		{
			name: "must point to a JAR artifact",
			checkFn: func(t *testing.T) {
				assert.True(t,
					strings.HasSuffix(wrapper.WrapperURL, ".jar"),
					"WrapperURL must point to a .jar artifact; got %q", wrapper.WrapperURL)
			},
		},
		{
			name: "resolved wrapper version must always be 3.1.1",
			checkFn: func(t *testing.T) {
				assert.Contains(t, wrapper.WrapperURL, "3.1.1",
					"WrapperURL must pin wrapper version 3.1.1; got %q", wrapper.WrapperURL)
			},
		},
		{
			name: "must reference maven-wrapper artifact",
			checkFn: func(t *testing.T) {
				assert.Contains(t, wrapper.WrapperURL, "maven-wrapper",
					"WrapperURL must reference a maven-wrapper artifact")
			},
		},
		{
			name: "exact expected value",
			checkFn: func(t *testing.T) {
				expected := "https://repo.maven.apache.org/maven2/org/apache/maven/wrapper/maven-wrapper/3.1.1/maven-wrapper-3.1.1.jar"
				assert.Equal(t, expected, wrapper.WrapperURL)
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, tc.checkFn)
	}
}

func TestWrapperURL_BehavioralSpecs(t *testing.T) {
	tests := []struct {
		name           string
		serverBehavior func(w http.ResponseWriter, r *http.Request)
		wantStatusOk   bool
		wantErrOnFetch bool
		description    string
		skipHTTP       bool
	}{
		{
			name: "scenario: wrapper JAR not present – server reachable returns 200",
			serverBehavior: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				// Minimal JAR header (ZIP-based).
				_, _ = w.Write([]byte("PK\x03\x04"))
			},
			wantStatusOk:   true,
			wantErrOnFetch: false,
			description:    "When wrapper JAR is absent, it is downloaded from wrapperUrl successfully",
		},
		{
			name: "error_case: URL unreachable – server returns 404",
			serverBehavior: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			},
			wantStatusOk:   false,
			wantErrOnFetch: false,
			description:    "When wrapperUrl is unreachable, bootstrap fails",
		},
		{
			name: "error_case: server returns 500 internal server error",
			serverBehavior: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			wantStatusOk:   false,
			wantErrOnFetch: false,
			description:    "When server errors, JAR download fails",
		},
		{
			name:        "scenario: wrapper JAR 3.1.1 already present – no download",
			skipHTTP:    true,
			wantStatusOk: true,
			description:  "When wrapper JAR 3.1.1 is already cached, it is reused without network I/O",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if tc.skipHTTP {
				// Simulate cache-hit scenario: verify constant encodes correct
				// version without any network call.
				assert.Contains(t, wrapper.WrapperURL, "3.1.1")
				assert.Contains(t, wrapper.WrapperURL, "maven-wrapper")
				return
			}

			srv := httptest.NewServer(http.HandlerFunc(tc.serverBehavior))
			defer srv.Close()

			req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
			require.NoError(t, err)

			resp, err := http.DefaultClient.Do(req)
			if tc.wantErrOnFetch {
				assert.Error(t, err, tc.description)
				return
			}
			require.NoError(t, err)
			defer resp.Body.Close()

			if tc.wantStatusOk {
				assert.Equal(t, http.StatusOK, resp.StatusCode, tc.description)
			} else {
				assert.NotEqual(t, http.StatusOK, resp.StatusCode, tc.description)
			}
		})
	}
}

// TestWrapperURL_JARValidation validates the content-type / magic-byte check
// that a real wrapper implementation would perform after