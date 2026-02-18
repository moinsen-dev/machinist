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

func TestNodeScanner_Name(t *testing.T) {
	s := NewNodeScanner("/home/user", &util.MockCommandRunner{})
	assert.Equal(t, "node", s.Name())
}

func TestNodeScanner_Description(t *testing.T) {
	s := NewNodeScanner("/home/user", &util.MockCommandRunner{})
	assert.NotEmpty(t, s.Description())
}

func TestNodeScanner_Category(t *testing.T) {
	s := NewNodeScanner("/home/user", &util.MockCommandRunner{})
	assert.Equal(t, "runtimes", s.Category())
}

func TestNodeScanner_Scan_WithNvm(t *testing.T) {
	homeDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(homeDir, ".nvm"), 0o755))

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"bash -l -c nvm list": {
				Output: "->     v20.10.0\n       v18.19.0\n       v16.20.2\ndefault -> v20.10.0",
			},
			"npm list -g --depth=0 --json": {
				Output: `{"dependencies":{"npm":{"version":"10.2.3"},"typescript":{"version":"5.3.3"},"eslint":{"version":"8.56.0"}}}`,
			},
		},
	}

	s := NewNodeScanner(homeDir, mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.Node)

	node := result.Node
	assert.Equal(t, "nvm", node.Manager)
	assert.Equal(t, "v20.10.0", node.DefaultVersion)
	assert.ElementsMatch(t, []string{"v20.10.0", "v18.19.0", "v16.20.2"}, node.Versions)

	// npm itself should be excluded from global packages
	assert.Len(t, node.GlobalPackages, 2)
	assertHasPackage(t, node.GlobalPackages, "typescript", "5.3.3")
	assertHasPackage(t, node.GlobalPackages, "eslint", "8.56.0")
}

func TestNodeScanner_Scan_WithFnm(t *testing.T) {
	homeDir := t.TempDir()
	// No .nvm dir, but fnm is installed via IsInstalled check.
	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"fnm": {Output: "fnm 1.35.1", Err: nil},
			"fnm list": {
				Output: "* v20.11.0 default\n  v18.19.1",
			},
			"npm list -g --depth=0 --json": {
				Output: `{"dependencies":{"yarn":{"version":"1.22.21"}}}`,
			},
		},
	}

	s := NewNodeScanner(homeDir, mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.Node)

	node := result.Node
	assert.Equal(t, "fnm", node.Manager)
	assert.Equal(t, "v20.11.0", node.DefaultVersion)
	assert.ElementsMatch(t, []string{"v20.11.0", "v18.19.1"}, node.Versions)
	assert.Len(t, node.GlobalPackages, 1)
	assertHasPackage(t, node.GlobalPackages, "yarn", "1.22.21")
}

func TestNodeScanner_Scan_SystemNode(t *testing.T) {
	homeDir := t.TempDir()
	// No nvm, no fnm, but node is installed directly.
	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"node --version": {Output: "v20.10.0"},
			"npm list -g --depth=0 --json": {
				Output: `{"dependencies":{}}`,
			},
		},
	}

	s := NewNodeScanner(homeDir, mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.Node)

	node := result.Node
	assert.Equal(t, "", node.Manager)
	assert.Equal(t, "v20.10.0", node.DefaultVersion)
	assert.Equal(t, []string{"v20.10.0"}, node.Versions)
	assert.Empty(t, node.GlobalPackages)
}

func TestNodeScanner_Scan_NoNodeInstalled(t *testing.T) {
	homeDir := t.TempDir()
	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{},
	}

	s := NewNodeScanner(homeDir, mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	assert.Nil(t, result.Node)
}

func TestNodeScanner_Scan_GlobalPackages(t *testing.T) {
	homeDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(homeDir, ".nvm"), 0o755))

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"bash -l -c nvm list": {
				Output: "->     v20.10.0\ndefault -> v20.10.0",
			},
			"npm list -g --depth=0 --json": {
				Output: `{"dependencies":{"typescript":{"version":"5.3.3"},"eslint":{"version":"8.56.0"}}}`,
			},
		},
	}

	s := NewNodeScanner(homeDir, mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.Node)

	pkgs := result.Node.GlobalPackages
	assert.Len(t, pkgs, 2)
	assertHasPackage(t, pkgs, "typescript", "5.3.3")
	assertHasPackage(t, pkgs, "eslint", "8.56.0")
}

// --- helpers ---

func assertHasPackage(t *testing.T, pkgs []domain.Package, name, version string) {
	t.Helper()
	for _, p := range pkgs {
		if p.Name == name {
			assert.Equal(t, version, p.Version, "version mismatch for package %s", name)
			return
		}
	}
	t.Errorf("package %s not found in %v", name, pkgs)
}
