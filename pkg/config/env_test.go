package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadEnvFile_BasicParsing(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")

	content := `MY_TEST_VAR_1=hello
MY_TEST_VAR_2=world
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Clean up after test
	defer os.Unsetenv("MY_TEST_VAR_1")
	defer os.Unsetenv("MY_TEST_VAR_2")

	if err := LoadEnvFile(path); err != nil {
		t.Fatalf("LoadEnvFile() error: %v", err)
	}

	if got := os.Getenv("MY_TEST_VAR_1"); got != "hello" {
		t.Errorf("MY_TEST_VAR_1 = %q, want %q", got, "hello")
	}
	if got := os.Getenv("MY_TEST_VAR_2"); got != "world" {
		t.Errorf("MY_TEST_VAR_2 = %q, want %q", got, "world")
	}
}

func TestLoadEnvFile_SkipsCommentsAndEmptyLines(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")

	content := `# This is a comment
MY_TEST_VAR_COMMENT=value

# Another comment

`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	defer os.Unsetenv("MY_TEST_VAR_COMMENT")

	if err := LoadEnvFile(path); err != nil {
		t.Fatalf("LoadEnvFile() error: %v", err)
	}

	if got := os.Getenv("MY_TEST_VAR_COMMENT"); got != "value" {
		t.Errorf("MY_TEST_VAR_COMMENT = %q, want %q", got, "value")
	}
}

func TestLoadEnvFile_RemovesQuotes(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")

	content := `MY_TEST_VAR_DQ="double quoted"
MY_TEST_VAR_SQ='single quoted'
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	defer os.Unsetenv("MY_TEST_VAR_DQ")
	defer os.Unsetenv("MY_TEST_VAR_SQ")

	if err := LoadEnvFile(path); err != nil {
		t.Fatalf("LoadEnvFile() error: %v", err)
	}

	if got := os.Getenv("MY_TEST_VAR_DQ"); got != "double quoted" {
		t.Errorf("MY_TEST_VAR_DQ = %q, want %q", got, "double quoted")
	}
	if got := os.Getenv("MY_TEST_VAR_SQ"); got != "single quoted" {
		t.Errorf("MY_TEST_VAR_SQ = %q, want %q", got, "single quoted")
	}
}

func TestLoadEnvFile_EnvTakesPrecedence(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")

	content := `MY_TEST_VAR_EXISTING=from-file
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Set env var before loading
	os.Setenv("MY_TEST_VAR_EXISTING", "from-env")
	defer os.Unsetenv("MY_TEST_VAR_EXISTING")

	if err := LoadEnvFile(path); err != nil {
		t.Fatalf("LoadEnvFile() error: %v", err)
	}

	if got := os.Getenv("MY_TEST_VAR_EXISTING"); got != "from-env" {
		t.Errorf("MY_TEST_VAR_EXISTING = %q, want %q (env should take precedence)", got, "from-env")
	}
}

func TestLoadEnvFile_MissingFileIsOptional(t *testing.T) {
	err := LoadEnvFile("/nonexistent/.env")
	if err != nil {
		t.Errorf("LoadEnvFile() should return nil for missing file, got: %v", err)
	}
}

func TestLoadEnvFile_ValueWithEquals(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")

	content := `MY_TEST_VAR_EQ=key=value=extra
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	defer os.Unsetenv("MY_TEST_VAR_EQ")

	if err := LoadEnvFile(path); err != nil {
		t.Fatalf("LoadEnvFile() error: %v", err)
	}

	if got := os.Getenv("MY_TEST_VAR_EQ"); got != "key=value=extra" {
		t.Errorf("MY_TEST_VAR_EQ = %q, want %q", got, "key=value=extra")
	}
}
