package runtimes

import (
	"context"
	"testing"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBunScanner_Name(t *testing.T) {
	s := NewBunScanner(&util.MockCommandRunner{})
	assert.Equal(t, "bun", s.Name())
}

func TestBunScanner_Description(t *testing.T) {
	s := NewBunScanner(&util.MockCommandRunner{})
	assert.NotEmpty(t, s.Description())
}

func TestBunScanner_Category(t *testing.T) {
	s := NewBunScanner(&util.MockCommandRunner{})
	assert.Equal(t, "runtimes", s.Category())
}

func TestBunScanner_Scan_HappyPath(t *testing.T) {
	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"bun": {Output: "", Err: nil}, // IsInstalled check
			"bun --version": {Output: "1.0.25"},
			"bun pm ls -g": {
				Output: "/home/user/.bun/install/global/node_modules (global)\n├── typescript@5.3.3\n└── eslint@8.56.0",
			},
		},
	}

	s := NewBunScanner(mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result.Bun)

	bun := result.Bun
	assert.Equal(t, "1.0.25", bun.Version)
	require.Len(t, bun.GlobalPackages, 2)
	assert.Equal(t, domain.Package{Name: "typescript", Version: "5.3.3"}, bun.GlobalPackages[0])
	assert.Equal(t, domain.Package{Name: "eslint", Version: "8.56.0"}, bun.GlobalPackages[1])
}

func TestBunScanner_Scan_NotInstalled(t *testing.T) {
	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			// "bun" key absent → IsInstalled returns false
		},
	}

	s := NewBunScanner(mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	assert.Nil(t, result.Bun)
}

func TestBunScanner_Scan_EmptyGlobalList(t *testing.T) {
	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"bun":            {Output: "", Err: nil},
			"bun --version":  {Output: "1.1.0"},
			"bun pm ls -g":   {Output: ""},
		},
	}

	s := NewBunScanner(mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result.Bun)

	assert.Equal(t, "1.1.0", result.Bun.Version)
	assert.Empty(t, result.Bun.GlobalPackages)
}

func TestBunScanner_Scan_ScopedPackages(t *testing.T) {
	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"bun":           {Output: "", Err: nil},
			"bun --version": {Output: "1.0.25"},
			"bun pm ls -g": {
				Output: "/home/user/.bun/install/global/node_modules (global)\n├── @biomejs/biome@1.4.1\n└── @antfu/ni@0.21.12",
			},
		},
	}

	s := NewBunScanner(mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result.Bun)

	pkgs := result.Bun.GlobalPackages
	require.Len(t, pkgs, 2)
	assert.Equal(t, domain.Package{Name: "@biomejs/biome", Version: "1.4.1"}, pkgs[0])
	assert.Equal(t, domain.Package{Name: "@antfu/ni", Version: "0.21.12"}, pkgs[1])
}

func TestBunScanner_Scan_NoVersionSuffix(t *testing.T) {
	// Some bun pm output may show just package names without versions.
	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"bun":           {Output: "", Err: nil},
			"bun --version": {Output: "1.0.25"},
			"bun pm ls -g": {
				Output: "/home/user/.bun/install/global/node_modules (global)\n└── serve",
			},
		},
	}

	s := NewBunScanner(mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result.Bun)

	pkgs := result.Bun.GlobalPackages
	require.Len(t, pkgs, 1)
	assert.Equal(t, domain.Package{Name: "serve", Version: ""}, pkgs[0])
}

func TestBunParseBunGlobalList_TreeChars(t *testing.T) {
	lines := []string{
		"/home/user/.bun/install/global/node_modules (global)",
		"├── typescript@5.3.3",
		"├── eslint@8.56.0",
		"└── prettier@3.2.4",
	}
	pkgs := parseBunGlobalList(lines)
	require.Len(t, pkgs, 3)
	assert.Equal(t, domain.Package{Name: "typescript", Version: "5.3.3"}, pkgs[0])
	assert.Equal(t, domain.Package{Name: "eslint", Version: "8.56.0"}, pkgs[1])
	assert.Equal(t, domain.Package{Name: "prettier", Version: "3.2.4"}, pkgs[2])
}

func TestSplitBunPackageEntry(t *testing.T) {
	cases := []struct {
		input   string
		name    string
		version string
	}{
		{"typescript@5.3.3", "typescript", "5.3.3"},
		{"@biomejs/biome@1.4.1", "@biomejs/biome", "1.4.1"},
		{"serve", "serve", ""},
		{"@scope/pkg", "@scope/pkg", ""},
	}
	for _, tc := range cases {
		name, version := splitBunPackageEntry(tc.input)
		assert.Equal(t, tc.name, name, "name for %q", tc.input)
		assert.Equal(t, tc.version, version, "version for %q", tc.input)
	}
}
