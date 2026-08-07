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
// Helper – parse a raw URL string and assert it is valid.
// ---------------------------------------------------------------------------

func mustParseURL(t *testing.T, raw string) *url.URL {
	t.Helper()
	u, err := url.Parse(raw)
	require.NoError(t, err, "url.Parse(%q) must not return an error", raw)
	return u
}

// ---------------------------------------------------------------------------
// Constant-value tests
// ---------------------------------------------------------------------------

func TestMavenDistributionVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		got      string
		wantExact string
	}{
		{
			name:      "version is pinned to 3.8.7",
			got:       wrapper.MavenDistributionVersion,
			wantExact: "3.8.7",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.wantExact, tc.got,
				"MavenDistributionVersion must be pinned to ensure reproducible builds")
		})
	}
}

func TestMavenWrapperVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		got       string
		wantExact string
	}{
		{
			name:      "wrapper version is pinned to 3.1.1",
			got:       wrapper.MavenWrapperVersion,
			wantExact: "3.1.1",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.wantExact, tc.got,
				"MavenWrapperVersion must be pinned to ensure reproducible builds")
		})
	}
}

// ---------------------------------------------------------------------------
// distributionUrl invariant tests
// ---------------------------------------------------------------------------

func TestDistributionURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		check       func(t *testing.T, u *url.URL, raw string)
	}{
		{
			name: "constant is non-empty",
			check: func(t *testing.T, _ *url.URL, raw string) {
				assert.NotEmpty(t, raw, "MavenDistributionURL must not be empty")
			},
		},
		{
			name: "URL scheme is https (secure)",
			check: func(t *testing.T, u *url.URL, _ string) {
				assert.Equal(t, "https", u.Scheme,
					"distributionUrl must use a secure (https) scheme")
			},
		},
		{
			name: "URL host is the trusted Maven central repository",
			check: func(t *testing.T, u *url.URL, _ string) {
				assert.Equal(t, "repo.maven.apache.org", u.Host,
					"distributionUrl must point to the trusted Maven Central repository")
			},
		},
		{
			name: "URL references Maven version 3.8.7",
			check: func(t *testing.T, _ *url.URL, raw string) {
				assert.Contains(t, raw, "3.8.7",
					"distributionUrl must contain the pinned Maven version 3.8.7")
			},
		},
		{
			name: "URL references a binary distribution archive (-bin.zip)",
			check: func(t *testing.T, _ *url.URL, raw string) {
				assert.True(t,
					strings.HasSuffix(raw, "-bin.zip"),
					"distributionUrl must point to a binary distribution zip archive; got %q", raw)
			},
		},
		{
			name: "URL path contains apache-maven artifact coordinates",
			check: func(t *testing.T, u *url.URL, _ string) {
				assert.Contains(t, u.Path, "apache-maven",
					"distributionUrl path must contain the apache-maven artifact")
			},
		},
		{
			name: "URL version segment matches MavenDistributionVersion constant",
			check: func(t *testing.T, _ *url.URL, raw string) {
				assert.Contains(t, raw, wrapper.MavenDistributionVersion,
					"distributionUrl must embed the value of MavenDistributionVersion")
			},
		},
		{
			name: "URL is a parseable absolute URL",
			check: func(t *testing.T, u *url.URL, _ string) {
				assert.True(t, u.IsAbs(), "distributionUrl must be an absolute URL")
			},
		},
	}

	raw := wrapper.MavenDistributionURL
	u := mustParseURL(t, raw)

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.check(t, u, raw)
		})
	}
}

// ---------------------------------------------------------------------------
// wrapperUrl invariant tests
// ---------------------------------------------------------------------------

func TestWrapperURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		check func(t *testing.T, u *url.URL, raw string)
	}{
		{
			name: "constant is non-empty",
			check: func(t *testing.T, _ *url.URL, raw string) {
				assert.NotEmpty(t, raw, "MavenWrapperURL must not be empty")
			},
		},
		{
			name: "URL scheme is https (secure)",
			check: func(t *testing.T, u *url.URL, _ string) {
				assert.Equal(t, "https", u.Scheme,
					"wrapperUrl must use a secure (https) scheme")
			},
		},
		{
			name: "URL host is the trusted Maven central repository",
			check: func(t *testing.T, u *url.URL, _ string) {
				assert.Equal(t, "repo.maven.apache.org", u.Host,
					"wrapperUrl must point to the trusted Maven Central repository")
			},
		},
		{
			name: "URL references wrapper version 3.1.1",
			check: func(t *testing.T, _ *url.URL, raw string) {
				assert.Contains(t, raw, "3.1.1",
					"wrapperUrl must contain the pinned wrapper version 3.1.1")
			},
		},
		{
			name: "URL references a JAR artifact",
			check: func(t *testing.T, _ *url.URL, raw string) {
				assert.True(t,
					strings.HasSuffix(raw, ".jar"),
					"wrapperUrl must point to a JAR file; got %q", raw)
			},
		},
		{
			name: "URL path contains maven-wrapper artifact coordinates",
			check: func(t *testing.T, u *url.URL, _ string) {
				assert.Contains(t, u.Path, "maven-wrapper",
					"wrapperUrl path must contain the maven-wrapper artifact")
			},
		},
		{
			name: "URL version segment matches MavenWrapperVersion constant",
			check: func(t *testing.T, _ *url.URL, raw string) {
				assert.Contains(t, raw, wrapper.MavenWrapperVersion,
					"wrapperUrl must embed the value of MavenWrapperVersion")
			},
		},
		{
			name: "URL is a parseable absolute URL",
			check: func(t *testing.T, u *url.URL, _ string) {
				assert.True(t, u.IsAbs(), "wrapperUrl must be an absolute URL")
			},
		},
	}

	raw := wrapper.MavenWrapperURL
	u := mustParseURL(t, raw)

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.check(t, u, raw)
		})
	}
}

// ---------------------------------------------------------------------------
// Global invariants – both URLs must be defined and coherent
// ---------------------------------------------------------------------------

func TestBothURLsDefined(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		check func(t *testing.T)
	}{
		{
			name: "both MavenDistributionURL and MavenWrapperURL are non-empty",
			check: func(t *testing.T) {
				assert.NotEmpty(t, wrapper.MavenDistributionURL,
					"distributionUrl must be defined for the Maven Wrapper to function")
				assert.NotEmpty(t, wrapper.MavenWrapperURL,
					"wrapperUrl must be defined for the Maven Wrapper to function")
			},
		},
		{
			name: "MavenDistributionURL and MavenWrapperURL are distinct",
			check: func(t *testing.T) {
				assert.NotEqual(t, wrapper.MavenDistributionURL, wrapper.MavenWrapperURL,
					"distributionUrl and wrapperUrl must be different artifacts")
			},
		},
		{
			name: "distribution version appears in distribution URL",
			check: func(t *testing.T) {
				assert.Contains(t, wrapper.MavenDistributionURL, wrapper.MavenDistributionVersion,
					"MavenDistributionURL must embed MavenDistributionVersion for consistency")
			},
		},
		{
			name: "wrapper version appears in wrapper URL",
			check: func(t *testing.T) {
				assert.Contains(t, wrapper.MavenWrapperURL, wrapper.MavenWrapperVersion,
					"MavenWrapperURL must embed MavenWrapperVersion for consistency")
			},
		},
		{
			name: "both URLs use a trusted host",
			check: func(t *testing.T) {
				for _, raw := range []string{wrapper.MavenDistributionURL, wrapper.MavenWrapperURL} {
					u, err := url.Parse(raw)
					require.NoError(t, err)
					assert.Equal(t, "repo.maven.apache.org", u.Host,
						"URL %q must be hosted on repo.maven.apache.org", raw)
				}
			},
		},
		{
			name: "both URLs use https scheme",
			check: func(t *testing.T) {
				for _, raw := range []string{wrapper.MavenDistributionURL, wrapper.MavenWrapperURL} {
					u, err := url.Parse(raw)
					require.NoError(t, err)
					assert.Equal(t, "https", u.Scheme,
						"URL %q must use https scheme", raw)
				}
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.check(t)
		})
	}
}

