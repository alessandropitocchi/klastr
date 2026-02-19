package main

import (
	"os/exec"
	"strings"
	"testing"
)

func TestCheckCmd_AllFound(t *testing.T) {
	origLookPath := checkLookPath
	origExecCommand := checkExecCommand
	defer func() {
		checkLookPath = origLookPath
		checkExecCommand = origExecCommand
	}()

	checkLookPath = func(file string) (string, error) {
		return "/usr/local/bin/" + file, nil
	}
	checkExecCommand = func(name string, args ...string) *exec.Cmd {
		return exec.Command("echo", "v1.0.0")
	}

	err := executeCommand("check")
	if err != nil {
		t.Fatalf("check should succeed when all tools found: %v", err)
	}
}

func TestCheckCmd_MissingTool(t *testing.T) {
	origLookPath := checkLookPath
	origExecCommand := checkExecCommand
	defer func() {
		checkLookPath = origLookPath
		checkExecCommand = origExecCommand
	}()

	checkLookPath = func(file string) (string, error) {
		if file == "kind" {
			return "", &exec.Error{Name: "kind", Err: exec.ErrNotFound}
		}
		return "/usr/local/bin/" + file, nil
	}
	checkExecCommand = func(name string, args ...string) *exec.Cmd {
		return exec.Command("echo", "v1.0.0")
	}

	err := executeCommand("check")
	if err == nil {
		t.Fatal("check should fail when a tool is missing")
	}
	if !strings.Contains(err.Error(), "missing") {
		t.Errorf("error should mention missing prerequisites, got: %v", err)
	}
}

func TestGetVersion_Clean(t *testing.T) {
	origExecCommand := checkExecCommand
	defer func() { checkExecCommand = origExecCommand }()

	tests := []struct {
		name     string
		output   string
		expected string
	}{
		{"docker", "Docker version 27.5.1, build abc123", "27.5.1"},
		{"kind", "kind v0.25.0 go1.23.4 darwin/arm64", "v0.25.0 go1.23.4 darwin/arm64"},
		{"kubectl", "Client Version: v1.31.4", "v1.31.4"},
		{"helm", "v3.16.3+gc4e3792", "v3.16.3+gc4e3792"},
		{"simple", "v1.0.0", "v1.0.0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkExecCommand = func(name string, args ...string) *exec.Cmd {
				return exec.Command("echo", tt.output)
			}
			got := getVersion([]string{"test", "--version"})
			if got != tt.expected {
				t.Errorf("getVersion() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestGetVersion_Error(t *testing.T) {
	origExecCommand := checkExecCommand
	defer func() { checkExecCommand = origExecCommand }()

	checkExecCommand = func(name string, args ...string) *exec.Cmd {
		return exec.Command("false")
	}

	got := getVersion([]string{"nonexistent", "--version"})
	if got != "(unknown version)" {
		t.Errorf("getVersion() = %q, want %q", got, "(unknown version)")
	}
}
