package registry

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

const (
	DefaultAPIVersion = "claim-registry.io/v1alpha1"
	DefaultKind       = "ClaimRegistry"
)

// ParseData parses raw YAML bytes into a ClaimRegistry.
func ParseData(data []byte) (*ClaimRegistry, error) {
	var reg ClaimRegistry
	if err := yaml.Unmarshal(data, &reg); err != nil {
		return nil, fmt.Errorf("parsing registry data: %w", err)
	}
	return &reg, nil
}

// FindEntry returns a pointer to the claim entry with the given name, or nil.
func FindEntry(reg *ClaimRegistry, name string) *ClaimEntry {
	for i, e := range reg.Claims {
		if e.Name == name {
			return &reg.Claims[i]
		}
	}
	return nil
}

// FilterEntries returns entries matching the given filters.
// Empty strings are treated as wildcards (match all).
func FilterEntries(reg *ClaimRegistry, category, template, status, source string) []ClaimEntry {
	var result []ClaimEntry
	for _, e := range reg.Claims {
		if category != "" && e.Category != category {
			continue
		}
		if template != "" && e.Template != template {
			continue
		}
		if status != "" && e.Status != status {
			continue
		}
		if source != "" && e.Source != source {
			continue
		}
		result = append(result, e)
	}
	return result
}
