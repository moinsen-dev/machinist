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
	restoreList = false
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
	// Should show groups, not sections
	if !strings.Contains(output, "Groups to execute") {
		t.Errorf("expected output to contain 'Groups to execute', got:\n%s", output)
	}
	if !strings.Contains(output, "homebrew") {
		t.Errorf("expected output to contain 'homebrew' group, got:\n%s", output)
	}
	// Should show stage count
	if !strings.Contains(output, "stages)") {
		t.Errorf("expected output to contain stage count, got:\n%s", output)
	}
}

func TestRestoreSkipAndOnlyMutuallyExclusive(t *testing.T) {
	resetRestoreFlags()
	_, err := executeCommand("restore", "some-file.toml", "--skip", "homebrew", "--only", "configs")
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

func TestRestoreList(t *testing.T) {
	resetRestoreFlags()
	output, err := executeCommand("restore", "--list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "Available restore groups") {
		t.Errorf("expected output to contain 'Available restore groups', got:\n%s", output)
	}
	// Should list all group names
	for _, name := range []string{"homebrew", "secrets", "configs", "runtimes", "repos", "macos"} {
		if !strings.Contains(output, name) {
			t.Errorf("expected output to contain group '%s', got:\n%s", name, output)
		}
	}
}

func TestRestoreDryRunWithOnly(t *testing.T) {
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
formulae = [{name = "git"}]
casks = []

[shell]
default_shell = "/bin/zsh"
`
	if err := os.WriteFile(manifest, []byte(content), 0644); err != nil {
		t.Fatalf("write test manifest: %v", err)
	}

	output, err := executeCommand("restore", manifest, "--dry-run", "--only", "homebrew")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "homebrew") {
		t.Errorf("expected output to contain 'homebrew', got:\n%s", output)
	}
	// configs should be filtered out by --only
	if strings.Contains(output, "configs") {
		t.Errorf("expected 'configs' to be filtered out, got:\n%s", output)
	}
}

func TestRestore_UnknownGroupName(t *testing.T) {
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

	// Test --only with unknown group
	_, err := executeCommand("restore", manifest, "--dry-run", "--only", "bogus,configs")
	if err == nil {
		t.Fatal("expected error for unknown group name, got nil")
	}
	if !strings.Contains(err.Error(), "unknown group(s)") {
		t.Errorf("expected error to contain 'unknown group(s)', got: %s", err.Error())
	}
	if !strings.Contains(err.Error(), "bogus") {
		t.Errorf("expected error to mention 'bogus', got: %s", err.Error())
	}

	// Test --skip with unknown group
	resetRestoreFlags()
	_, err = executeCommand("restore", manifest, "--dry-run", "--skip", "nope")
	if err == nil {
		t.Fatal("expected error for unknown group name in --skip, got nil")
	}
	if !strings.Contains(err.Error(), "unknown group(s)") {
		t.Errorf("expected error to contain 'unknown group(s)', got: %s", err.Error())
	}
	if !strings.Contains(err.Error(), "nope") {
		t.Errorf("expected error to mention 'nope', got: %s", err.Error())
	}
}

func TestRestoreDryRunWithSkip(t *testing.T) {
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
formulae = [{name = "git"}]
casks = []

[git_repos]
[[git_repos.repositories]]
remote = "git@github.com:user/repo.git"
path = "~/work/repo"
`
	if err := os.WriteFile(manifest, []byte(content), 0644); err != nil {
		t.Fatalf("write test manifest: %v", err)
	}

	output, err := executeCommand("restore", manifest, "--dry-run", "--skip", "repos")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "homebrew") {
		t.Errorf("expected output to contain 'homebrew', got:\n%s", output)
	}
	// repos should be filtered out by --skip
	if strings.Contains(output, "repos") {
		t.Errorf("expected 'repos' to be filtered out, got:\n%s", output)
	}
}
