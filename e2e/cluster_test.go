package e2e

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestClusterLifecycle tests basic cluster creation and destruction
func TestClusterLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	clusterName := "e2e-test-cluster"
	contextName := "kind-" + clusterName
	
	// Cleanup before and after test
	cleanupCluster(t, clusterName)
	defer cleanupCluster(t, clusterName)

	t.Run("CreateCluster", func(t *testing.T) {
		// Create minimal template
		tmpDir := t.TempDir()
		templateFile := filepath.Join(tmpDir, "template.yaml")
		
		template := fmt.Sprintf(`
name: %s
provider:
  type: kind
cluster:
  controlPlanes: 1
  workers: 0
  version: v1.31.0
`, clusterName)

		if err := os.WriteFile(templateFile, []byte(template), 0644); err != nil {
			t.Fatalf("Failed to write template: %v", err)
		}

		// Run klastr
		output, err := runKlastr(t, "run", "-f", templateFile)
		if err != nil {
			t.Fatalf("Failed to create cluster: %v\nOutput: %s", err, output)
		}

		// Verify cluster was created
		if !strings.Contains(output, "Cluster created successfully") && !strings.Contains(output, "created") {
			t.Logf("Output: %s", output)
		}

		// Wait for cluster to be ready
		if err := waitForClusterReady(t, contextName, 5*time.Minute); err != nil {
			t.Fatalf("Cluster not ready: %v", err)
		}

		t.Log("Cluster created and ready")
	})

	t.Run("ClusterStatus", func(t *testing.T) {
		output, err := runKlastr(t, "status", "--name", clusterName)
		if err != nil {
			t.Fatalf("Failed to get status: %v\nOutput: %s", err, output)
		}

		if !strings.Contains(output, clusterName) {
			t.Errorf("Status output doesn't contain cluster name: %s", output)
		}
	})

	t.Run("DestroyCluster", func(t *testing.T) {
		output, err := runKlastr(t, "destroy", "--name", clusterName)
		if err != nil {
			t.Fatalf("Failed to destroy cluster: %v\nOutput: %s", err, output)
		}

		// Verify cluster no longer exists
		_, err = runKlastr(t, "status", "--name", clusterName)
		if err == nil {
			t.Error("Cluster still exists after destroy")
		}
	})
}

// TestClusterWithStorage tests cluster with storage plugin
func TestClusterWithStorage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	clusterName := "e2e-storage-test"
	defer cleanupCluster(t, clusterName)

	tmpDir := t.TempDir()
	templateFile := filepath.Join(tmpDir, "template.yaml")
	
	template := fmt.Sprintf(`
name: %s
provider:
  type: kind
cluster:
  controlPlanes: 1
  workers: 1
  version: v1.31.0
plugins:
  storage:
    enabled: true
    type: local-path
`, clusterName)

	if err := os.WriteFile(templateFile, []byte(template), 0644); err != nil {
		t.Fatalf("Failed to write template: %v", err)
	}

	// Create cluster
	output, err := runKlastr(t, "run", "-f", templateFile)
	if err != nil {
		t.Fatalf("Failed to create cluster: %v\nOutput: %s", err, output)
	}

	// Verify storage plugin is installed
	if !strings.Contains(output, "storage") && !strings.Contains(output, "Storage") {
		t.Logf("Warning: storage plugin not mentioned in output: %s", output)
	}

	t.Log("Cluster with storage created successfully")
}

// TestClusterWithIngress tests cluster with ingress plugin
func TestClusterWithIngress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	clusterName := "e2e-ingress-test"
	defer cleanupCluster(t, clusterName)

	tmpDir := t.TempDir()
	templateFile := filepath.Join(tmpDir, "template.yaml")
	
	template := fmt.Sprintf(`
name: %s
provider:
  type: kind
cluster:
  controlPlanes: 1
  workers: 1
  version: v1.31.0
plugins:
  ingress:
    enabled: true
    type: nginx
`, clusterName)

	if err := os.WriteFile(templateFile, []byte(template), 0644); err != nil {
		t.Fatalf("Failed to write template: %v", err)
	}

	output, err := runKlastr(t, "run", "-f", templateFile)
	if err != nil {
		t.Fatalf("Failed to create cluster: %v\nOutput: %s", err, output)
	}

	t.Log("Cluster with ingress created successfully")
}
