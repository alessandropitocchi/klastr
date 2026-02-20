package snapshot

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/alepito/deploy-cluster/pkg/logger"
)

// snapshotsDir returns the base directory for all snapshots.
func snapshotsDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".deploy-cluster", "snapshots"), nil
}

// snapshotDir returns the directory for a specific snapshot.
func snapshotDir(name string) (string, error) {
	base, err := snapshotsDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, name), nil
}

// Save creates a new snapshot by exporting cluster resources to disk.
func Save(name, kubecontext, clusterName, provider, templateFile string, namespaces []string, log *logger.Logger) error {
	dir, err := snapshotDir(name)
	if err != nil {
		return err
	}

	// Check if snapshot already exists
	if _, err := os.Stat(dir); err == nil {
		return fmt.Errorf("snapshot %q already exists, delete it first", name)
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create snapshot directory: %w", err)
	}

	// Export resources
	log.Info("Exporting resources from cluster %q...\n", clusterName)
	count, err := ExportResources(dir, ExportOptions{
		Kubecontext: kubecontext,
		Namespaces:  namespaces,
		Log:         log,
	})
	if err != nil {
		// Clean up on failure
		os.RemoveAll(dir)
		return fmt.Errorf("failed to export resources: %w", err)
	}

	// Save metadata
	meta := &Metadata{
		Name:          name,
		ClusterName:   clusterName,
		Provider:      provider,
		Kubecontext:   kubecontext,
		Namespaces:    namespaces,
		CreatedAt:     time.Now(),
		TemplateFile:  templateFile,
		ResourceCount: count,
	}
	if err := SaveMetadata(dir, meta); err != nil {
		os.RemoveAll(dir)
		return fmt.Errorf("failed to save metadata: %w", err)
	}

	log.Success("Snapshot %q saved (%d resources)\n", name, count)
	return nil
}

// Restore applies a snapshot to the cluster.
func Restore(name, kubecontext string, dryRun bool, log *logger.Logger) error {
	dir, err := snapshotDir(name)
	if err != nil {
		return err
	}

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("snapshot %q not found", name)
	}

	meta, err := LoadMetadata(dir)
	if err != nil {
		return err
	}

	mode := ""
	if dryRun {
		mode = " (dry-run)"
	}
	log.Info("Restoring snapshot %q (%d resources)%s...\n", meta.Name, meta.ResourceCount, mode)

	if err := RestoreSnapshot(dir, RestoreOptions{
		Kubecontext: kubecontext,
		Log:         log,
		DryRun:      dryRun,
	}); err != nil {
		return fmt.Errorf("failed to restore snapshot: %w", err)
	}

	log.Success("Snapshot %q restored successfully%s\n", name, mode)
	return nil
}

// List returns metadata for all snapshots.
func List() ([]Metadata, error) {
	base, err := snapshotsDir()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(base); os.IsNotExist(err) {
		return nil, nil
	}

	entries, err := os.ReadDir(base)
	if err != nil {
		return nil, fmt.Errorf("failed to read snapshots directory: %w", err)
	}

	var snapshots []Metadata
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		meta, err := LoadMetadata(filepath.Join(base, entry.Name()))
		if err != nil {
			continue
		}
		snapshots = append(snapshots, *meta)
	}
	return snapshots, nil
}

// Delete removes a snapshot from disk.
func Delete(name string) error {
	dir, err := snapshotDir(name)
	if err != nil {
		return err
	}

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("snapshot %q not found", name)
	}

	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("failed to delete snapshot: %w", err)
	}
	return nil
}
