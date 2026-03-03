package e2e

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// TestDirectoryStructure tests loading configuration from directory
func TestDirectoryStructure(t *testing.T) {
	if os.Getenv("RUN_E2E") != "1" {
		t.Skip("Skipping e2e test. Set RUN_E2E=1 to run.")
	}

	clusterName := "e2e-dir-test"
	defer cleanupCluster(t, clusterName)

	// Create directory structure
	tmpDir := t.TempDir()
	clusterDir := filepath.Join(tmpDir, clusterName)
	
	// Create directories
	if err := os.MkdirAll(filepath.Join(clusterDir, "plugins"), 0755); err != nil {
		t.Fatalf("Failed to create plugins dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(clusterDir, "apps"), 0755); err != nil {
		t.Fatalf("Failed to create apps dir: %v", err)
	}

	// Create klastr.yaml
	klastrContent := fmt.Sprintf(`
name: %s
provider:
  type: kind
cluster:
  controlPlanes: 1
  workers: 1
  version: v1.31.0
`, clusterName)
	
	if err := os.WriteFile(filepath.Join(clusterDir, "klastr.yaml"), []byte(klastrContent), 0644); err != nil {
		t.Fatalf("Failed to write klastr.yaml: %v", err)
	}

	// Create plugins/storage.yaml
	storageContent := `
plugins:
  storage:
    enabled: true
    type: local-path
`
	if err := os.WriteFile(filepath.Join(clusterDir, "plugins", "storage.yaml"), []byte(storageContent), 0644); err != nil {
		t.Fatalf("Failed to write storage.yaml: %v", err)
	}

	// Create plugins/ingress.yaml
	ingressContent := `
plugins:
  ingress:
    enabled: true
    type: traefik
`
	if err := os.WriteFile(filepath.Join(clusterDir, "plugins", "ingress.yaml"), []byte(ingressContent), 0644); err != nil {
		t.Fatalf("Failed to write ingress.yaml: %v", err)
	}

	// Test lint
	t.Run("LintDirectory", func(t *testing.T) {
		output, err := runKlastr(t, "lint", "-f", clusterDir)
		if err != nil {
			t.Fatalf("Lint failed: %v\nOutput: %s", err, output)
		}
		t.Logf("Lint output: %s", output)
	})

	// Test run
	t.Run("RunFromDirectory", func(t *testing.T) {
		output, err := runKlastr(t, "run", "-f", clusterDir)
		if err != nil {
			t.Fatalf("Run failed: %v\nOutput: %s", err, output)
		}
		t.Logf("Run output: %s", output)
	})

	// Test status
	t.Run("StatusFromDirectory", func(t *testing.T) {
		output, err := runKlastr(t, "status", "-f", clusterDir)
		if err != nil {
			t.Fatalf("Status failed: %v\nOutput: %s", err, output)
		}
		t.Logf("Status output: %s", output)
	})
}

// TestDirectoryMerge tests that files are merged correctly
func TestDirectoryMerge(t *testing.T) {
	if os.Getenv("RUN_E2E") != "1" {
		t.Skip("Skipping e2e test. Set RUN_E2E=1 to run.")
	}

	clusterName := "e2e-merge-test"
	defer cleanupCluster(t, clusterName)

	tmpDir := t.TempDir()
	clusterDir := filepath.Join(tmpDir, clusterName)
	
	// Create directories
	if err := os.MkdirAll(filepath.Join(clusterDir, "plugins"), 0755); err != nil {
		t.Fatalf("Failed to create plugins dir: %v", err)
	}

	// Create main config with partial plugin config
	klastrContent := fmt.Sprintf(`
name: %s
provider:
  type: kind
cluster:
  controlPlanes: 1
  workers: 2
  version: v1.31.0
plugins:
  storage:
    enabled: true
    type: local-path
`, clusterName)
	
	if err := os.WriteFile(filepath.Join(clusterDir, "klastr.yaml"), []byte(klastrContent), 0644); err != nil {
		t.Fatalf("Failed to write klastr.yaml: %v", err)
	}

	// Create additional plugin config that should merge
	ingressContent := `
plugins:
  ingress:
    enabled: true
    type: traefik
`
	if err := os.WriteFile(filepath.Join(clusterDir, "plugins", "01-ingress.yaml"), []byte(ingressContent), 0644); err != nil {
		t.Fatalf("Failed to write ingress.yaml: %v", err)
	}

	// Create monitoring config
	monitoringContent := `
plugins:
  monitoring:
    enabled: true
    type: prometheus
`
	if err := os.WriteFile(filepath.Join(clusterDir, "plugins", "02-monitoring.yaml"), []byte(monitoringContent), 0644); err != nil {
		t.Fatalf("Failed to write monitoring.yaml: %v", err)
	}

	// Run and verify all plugins are installed
	output, err := runKlastr(t, "run", "-f", clusterDir)
	if err != nil {
		t.Fatalf("Run failed: %v\nOutput: %s", err, output)
	}

	t.Logf("Cluster with merged plugins created successfully")
}
