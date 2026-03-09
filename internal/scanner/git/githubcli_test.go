package git

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/moinsen-dev/machinist/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGitHubCLIScanner_Name(t *testing.T) {
	s := NewGitHubCLIScanner("/tmp", &util.MockCommandRunner{Responses: map[string]util.MockResponse{}})
	assert.Equal(t, "github-cli", s.Name())
}

func TestGitHubCLIScanner_Description(t *testing.T) {
	s := NewGitHubCLIScanner("/tmp", &util.MockCommandRunner{Responses: map[string]util.MockResponse{}})
	assert.NotEmpty(t, s.Description())
}

func TestGitHubCLIScanner_Category(t *testing.T) {
	s := NewGitHubCLIScanner("/tmp", &util.MockCommandRunner{Responses: map[string]util.MockResponse{}})
	assert.Equal(t, "git", s.Category())
}

func TestGitHubCLIScanner_Scan_NotInstalled(t *testing.T) {
	// No "gh" key → IsInstalled returns false.
	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{},
	}
	s := NewGitHubCLIScanner("/tmp", mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	assert.Nil(t, result.GitHubCLI)
}

func TestGitHubCLIScanner_Scan_HappyPath(t *testing.T) {
	homeDir := t.TempDir()

	// Create ~/.config/gh directory.
	ghConfigDir := filepath.Join(homeDir, ".config", "gh")
	require.NoError(t, os.MkdirAll(ghConfigDir, 0o755))

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"gh": {Output: "", Err: nil}, // IsInstalled
			"gh extension list": {
				Output: "gh-copilot\thttps://github.com/github/gh-copilot\tv1.0.0\ngh-dash\thttps://github.com/dlvhdr/gh-dash\tv3.12.0",
			},
		},
	}

	s := NewGitHubCLIScanner(homeDir, mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.GitHubCLI)

	cli := result.GitHubCLI
	assert.Equal(t, filepath.Join(".config", "gh"), cli.ConfigDir)
	require.Len(t, cli.Extensions, 2)
	assert.Equal(t, "gh-copilot", cli.Extensions[0])
	assert.Equal(t, "gh-dash", cli.Extensions[1])
}

func TestGitHubCLIScanner_Scan_NoExtensions(t *testing.T) {
	homeDir := t.TempDir()

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"gh": {Output: "", Err: nil},
			"gh extension list": {
				Output: "",
			},
		},
	}

	s := NewGitHubCLIScanner(homeDir, mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	// No config dir, no extensions → section should be nil.
	assert.Nil(t, result.GitHubCLI)
}

func TestGitHubCLIScanner_Scan_ExtensionListError(t *testing.T) {
	homeDir := t.TempDir()

	// Create the config dir so it's detected.
	ghConfigDir := filepath.Join(homeDir, ".config", "gh")
	require.NoError(t, os.MkdirAll(ghConfigDir, 0o755))

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"gh": {Output: "", Err: nil},
			"gh extension list": {
				Output: "", Err: fmt.Errorf("could not connect to GitHub"),
			},
		},
	}

	s := NewGitHubCLIScanner(homeDir, mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.GitHubCLI)
	// ConfigDir is still recorded even when extension list fails.
	assert.Equal(t, filepath.Join(".config", "gh"), result.GitHubCLI.ConfigDir)
	assert.Empty(t, result.GitHubCLI.Extensions)
}

func TestGitHubCLIScanner_Scan_SpaceSeparatedExtensions(t *testing.T) {
	homeDir := t.TempDir()

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"gh": {Output: "", Err: nil},
			// Extensions separated by spaces rather than tabs.
			"gh extension list": {
				Output: "gh-actions-toolkit  https://github.com/nicholasgasior/gh-actions-toolkit  v0.2.1\ngh-notify  https://github.com/meiji163/gh-notify  v1.2.3",
			},
		},
	}

	s := NewGitHubCLIScanner(homeDir, mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.GitHubCLI)
	require.Len(t, result.GitHubCLI.Extensions, 2)
	assert.Equal(t, "gh-actions-toolkit", result.GitHubCLI.Extensions[0])
	assert.Equal(t, "gh-notify", result.GitHubCLI.Extensions[1])
}
