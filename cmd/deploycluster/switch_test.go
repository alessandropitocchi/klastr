package main

import (
	"os/exec"
	"strings"
	"testing"
)

type switchCapturedCmd struct {
	name string
	args []string
}

func setupSwitchFakeExec(t *testing.T, kindClusters string, currentCtx string) *[]switchCapturedCmd {
	t.Helper()
	var cmds []switchCapturedCmd
	orig := switchExecCommand
	switchExecCommand = func(name string, args ...string) *exec.Cmd {
		cmds = append(cmds, switchCapturedCmd{name, args})
		// Return appropriate output based on command
		if name == "kind" && len(args) > 0 && args[0] == "get" {
			return exec.Command("echo", kindClusters)
		}
		if name == "kubectl" && len(args) > 1 && args[1] == "current-context" {
			return exec.Command("echo", currentCtx)
		}
		// Default: succeed silently
		return exec.Command("true")
	}
	t.Cleanup(func() { switchExecCommand = orig })
	return &cmds
}

func TestSwitchCmd_ListClusters(t *testing.T) {
	setupSwitchFakeExec(t, "my-cluster\ndev-cluster", "kind-dev-cluster")

	err := executeCommand("switch")
	if err != nil {
		t.Fatalf("switch (list) should succeed: %v", err)
	}
}

func TestSwitchCmd_ListEmpty(t *testing.T) {
	setupSwitchFakeExec(t, "", "")

	err := executeCommand("switch")
	if err != nil {
		t.Fatalf("switch (list empty) should succeed: %v", err)
	}
}

func TestSwitchCmd_SwitchToCluster(t *testing.T) {
	cmds := setupSwitchFakeExec(t, "my-cluster\ndev-cluster", "kind-dev-cluster")

	err := executeCommand("switch", "my-cluster")
	if err != nil {
		t.Fatalf("switch to cluster should succeed: %v", err)
	}

	// Should have: kind get clusters + kubectl config use-context
	found := false
	for _, cmd := range *cmds {
		if cmd.name == "kubectl" && len(cmd.args) >= 3 && cmd.args[1] == "use-context" {
			if cmd.args[2] != "kind-my-cluster" {
				t.Errorf("use-context arg = %q, want %q", cmd.args[2], "kind-my-cluster")
			}
			found = true
		}
	}
	if !found {
		t.Error("should call kubectl config use-context")
	}
}

func TestSwitchCmd_ClusterNotFound(t *testing.T) {
	setupSwitchFakeExec(t, "my-cluster\ndev-cluster", "kind-dev-cluster")

	err := executeCommand("switch", "nonexistent")
	if err == nil {
		t.Fatal("switch to nonexistent cluster should fail")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found', got: %v", err)
	}
}

func TestSwitchCmd_TooManyArgs(t *testing.T) {
	setupSwitchFakeExec(t, "", "")

	err := executeCommand("switch", "a", "b")
	if err == nil {
		t.Fatal("switch with 2 args should fail")
	}
}
