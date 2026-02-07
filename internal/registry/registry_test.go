package registry

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testYAML = `
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

func TestParseData(t *testing.T) {
	reg, err := ParseData([]byte(testYAML))
	require.NoError(t, err)

	assert.Equal(t, "claim-registry.io/v1alpha1", reg.APIVersion)
	assert.Equal(t, "ClaimRegistry", reg.Kind)
	assert.Len(t, reg.Claims, 3)
	assert.Equal(t, "hacky", reg.Claims[0].Name)
	assert.Equal(t, "volumeclaim", reg.Claims[0].Template)
}

func TestParseDataInvalid(t *testing.T) {
	_, err := ParseData([]byte(":::invalid"))
	assert.Error(t, err)
}

func TestFindEntry(t *testing.T) {
	reg, err := ParseData([]byte(testYAML))
	require.NoError(t, err)

	entry := FindEntry(reg, "hacky")
	require.NotNil(t, entry)
	assert.Equal(t, "volumeclaim", entry.Template)

	entry = FindEntry(reg, "nonexistent")
	assert.Nil(t, entry)
}

func TestFilterEntries(t *testing.T) {
	reg, err := ParseData([]byte(testYAML))
	require.NoError(t, err)

	// No filters — return all
	result := FilterEntries(reg, "", "", "", "")
	assert.Len(t, result, 3)

	// Filter by category
	result = FilterEntries(reg, "cli", "", "", "")
	assert.Len(t, result, 2)

	// Filter by template
	result = FilterEntries(reg, "", "volumeclaim", "", "")
	assert.Len(t, result, 1)
	assert.Equal(t, "hacky", result[0].Name)

	// Filter by status
	result = FilterEntries(reg, "", "", "inactive", "")
	assert.Len(t, result, 1)
	assert.Equal(t, "demo-project", result[0].Name)

	// Filter by source
	result = FilterEntries(reg, "", "", "", "gitops")
	assert.Len(t, result, 1)
	assert.Equal(t, "demo-project", result[0].Name)

	// Combined filters
	result = FilterEntries(reg, "cli", "harvestervm", "active", "cli")
	assert.Len(t, result, 1)
	assert.Equal(t, "harvestervm-developer-martin", result[0].Name)

	// No match
	result = FilterEntries(reg, "cli", "", "inactive", "")
	assert.Len(t, result, 0)
}
