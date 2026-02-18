package bundler

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrepareBundleDir(t *testing.T) {
	snap := &domain.Snapshot{
		Meta: newMeta(),
		Homebrew: &domain.HomebrewSection{
			Taps: []string{"homebrew/core"},
			Formulae: []domain.Package{
				{Name: "git", Version: "2.40"},
			},
			Casks: []domain.Package{
				{Name: "firefox"},
			},
		},
		Shell: &domain.ShellSection{
			DefaultShell: "/bin/zsh",
		},
	}

	outputDir := t.TempDir()
	bundleDir := filepath.Join(outputDir, "bundle")

	err := PrepareBundleDir(snap, bundleDir, "")
	require.NoError(t, err)

	// Directory created with correct structure
	info, err := os.Stat(bundleDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())

	// manifest.toml exists and contains valid TOML
	manifestPath := filepath.Join(bundleDir, "manifest.toml")
	manifestData, err := os.ReadFile(manifestPath)
	require.NoError(t, err)
	assert.True(t, len(manifestData) > 0, "manifest.toml should not be empty")
	// Verify it's valid TOML by unmarshalling
	parsed, err := domain.UnmarshalManifest(manifestData)
	require.NoError(t, err)
	assert.Equal(t, "test-mac", parsed.Meta.SourceHostname)

	// install.command exists and starts with #!/bin/bash
	installPath := filepath.Join(bundleDir, "install.command")
	installData, err := os.ReadFile(installPath)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(string(installData), "#!/bin/bash"), "install.command should start with #!/bin/bash")
	// Check it is executable
	installInfo, err := os.Stat(installPath)
	require.NoError(t, err)
	assert.True(t, installInfo.Mode().Perm()&0100 != 0, "install.command should be executable")

	// configs/ directory exists
	configsDir := filepath.Join(bundleDir, "configs")
	configsInfo, err := os.Stat(configsDir)
	require.NoError(t, err)
	assert.True(t, configsInfo.IsDir())
}

func TestPrepareBundleDir_WithConfigFiles(t *testing.T) {
	// Create source config files in a temp dir
	configSourceDir := t.TempDir()
	zshrcContent := "# zshrc config\nexport PATH=$HOME/bin:$PATH\n"
	zprofileContent := "# zprofile config\n"
	require.NoError(t, os.WriteFile(filepath.Join(configSourceDir, ".zshrc"), []byte(zshrcContent), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(configSourceDir, ".zprofile"), []byte(zprofileContent), 0644))

	snap := &domain.Snapshot{
		Meta: newMeta(),
		Shell: &domain.ShellSection{
			DefaultShell: "/bin/zsh",
			ConfigFiles: []domain.ConfigFile{
				{Source: filepath.Join(configSourceDir, ".zshrc"), BundlePath: "configs/.zshrc"},
				{Source: filepath.Join(configSourceDir, ".zprofile"), BundlePath: "configs/.zprofile"},
			},
		},
	}

	outputDir := t.TempDir()
	bundleDir := filepath.Join(outputDir, "bundle")

	err := PrepareBundleDir(snap, bundleDir, configSourceDir)
	require.NoError(t, err)

	// Config files should be copied to configs/ subdirectory
	copiedZshrc, err := os.ReadFile(filepath.Join(bundleDir, "configs", ".zshrc"))
	require.NoError(t, err)
	assert.Equal(t, zshrcContent, string(copiedZshrc))

	copiedZprofile, err := os.ReadFile(filepath.Join(bundleDir, "configs", ".zprofile"))
	require.NoError(t, err)
	assert.Equal(t, zprofileContent, string(copiedZprofile))
}

func TestCreateDMG(t *testing.T) {
	sourceDir := t.TempDir()
	outputPath := filepath.Join(t.TempDir(), "test.dmg")
	volumeName := "Machinist Restore"

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			fmt.Sprintf("hdiutil create -volname %s -srcfolder %s -ov -format UDZO %s",
				volumeName, sourceDir, outputPath): {Output: "", Err: nil},
		},
	}

	err := CreateDMG(context.Background(), mock, sourceDir, outputPath, volumeName, "")
	require.NoError(t, err)

	// Verify hdiutil was called with correct args
	require.Len(t, mock.Calls, 1)
	call := mock.Calls[0]
	assert.Contains(t, call, "hdiutil create")
	assert.Contains(t, call, "-volname "+volumeName)
	assert.Contains(t, call, "-srcfolder "+sourceDir)
	assert.Contains(t, call, "-ov")
	assert.Contains(t, call, "-format UDZO")
	assert.Contains(t, call, outputPath)
}

func TestCreateDMG_HdiutilError(t *testing.T) {
	sourceDir := t.TempDir()
	outputPath := filepath.Join(t.TempDir(), "test.dmg")
	volumeName := "Machinist Restore"

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			fmt.Sprintf("hdiutil create -volname %s -srcfolder %s -ov -format UDZO %s",
				volumeName, sourceDir, outputPath): {Output: "", Err: fmt.Errorf("hdiutil: resource busy")},
		},
	}

	err := CreateDMG(context.Background(), mock, sourceDir, outputPath, volumeName, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "hdiutil: resource busy")
}

func TestCreateDMG_WithPassword(t *testing.T) {
	sourceDir := t.TempDir()
	outputPath := filepath.Join(t.TempDir(), "test.dmg")
	volumeName := "Machinist Restore"
	password := "s3cret"

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			fmt.Sprintf("hdiutil create -volname %s -srcfolder %s -ov -format UDZO -encryption AES-256 -stdinpass %s",
				volumeName, sourceDir, outputPath): {Output: "", Err: nil},
		},
	}

	err := CreateDMG(context.Background(), mock, sourceDir, outputPath, volumeName, password)
	require.NoError(t, err)

	// Verify hdiutil was called with encryption flags
	require.Len(t, mock.Calls, 1)
	call := mock.Calls[0]
	assert.Contains(t, call, "-encryption AES-256")
	assert.Contains(t, call, "-stdinpass")
}
