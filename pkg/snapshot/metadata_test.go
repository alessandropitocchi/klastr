package snapshot

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSaveLoadMetadata_RoundTrip(t *testing.T) {
	dir := t.TempDir()

	original := &Metadata{
		Name:          "test-snap",
		ClusterName:   "my-cluster",
		Provider:      "kind",
		Kubecontext:   "kind-my-cluster",
		Namespaces:    []string{"default", "app"},
		CreatedAt:     time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
		TemplateFile:  "template.yaml",
		ResourceCount: 42,
	}

	if err := SaveMetadata(dir, original); err != nil {
		t.Fatalf("SaveMetadata() error = %v", err)
	}

	loaded, err := LoadMetadata(dir)
	if err != nil {
		t.Fatalf("LoadMetadata() error = %v", err)
	}

	if loaded.Name != original.Name {
		t.Errorf("Name = %q, want %q", loaded.Name, original.Name)
	}
	if loaded.ClusterName != original.ClusterName {
		t.Errorf("ClusterName = %q, want %q", loaded.ClusterName, original.ClusterName)
	}
	if loaded.Provider != original.Provider {
		t.Errorf("Provider = %q, want %q", loaded.Provider, original.Provider)
	}
	if loaded.Kubecontext != original.Kubecontext {
		t.Errorf("Kubecontext = %q, want %q", loaded.Kubecontext, original.Kubecontext)
	}
	if loaded.ResourceCount != original.ResourceCount {
		t.Errorf("ResourceCount = %d, want %d", loaded.ResourceCount, original.ResourceCount)
	}
	if loaded.TemplateFile != original.TemplateFile {
		t.Errorf("TemplateFile = %q, want %q", loaded.TemplateFile, original.TemplateFile)
	}
	if len(loaded.Namespaces) != len(original.Namespaces) {
		t.Errorf("Namespaces len = %d, want %d", len(loaded.Namespaces), len(original.Namespaces))
	}
}

func TestSaveMetadata_EmptyNamespaces(t *testing.T) {
	dir := t.TempDir()

	original := &Metadata{
		Name:        "no-ns-snap",
		ClusterName: "test",
		Provider:    "kind",
		Kubecontext: "kind-test",
		CreatedAt:   time.Now(),
	}

	if err := SaveMetadata(dir, original); err != nil {
		t.Fatalf("SaveMetadata() error = %v", err)
	}

	loaded, err := LoadMetadata(dir)
	if err != nil {
		t.Fatalf("LoadMetadata() error = %v", err)
	}

	if loaded.Namespaces != nil {
		t.Errorf("Namespaces = %v, want nil", loaded.Namespaces)
	}
}

func TestLoadMetadata_FileNotFound(t *testing.T) {
	_, err := LoadMetadata("/nonexistent/path")
	if err == nil {
		t.Fatal("expected error for missing metadata file")
	}
}

func TestLoadMetadata_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "metadata.yaml"), []byte("{{{{invalid"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadMetadata(dir)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}
