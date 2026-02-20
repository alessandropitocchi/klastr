package snapshot

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSnapshotDelete_NotFound(t *testing.T) {
	err := Delete("nonexistent-snapshot-xyz")
	if err == nil {
		t.Fatal("expected error for nonexistent snapshot")
	}
}

func TestSnapshotList_EmptyDir(t *testing.T) {
	snapshots, err := List()
	if err != nil {
		// If no snapshots dir exists, should return nil
		t.Fatalf("List() error = %v", err)
	}
	// May or may not be nil depending on whether .deploy-cluster/snapshots exists
	_ = snapshots
}

func TestSnapshotDir(t *testing.T) {
	dir, err := snapshotDir("test-snap")
	if err != nil {
		t.Fatalf("snapshotDir() error = %v", err)
	}
	if !filepath.IsAbs(dir) {
		t.Errorf("expected absolute path, got %q", dir)
	}
	if filepath.Base(dir) != "test-snap" {
		t.Errorf("expected dir to end with 'test-snap', got %q", dir)
	}
}

func TestSnapshotDelete_Existing(t *testing.T) {
	// Create a temp snapshot
	base, err := snapshotsDir()
	if err != nil {
		t.Fatal(err)
	}

	name := "test-delete-snap"
	dir := filepath.Join(base, name)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	// Write metadata
	meta := &Metadata{Name: name, ClusterName: "test"}
	if err := SaveMetadata(dir, meta); err != nil {
		t.Fatal(err)
	}

	// Delete should succeed
	if err := Delete(name); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify deleted
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Error("snapshot directory should be deleted")
	}
}
