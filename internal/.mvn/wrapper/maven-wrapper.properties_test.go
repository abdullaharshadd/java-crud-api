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

	wrapper "github.com/yourorg/yourrepo/internal/.mvn/wrapper"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// serveArtifact returns a test server that responds 200 OK with a minimal
// body when the given path suffix is requested, and 404 otherwise.
func serveArtifact(t *testing.T, pathSuffix string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, pathSuffix) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("artifact-bytes"))
			return
		}
		http.NotFound(w, r)
	}))
}

// ---------------------------------------------------------------------------
// Table-driven constant value tests
// ---------------------------------------------------------------------------

func TestConstants_Values(t *testing.T) {
	tests := []struct {
		name     string
		got      string
		wantExact string
	}{
		{
			name:      "MavenDistributionVersion is 3.8.7",
			got:       wrapper.MavenDistributionVersion,
			wantExact: "3.8.7",
		},
		{
			name:      "MavenWrapperVersion is 3.1.1",
			got:       wrapper.MavenWrapperVersion,
			wantExact: "3.1.1",
		},
		{
			name:      "DistributionURL exact value",
			got:       wrapper.DistributionURL,
			wantExact: "https://repo.maven.apache.org/maven2/org/apache/maven/apache-maven/3.8.7/apache-maven-3.8.7-bin.zip",
		},
		{
			name:      "WrapperURL exact value",
			got:       wrapper.WrapperURL,
			wantExact: "https://repo.maven.apache.org/maven2/org/apache/maven/wrapper/maven-wrapper/3.1.1/maven-wrapper-3.1.1.jar",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.wantExact, tc.got)
		})
	}
}

// ---------------------------------------------------------------------------
// distributionUrl behavioral specs
// ---------------------------------------------------------------------------

func TestDistributionURL_BehavioralSpecs(t *testing.T) {
	tests := []struct {
		name          string
		scenario      string
		rawURL        string
		wantScheme    string
		wantHost      string
		wantSuffix    string
		wantVersion   string
		wantExtension string
	}{
		{
			name:          "resolves to correct distribution URL",
			scenario:      "Maven Wrapper reads the properties file to resolve the Maven distribution",
			rawURL:        wrapper.DistributionURL,
			wantScheme:    "https",
			wantHost:      "repo.maven.apache.org",
			wantSuffix:    "apache-maven-3.8.7-bin.zip",
			wantVersion:   "3.8.7",
			wantExtension: ".zip",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			parsed, err := url.Parse(tc.rawURL)
			require.NoError(t, err, "DistributionURL must be a parseable URL")

			assert.Equal(t, tc.wantScheme, parsed.Scheme,
				"DistributionURL must use HTTPS")
			assert.Equal(t, tc.wantHost, parsed.Host,
				"DistributionURL must point to the official Maven Central host")
			assert.True(t, strings.HasSuffix(parsed.Path, tc.wantSuffix),
				"DistributionURL path must end with %s, got %s", tc.wantSuffix, parsed.Path)
			assert.Contains(t, parsed.Path, tc.wantVersion,
				"DistributionURL must reference Maven version %s", tc.wantVersion)
			assert.True(t, strings.HasSuffix(parsed.Path, tc.wantExtension),
				"DistributionURL must point to a zip archive")
		})
	}
}

func TestDistributionURL_Invariants(t *testing.T) {
	t.Run("value must not be empty", func(t *testing.T) {
		assert.NotEmpty(t, wrapper.DistributionURL)
	})

	t.Run("must be a valid URL", func(t *testing.T) {
		parsed, err := url.ParseRequestURI(wrapper.DistributionURL)
		require.NoError(t, err)
		assert.NotEmpty(t, parsed.Host)
	})

	t.Run("points to official Maven Central repository", func(t *testing.T) {
		assert.True(t, strings.HasPrefix(wrapper.DistributionURL, "https://repo.maven.apache.org/"),
			"DistributionURL must start with https://repo.maven.apache.org/")
	})

	t.Run("references Apache Maven 3.8.7", func(t *testing.T) {
		assert.Contains(t, wrapper.DistributionURL, "3.8.7")
		assert.Contains(t, wrapper.DistributionURL, "apache-maven")
	})

	t.Run("points to a binary zip archive", func(t *testing.T) {
		assert.True(t, strings.HasSuffix(wrapper.DistributionURL, ".zip"),
			"DistributionURL must end with .zip")
	})

	t.Run("distribution version constant matches URL", func(t *testing.T) {
		assert.Contains(t, wrapper.DistributionURL, wrapper.MavenDistributionVersion,
			"DistributionURL must embed MavenDistributionVersion")
	})
}

