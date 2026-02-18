package runtimes

import (
	"context"
	"testing"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRustScanner_Name(t *testing.T) {
	s := NewRustScanner(&util.MockCommandRunner{})
	assert.Equal(t, "rust", s.Name())
}

func TestRustScanner_Description(t *testing.T) {
	s := NewRustScanner(&util.MockCommandRunner{})
	assert.NotEmpty(t, s.Description())
}

func TestRustScanner_Category(t *testing.T) {
	s := NewRustScanner(&util.MockCommandRunner{})
	assert.Equal(t, "runtimes", s.Category())
}

func TestRustScanner_Scan_Full(t *testing.T) {
	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"rustup": {Output: "", Err: nil}, // IsInstalled check
			"rustup toolchain list": {
				Output: "stable-aarch64-apple-darwin (default)\nnightly-aarch64-apple-darwin",
			},
			"rustup component list --installed": {
				Output: "cargo\nclippy\nrust-analyzer\nrust-src\nrust-std\nrustc\nrustfmt",
			},
			"cargo install --list": {
				Output: "bat v0.24.0:\n    bat\nripgrep v14.1.0:\n    rg\ntokei v12.1.2:\n    tokei",
			},
		},
	}

	s := NewRustScanner(mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result.Rust)

	rust := result.Rust

	assert.Equal(t, []string{"stable-aarch64-apple-darwin", "nightly-aarch64-apple-darwin"}, rust.Toolchains)
	assert.Equal(t, "stable-aarch64-apple-darwin", rust.DefaultToolchain)

	assert.Len(t, rust.Components, 7)
	assert.Contains(t, rust.Components, "cargo")
	assert.Contains(t, rust.Components, "rustfmt")
	assert.Contains(t, rust.Components, "rust-analyzer")

	require.Len(t, rust.CargoPackages, 3)
	assert.Equal(t, domain.Package{Name: "bat", Version: "v0.24.0"}, rust.CargoPackages[0])
	assert.Equal(t, domain.Package{Name: "ripgrep", Version: "v14.1.0"}, rust.CargoPackages[1])
	assert.Equal(t, domain.Package{Name: "tokei", Version: "v12.1.2"}, rust.CargoPackages[2])
}

func TestRustScanner_Scan_NoRust(t *testing.T) {
	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			// "rustup" key absent → IsInstalled returns false
		},
	}

	s := NewRustScanner(mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	assert.Nil(t, result.Rust)
}

func TestRustScanner_Scan_NoCargoPackages(t *testing.T) {
	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"rustup": {Output: "", Err: nil},
			"rustup toolchain list": {
				Output: "stable-aarch64-apple-darwin (default)",
			},
			"rustup component list --installed": {
				Output: "cargo\nrustc\nrustfmt",
			},
			"cargo install --list": {
				Output: "",
			},
		},
	}

	s := NewRustScanner(mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result.Rust)

	rust := result.Rust
	assert.Equal(t, []string{"stable-aarch64-apple-darwin"}, rust.Toolchains)
	assert.Equal(t, "stable-aarch64-apple-darwin", rust.DefaultToolchain)
	assert.Len(t, rust.Components, 3)
	assert.Empty(t, rust.CargoPackages)
}

func TestRustScanner_Scan_DefaultToolchainParsing(t *testing.T) {
	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"rustup": {Output: "", Err: nil},
			"rustup toolchain list": {
				Output: "nightly-aarch64-apple-darwin\nstable-aarch64-apple-darwin (default)\nbeta-aarch64-apple-darwin",
			},
			"rustup component list --installed": {
				Output: "rustc",
			},
			"cargo install --list": {
				Output: "",
			},
		},
	}

	s := NewRustScanner(mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result.Rust)

	rust := result.Rust
	// "(default)" suffix should be stripped from the toolchain name in the list
	assert.Equal(t, []string{
		"nightly-aarch64-apple-darwin",
		"stable-aarch64-apple-darwin",
		"beta-aarch64-apple-darwin",
	}, rust.Toolchains)
	// DefaultToolchain should be set from the one marked "(default)"
	assert.Equal(t, "stable-aarch64-apple-darwin", rust.DefaultToolchain)
}

func TestRustScanner_Scan_CargoInstallParsing(t *testing.T) {
	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"rustup": {Output: "", Err: nil},
			"rustup toolchain list": {
				Output: "stable-aarch64-apple-darwin (default)",
			},
			"rustup component list --installed": {
				Output: "rustc",
			},
			"cargo install --list": {
				Output: "package-name v1.2.3:\n    binary-name\nanother-pkg v0.5.0:\n    bin1\n    bin2",
			},
		},
	}

	s := NewRustScanner(mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result.Rust)

	// Only lines with version (ending in `:`) create Package entries
	pkgs := result.Rust.CargoPackages
	require.Len(t, pkgs, 2)
	assert.Equal(t, domain.Package{Name: "package-name", Version: "v1.2.3"}, pkgs[0])
	assert.Equal(t, domain.Package{Name: "another-pkg", Version: "v0.5.0"}, pkgs[1])
}
