package main

import (
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
)

type Dependency struct {
	Type    string                 `yaml:"type"`
	Size    string                 `yaml:"size,omitempty"`
	Version int                    `yaml:"version,omitempty"`
	Config  map[string]interface{} `yaml:"config,omitempty"` // For future expansion
}

// AppSpec represents the application configuration from app.yaml
type AppSpec struct {
	APIVersion   string                `yaml:"apiVersion"`
	Name         string                `yaml:"name"`
	Tribe        string                `yaml:"tribe"`
	Team         string                `yaml:"team"`
	Type         string                `yaml:"type"`
	Public       bool                  `yaml:"public"`
	Size         string                `yaml:"size"`
	Dependencies map[string]Dependency `yaml:"dependencies,omitempty"`
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

func (a *AppSpec) GetDependency(name string) (Dependency, bool) {
	dep, exists := a.Dependencies[name]
	return dep, exists
}

func (a *AppSpec) HasDatabase() bool {
	db, exists := a.GetDependency("database")
	return exists && db.Type == "postgres"
}

func (a *AppSpec) GetDatabaseConfig() (size string, version int, err error) {
	db, exists := a.GetDependency("database")
	if !exists {
		return "", 0, fmt.Errorf("no database dependency found")
	}

	size = db.Size
	if size == "" {
		size = "small" // default
	}

	version = db.Version
	if version == 0 {
		version = 17 // default
	}

	return size, version, nil
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