// ---------------------------------------------------------------------------
// wrapperUrl behavioral specs
// ---------------------------------------------------------------------------

func TestWrapperURL_BehavioralSpecs(t *testing.T) {
	tests := []struct {
		name          string
		scenario      string
		rawURL        string
		wantScheme    string
		wantHost      string
		wantSuffix    string
		wantVersion   string
		wantExtension string
	}{
		{
			name:          "resolves to correct wrapper JAR URL",
			scenario:      "Maven Wrapper reads the properties file to resolve the wrapper JAR",
			rawURL:        wrapper.WrapperURL,
			wantScheme:    "https",
			wantHost:      "repo.maven.apache.org",
			wantSuffix:    "maven-wrapper-3.1.1.jar",
			wantVersion:   "3.1.1",
			wantExtension: ".jar",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			parsed, err := url.Parse(tc.rawURL)
			require.NoError(t, err, "WrapperURL must be a parseable URL")

			assert.Equal(t, tc.wantScheme, parsed.Scheme,
				"WrapperURL must use HTTPS")
			assert.Equal(t, tc.wantHost, parsed.Host,
				"WrapperURL must point to the official Maven Central host")
			assert.True(t, strings.HasSuffix(parsed.Path, tc.wantSuffix),
				"WrapperURL path must end with %s, got %s", tc.wantSuffix, parsed.Path)
			assert.Contains(t, parsed.Path, tc.wantVersion,
				"WrapperURL must reference maven-wrapper version %s", tc.wantVersion)
			assert.True(t, strings.HasSuffix(parsed.Path, tc.wantExtension),
				"WrapperURL must point to a JAR artifact")
		})
	}
}

func TestWrapperURL_Invariants(t *testing.T) {
	t.Run("value must not be empty", func(t *testing.T) {
		assert.NotEmpty(t, wrapper.WrapperURL)
	})

	t.Run("must be a valid URL", func(t *testing.T) {
		parsed, err := url.ParseRequestURI(wrapper.WrapperURL)
		require.NoError(t, err)
		assert.NotEmpty(t, parsed.Host)
	})

	t.Run("points to official Maven Central repository", func(t *testing.T) {
		assert.True(t, strings.HasPrefix(wrapper.WrapperURL, "https://repo.maven.apache.org/"),
			"WrapperURL must start with https://repo.maven.apache.org/")
	})

	t.Run("references maven-wrapper 3.1.1", func(t *testing.T) {
		assert.Contains(t, wrapper.WrapperURL, "3.1.1")
		assert.Contains(t, wrapper.WrapperURL, "maven-wrapper")
	})

	t.Run("points to a JAR artifact", func(t *testing.T) {
		assert.True(t, strings.HasSuffix(wrapper.WrapperURL, ".jar"),
			"WrapperURL must end with .jar")
	})

	t.Run("wrapper version constant matches URL", func(t *testing.T) {
		assert.Contains(t, wrapper.WrapperURL, wrapper.MavenWrapperVersion,
			"WrapperURL must embed MavenWrapperVersion")
	})
}

// ---------------------------------------------------------------------------
// Global invariants
// ---------------------------------------------------------------------------

func TestGlobalInvariants_BothURLsPresent(t *testing.T) {
	t.Run("DistributionURL must be non-empty for wrapper to function", func(t *testing.T) {
		assert.NotEmpty(t, wrapper.DistributionURL,
			"Both distributionUrl and wrapperUrl keys must be present for the wrapper to function")
	})

	t.Run("WrapperURL must be non-empty for wrapper to function", func(t *testing.T) {
		assert.NotEmpty(t, wrapper.WrapperURL,
			"Both distributionUrl and wrapperUrl keys must be present for the wrapper to function")
	})
}

