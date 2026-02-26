// Package e2e provides end-to-end tests for klastr.
// These tests create real kind clusters and may take several minutes to complete.
// Run with: go test -v ./e2e/... -timeout 30m
package e2e

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const (
	testTimeout     = 30 * time.Minute
	clusterTimeout  = 10 * time.Minute
	pluginTimeout   = 5 * time.Minute
)

// klastrBinary returns the path to the built binary
func klastrBinary() string {
	wd, _ := os.Getwd()
	return filepath.Join(filepath.Dir(wd), "klastr")
}

// TestMain builds the binary before running e2e tests
func TestMain(m *testing.M) {
	// Get project root
	wd, _ := os.Getwd()
	projectRoot := filepath.Dir(wd)
	
	// Build klastr binary
	binaryPath := filepath.Join(projectRoot, "klastr")
	cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/deploycluster")
	cmd.Dir = projectRoot
	if output, err := cmd.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to build klastr: %v\n%s\n", err, output)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

// runKlastr executes a klastr command and returns output
func runKlastr(t *testing.T, args ...string) (string, error) {
	t.Helper()
	
	ctx, cancel := context.WithTimeout(context.Background(), clusterTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, klastrBinary(), args...)
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		return string(output), fmt.Errorf("klastr %s failed: %w\nOutput: %s", 
			strings.Join(args, " "), err, output)
	}
	
	return string(output), nil
}

// cleanupCluster destroys a cluster if it exists
func cleanupCluster(t *testing.T, name string) {
	t.Helper()
	
	// Try to destroy, ignore errors (cluster might not exist)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, klastrBinary(), "destroy", "--name", name)
	cmd.Run() // Ignore error
	
	// Also try with kind directly as fallback
	ctx2, cancel2 := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel2()
	
	cmd2 := exec.CommandContext(ctx2, "kind", "delete", "cluster", "--name", name)
	cmd2.Run()
}

// waitForClusterReady waits for cluster to be ready
func waitForClusterReady(t *testing.T, kubeContext string, timeout time.Duration) error {
	t.Helper()
	
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for cluster to be ready")
		default:
		}

		cmd := exec.CommandContext(ctx, "kubectl", "--context", kubeContext, "get", "nodes")
		output, err := cmd.CombinedOutput()
		if err == nil && strings.Contains(string(output), "Ready") {
			return nil
		}

		time.Sleep(5 * time.Second)
	}
}
