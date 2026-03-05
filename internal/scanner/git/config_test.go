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

func TestGitConfigScanner_Name(t *testing.T) {
	s := NewGitConfigScanner("/tmp", &util.MockCommandRunner{Responses: map[string]util.MockResponse{}})
	assert.Equal(t, "git-config", s.Name())
}

func TestGitConfigScanner_Description(t *testing.T) {
	s := NewGitConfigScanner("/tmp", &util.MockCommandRunner{Responses: map[string]util.MockResponse{}})
	assert.NotEmpty(t, s.Description())
}

func TestGitConfigScanner_Category(t *testing.T) {
	s := NewGitConfigScanner("/tmp", &util.MockCommandRunner{Responses: map[string]util.MockResponse{}})
	assert.Equal(t, "git", s.Category())
}

func TestGitConfigScanner_Scan_NotInstalled(t *testing.T) {
	// No "git" key → IsInstalled returns false.
	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{},
	}
	s := NewGitConfigScanner("/tmp", mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	assert.Nil(t, result.Git)
}

func TestGitConfigScanner_Scan_HappyPath(t *testing.T) {
	homeDir := t.TempDir()

	// Create ~/.gitconfig so FileExists returns true.
	require.NoError(t, os.WriteFile(filepath.Join(homeDir, ".gitconfig"), []byte("[user]\n\tname = Test User\n"), 0o644))

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"git": {Output: "", Err: nil}, // IsInstalled
			"git config --global core.excludesfile": {
				Output: filepath.Join(homeDir, ".gitignore_global"),
			},
			"git config --global user.signingkey": {
				Output: "ABCDEF1234567890",
			},
			"git config --global gpg.format": {
				Output: "ssh",
			},
			"git config --global credential.helper": {
				Output: "osxkeychain",
			},
			"git config --global init.templateDir": {
				Output: filepath.Join(homeDir, ".git-templates"),
			},
		},
	}

	s := NewGitConfigScanner(homeDir, mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.Git)

	git := result.Git

	// Should have two ConfigFiles: .gitconfig and the excludes file.
	require.Len(t, git.ConfigFiles, 2)
	assert.Equal(t, ".gitconfig", git.ConfigFiles[0].Source)
	assert.Equal(t, "configs/.gitconfig", git.ConfigFiles[0].BundlePath)
	assert.Equal(t, filepath.Join(homeDir, ".gitignore_global"), git.ConfigFiles[1].Source)

	// gpg.format overrides the signing method.
	assert.Equal(t, "ssh", git.SigningMethod)
	assert.Equal(t, "osxkeychain", git.CredentialHelper)
	assert.Equal(t, filepath.Join(homeDir, ".git-templates"), git.TemplateDir)
}

func TestGitConfigScanner_Scan_GPGSigningMethod(t *testing.T) {
	homeDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(homeDir, ".gitconfig"), []byte("[user]\n\tname = Test\n"), 0o644))

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"git": {Output: "", Err: nil},
			"git config --global core.excludesfile": {
				Output: "", Err: fmt.Errorf("exit status 1"),
			},
			"git config --global user.signingkey": {
				Output: "ABCDEF1234567890",
			},
			"git config --global gpg.format": {
				Output: "", Err: fmt.Errorf("exit status 1"),
			},
			"git config --global credential.helper": {
				Output: "", Err: fmt.Errorf("exit status 1"),
			},
			"git config --global init.templateDir": {
				Output: "", Err: fmt.Errorf("exit status 1"),
			},
		},
	}

	s := NewGitConfigScanner(homeDir, mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.Git)
	// When gpg.format is missing but signingkey is present, method defaults to "gpg".
	assert.Equal(t, "gpg", result.Git.SigningMethod)
}

func TestGitConfigScanner_Scan_PartialConfig(t *testing.T) {
	// ~/.gitconfig does not exist, only credential.helper is set.
	homeDir := t.TempDir()

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"git": {Output: "", Err: nil},
			"git config --global core.excludesfile": {
				Output: "", Err: fmt.Errorf("exit status 1"),
			},
			"git config --global user.signingkey": {
				Output: "", Err: fmt.Errorf("exit status 1"),
			},
			"git config --global gpg.format": {
				Output: "", Err: fmt.Errorf("exit status 1"),
			},
			"git config --global credential.helper": {
				Output: "store",
			},
			"git config --global init.templateDir": {
				Output: "", Err: fmt.Errorf("exit status 1"),
			},
		},
	}

	s := NewGitConfigScanner(homeDir, mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.Git)

	// No config files (gitconfig doesn't exist), but credential helper set.
	assert.Empty(t, result.Git.ConfigFiles)
	assert.Equal(t, "store", result.Git.CredentialHelper)
	assert.Empty(t, result.Git.SigningMethod)
	assert.Empty(t, result.Git.TemplateDir)
}

func TestGitConfigScanner_Scan_NothingFound(t *testing.T) {
	// git is installed but nothing is configured — result.Git should be nil.
	homeDir := t.TempDir()

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"git": {Output: "", Err: nil},
			"git config --global core.excludesfile": {
				Output: "", Err: fmt.Errorf("exit status 1"),
			},
			"git config --global user.signingkey": {
				Output: "", Err: fmt.Errorf("exit status 1"),
			},
			"git config --global gpg.format": {
				Output: "", Err: fmt.Errorf("exit status 1"),
			},
			"git config --global credential.helper": {
				Output: "", Err: fmt.Errorf("exit status 1"),
			},
			"git config --global init.templateDir": {
				Output: "", Err: fmt.Errorf("exit status 1"),
			},
		},
	}

	s := NewGitConfigScanner(homeDir, mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	assert.Nil(t, result.Git)
}