func TestGlobalInvariants_OfficialRepository(t *testing.T) {
	tests := []struct {
		name   string
		rawURL string
	}{
		{
			name:   "DistributionURL is on official Maven Central",
			rawURL: wrapper.DistributionURL,
		},
		{
			name:   "WrapperURL is on official Maven Central",
			rawURL: wrapper.WrapperURL,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			parsed, err := url.Parse(tc.rawURL)
			require.NoError(t, err)
			assert.Equal(t, "repo.maven.apache.org", parsed.Host,
				"URLs must point to the official Maven Central repository")
		})
	}
}

func TestGlobalInvariants_ReproducibleBuilds(t *testing.T) {
	t.Run("constants are immutable (same value on repeated access)", func(t *testing.T) {
		first := wrapper.DistributionURL
		second := wrapper.DistributionURL
		assert.Equal(t, first, second, "constant value must be stable across accesses")

		firstW := wrapper.WrapperURL
		secondW := wrapper.WrapperURL
		assert.Equal(t, firstW, secondW, "constant value must be stable across accesses")
	})
}

// ---------------------------------------------------------------------------
// HTTP simulation tests (httptest) – error cases
// ---------------------------------------------------------------------------

// artifactFetchSimulator simulates the Maven Wrapper downloading an artifact.
// In production this would be Java code; here we model it as a Go function
// that performs an HTTP GET against whatever server is injected so that we
// can use httptest.Server to exercise the error cases described in the spec.
type artifactFetchSimulator struct {
	client  *http.Client
	baseURL string // injected by tests – overrides the real host
}

func (s *artifactFetchSimulator) fetchDistribution() (*http.Response, error) {
	target := rewriteURL(wrapper.DistributionURL, s.baseURL)
	return s.client.Get(target)
}

func (s *artifactFetchSimulator) fetchWrapper() (*http.Response, error) {
	target := rewriteURL(wrapper.WrapperURL, s.baseURL)
	return s.client.Get(target)
}

// rewriteURL replaces the scheme+host of rawURL with baseURL so we can point
// requests at an httptest.Server instead of the real internet.
func rewriteURL(rawURL, baseURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	base, err := url.Parse(baseURL)
	if err != nil {
		return rawURL
	}
	parsed.Scheme = base.Scheme
	parsed.Host = base.Host
	return parsed.String()
}

func TestDistributionURL_HTTPSimulation(t *testing.T) {
	tests := []struct {
		name           string
		serverHandler  http.HandlerFunc
		wantStatusCode int
		wantErr        bool
		description    string
	}{
		{
			name: "artifact available – wrapper bootstraps successfully",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				if strings.HasSuffix(r.URL.Path, "apache-maven-3.8.7-bin.zip") {
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte("PK")) // minimal zip magic bytes
					return
				}
				http.NotFound(w, r)
			},
			wantStatusCode: http.StatusOK,
			wantErr:        false,
			description:    "Maven Wrapper downloads Apache Maven 3.8.7 binary zip from the specified URL",
		},
		{
			name: "artifact not found – wrapper fails to bootstrap",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				http.NotFound(w, r)
			},
			wantStatusCode: http.StatusNotFound,
			wantErr:        false, // HTTP round-trip succeeds; application-level error is 404
			description:    "If the artifact does not exist, the wrapper fails to bootstrap Maven",
		},
		{
			name: "server error – wrapper fails to bootstrap",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "internal server error", http.StatusInternalServerError)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantErr:        false,
			description:    "If the URL is unreachable or the artifact does not exist, the wrapper fails to bootstrap Maven",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(tc.serverHandler)
			defer srv.Close()

			sim := &artifactFetchSimulator{
				client:  srv.Client(),
				baseURL: srv.URL,
			}

			resp, err := sim.fetchDistribution()
			if tc.wantErr {
				assert.Error(t, err, tc.description)
				return
			}
			require.NoError(t, err, tc.description)
			defer resp.Body.Close()
			assert.Equal(t, tc.wantStatusCode, resp.Status