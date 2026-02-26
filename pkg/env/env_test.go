package env

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alessandropitocchi/deploy-cluster/pkg/template"
)

func TestManager_Create(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManager(tmpDir)

	// Create environment
	if err := manager.Create("dev", "../../"); err != nil {
		t.Fatalf("Failed to create environment: %v", err)
	}

	// Verify directory was created
	envDir := filepath.Join(tmpDir, "environments", "dev")
	if _, err := os.Stat(envDir); os.IsNotExist(err) {
		t.Error("Environment directory was not created")
	}

	// Verify overlay.yaml was created
	overlayFile := filepath.Join(envDir, "overlay.yaml")
	if _, err := os.Stat(overlayFile); os.IsNotExist(err) {
		t.Error("overlay.yaml was not created")
	}

	// Read and verify content
	data, err := os.ReadFile(overlayFile)
	if err != nil {
		t.Fatalf("Failed to read overlay.yaml: %v", err)
	}

	content := string(data)
	if !contains(content, "name: dev") {
		t.Error("overlay.yaml missing name field")
	}
	if !contains(content, "base: ../../") {
		t.Error("overlay.yaml missing base field")
	}
}

func TestManager_List(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManager(tmpDir)

	// Create environments directory and some environments
	envDir := filepath.Join(tmpDir, "environments")
	if err := os.MkdirAll(envDir, 0755); err != nil {
		t.Fatalf("Failed to create environments dir: %v", err)
	}

	// Create dev environment
	if err := os.MkdirAll(filepath.Join(envDir, "dev"), 0755); err != nil {
		t.Fatalf("Failed to create dev env: %v", err)
	}

	// Create prod environment
	if err := os.MkdirAll(filepath.Join(envDir, "prod"), 0755); err != nil {
		t.Fatalf("Failed to create prod env: %v", err)
	}

	// List environments
	envs, err := manager.List()
	if err != nil {
		t.Fatalf("Failed to list environments: %v", err)
	}

	if len(envs) != 2 {
		t.Errorf("Expected 2 environments, got %d", len(envs))
	}

	// Check that both environments are listed
	hasDev := false
	hasProd := false
	for _, env := range envs {
		if env == "dev" {
			hasDev = true
		}
		if env == "prod" {
			hasProd = true
		}
	}

	if !hasDev {
		t.Error("dev environment not found in list")
	}
	if !hasProd {
		t.Error("prod environment not found in list")
	}
}

func TestManager_ListEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManager(tmpDir)

	envs, err := manager.List()
	if err != nil {
		t.Fatalf("Failed to list environments: %v", err)
	}

	if len(envs) != 0 {
		t.Errorf("Expected 0 environments, got %d", len(envs))
	}
}

func TestManager_Load(t *testing.T) {
	tmpDir := t.TempDir()

	// Create base configuration
	baseDir := filepath.Join(tmpDir, "base")
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		t.Fatalf("Failed to create base dir: %v", err)
	}

	baseConfig := `
name: myapp
provider:
  type: kind
cluster:
  controlPlanes: 1
  workers: 1
  version: v1.31.0
`
	if err := os.WriteFile(filepath.Join(baseDir, "klastr.yaml"), []byte(baseConfig), 0644); err != nil {
		t.Fatalf("Failed to write base config: %v", err)
	}

	// Create environment with overlay
	envDir := filepath.Join(tmpDir, "environments", "prod")
	if err := os.MkdirAll(envDir, 0755); err != nil {
		t.Fatalf("Failed to create env dir: %v", err)
	}

	overlayConfig := `
name: production
base: ../../base
patches:
  - target: name
    value: myapp-prod
  - target: cluster.workers
    value: 5
`
	if err := os.WriteFile(filepath.Join(envDir, "overlay.yaml"), []byte(overlayConfig), 0644); err != nil {
		t.Fatalf("Failed to write overlay: %v", err)
	}

	// Load environment
	manager := NewManager(tmpDir)
	cfg, err := manager.Load("prod")
	if err != nil {
		t.Fatalf("Failed to load environment: %v", err)
	}

	// Verify patches were applied
	if cfg.Name != "myapp-prod" {
		t.Errorf("Name = %q, want %q", cfg.Name, "myapp-prod")
	}
	if cfg.Cluster.Workers != 5 {
		t.Errorf("Workers = %d, want 5", cfg.Cluster.Workers)
	}
	if cfg.Cluster.ControlPlanes != 1 {
		t.Errorf("ControlPlanes = %d, want 1", cfg.Cluster.ControlPlanes)
	}
	if cfg.Provider.Type != "kind" {
		t.Errorf("Provider.Type = %q, want kind", cfg.Provider.Type)
	}
}

