package scanner

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockScanner implements the Scanner interface for testing purposes.
type mockScanner struct {
	name     string
	desc     string
	category string
	result   *ScanResult
	err      error
}

func (m *mockScanner) Name() string        { return m.name }
func (m *mockScanner) Description() string  { return m.desc }
func (m *mockScanner) Category() string     { return m.category }
func (m *mockScanner) Scan(ctx context.Context) (*ScanResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.result, nil
}

func TestRegistry_RegisterAndGet(t *testing.T) {
	reg := NewRegistry()
	mock := &mockScanner{
		name:     "brew",
		desc:     "Scans Homebrew",
		category: "packages",
	}

	err := reg.Register(mock)
	require.NoError(t, err)

	got, err := reg.Get("brew")
	require.NoError(t, err)
	assert.Same(t, mock, got, "Get should return the exact same scanner instance")
}

func TestRegistry_RegisterDuplicate(t *testing.T) {
	reg := NewRegistry()
	mock1 := &mockScanner{name: "brew", desc: "first", category: "packages"}
	mock2 := &mockScanner{name: "brew", desc: "second", category: "packages"}

	err := reg.Register(mock1)
	require.NoError(t, err)

	err = reg.Register(mock2)
	require.Error(t, err, "registering a duplicate name should return an error")
	assert.Contains(t, err.Error(), "brew")
}

func TestRegistry_GetUnknown(t *testing.T) {
	reg := NewRegistry()

	got, err := reg.Get("nonexistent")
	require.Error(t, err, "getting an unregistered scanner should return an error")
	assert.Nil(t, got)
	assert.Contains(t, err.Error(), "nonexistent")
}

func TestRegistry_List(t *testing.T) {
	reg := NewRegistry()
	names := []string{"zsh", "brew", "git"}
	for _, n := range names {
		err := reg.Register(&mockScanner{name: n, desc: n, category: "test"})
		require.NoError(t, err)
	}

	list := reg.List()
	require.Len(t, list, 3)

	// Verify alphabetical order
	assert.Equal(t, "brew", list[0].Name())
	assert.Equal(t, "git", list[1].Name())
	assert.Equal(t, "zsh", list[2].Name())
}

func TestRegistry_ScanAll(t *testing.T) {
	reg := NewRegistry()

	brewSection := &domain.HomebrewSection{
		Taps: []string{"homebrew/core"},
		Formulae: []domain.Package{
			{Name: "go", Version: "1.25"},
		},
	}
	shellSection := &domain.ShellSection{
		DefaultShell: "/bin/zsh",
		Framework:    "oh-my-zsh",
	}

	brewScanner := &mockScanner{
		name:     "brew",
		desc:     "Homebrew scanner",
		category: "packages",
		result: &ScanResult{
			ScannerName: "brew",
			Duration:    100 * time.Millisecond,
			Homebrew:    brewSection,
		},
	}
	shellScanner := &mockScanner{
		name:     "shell",
		desc:     "Shell scanner",
		category: "shell",
		result: &ScanResult{
			ScannerName: "shell",
			Duration:    50 * time.Millisecond,
			Shell:       shellSection,
		},
	}

	require.NoError(t, reg.Register(brewScanner))
	require.NoError(t, reg.Register(shellScanner))

	snap, errs := reg.ScanAll(context.Background())
	require.NotNil(t, snap)
	assert.Empty(t, errs, "ScanAll should have no errors when all scanners succeed")

	// Verify both sections are populated in the snapshot
	require.NotNil(t, snap.Homebrew)
	assert.Equal(t, brewSection, snap.Homebrew)

	require.NotNil(t, snap.Shell)
	assert.Equal(t, shellSection, snap.Shell)

	// Verify Meta is populated
	assert.NotEmpty(t, snap.Meta.SourceArch)
	assert.NotZero(t, snap.Meta.ScanDurationSecs)
}

func TestRegistry_ScanAll_WithError(t *testing.T) {
	reg := NewRegistry()

	shellSection := &domain.ShellSection{DefaultShell: "/bin/zsh"}
	goodScanner := &mockScanner{
		name:     "shell",
		desc:     "Shell scanner",
		category: "shell",
		result: &ScanResult{
			ScannerName: "shell",
			Duration:    50 * time.Millisecond,
			Shell:       shellSection,
		},
	}
	badScanner := &mockScanner{
		name:     "brew",
		desc:     "Homebrew scanner",
		category: "packages",
		err:      fmt.Errorf("homebrew not installed"),
	}

	require.NoError(t, reg.Register(goodScanner))
	require.NoError(t, reg.Register(badScanner))

	snap, errs := reg.ScanAll(context.Background())
	require.NotNil(t, snap, "ScanAll should still return a snapshot even when some scanners fail")

	// The good scanner's result should still be applied
	require.NotNil(t, snap.Shell)
	assert.Equal(t, "/bin/zsh", snap.Shell.DefaultShell)

	// The bad scanner's section should be nil
	assert.Nil(t, snap.Homebrew)

	// There should be exactly one error
	require.Len(t, errs, 1)
	assert.Contains(t, errs[0].Error(), "homebrew not installed")
}

