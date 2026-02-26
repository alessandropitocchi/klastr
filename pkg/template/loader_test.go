package template

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoader_LoadFromDirectory(t *testing.T) {
	// Create temp directory structure
	tmpDir := t.TempDir()
	
	// Create klastr.yaml
	klastrContent := `
name: test-cluster
provider:
  type: kind
cluster:
  controlPlanes: 1
  workers: 2
  version: v1.31.0
`
	if err := os.WriteFile(filepath.Join(tmpDir, "klastr.yaml"), []byte(klastrContent), 0644); err != nil {
		t.Fatalf("Failed to write klastr.yaml: %v", err)
	}

	// Create plugins directory
	if err := os.MkdirAll(filepath.Join(tmpDir, "plugins"), 0755); err != nil {
		t.Fatalf("Failed to create plugins dir: %v", err)
	}

	// Create storage.yaml in plugins
	storageContent := `
plugins:
  storage:
    enabled: true
    type: local-path
`
	if err := os.WriteFile(filepath.Join(tmpDir, "plugins", "storage.yaml"), []byte(storageContent), 0644); err != nil {
		t.Fatalf("Failed to write storage.yaml: %v", err)
	}

	// Test loading
	loader := NewLoader()
	tmpl, err := loader.Load(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load from directory: %v", err)
	}

	// Verify merged config
	if tmpl.Name != "test-cluster" {
		t.Errorf("Name = %q, want %q", tmpl.Name, "test-cluster")
	}
	if tmpl.Provider.Type != "kind" {
		t.Errorf("Provider.Type = %q, want %q", tmpl.Provider.Type, "kind")
	}
	if tmpl.Cluster.ControlPlanes != 1 {
		t.Errorf("Cluster.ControlPlanes = %d, want 1", tmpl.Cluster.ControlPlanes)
	}
	if tmpl.Cluster.Workers != 2 {
		t.Errorf("Cluster.Workers = %d, want 2", tmpl.Cluster.Workers)
	}
	if tmpl.Plugins.Storage == nil || !tmpl.Plugins.Storage.Enabled {
		t.Error("Storage plugin should be enabled from plugins/storage.yaml")
	}
}

func TestLoader_MergeTemplates(t *testing.T) {
	loader := NewLoader()

	// Create base template
	base := &Template{
		Name: "base-cluster",
		Provider: ProviderTemplate{
			Type: "kind",
		},
		Cluster: ClusterTemplate{
			ControlPlanes: 1,
			Workers:       1,
			Version:       "v1.30.0",
		},
	}

	// Create overlay template
	overlay := &Template{
		Name: "overlay-cluster",
		Cluster: ClusterTemplate{
			Workers: 3,
			Version: "v1.31.0",
		},
		Plugins: PluginsTemplate{
			Storage: &StorageTemplate{
				Enabled: true,
				Type:    "local-path",
			},
		},
	}

	// Merge
	loader.mergeTemplates(base, overlay)

	// Verify merge results
	if base.Name != "overlay-cluster" {
		t.Errorf("Name should be overridden: got %q, want %q", base.Name, "overlay-cluster")
	}
	if base.Provider.Type != "kind" {
		t.Errorf("Provider.Type should be preserved: got %q", base.Provider.Type)
	}
	if base.Cluster.ControlPlanes != 1 {
		t.Errorf("ControlPlanes should be preserved: got %d", base.Cluster.ControlPlanes)
	}
	if base.Cluster.Workers != 3 {
		t.Errorf("Workers should be overridden: got %d, want 3", base.Cluster.Workers)
	}
	if base.Cluster.Version != "v1.31.0" {
		t.Errorf("Version should be overridden: got %q, want v1.31.0", base.Cluster.Version)
	}
	if base.Plugins.Storage == nil || !base.Plugins.Storage.Enabled {
		t.Error("Storage plugin should be merged from overlay")
	}
}

func TestLoader_LoadFromFile(t *testing.T) {
	tmpDir := t.TempDir()
	templateFile := filepath.Join(tmpDir, "template.yaml")

	content := `
name: file-test
provider:
  type: k3d
cluster:
  controlPlanes: 1
  workers: 0
  version: v1.31.0
`
	if err := os.WriteFile(templateFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write template: %v", err)
	}

	loader := NewLoader()
	tmpl, err := loader.Load(templateFile)
	if err != nil {
		t.Fatalf("Failed to load from file: %v", err)
	}

	if tmpl.Name != "file-test" {
		t.Errorf("Name = %q, want %q", tmpl.Name, "file-test")
	}
	if tmpl.Provider.Type != "k3d" {
		t.Errorf("Provider.Type = %q, want k3d", tmpl.Provider.Type)
	}
}

func TestLoader_LoadNonExistent(t *testing.T) {
	loader := NewLoader()
	_, err := loader.Load("/non/existent/path")
	if err == nil {
		t.Error("Expected error for non-existent path")
	}
}

func TestLoader_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	loader := NewLoader()
	_, err := loader.Load(tmpDir)
	if err == nil {
		t.Error("Expected error for empty directory")
	}
}