func TestManager_LoadNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManager(tmpDir)

	_, err := manager.Load("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent environment")
	}
}

func TestPatchCluster(t *testing.T) {
	manager := NewManager(".")

	tests := []struct {
		name     string
		path     []string
		value    interface{}
		expected interface{}
		getter   func(c template.ClusterTemplate) interface{}
	}{
		{
			name:     "controlPlanes",
			path:     []string{"controlPlanes"},
			value:    3,
			expected: 3,
			getter:   func(c template.ClusterTemplate) interface{} { return c.ControlPlanes },
		},
		{
			name:     "workers",
			path:     []string{"workers"},
			value:    5,
			expected: 5,
			getter:   func(c template.ClusterTemplate) interface{} { return c.Workers },
		},
		{
			name:     "version",
			path:     []string{"version"},
			value:    "v1.32.0",
			expected: "v1.32.0",
			getter:   func(c template.ClusterTemplate) interface{} { return c.Version },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cluster := template.ClusterTemplate{
				ControlPlanes: 1,
				Workers:       1,
				Version:       "v1.31.0",
			}

			err := manager.patchCluster(&cluster, tt.path, tt.value)
			if err != nil {
				t.Fatalf("patchCluster failed: %v", err)
			}

			got := tt.getter(cluster)
			if got != tt.expected {
				t.Errorf("got %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestPatchProvider(t *testing.T) {
	manager := NewManager(".")

	tests := []struct {
		name     string
		path     []string
		value    interface{}
		expected interface{}
		getter   func(p template.ProviderTemplate) interface{}
	}{
		{
			name:     "type",
			path:     []string{"type"},
			value:    "existing",
			expected: "existing",
			getter:   func(p template.ProviderTemplate) interface{} { return p.Type },
		},
		{
			name:     "context",
			path:     []string{"context"},
			value:    "prod-cluster",
			expected: "prod-cluster",
			getter:   func(p template.ProviderTemplate) interface{} { return p.Context },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := template.ProviderTemplate{
				Type: "kind",
			}

			err := manager.patchProvider(&provider, tt.path, tt.value)
			if err != nil {
				t.Fatalf("patchProvider failed: %v", err)
			}

			got := tt.getter(provider)
			if got != tt.expected {
				t.Errorf("got %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestPatchPlugins(t *testing.T) {
	manager := NewManager(".")

	tests := []struct {
		name     string
		path     []string
		value    interface{}
		check    func(p template.PluginsTemplate) bool
	}{
		{
			name:  "storage enabled",
			path:  []string{"storage", "enabled"},
			value: true,
			check: func(p template.PluginsTemplate) bool {
				return p.Storage != nil && p.Storage.Enabled
			},
		},
		{
			name:  "ingress enabled",
			path:  []string{"ingress", "enabled"},
			value: true,
			check: func(p template.PluginsTemplate) bool {
				return p.Ingress != nil && p.Ingress.Enabled
			},
		},
		{
			name:  "monitoring enabled",
			path:  []string{"monitoring", "enabled"},
			value: true,
			check: func(p template.PluginsTemplate) bool {
				return p.Monitoring != nil && p.Monitoring.Enabled
			},
		},
		{
			name:  "dashboard enabled",
			path:  []string{"dashboard", "enabled"},
			value: true,
			check: func(p template.PluginsTemplate) bool {
				return p.Dashboard != nil && p.Dashboard.Enabled
			},
		},
		{
			name:  "certManager enabled",
			path:  []string{"certManager", "enabled"},
			value: true,
			check: func(p template.PluginsTemplate) bool {
				return p.CertManager != nil && p.CertManager.Enabled
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugins := template.PluginsTemplate{}

			err := manager.patchPlugins(&plugins, tt.path, tt.value)
			if err != nil {
				t.Fatalf("patchPlugins failed: %v", err)
			}

			if !tt.check(plugins) {
				t.Errorf("patch not applied correctly")
			}
		})
	}
}

func TestPatchSnapshot(t *testing.T) {
	manager := NewManager(".")

	tests := []struct {
		name     string
		path     []string
		value    interface{}
		expected interface{}
		getter   func(s template.SnapshotConfig) interface{}
	}{
		{
			name:     "enabled",
			path:     []string{"enabled"},
			value:    true,
			expected: true,
			getter:   func(s template.SnapshotConfig) interface{} { return s.Enabled },
		},
		{
			name:     "bucket",
			path:     []string{"bucket"},
			value:    "my-backups",
			expected: "my-backups",
			getter:   func(s template.SnapshotConfig) interface{} { return s.Bucket },
		},
		{
			name:     "prefix",
			path:     []string{"prefix"},
			value:    "clusters/prod/",
			expected: "clusters/prod/",
			getter:   func(s template.SnapshotConfig) interface{} { return s.Prefix },
		},
		{
			name:     "region",
			path:     []string{"region"},
			value:    "us-west-2",
			expected: "us-west-2",
			getter:   func(s template.SnapshotConfig) interface{} { return s.Region },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			snapshot := template.SnapshotConfig{}

			err := manager.patchSnapshot(&snapshot, tt.path, tt.value)
			if err != nil {
				t.Fatalf("patchSnapshot failed: %v", err)
			}

			got := tt.getter(snapshot)
			if got != tt.expected {
				t.Errorf("got %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestApplyPatch(t *testing.T) {
	manager := NewManager(".")

	tests := []struct {
		name     string
		target   string
		value    interface{}
		verify   func(cfg *template.Template) bool
	}{
		{
			name:   "name",
			target: "name",
			value:  "new-name",
			verify: func(cfg *template.Template) bool { return cfg.Name == "new-name" },
		},
		{
			name:   "cluster.workers",
			target: "cluster.workers",
			value:  10,
			verify: func(cfg *template.Template) bool { return cfg.Cluster.Workers == 10 },
		},
		{
			name:   "provider.type",
			target: "provider.type",
			value:  "k3d",
			verify: func(cfg *template.Template) bool { return cfg.Provider.Type == "k3d" },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &template.Template{
				Name: "test",
				Provider: template.ProviderTemplate{
					Type: "kind",
				},
				Cluster: template.ClusterTemplate{
					ControlPlanes: 1,
					Workers:       1,
					Version:       "v1.31.0",
				},
			}

			patch := Patch{
				Target: tt.target,
				Value:  tt.value,
			}

			err := manager.applyPatch(cfg, patch)
			if err != nil {
				t.Fatalf("applyPatch failed: %v", err)
			}

			if !tt.verify(cfg) {
				t.Errorf("patch not applied correctly for target %s", tt.target)
			}
		})
	}
}

func TestApplyPatchInvalid(t *testing.T) {
	manager := NewManager(".")

	cfg := &template.Template{
		Name: "test",
	}

	// Test empty target
	patch := Patch{
		Target: "",
		Value:  "value",
	}
	err := manager.applyPatch(cfg, patch)
	if err == nil {
		t.Error("Expected error for empty target")
	}

	// Test unknown target
	patch = Patch{
		Target: "unknown.field",
		Value:  "value",
	}
	err = manager.applyPatch(cfg, patch)
	if err == nil {
		t.Error("Expected error for unknown target")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
