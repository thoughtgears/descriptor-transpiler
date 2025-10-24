package main

import (
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
)

// AppSpec represents the application configuration from app.yaml
type AppSpec struct {
	APIVersion   string                 `yaml:"apiVersion"`
	Name         string                 `yaml:"name"`
	Tribe        string                 `yaml:"tribe"`
	Team         string                 `yaml:"team"`
	Type         string                 `yaml:"type"`
	Public       bool                   `yaml:"public"`
	Size         string                 `yaml:"size"`
	Dependencies map[string]interface{} `yaml:"dependencies,omitempty"`
	Components   []interface{}          `yaml:"components,omitempty"`
}

// LoadAppSpec reads and parses the app.yaml file
func LoadAppSpec(path string) (*AppSpec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read app.yaml: %w", err)
	}

	var spec AppSpec

	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("failed to parse app.yaml: %w", err)
	}

	if err := spec.validateSize(); err != nil {
		return nil, err
	}

	return &spec, nil
}

// validateSize checks if the size is valid and normalizes it
func (a *AppSpec) validateSize() error {
	validSizes := map[string]bool{
		"tiny":   true,
		"small":  true,
		"medium": true,
		"large":  true,
	}

	// Default to tiny if not specified
	if a.Size == "" {
		a.Size = "tiny"
		return nil
	}

	if !validSizes[a.Size] {
		return fmt.Errorf("invalid size '%s': must be one of: tiny, small, medium, large", a.Size)
	}

	return nil
}

// ToChartSize converts the string size to the Size type from charts.go
func (a *AppSpec) ToChartSize() Size {
	return Size(a.Size)
}

// BuildLabels creates Kubernetes labels from AppSpec
func (a *AppSpec) BuildLabels() map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":      a.Name,
		"app.kubernetes.io/team":      a.Team,
		"app.kubernetes.io/tribe":     a.Tribe,
		"app.kubernetes.io/type":      a.Type,
		"app.kubernetes.io/component": "web",
	}
}