// ---------------------------------------------------------------------------
// Scenario: distribution already cached – URL still resolves correctly
// (simulated via httptest so no real network call is made)
// ---------------------------------------------------------------------------

func TestDistributionURLStructureMatchesMavenCentralLayout(t *testing.T) {
	t.Parallel()

	// Maven Central layout:
	// /maven2/<groupPath>/<artifactId>/<version>/<artifactId>-<version>-<classifier>.<ext>
	// groupId = org.apache.maven  → org/apache/maven
	// artifactId = apache-maven
	// version = 3.8.7
	// classifier = bin
	// ext = zip
	tests := []struct {
		name    string
		segment string
	}{
		{"path contains maven2 repository prefix", "maven2"},
		{"path contains org/apache/maven group", "org/apache/maven"},
		{"path contains apache-maven artifactId", "apache-maven"},
		{"path contains version directory", "3.8.7"},
		{"filename contains expected artifact name", "apache-maven-3.8.7-bin.zip"},
	}

	u := mustParseURL(t, wrapper.MavenDistributionURL)

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Contains(t, u.Path, tc.segment,
				"MavenDistributionURL path must contain %q to match Maven Central layout", tc.segment)
		})
	}
}

func TestWrapperURLStructureMatchesMavenCentralLayout(t *testing.T) {
	t.Parallel()

	// Maven Central layout for the wrapper JAR:
	// groupId = org.apache.maven.wrapper  → org/apache/maven/wrapper
	// artifactId = maven-wrapper
	// version = 3.1.1
	// ext = jar
	tests := []struct {
		name    string
		segment string
	}{
		{"path contains maven2 repository prefix", "maven2"},
		{"path contains org/apache/maven/wrapper group", "org/apache/maven/wrapper"},
		{"path contains maven-wrapper artifactId", "maven-wrapper"},
		{"path contains version directory", "3.1.1"},
		{"filename contains expected artifact name", "maven-wrapper-3.1.1.jar"},
	}

	u := mustParseURL(t, wrapper.MavenWrapperURL)

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Contains(t, u.Path, tc.segment,
				"MavenWrapperURL path must contain %q to match Maven Central layout", tc.segment)
		})
	}
}

// ---------------------------------------------------------------------------
// Simulate error cases: verify the constants are NOT empty / NOT malformed
// (represents: "build fails if URL is missing or invalid")
// ---------------------------------------------------------------------------

func TestURLsAreNotMalformed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
	}{
		{
			name: "MavenDistributionURL parses without error",
			raw:  wrapper.MavenDistributionURL,
		},
		{
			name: "MavenWrapperURL parses without error",
			raw:  wrapper.MavenWrapperURL,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			u, err := url.Parse(tc.raw)
			assert.NoError(t, err, "URL %q must be parseable", tc.raw)
			assert.NotNil(t, u, "Parsed URL must not be nil")
			assert.True(t, u.IsAbs(), "URL %q must be absolute (scheme + host)", tc.raw)
			assert.NotEmpty(t, u.Host, "URL %q must have a non-empty host", tc.raw)
			assert.NotEmpty(t, u.Path, "URL %q must have a non-empty path", tc.raw)
		})
	}
}

// ---------------------------------------------------------------------------
// Reproducibility invariant: constants are immutable (no pointer indirection)
// ---------------------------------------------------------------------------

func TestConstantsAreReproducible(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		read1 func() string
		read2 func() string
	}{
		{
			name:  "MavenDistributionVersion reads the same value twice",
			read1: func() string { return wrapper.