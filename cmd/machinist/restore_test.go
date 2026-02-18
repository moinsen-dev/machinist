package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// resetRestoreFlags resets package-level restore flag vars to defaults so
// tests don't leak state into each other.
func resetRestoreFlags() {
	restoreSkip = ""
	restoreOnly = ""
	restoreDryRun = false
	restoreYes = false
}

func TestRestoreNonExistentFile(t *testing.T) {
	resetRestoreFlags()
	_, err := executeCommand("restore", "/tmp/does-not-exist-machinist/manifest.toml", "--yes")
	if err == nil {
		t.Fatal("expected error for non-existent manifest, got nil")
	}
	if !strings.Contains(err.Error(), "read manifest") {
		t.Errorf("expected error to contain 'read manifest', got: %s", err.Error())
	}
}

func TestRestoreDryRun(t *testing.T) {
	resetRestoreFlags()
	dir := t.TempDir()
	manifest := filepath.Join(dir, "manifest.toml")
	content := `[meta]
source_hostname = "test-mac"
source_arch = "arm64"
snapshot_date = "2025-01-01"
machinist_version = "0.1.0"

[homebrew]
taps = []
formulae = []
casks = []
`
	if err := os.WriteFile(manifest, []byte(content), 0644); err != nil {
		t.Fatalf("write test manifest: %v", err)
	}

	output, err := executeCommand("restore", manifest, "--dry-run")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "Dry-run mode") {
		t.Errorf("expected output to contain 'Dry-run mode', got:\n%s", output)
	}
	if !strings.Contains(output, "No changes were made") {
		t.Errorf("expected output to contain 'No changes were made', got:\n%s", output)
	}
	if !strings.Contains(output, "test-mac") {
		t.Errorf("expected output to contain hostname 'test-mac', got:\n%s", output)
	}
}

func TestRestoreSkipAndOnlyMutuallyExclusive(t *testing.T) {
	resetRestoreFlags()
	_, err := executeCommand("restore", "some-file.toml", "--skip", "homebrew", "--only", "shell")
	if err == nil {
		t.Fatal("expected error when both --skip and --only are specified, got nil")
	}
	if !strings.Contains(err.Error(), "mutually exclusive") {
		t.Errorf("expected error to mention 'mutually exclusive', got: %s", err.Error())
	}
}

func TestRestoreWithoutYesShowsPrompt(t *testing.T) {
	resetRestoreFlags()
	dir := t.TempDir()
	manifest := filepath.Join(dir, "manifest.toml")
	content := `[meta]
source_hostname = "test-mac"
source_arch = "arm64"
snapshot_date = "2025-01-01"
machinist_version = "0.1.0"

[homebrew]
taps = []
formulae = []
casks = []
`
	if err := os.WriteFile(manifest, []byte(content), 0644); err != nil {
		t.Fatalf("write test manifest: %v", err)
	}

	output, err := executeCommand("restore", manifest)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "Use --yes to confirm execution") {
		t.Errorf("expected output to contain confirmation prompt, got:\n%s", output)
	}
}
