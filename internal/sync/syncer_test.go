package sync

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testRegistryYAML = `
apiVersion: claim-registry.io/v1alpha1
kind: ClaimRegistry
claims:
  - name: hacky
    template: volumeclaim
    category: cli
    status: active
  - name: demo
    template: harborproject
    category: infra
    status: inactive
`

func newTestServer(t *testing.T, response string, statusCode int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		w.Write([]byte(response))
	}))
}

func TestInitialSync(t *testing.T) {
	ts := newTestServer(t, testRegistryYAML, http.StatusOK)
	defer ts.Close()

	s := NewSyncer(Config{
		Repo:    "test/repo",
		BaseURL: ts.URL,
	})

	err := s.InitialSync(context.Background())
	require.NoError(t, err)

	reg := s.GetRegistry()
	require.NotNil(t, reg)
	assert.Len(t, reg.Claims, 2)
	assert.Equal(t, "hacky", reg.Claims[0].Name)
}

func TestInitialSyncFailure(t *testing.T) {
	ts := newTestServer(t, "not found", http.StatusNotFound)
	defer ts.Close()

	s := NewSyncer(Config{
		Repo:    "test/repo",
		BaseURL: ts.URL,
	})

	err := s.InitialSync(context.Background())
	assert.Error(t, err)
	assert.Nil(t, s.GetRegistry())
}

func TestTokenHeader(t *testing.T) {
	var gotAuth string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Write([]byte(testRegistryYAML))
	}))
	defer ts.Close()

	s := NewSyncer(Config{
		Repo:    "test/repo",
		Token:   "ghp_test123",
		BaseURL: ts.URL,
	})

	err := s.InitialSync(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "token ghp_test123", gotAuth)
}

func TestBackgroundSync(t *testing.T) {
	callCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Write([]byte(testRegistryYAML))
	}))
	defer ts.Close()

	s := NewSyncer(Config{
		Repo:     "test/repo",
		BaseURL:  ts.URL,
		Interval: 50 * time.Millisecond,
	})

	err := s.InitialSync(context.Background())
	require.NoError(t, err)

	ctx := context.Background()
	s.Start(ctx)

	// Wait for at least 2 background syncs
	time.Sleep(150 * time.Millisecond)
	s.Stop()

	// Initial sync + at least 2 background polls
	assert.GreaterOrEqual(t, callCount, 3)
}

func TestDefaultConfig(t *testing.T) {
	s := NewSyncer(Config{Repo: "test/repo"})
	assert.Equal(t, "claims/registry.yaml", s.cfg.Path)
	assert.Equal(t, "main", s.cfg.Branch)
	assert.Equal(t, 60*time.Second, s.cfg.Interval)
	assert.Equal(t, "https://raw.githubusercontent.com", s.cfg.BaseURL)
}
