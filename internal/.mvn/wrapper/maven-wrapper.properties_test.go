```go
package wrapper_test

import (
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

// parseURL is a small helper that returns (scheme, host, path) without failing
// the test itself; callers do their own assertions.
func mustParseURL(t *testing.T, raw string) *url.URL {
	t.Helper()
	u, err := url.Parse(raw)
	require.NoError(t, err, "url.Parse(%q) must not return an error", raw)
	return u
}

// ---------------------------------------------------------------------------
// Table-driven tests for DistributionURL
// ---------------------------------------------------------------------------

func TestDistributionURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		checkFunc func(t *testing.T, raw string)
	}{
		{
			name: "constant is non-empty",
			checkFunc: func(t *testing.T, raw string) {
				assert.NotEmpty(t, raw, "DistributionURL must not be empty")
			},
		},
		{
			name: "is a well-formed URL",
			checkFunc: func(t *testing.T, raw string) {
				u := mustParseURL(t, raw)
				assert.NotEmpty(t, u.Scheme, "DistributionURL must have a scheme")
				assert.NotEmpty(t, u.Host, "DistributionURL must have a host")
				assert.NotEmpty(t, u.Path, "DistributionURL must have a path")
			},
		},
		{
			name: "scheme is https",
			checkFunc: func(t *testing.T, raw string) {
				u := mustParseURL(t, raw)
				assert.Equal(t, "https", u.Scheme, "DistributionURL scheme must be https")
			},
		},
		{
			name: "host is official Maven Central",
			checkFunc: func(t *testing.T, raw string) {
				u := mustParseURL(t, raw)
				assert.Equal(t, "repo.maven.apache.org", u.Host,
					"DistributionURL host must be repo.maven.apache.org")
			},
		},
		{
			name: "points to a Maven binary distribution archive (zip)",
			checkFunc: func(t *testing.T, raw string) {
				assert.True(t, strings.HasSuffix(raw, ".zip"),
					"DistributionURL must point to a .zip archive, got: %s", raw)
			},
		},
		{
			name: "encodes exactly Maven version 3.8.7",
			checkFunc: func(t *testing.T, raw string) {
				assert.Contains(t, raw, "3.8.7",
					"DistributionURL must encode Maven version 3.8.7")
			},
		},
		{
			name: "points to apache-maven artifact",
			checkFunc: func(t *testing.T, raw string) {
				assert.Contains(t, raw, "apache-maven",
					"DistributionURL must reference the apache-maven artifact")
			},
		},
		{
			name: "path contains bin qualifier indicating binary distribution",
			checkFunc: func(t *testing.T, raw string) {
				assert.Contains(t, raw, "-bin",
					"DistributionURL must point to a -bin distribution archive")
			},
		},
		{
			name: "expected exact value matches",
			checkFunc: func(t *testing.T, raw string) {
				const want = "https://repo.maven.apache.org/maven2/org/apache/maven/apache-maven/3.8.7/apache-maven-3.8.7-bin.zip"
				assert.Equal(t, want, raw, "DistributionURL must equal the canonical value")
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.checkFunc(t, wrapper.DistributionURL)
		})
	}
}

// ---------------------------------------------------------------------------
// Table-driven tests for WrapperURL
// ---------------------------------------------------------------------------

func TestWrapperURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		checkFunc func(t *testing.T, raw string)
	}{
		{
			name: "constant is non-empty",
			checkFunc: func(t *testing.T, raw string) {
				assert.NotEmpty(t, raw, "WrapperURL must not be empty")
			},
		},
		{
			name: "is a well-formed URL",
			checkFunc: func(t *testing.T, raw string) {
				u := mustParseURL(t, raw)
				assert.NotEmpty(t, u.Scheme, "WrapperURL must have a scheme")
				assert.NotEmpty(t, u.Host, "WrapperURL must have a host")
				assert.NotEmpty(t, u.Path, "WrapperURL must have a path")
			},
		},
		{
			name: "scheme is https",
			checkFunc: func(t *testing.T, raw string) {
				u := mustParseURL(t, raw)
				assert.Equal(t, "https", u.Scheme, "WrapperURL scheme must be https")
			},
		},
		{
			name: "host is official Maven Central",
			checkFunc: func(t *testing.T, raw string) {
				u := mustParseURL(t, raw)
				assert.Equal(t, "repo.maven.apache.org", u.Host,
					"WrapperURL host must be repo.maven.apache.org")
			},
		},
		{
			name: "points to a JAR artifact",
			checkFunc: func(t *testing.T, raw string) {
				assert.True(t, strings.HasSuffix(raw, ".jar"),
					"WrapperURL must point to a .jar file, got: %s", raw)
			},
		},
		{
			name: "encodes exactly wrapper version 3.1.1",
			checkFunc: func(t *testing.T, raw string) {
				assert.Contains(t, raw, "3.1.1",
					"WrapperURL must encode maven-wrapper version 3.1.1")
			},
		},
		{
			name: "points to maven-wrapper artifact",
			checkFunc: func(t *testing.T, raw string) {
				assert.Contains(t, raw, "maven-wrapper",
					"WrapperURL must reference the maven-wrapper artifact")
			},
		},
		{
			name: "expected exact value matches",
			checkFunc: func(t *testing.T, raw string) {
				const want = "https://repo.maven.apache.org/maven2/org/apache/maven/wrapper/maven-wrapper/3.1.1/maven-wrapper-3.1.1.jar"
				assert.Equal(t, want, raw, "WrapperURL must equal the canonical value")
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.checkFunc(t, wrapper.WrapperURL)
		})
	}
}

// ---------------------------------------------------------------------------
// Global invariants: both constants must be defined and consistent
// ---------------------------------------------------------------------------

func TestGlobalInvariants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		fn   func(t *testing.T)
	}{
		{
			name: "both DistributionURL and WrapperURL are defined",
			fn: func(t *testing.T) {
				assert.NotEmpty(t, wrapper.DistributionURL,
					"DistributionURL must be defined for the wrapper to bootstrap")
				assert.NotEmpty(t, wrapper.WrapperURL,
					"WrapperURL must be defined for the wrapper to bootstrap")
			},
		},
		{
			name: "both URLs resolve to the official Maven Central host",
			fn: func(t *testing.T) {
				const mavenCentral = "repo.maven.apache.org"
				for _, raw := range []string{wrapper.DistributionURL, wrapper.WrapperURL} {
					u := mustParseURL(t, raw)
					assert.Equal(t, mavenCentral, u.Host,
						"URL %q must resolve to %s", raw, mavenCentral)
				}
			},
		},
		{
			name: "URLs are distinct (distribution vs wrapper JAR)",
			fn: func(t *testing.T) {
				assert.NotEqual(t, wrapper.DistributionURL, wrapper.WrapperURL,
					"DistributionURL and WrapperURL must be different")
			},
		},
		{
			name: "DistributionURL encodes a different version than WrapperURL",
			fn: func(t *testing.T) {
				assert.Contains(t, wrapper.DistributionURL, "3.8.7")
				assert.Contains(t, wrapper.WrapperURL, "3.1.1")
			},
		},
		{
			name: "DistributionURL version does not appear in WrapperURL",
			fn: func(t *testing.T) {
				assert.False(t, strings.Contains(wrapper.WrapperURL, "3.8.7"),
					"WrapperURL must not embed the Maven distribution version 3.8.7")
			},
		},
		{
			name: "WrapperURL version does not appear in DistributionURL",
			fn: func(t *testing.T) {
				// 3.1.1 is not part of the distribution URL
				assert.False(t, strings.Contains(wrapper.DistributionURL, "3.1.1"),
					"DistributionURL must not embed the wrapper version 3.1.1")
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.fn(t)
		})
	}
}

// ---------------------------------------------------------------------------
// Simulated download behaviour using httptest
// (validates the URLs are structurally usable in an HTTP client context)
// ---------------------------------------------------------------------------

func TestURLsAreHTTPClientCompatible(t *testing.T) {
	t.Parallel()

	// We do NOT make real network calls. Instead we verify that url.Parse
	// produces a URL that, if a server existed at that host, an http.Client
	// could use without modification. Real connectivity is out of scope.

	tests := []struct {
		name string
		raw  string
	}{
		{name: "DistributionURL is HTTP-client-compatible", raw: wrapper.DistributionURL},
		{name: "WrapperURL is HTTP-client-compatible", raw: wrapper.WrapperURL},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			u := mustParseURL(t, tc.raw)

			// A URL usable by net/http must have a scheme of http or https and a
			// non-empty host.
			assert.True(t,
				u.Scheme == "http" || u.Scheme == "https",
				"URL scheme must be http or https, got %q", u.Scheme,
			)
			assert.NotEmpty(t, u.Host, "URL must have a non-empty host")

			// The path must be absolute (starts with /)
			assert.True(t, strings.HasPrefix(u.Path, "/"),
				"URL path must be absolute (start with /), got %q", u.Path)
		})
	}
}

// ---------------------------------------------------------------------------
// Simulated error scenarios (documentation: what would fail in Maven)
// ---------------------------------------------------------------------------

// TestDistributionURLErrorCases documents the failure modes described in the
// behavioural spec. Because these are documentation-only constants the tests
// verify the properties of the URL strings that would cause those failures
// if the URL were wrong.
func TestDistributionURLErrorCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		mutate      func(string) string
		expectValid bool
		reason      string
	}{
		{
			name:        "original URL is valid and would not trigger a download failure",
			mutate:      func(s string) string { return s },
			expectValid: true,
			reason:      "unmodified URL passes all validity checks",
		},
		{
			name:        "empty URL would cause Maven Wrapper to fail the build",
			mutate:      func(_ string) string { return "" },
			expectValid: false,
			reason:      "empty URL is not a well-formed URL",
		},
		{
			name:        "URL without scheme would cause Maven Wrapper to fail",
			mutate:      func(s string) string { return strings.TrimPrefix(s, "https://") },
			expectValid: false,
			reason:      "URL without scheme is not well-formed for net/http",
		},
		{
			name:        "URL pointing to wrong host would cause Maven Wrapper to download from unverified source",
			mutate:      func(s string) string { return strings.Replace(s, "repo.maven.apache.org", "evil.example.com", 1) },
			expectValid: false, // invalid by our invariant (wrong host)
			reason:      "host is not repo.maven.apache.org",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			candidate := tc.mutate(wrapper.DistributionURL)

			if tc.expectValid {
				u, err := url.Parse(candidate)
				assert.NoError(t, err)
				assert.Equal(t, "https", u.Scheme)
				assert.Equal(t, "repo.maven.apache.org", u.Host)
				assert.NotEmpty(t, u.Path)
			} else {
				// At least one invariant must be violated.
				u, err := url.Parse(candidate)
				violated := err != nil ||
					(u != nil && u.Scheme != "https") ||
					(u != nil && u.Host != "repo.maven.apache.org") ||
					candidate == ""
				assert.True(t, violated,
					"expected an invariant violation for mutated URL %q: %s", candidate, tc.reason)
			}
		})
	}
}

func TestWrapperURLErrorCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		mutate      func(string) string
		expectValid bool
		reason      string
	}{
		{
			name:        "original WrapperURL is valid",
			mutate:      func(s string) string { return s },
			expectValid: true,
			reason:      "unmodified URL passes all validity checks",
		},
		{
			name:        "empty WrapperURL would fail the wrapper bootstrap",
			mutate:      func(_ string) string { return "" },
			expectValid: false,
			reason:      "empty URL is not well-formed",
		},
		{
			name:        "WrapperURL without .jar extension would download wrong artifact",
			mutate:      func(s string) string { return strings.TrimSuffix(s, ".jar") },
			expectValid: false,
			reason:      "missing .jar suffix violates artifact type invariant",
		},
		{
			name:        "WrapperURL with wrong version would break reproducibility",
			mutate:      func(s string) string { return strings.Replace(s, "3.1.1", "9.9.9", -1) },
			expectValid: false,
			reason:      "version 3.1.1 is required for reproducible builds",
		