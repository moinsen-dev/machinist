package git

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/moinsen-dev/machinist/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGitReposScanner_Name(t *testing.T) {
	s := NewGitReposScanner(nil, &util.MockCommandRunner{})
	assert.Equal(t, "git-repos", s.Name())
}

func TestGitReposScanner_Description(t *testing.T) {
	s := NewGitReposScanner(nil, &util.MockCommandRunner{})
	assert.NotEmpty(t, s.Description())
}

func TestGitReposScanner_Category(t *testing.T) {
	s := NewGitReposScanner(nil, &util.MockCommandRunner{})
	assert.Equal(t, "git", s.Category())
}

func TestGitReposScanner_Scan_FindsRepos(t *testing.T) {
	tmpDir := t.TempDir()

	// Create repo directories with .git subdirs
	projectA := filepath.Join(tmpDir, "project-a")
	projectB := filepath.Join(tmpDir, "project-b")
	notARepo := filepath.Join(tmpDir, "not-a-repo")

	require.NoError(t, os.MkdirAll(filepath.Join(projectA, ".git"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(projectB, ".git"), 0o755))
	require.NoError(t, os.MkdirAll(notARepo, 0o755))

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"git": {Output: "", Err: nil}, // IsInstalled
			fmt.Sprintf("git -C %s remote get-url origin", projectA): {
				Output: "https://github.com/user/project-a.git",
			},
			fmt.Sprintf("git -C %s branch --show-current", projectA): {
				Output: "main",
			},
			fmt.Sprintf("git -C %s remote get-url origin", projectB): {
				Output: "https://github.com/user/project-b.git",
			},
			fmt.Sprintf("git -C %s branch --show-current", projectB): {
				Output: "main",
			},
		},
	}

	s := NewGitReposScanner([]string{tmpDir}, mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.GitRepos)
	assert.Equal(t, []string{tmpDir}, result.GitRepos.SearchPaths)
	require.Len(t, result.GitRepos.Repositories, 2)

	// Sort repos by path for deterministic assertion
	repos := result.GitRepos.Repositories
	sort.Slice(repos, func(i, j int) bool { return repos[i].Path < repos[j].Path })

	assert.Equal(t, projectA, repos[0].Path)
	assert.Equal(t, "https://github.com/user/project-a.git", repos[0].Remote)
	assert.Equal(t, "main", repos[0].Branch)

	assert.Equal(t, projectB, repos[1].Path)
	assert.Equal(t, "https://github.com/user/project-b.git", repos[1].Remote)
	assert.Equal(t, "main", repos[1].Branch)
}

func TestGitReposScanner_Scan_NoSearchPaths(t *testing.T) {
	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"git": {Output: "", Err: nil},
		},
	}

	s := NewGitReposScanner([]string{}, mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	assert.Nil(t, result.GitRepos)
}

func TestGitReposScanner_Scan_NoGitRepos(t *testing.T) {
	tmpDir := t.TempDir()

	// Create directories that are NOT git repos
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "some-dir"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "another-dir"), 0o755))

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"git": {Output: "", Err: nil},
		},
	}

	s := NewGitReposScanner([]string{tmpDir}, mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.GitRepos)
	assert.Empty(t, result.GitRepos.Repositories)
}

func TestGitReposScanner_Scan_NestedRepos(t *testing.T) {
	tmpDir := t.TempDir()

	parent := filepath.Join(tmpDir, "parent")
	child := filepath.Join(parent, "child")

	require.NoError(t, os.MkdirAll(filepath.Join(parent, ".git"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(child, ".git"), 0o755))

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"git": {Output: "", Err: nil},
			fmt.Sprintf("git -C %s remote get-url origin", parent): {
				Output: "https://github.com/user/parent.git",
			},
			fmt.Sprintf("git -C %s branch --show-current", parent): {
				Output: "main",
			},
			fmt.Sprintf("git -C %s remote get-url origin", child): {
				Output: "https://github.com/user/child.git",
			},
			fmt.Sprintf("git -C %s branch --show-current", child): {
				Output: "develop",
			},
		},
	}

	s := NewGitReposScanner([]string{tmpDir}, mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.GitRepos)
	require.Len(t, result.GitRepos.Repositories, 2)

	repos := result.GitRepos.Repositories
	sort.Slice(repos, func(i, j int) bool { return repos[i].Path < repos[j].Path })

	assert.Equal(t, parent, repos[0].Path)
	assert.Equal(t, child, repos[1].Path)
}

func TestGitReposScanner_Scan_RemoteError(t *testing.T) {
	tmpDir := t.TempDir()

	repoPath := filepath.Join(tmpDir, "no-remote")
	require.NoError(t, os.MkdirAll(filepath.Join(repoPath, ".git"), 0o755))

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"git": {Output: "", Err: nil},
			fmt.Sprintf("git -C %s remote get-url origin", repoPath): {
				Output: "",
				Err:    fmt.Errorf("fatal: No such remote 'origin'"),
			},
			fmt.Sprintf("git -C %s branch --show-current", repoPath): {
				Output: "main",
			},
		},
	}

	s := NewGitReposScanner([]string{tmpDir}, mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.GitRepos)
	require.Len(t, result.GitRepos.Repositories, 1)

	repo := result.GitRepos.Repositories[0]
	assert.Equal(t, repoPath, repo.Path)
	assert.Empty(t, repo.Remote)
	assert.Equal(t, "main", repo.Branch)
}

func TestGitReposScanner_Scan_GitNotInstalled(t *testing.T) {
	// No "git" key in Responses means IsInstalled returns false
	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{},
	}

	s := NewGitReposScanner([]string{"/some/path"}, mock)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	assert.Nil(t, result.GitRepos)
	assert.Empty(t, result.GitRepos)
}
