package snapshot

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Metadata holds information about a snapshot.
type Metadata struct {
	Name          string    `yaml:"name"`
	ClusterName   string    `yaml:"clusterName"`
	Provider      string    `yaml:"provider"`
	Kubecontext   string    `yaml:"kubecontext"`
	Namespaces    []string  `yaml:"namespaces,omitempty"`
	CreatedAt     time.Time `yaml:"createdAt"`
	TemplateFile  string    `yaml:"templateFile,omitempty"`
	ResourceCount int       `yaml:"resourceCount"`
}

// SaveMetadata writes metadata to metadata.yaml in the given directory.
func SaveMetadata(dir string, m *Metadata) error {
	data, err := yaml.Marshal(m)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}
	return os.WriteFile(filepath.Join(dir, "metadata.yaml"), data, 0644)
}

// LoadMetadata reads metadata.yaml from the given directory.
func LoadMetadata(dir string) (*Metadata, error) {
	data, err := os.ReadFile(filepath.Join(dir, "metadata.yaml"))
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata: %w", err)
	}
	var m Metadata
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("failed to parse metadata: %w", err)
	}
	return &m, nil
}
