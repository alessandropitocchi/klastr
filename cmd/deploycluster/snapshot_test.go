package main

import (
	"strings"
	"testing"
)

func TestSnapshotCmd_Subcommands(t *testing.T) {
	expected := []string{"save", "restore", "list", "delete", "diff"}
	commands := snapshotCmd.Commands()

	registered := make(map[string]bool)
	for _, cmd := range commands {
		registered[cmd.Name()] = true
	}

	for _, name := range expected {
		if !registered[name] {
			t.Errorf("snapshot subcommand %q not registered", name)
		}
	}
}

func TestSnapshotSaveCmd_Flags(t *testing.T) {
	f := snapshotSaveCmd.Flags()

	tf := f.Lookup("template")
	if tf == nil {
		t.Fatal("save should have --template flag")
	}
	if tf.DefValue != "template.yaml" {
		t.Errorf("--template default = %q, want %q", tf.DefValue, "template.yaml")
	}
	if tf.Shorthand != "t" {
		t.Errorf("--template shorthand = %q, want %q", tf.Shorthand, "t")
	}

	nf := f.Lookup("namespace")
	if nf == nil {
		t.Fatal("save should have --namespace flag")
	}

	ef := f.Lookup("env")
	if ef == nil {
		t.Fatal("save should have --env flag")
	}

	es := f.Lookup("exclude-secrets")
	if es == nil {
		t.Fatal("save should have --exclude-secrets flag")
	}
	if es.DefValue != "false" {
		t.Errorf("--exclude-secrets default = %q, want %q", es.DefValue, "false")
	}
}

func TestSnapshotRestoreCmd_Flags(t *testing.T) {
	f := snapshotRestoreCmd.Flags()

	tf := f.Lookup("template")
	if tf == nil {
		t.Fatal("restore should have --template flag")
	}

	dr := f.Lookup("dry-run")
	if dr == nil {
		t.Fatal("restore should have --dry-run flag")
	}
	if dr.DefValue != "false" {
		t.Errorf("--dry-run default = %q, want %q", dr.DefValue, "false")
	}

	ef := f.Lookup("env")
	if ef == nil {
		t.Fatal("restore should have --env flag")
	}
}

func TestSnapshotSave_MissingTemplate(t *testing.T) {
	err := executeCommand("snapshot", "save", "test-snap", "--template", "nonexistent-template-xyz.yaml")
	if err == nil {
		t.Fatal("expected error for missing template")
	}
	if !strings.Contains(err.Error(), "failed to load template") {
		t.Errorf("error = %q, want it to contain 'failed to load template'", err.Error())
	}
}

func TestSnapshotRestore_MissingTemplate(t *testing.T) {
	err := executeCommand("snapshot", "restore", "test-snap", "--template", "nonexistent-template-xyz.yaml")
	if err == nil {
		t.Fatal("expected error for missing template")
	}
	if !strings.Contains(err.Error(), "failed to load template") {
		t.Errorf("error = %q, want it to contain 'failed to load template'", err.Error())
	}
}

func TestSnapshotSave_MissingName(t *testing.T) {
	err := executeCommand("snapshot", "save")
	if err == nil {
		t.Fatal("expected error for missing name argument")
	}
}

func TestSnapshotRestore_MissingName(t *testing.T) {
	err := executeCommand("snapshot", "restore")
	if err == nil {
		t.Fatal("expected error for missing name argument")
	}
}

func TestSnapshotDelete_MissingName(t *testing.T) {
	err := executeCommand("snapshot", "delete")
	if err == nil {
		t.Fatal("expected error for missing name argument")
	}
}

func TestSnapshotDiffCmd_Flags(t *testing.T) {
	f := snapshotDiffCmd.Flags()

	tf := f.Lookup("template")
	if tf == nil {
		t.Fatal("diff should have --template flag")
	}
	if tf.DefValue != "template.yaml" {
		t.Errorf("--template default = %q, want %q", tf.DefValue, "template.yaml")
	}

	ef := f.Lookup("env")
	if ef == nil {
		t.Fatal("diff should have --env flag")
	}
}

func TestSnapshotDiff_MissingName(t *testing.T) {
	err := executeCommand("snapshot", "diff")
	if err == nil {
		t.Fatal("expected error for missing name argument")
	}
}

func TestSnapshotDiff_MissingTemplate(t *testing.T) {
	err := executeCommand("snapshot", "diff", "test-snap", "--template", "nonexistent-template-xyz.yaml")
	if err == nil {
		t.Fatal("expected error for missing template")
	}
	if !strings.Contains(err.Error(), "failed to load template") {
		t.Errorf("error = %q, want it to contain 'failed to load template'", err.Error())
	}
}

func TestSnapshotDelete_NotFound(t *testing.T) {
	err := executeCommand("snapshot", "delete", "nonexistent-snapshot-xyz-test")
	if err == nil {
		t.Fatal("expected error for nonexistent snapshot")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %q, want it to contain 'not found'", err.Error())
	}
}