func TestRegistry_ScanOne(t *testing.T) {
	reg := NewRegistry()

	brewSection := &domain.HomebrewSection{
		Taps: []string{"homebrew/core"},
	}
	mock := &mockScanner{
		name:     "brew",
		desc:     "Homebrew scanner",
		category: "packages",
		result: &ScanResult{
			ScannerName: "brew",
			Duration:    100 * time.Millisecond,
			Homebrew:    brewSection,
		},
	}

	require.NoError(t, reg.Register(mock))

	result, err := reg.ScanOne(context.Background(), "brew")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "brew", result.ScannerName)
	assert.Equal(t, brewSection, result.Homebrew)

	// ScanOne with unknown name should error
	result, err = reg.ScanOne(context.Background(), "nonexistent")
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "nonexistent")
}

func TestRegistry_ScanOne_ScannerError(t *testing.T) {
	reg := NewRegistry()

	badScanner := &mockScanner{
		name:     "failing",
		desc:     "A scanner that always fails",
		category: "test",
		err:      fmt.Errorf("disk read error"),
	}

	require.NoError(t, reg.Register(badScanner))

	result, err := reg.ScanOne(context.Background(), "failing")
	require.Error(t, err, "ScanOne should propagate scanner errors")
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "disk read error")
	assert.Contains(t, err.Error(), "failing")
}

func TestRegistry_ScanAll_AppliesMoreSections(t *testing.T) {
	reg := NewRegistry()

	gitSection := &domain.GitSection{
		SigningMethod: "gpg",
	}
	dockerSection := &domain.DockerSection{
		Runtime: "containerd",
	}
	vscodeSection := &domain.VSCodeSection{
		Extensions: []string{"golang.go", "ms-python.python"},
	}
	nodeSection := &domain.NodeSection{
		Manager:        "nvm",
		DefaultVersion: "20.11.0",
	}
	pythonSection := &domain.PythonSection{
		Manager:        "pyenv",
		DefaultVersion: "3.12.1",
	}
	rustSection := &domain.RustSection{
		DefaultToolchain: "stable",
		Components:       []string{"rustfmt", "clippy"},
	}

	gitScanner := &mockScanner{
		name:     "git",
		desc:     "Git scanner",
		category: "vcs",
		result: &ScanResult{
			ScannerName: "git",
			Git:         gitSection,
		},
	}
	dockerScanner := &mockScanner{
		name:     "docker",
		desc:     "Docker scanner",
		category: "containers",
		result: &ScanResult{
			ScannerName: "docker",
			Docker:      dockerSection,
		},
	}
	vscodeScanner := &mockScanner{
		name:     "vscode",
		desc:     "VSCode scanner",
		category: "editors",
		result: &ScanResult{
			ScannerName: "vscode",
			VSCode:      vscodeSection,
		},
	}
	nodeScanner := &mockScanner{
		name:     "node",
		desc:     "Node scanner",
		category: "languages",
		result: &ScanResult{
			ScannerName: "node",
			Node:        nodeSection,
		},
	}
	pythonScanner := &mockScanner{
		name:     "python",
		desc:     "Python scanner",
		category: "languages",
		result: &ScanResult{
			ScannerName: "python",
			Python:      pythonSection,
		},
	}
	rustScanner := &mockScanner{
		name:     "rust",
		desc:     "Rust scanner",
		category: "languages",
		result: &ScanResult{
			ScannerName: "rust",
			Rust:        rustSection,
		},
	}

	require.NoError(t, reg.Register(gitScanner))
	require.NoError(t, reg.Register(dockerScanner))
	require.NoError(t, reg.Register(vscodeScanner))
	require.NoError(t, reg.Register(nodeScanner))
	require.NoError(t, reg.Register(pythonScanner))
	require.NoError(t, reg.Register(rustScanner))

	snap, errs := reg.ScanAll(context.Background())
	require.NotNil(t, snap)
	assert.Empty(t, errs)

	// Verify all sections are applied to the snapshot.
	require.NotNil(t, snap.Git)
	assert.Equal(t, "gpg", snap.Git.SigningMethod)

	require.NotNil(t, snap.Docker)
	assert.Equal(t, "containerd", snap.Docker.Runtime)

	require.NotNil(t, snap.VSCode)
	assert.Equal(t, []string{"golang.go", "ms-python.python"}, snap.VSCode.Extensions)

	require.NotNil(t, snap.Node)
	assert.Equal(t, "nvm", snap.Node.Manager)
	assert.Equal(t, "20.11.0", snap.Node.DefaultVersion)

	require.NotNil(t, snap.Python)
	assert.Equal(t, "pyenv", snap.Python.Manager)
	assert.Equal(t, "3.12.1", snap.Python.DefaultVersion)

	require.NotNil(t, snap.Rust)
	assert.Equal(t, "stable", snap.Rust.DefaultToolchain)
	assert.Equal(t, []string{"rustfmt", "clippy"}, snap.Rust.Components)
}
