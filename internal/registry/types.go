package registry

// ClaimRegistry represents the claims/registry.yaml file
type ClaimRegistry struct {
	APIVersion string       `yaml:"apiVersion" json:"apiVersion"`
	Kind       string       `yaml:"kind" json:"kind"`
	Claims     []ClaimEntry `yaml:"claims" json:"claims"`
}

// ClaimEntry represents a single claim in the registry
type ClaimEntry struct {
	Name       string `yaml:"name" json:"name"`
	Template   string `yaml:"template" json:"template"`
	Category   string `yaml:"category" json:"category"`
	Namespace  string `yaml:"namespace" json:"namespace"`
	CreatedAt  string `yaml:"createdAt" json:"createdAt"`
	CreatedBy  string `yaml:"createdBy" json:"createdBy"`
	Source     string `yaml:"source" json:"source"`
	Repository string `yaml:"repository" json:"repository"`
	Path       string `yaml:"path" json:"path"`
	Status     string `yaml:"status" json:"status"`
}
