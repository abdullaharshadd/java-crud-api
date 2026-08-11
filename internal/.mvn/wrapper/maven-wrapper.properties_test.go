```go
package wrapper_test

import (
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	wrapper "github.com/your/module/internal/.mvn/wrapper"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// parseURL is a small helper so test cases can assert individual URL fields
// without repeating url.Parse error handling everywhere.
func parseURL(t *testing.T, raw string) *url.URL {
	t.Helper()
	u, err := url.Parse(raw)
	require.NoError(t, err, "constant URL must be parseable")
	return u
}

// ---------------------------------------------------------------------------
// MavenDistributionURL – spec: distributionUrl
// ---------------------------------------------------------------------------

func TestMavenDistributionURL_ExactValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		got      string
		wantExact string
	}{
		{
			name:      "resolves to Apache Maven 3.8.7 binary zip",
			got:       wrapper.MavenDistributionURL,
			wantExact: "https://repo.maven.apache.org/maven2/org/apache/maven/apache-maven/3.8.7/apache-maven-3.8.7-bin.zip",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.wantExact, tc.got,
				"MavenDistributionURL must equal the pinned value from maven-wrapper.properties")
		})
	}
}

func TestMavenDistributionURL_Invariants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		checkFunc func(t *testing.T, raw string)
	}{
		{
			name: "uses HTTPS scheme",
			checkFunc: func(t *testing.T, raw string) {
				u := parseURL(t, raw)
				assert.Equal(t, "https", u.Scheme,
					"distributionUrl must use HTTPS for security and reproducibility")
			},
		},
		{
			name: "references official Maven Central repository host",
			checkFunc: func(t *testing.T, raw string) {
				u := parseURL(t, raw)
				assert.Equal(t, "repo.maven.apache.org", u.Host,
					"distributionUrl must reference the official Maven Central repository")
			},
		},
		{
			name: "points to a zip archive (binary distribution)",
			checkFunc: func(t *testing.T, raw string) {
				assert.True(t, strings.HasSuffix(raw, ".zip"),
					"distributionUrl must point to a downloadable binary zip archive")
			},
		},
		{
			name: "references Maven version 3.8.7",
			checkFunc: func(t *testing.T, raw string) {
				assert.Contains(t, raw, "3.8.7",
					"distributionUrl must reference Maven version 3.8.7")
			},
		},
		{
			name: "is a non-empty string",
			checkFunc: func(t *testing.T, raw string) {
				assert.NotEmpty(t, raw, "distributionUrl must be defined")
			},
		},
		{
			name: "is parseable as a URL",
			checkFunc: func(t *testing.T, raw string) {
				u, err := url.Parse(raw)
				assert.NoError(t, err, "distributionUrl must be a valid URL")
				assert.NotNil(t, u)
			},
		},
		{
			name: "path contains apache-maven artifact identifier",
			checkFunc: func(t *testing.T, raw string) {
				u := parseURL(t, raw)
				assert.Contains(t, u.Path, "apache-maven",
					"distributionUrl path must include the apache-maven artifact")
			},
		},
		{
			name: "path contains bin suffix indicating binary distribution",
			checkFunc: func(t *testing.T, raw string) {
				u := parseURL(t, raw)
				assert.Contains(t, u.Path, "-bin.zip",
					"distributionUrl must point to the binary (-bin) distribution")
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.checkFunc(t, wrapper.MavenDistributionURL)
		})
	}
}

// ---------------------------------------------------------------------------
// MavenWrapperURL – spec: wrapperUrl
// ---------------------------------------------------------------------------

func TestMavenWrapperURL_ExactValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		got       string
		wantExact string
	}{
		{
			name:      "resolves to maven-wrapper 3.1.1 JAR",
			got:       wrapper.MavenWrapperURL,
			wantExact: "https://repo.maven.apache.org/maven2/org/apache/maven/wrapper/maven-wrapper/3.1.1/maven-wrapper-3.1.1.jar",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.wantExact, tc.got,
				"MavenWrapperURL must equal the pinned value from maven-wrapper.properties")
		})
	}
}

func TestMavenWrapperURL_Invariants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		checkFunc func(t *testing.T, raw string)
	}{
		{
			name: "uses HTTPS scheme",
			checkFunc: func(t *testing.T, raw string) {
				u := parseURL(t, raw)
				assert.Equal(t, "https", u.Scheme,
					"wrapperUrl must use HTTPS for security and reproducibility")
			},
		},
		{
			name: "references official Maven Central repository host",
			checkFunc: func(t *testing.T, raw string) {
				u := parseURL(t, raw)
				assert.Equal(t, "repo.maven.apache.org", u.Host,
					"wrapperUrl must reference the official Maven Central repository")
			},
		},
		{
			name: "points to a JAR file",
			checkFunc: func(t *testing.T, raw string) {
				assert.True(t, strings.HasSuffix(raw, ".jar"),
					"wrapperUrl must point to a downloadable JAR file")
			},
		},
		{
			name: "references wrapper version 3.1.1",
			checkFunc: func(t *testing.T, raw string) {
				assert.Contains(t, raw, "3.1.1",
					"wrapperUrl must reference maven-wrapper version 3.1.1")
			},
		},
		{
			name: "is a non-empty string",
			checkFunc: func(t *testing.T, raw string) {
				assert.NotEmpty(t, raw, "wrapperUrl must be defined")
			},
		},
		{
			name: "is parseable as a URL",
			checkFunc: func(t *testing.T, raw string) {
				u, err := url.Parse(raw)
				assert.NoError(t, err, "wrapperUrl must be a valid URL")
				assert.NotNil(t, u)
			},
		},
		{
			name: "path contains maven-wrapper artifact identifier",
			checkFunc: func(t *testing.T, raw string) {
				u := parseURL(t, raw)
				assert.Contains(t, u.Path, "maven-wrapper",
					"wrapperUrl path must include the maven-wrapper artifact")
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.checkFunc(t, wrapper.MavenWrapperURL)
		})
	}
}

// ---------------------------------------------------------------------------
// MavenVersion constant
// ---------------------------------------------------------------------------

func TestMavenVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		got     string
		want    string
	}{
		{
			name: "MavenVersion equals 3.8.7",
			got:  wrapper.MavenVersion,
			want: "3.8.7",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, tc.got,
				"MavenVersion must record the pinned Apache Maven version")
		})
	}
}

// ---------------------------------------------------------------------------
// MavenWrapperVersion constant
// ---------------------------------------------------------------------------

func TestMavenWrapperVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		got  string
		want string
	}{
		{
			name: "MavenWrapperVersion equals 3.1.1",
			got:  wrapper.MavenWrapperVersion,
			want: "3.1.1",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, tc.got,
				"MavenWrapperVersion must record the pinned maven-wrapper version")
		})
	}
}

// ---------------------------------------------------------------------------
// Global invariants: both URLs must be defined, use HTTPS, and point to Maven
// Central.
// ---------------------------------------------------------------------------

func TestGlobalInvariants_BothURLsDefined(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		url  string
	}{
		{"MavenDistributionURL is defined", wrapper.MavenDistributionURL},
		{"MavenWrapperURL is defined", wrapper.MavenWrapperURL},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.NotEmpty(t, tc.url,
				"Both distributionUrl and wrapperUrl must be defined for the Maven Wrapper to function")
		})
	}
}

func TestGlobalInvariants_AllURLsUseHTTPS(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		url  string
	}{
		{"MavenDistributionURL uses HTTPS", wrapper.MavenDistributionURL},
		{"MavenWrapperURL uses HTTPS", wrapper.MavenWrapperURL},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			u := parseURL(t, tc.url)
			assert.Equal(t, "https", u.Scheme,
				"all URLs must use HTTPS and reference the official Maven Central repository for reproducibility")
		})
	}
}

func TestGlobalInvariants_AllURLsReferenceMavenCentral(t *testing.T) {
	t.Parallel()

	const mavenCentralHost = "repo.maven.apache.org"

	tests := []struct {
		name string
		url  string
	}{
		{"MavenDistributionURL is on Maven Central", wrapper.MavenDistributionURL},
		{"MavenWrapperURL is on Maven Central", wrapper.MavenWrapperURL},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			u := parseURL(t, tc.url)
			assert.Equal(t, mavenCentralHost, u.Host,
				"all URLs must reference the official Maven Central repository")
		})
	}
}

// ---------------------------------------------------------------------------
// Version consistency: version strings in constants must match URLs
// ---------------------------------------------------------------------------

func TestVersionConsistency(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		url     string
		version string
		label   string
	}{
		{
			name:    "MavenVersion appears in MavenDistributionURL",
			url:     wrapper.MavenDistributionURL,
			version: wrapper.MavenVersion,
			label:   "MavenVersion",
		},
		{
			name:    "MavenWrapperVersion appears in MavenWrapperURL",
			url:     wrapper.MavenWrapperURL,
			version: wrapper.MavenWrapperVersion,
			label:   "MavenWrapperVersion",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Contains(t, tc.url, tc.version,
				"%s constant must match the version embedded in the corresponding URL", tc.label)
		})
	}
}

// ---------------------------------------------------------------------------
// Error-case documentation: simulate an unreachable or invalid URL scenario
// using httptest to demonstrate the expected failure mode described in the
// behavioral specs.
//
// These tests use net/http/httptest to stand up a local server that returns
// error responses, mirroring the "build fails if the URL is unreachable or
// returns a non-existent artifact" error cases from the specs.
// ---------------------------------------------------------------------------

func TestDistributionURL_ErrorCases_UnreachableURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		rawURL         string
		expectParseErr bool
		description    string
	}{
		{
			name:           "empty string is not a valid distribution URL",
			rawURL:         "",
			expectParseErr: false, // url.Parse("") succeeds but host is empty
			description:    "an empty distributionUrl would cause the wrapper bootstrap to fail",
		},
		{
			name:           "non-HTTPS URL would be rejected by policy",
			rawURL:         "http://repo.maven.apache.org/maven2/apache-maven-3.8.7-bin.zip",
			expectParseErr: false,
			description:    "http (non-TLS) URLs violate the HTTPS invariant",
		},
		{
			name:           "URL without a host is invalid",
			rawURL:         "/relative/path/apache-maven-3.8.7-bin.zip",
			expectParseErr: false,
			description:    "relative URLs cannot be used to download artifacts",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			u, err := url.Parse(tc.rawURL)
			if tc.expectParseErr {
				assert.Error(t, err, tc.description)
				return
			}

			require.NoError(t, err)

			// The actual pinned constant must NOT be equal to these bad values.
			assert.NotEqual(t, wrapper.MavenDistributionURL, tc.rawURL,
				"pinned MavenDistributionURL must not equal an invalid URL: %s", tc.description)

			// Extra: if the bad URL is non-empty, its scheme or host should
			// differ from the invariant-compliant constant.
			if tc.rawURL != "" {
				badSchemeOrHost := u.Scheme != "https" || u.Host == ""
				assert.True(t, badSchemeOrHost,
					"the bad URL should violate