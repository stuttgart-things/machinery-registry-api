package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/stuttgart-things/machinery-registry-api/internal/registry"
	isync "github.com/stuttgart-things/machinery-registry-api/internal/sync"
)

const testRegistryYAML = `
apiVersion: claim-registry.io/v1alpha1
kind: ClaimRegistry
claims:
  - name: hacky
    template: volumeclaim
    category: cli
    namespace: default
    createdAt: "2026-02-05T10:58:33Z"
    createdBy: patrick
    source: cli
    repository: stuttgart-things/harvester
    path: claims/cli/hacky.yaml
    status: active
  - name: harvestervm-developer-martin
    template: harvestervm
    category: cli
    namespace: default
    createdAt: "2026-02-05T17:57:18Z"
    createdBy: patrick
    source: cli
    repository: stuttgart-things/harvester
    path: claims/cli/harvestervm-developer-martin.yaml
    status: active
  - name: demo-project
    template: harborproject
    category: infra
    namespace: harbor
    createdAt: "2026-02-07T20:17:19Z"
    createdBy: cli
    source: gitops
    repository: stuttgart-things/harvester
    path: claims/infra/demo-project.yaml
    status: inactive
`

// setupTestServer creates a Server backed by an httptest mock serving testRegistryYAML.
func setupTestServer(t *testing.T) *Server {
	t.Helper()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(testRegistryYAML))
	}))
	t.Cleanup(ts.Close)

	syncer := isync.NewSyncer(isync.Config{
		Repo:    "test/repo",
		BaseURL: ts.URL,
	})
	err := syncer.InitialSync(context.Background())
	require.NoError(t, err)

	return NewServer(syncer)
}

func TestHealthEndpoint(t *testing.T) {
	srv := setupTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	srv.router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var body map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &body)
	require.NoError(t, err)
	assert.Equal(t, "healthy", body["status"])
}

func TestVersionEndpoint(t *testing.T) {
	srv := setupTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/version", nil)
	rr := httptest.NewRecorder()
	srv.router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "version")
}

func TestListClaimsNoFilter(t *testing.T) {
	srv := setupTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/claims", nil)
	rr := httptest.NewRecorder()
	srv.router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp ClaimListResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "claim-registry.io/v1alpha1", resp.APIVersion)
	assert.Equal(t, "ClaimList", resp.Kind)
	assert.Len(t, resp.Items, 3)
}

func TestListClaimsWithCategoryFilter(t *testing.T) {
	srv := setupTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/claims?category=cli", nil)
	rr := httptest.NewRecorder()
	srv.router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp ClaimListResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Len(t, resp.Items, 2)
}

func TestListClaimsWithMultipleFilters(t *testing.T) {
	srv := setupTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/claims?category=cli&template=volumeclaim", nil)
	rr := httptest.NewRecorder()
	srv.router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp ClaimListResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Len(t, resp.Items, 1)
	assert.Equal(t, "hacky", resp.Items[0].Name)
}

func TestListClaimsWithStatusFilter(t *testing.T) {
	srv := setupTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/claims?status=inactive", nil)
	rr := httptest.NewRecorder()
	srv.router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp ClaimListResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Len(t, resp.Items, 1)
	assert.Equal(t, "demo-project", resp.Items[0].Name)
}

func TestListClaimsNoMatch(t *testing.T) {
	srv := setupTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/claims?category=nonexistent", nil)
	rr := httptest.NewRecorder()
	srv.router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp ClaimListResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Len(t, resp.Items, 0)
}

func TestGetClaimFound(t *testing.T) {
	srv := setupTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/claims/hacky", nil)
	rr := httptest.NewRecorder()
	srv.router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var entry registry.ClaimEntry
	err := json.Unmarshal(rr.Body.Bytes(), &entry)
	require.NoError(t, err)
	assert.Equal(t, "hacky", entry.Name)
	assert.Equal(t, "volumeclaim", entry.Template)
}

func TestGetClaimNotFound(t *testing.T) {
	srv := setupTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/claims/nonexistent", nil)
	rr := httptest.NewRecorder()
	srv.router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)

	var body map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &body)
	require.NoError(t, err)
	assert.Equal(t, "claim not found", body["error"])
}
