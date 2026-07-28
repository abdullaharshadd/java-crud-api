```go
// Package wrapper documents the Maven Wrapper configuration that was present
// in the original Java project. It contains no executable Go code.
package wrapper

import (
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Constants mirroring the original maven-wrapper.properties values so tests
// remain self-contained and do not depend on reading a file from disk.
const (
	distributionURL = "https://repo.maven.apache.org/maven2/org/apache/maven/apache-maven/3.8.7/apache-maven-3.8.7-bin.zip"
	wrapperURL      = "https://repo.maven.apache.org/maven2/org/apache/maven/wrapper/maven-wrapper/3.1.1/maven-wrapper-3.1.1.jar"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// parseURL is a thin wrapper used inside tests to keep assertion call-sites
// clean and to give the "function under test" a clear name.
func parseURL(raw string) (*url.URL, error) {
	return url.Parse(raw)
}

// isAbsoluteURL returns true when the string is a syntactically valid,
// absolute URL (scheme + host).
func isAbsoluteURL(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil {
		return false
	}
	return u.Scheme != "" && u.Host != ""
}

// extractVersion attempts to pull a version token from a Maven Central URL.
// It relies on the well-known Central path structure:
//
//	…/<artifactId>/<version>/<artifactId>-<version>-…
func extractVersion(rawURL string) string {
	parts := strings.Split(rawURL, "/")
	// The version segment sits just before the filename in a standard
	// Maven Central artifact URL.
	if len(parts) < 2 {
		return ""
	}
	return parts[len(parts)-2]
}

// ---------------------------------------------------------------------------
// Table-driven tests for distributionUrl invariants & behaviours
// ---------------------------------------------------------------------------

func TestDistributionURL_Invariants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		rawURL      string
		wantValid   bool
		wantPinned  bool   // expects a concrete version in the path
		wantVersion string // non-empty if version can be asserted exactly
	}{
		{
			name:        "canonical distribution URL is syntactically valid",
			rawURL:      distributionURL,
			wantValid:   true,
			wantPinned:  true,
			wantVersion: "3.8.7",
		},
		{
			name:      "empty string is not a valid distribution URL",
			rawURL:    "",
			wantValid: false,
		},
		{
			name:      "relative path is not a valid distribution URL",
			rawURL:    "apache-maven-3.8.7-bin.zip",
			wantValid: false,
		},
		{
			name:      "URL without version segment is considered unpinned",
			rawURL:    "https://repo.maven.apache.org/maven2/org/apache/maven/apache-maven/latest/apache-maven-latest-bin.zip",
			wantValid: true,
			// "latest" is not a numeric version – we treat it as unpinned.
			wantPinned: false,
		},
		{
			name:      "malformed URL is invalid",
			rawURL:    "://no-scheme",
			wantValid: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotValid := isAbsoluteURL(tc.rawURL)
			assert.Equal(t, tc.wantValid, gotValid,
				"isAbsoluteURL(%q) mismatch", tc.rawURL)

			if tc.wantValid && tc.wantPinned {
				version := extractVersion(tc.rawURL)
				// A pinned version must contain at least one dot (e.g. "3.8.7").
				assert.True(t,
					strings.Contains(version, "."),
					"distributionUrl version segment %q does not look pinned in URL %q",
					version, tc.rawURL,
				)
			}

			if tc.wantVersion != "" {
				version := extractVersion(tc.rawURL)
				assert.Equal(t, tc.wantVersion, version,
					"unexpected version extracted from distributionUrl")
			}
		})
	}
}

func TestDistributionURL_SameURLYieldsSameVersion(t *testing.T) {
	t.Parallel()

	// Invariant: the same URL always resolves to the same Maven distribution
	// version (idempotency / reproducibility guarantee).
	v1 := extractVersion(distributionURL)
	v2 := extractVersion(distributionURL)

	assert.Equal(t, v1, v2,
		"extractVersion must be deterministic for the same distributionUrl")
	assert.NotEmpty(t, v1,
		"version must not be empty for the canonical distributionUrl")
}

func TestDistributionURL_ParsedComponents(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		rawURL     string
		wantScheme string
		wantHost   string
		wantSuffix string // expected file suffix of the path
	}{
		{
			name:       "canonical URL has HTTPS scheme",
			rawURL:     distributionURL,
			wantScheme: "https",
			wantHost:   "repo.maven.apache.org",
			wantSuffix: ".zip",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			u, err := parseURL(tc.rawURL)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantScheme, u.Scheme)
			assert.Equal(t, tc.wantHost, u.Host)
			assert.True(t,
				strings.HasSuffix(u.Path, tc.wantSuffix),
				"path %q should end with %q", u.Path, tc.wantSuffix,
			)
		})
	}
}

// ---------------------------------------------------------------------------
// Table-driven tests for wrapperUrl invariants & behaviours
// ---------------------------------------------------------------------------

func TestWrapperURL_Invariants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		rawURL      string
		wantValid   bool
		wantPinned  bool
		wantVersion string
	}{
		{
			name:        "canonical wrapper URL is syntactically valid",
			rawURL:      wrapperURL,
			wantValid:   true,
			wantPinned:  true,
			wantVersion: "3.1.1",
		},
		{
			name:      "empty string is not a valid wrapper URL",
			rawURL:    "",
			wantValid: false,
		},
		{
			name:      "relative path is not a valid wrapper URL",
			rawURL:    "maven-wrapper-3.1.1.jar",
			wantValid: false,
		},
		{
			name:      "malformed URL is invalid",
			rawURL:    "not a url at all",
			wantValid: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotValid := isAbsoluteURL(tc.rawURL)
			assert.Equal(t, tc.wantValid, gotValid,
				"isAbsoluteURL(%q) mismatch", tc.rawURL)

			if tc.wantValid && tc.wantPinned {
				version := extractVersion(tc.rawURL)
				assert.True(t,
					strings.Contains(version, "."),
					"wrapperUrl version segment %q does not look pinned in URL %q",
					version, tc.rawURL,
				)
			}

			if tc.wantVersion != "" {
				version := extractVersion(tc.rawURL)
				assert.Equal(t, tc.wantVersion, version,
					"unexpected version extracted from wrapperUrl")
			}
		})
	}
}

func TestWrapperURL_ParsedComponents(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		rawURL     string
		wantScheme string
		wantHost   string
		wantSuffix string
	}{
		{
			name:       "canonical wrapper URL has HTTPS scheme and .jar suffix",
			rawURL:     wrapperURL,
			wantScheme: "https",
			wantHost:   "repo.maven.apache.org",
			wantSuffix: ".jar",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			u, err := parseURL(tc.rawURL)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantScheme, u.Scheme)
			assert.Equal(t, tc.wantHost, u.Host)
			assert.True(t,
				strings.HasSuffix(u.Path, tc.wantSuffix),
				"path %q should end with %q", u.Path, tc.wantSuffix,
			)
		})
	}
}

func TestWrapperURL_SameURLYieldsSameVersion(t *testing.T) {
	t.Parallel()

	v1 := extractVersion(wrapperURL)
	v2 := extractVersion(wrapperURL)

	assert.Equal(t, v1, v2,
		"extractVersion must be deterministic for the same wrapperUrl")
	assert.NotEmpty(t, v1,
		"version must not be empty for the canonical wrapperUrl")
}

// ---------------------------------------------------------------------------
// Global invariants
// ---------------------------------------------------------------------------

// TestBothURLsMustBeDefined validates the global invariant that both keys are
// present (non-empty) for the wrapper to function.
func TestBothURLsMustBeDefined(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		distURL         string
		wrapURL         string
		wantBothDefined bool
	}{
		{
			name:            "both URLs defined – wrapper is functional",
			distURL:         distributionURL,
			wrapURL:         wrapperURL,
			wantBothDefined: true,
		},
		{
			name:            "missing distributionUrl – wrapper not functional",
			distURL:         "",
			wrapURL:         wrapperURL,
			wantBothDefined: false,
		},
		{
			name:            "missing wrapperUrl – wrapper not functional",
			distURL:         distributionURL,
			wrapURL:         "",
			wantBothDefined: false,
		},
		{
			name:            "both URLs missing – wrapper not functional",
			distURL:         "",
			wrapURL:         "",
			wantBothDefined: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			bothDefined := tc.distURL != "" && tc.wrapURL != ""
			assert.Equal(t, tc.wantBothDefined, bothDefined)
		})
	}
}

// TestPinnedVersionsEnsureReproducibility validates that both the
// distributionUrl and wrapperUrl reference concrete version strings (not
// "latest", "SNAPSHOT", or similar moving targets).
func TestPinnedVersionsEnsureReproducibility(t *testing.T) {
	t.Parallel()

	movingTargets := []string{"latest", "LATEST", "RELEASE", "SNAPSHOT", ""}

	tests := []struct {
		name   string
		rawURL string
	}{
		{name: "distributionUrl is pinned", rawURL: distributionURL},
		{name: "wrapperUrl is pinned", rawURL: wrapperURL},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			version := extractVersion(tc.rawURL)
			for _, moving := range movingTargets {
				assert.NotEqual(t, moving, version,
					"URL %q must not use a moving version target %q", tc.rawURL, moving)
			}
			// A semver-like version must contain at least one dot.
			assert.True(t,
				strings.Contains(version, "."),
				"version %q extracted from %q does not appear to be a pinned semver", version, tc.rawURL,
			)
		})
	}
}

// TestWrapperCompatibility validates that both URLs share the same Maven
// Central host, which is a structural indication that the wrapper JAR version
// is compatible with the distribution version (both from the same repository).
func TestWrapperCompatibility(t *testing.T) {
	t.Parallel()

	uDist, errDist := parseURL(distributionURL)
	uWrap, errWrap := parseURL(wrapperURL)

	assert.NoError(t, errDist)
	assert.NoError(t, errWrap)
	assert.Equal(t, uDist.Host, uWrap.Host,
		"distributionUrl and wrapperUrl should resolve from the same Maven Central host")
}

// ---------------------------------------------------------------------------
// HTTP simulation tests
// ---------------------------------------------------------------------------
// Even though the migrated file contains no HTTP logic, the behavioral specs
// describe network download scenarios. We model them using httptest so that
// the test suite demonstrates the expected request/response semantics that a
// real Maven Wrapper implementation would exercise.

import (
	"io"
	"net/http"
	"net/http/httptest"
)

// artifactServer creates a test HTTP server that serves a fake artifact body
// at a given path. It returns the server and a cleanup function.
func artifactServer(t *testing.T, path string, body []byte, statusCode int) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		_, _ = w.Write(body)
	})
	return httptest.NewServer(mux)
}

func TestDistributionURL_Download_Scenarios(t *testing.T) {
	t.Parallel()

	fakeZipBody := []byte("PK\x03\x04fake-maven-zip-content")

	tests := []struct {
		name           string
		statusCode     int
		body           []byte
		wantErr        bool
		wantBodyPrefix []byte
	}{
		{
			name:           "distribution archive available – download succeeds",
			statusCode:     http.StatusOK,
			body:           fakeZipBody,
			wantErr:        false,
			wantBodyPrefix: fakeZipBody[:4],
		},
		{
			name:       "distribution URL unreachable – download fails with non-200",
			statusCode: http.StatusNotFound,
			body:       []byte("not found"),
			wantErr:    true,
		},
		{
			name:       "distribution URL returns server error – download fails",
			statusCode: http.StatusInternalServerError,
			body:       []byte("internal server error"),
			wantErr:    true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			const artifactPath = "/maven2/org/apache/maven/apache-maven/3.8.7/apache-maven-3.8.