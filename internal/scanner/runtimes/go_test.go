package runtimes

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoScanner_Name(t *testing.T) {
	s := NewGoScanner(&util.MockCommandRunner{})
	assert.Equal(t, "go", s.Name())
}

func TestGoScanner_Description(t *testing.T) {
	s := NewGoScanner(&util.MockCommandRunner{})
	assert.NotEmpty(t, s.Description())
}

func TestGoScanner_Category(t *testing.T) {
	s := NewGoScanner(&util.MockCommandRunner{})
	assert.Equal(t, "runtimes", s.Category())
}

func TestGoScanner_Scan_Full(t *testing.T) {
	tmpDir := t.TempDir()
	binDir := filepath.Join(tmpDir, "bin")
	require.NoError(t, os.MkdirAll(binDir, 0o755))

	// Create fake binaries in GOPATH/bin.
	for _, name := range []string{"gopls", "golangci-lint", "air"} {
		f, err := os.Create(filepath.Join(binDir, name))
		require.NoError(t, err)
		f.Close()
	}
	// Create a subdirectory that should be ignored.
	require.NoError(t, os.MkdirAll(filepath.Join(binDir, "subdir"), 0o755))

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"go":               {Output: "", Err: nil}, // IsInstalled check
			"go version":       {Output: "go version go1.24.1 darwin/arm64"},
			"go env GOPATH":    {Output: tmpDir},
		},
	}

	s := NewGoScanner(mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result.GoLang)

	goSection := result.GoLang
	assert.Equal(t, "1.24.1", goSection.Version)

	// Collect names from global packages.
	var names []string
	for _, pkg := range goSection.GlobalPackages {
		names = append(names, pkg.Name)
	}
	assert.Contains(t, names, "gopls")
	assert.Contains(t, names, "golangci-lint")
	assert.Contains(t, names, "air")
	// Subdirectory should not be included.
	assert.NotContains(t, names, "subdir")
	// Packages have no version (bin listing provides names only).
	for _, pkg := range goSection.GlobalPackages {
		assert.Equal(t, domain.Package{Name: pkg.Name}, pkg)
	}
}

func TestGoScanner_Scan_NotInstalled(t *testing.T) {
	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			// "go" key absent → IsInstalled returns false
		},
	}

	s := NewGoScanner(mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	assert.Nil(t, result.GoLang)
}

func TestGoScanner_Scan_EmptyGOPATHBin(t *testing.T) {
	tmpDir := t.TempDir()
	// No bin directory created.

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"go":            {Output: "", Err: nil},
			"go version":    {Output: "go version go1.22.0 linux/amd64"},
			"go env GOPATH": {Output: tmpDir},
		},
	}

	s := NewGoScanner(mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result.GoLang)

	assert.Equal(t, "1.22.0", result.GoLang.Version)
	assert.Empty(t, result.GoLang.GlobalPackages)
}

func TestParseGoVersion(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"go version go1.24.1 darwin/arm64", "1.24.1"},
		{"go version go1.22.0 linux/amd64", "1.22.0"},
		{"go version go1.21.5 windows/amd64", "1.21.5"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, parseGoVersion(tt.input))
		})
	}
}
