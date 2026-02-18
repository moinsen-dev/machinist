package git

import (
	"context"
	"os"
	"path/filepath"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/scanner"
	"github.com/moinsen-dev/machinist/internal/util"
)

// GitReposScanner scans for Git repositories in specified paths.
type GitReposScanner struct {
	searchPaths []string
	cmd         util.CommandRunner
}

// NewGitReposScanner creates a new GitReposScanner with the given search paths and CommandRunner.
func NewGitReposScanner(searchPaths []string, cmd util.CommandRunner) *GitReposScanner {
	return &GitReposScanner{
		searchPaths: searchPaths,
		cmd:         cmd,
	}
}

func (g *GitReposScanner) Name() string        { return "git-repos" }
func (g *GitReposScanner) Description() string  { return "Scans for Git repositories in specified paths" }
func (g *GitReposScanner) Category() string     { return "git" }

// Scan walks the search paths looking for .git directories and collects repository metadata.
func (g *GitReposScanner) Scan(ctx context.Context) (*scanner.ScanResult, error) {
	result := &scanner.ScanResult{
		ScannerName: g.Name(),
	}

	if !g.cmd.IsInstalled(ctx, "git") {
		return result, nil
	}

	if len(g.searchPaths) == 0 {
		return result, nil
	}

	section := &domain.GitReposSection{
		SearchPaths: g.searchPaths,
	}

	for _, searchPath := range g.searchPaths {
		repos, err := g.findRepos(ctx, searchPath)
		if err != nil {
			continue
		}
		section.Repositories = append(section.Repositories, repos...)
	}

	if section.Repositories == nil {
		section.Repositories = []domain.Repository{}
	}

	result.GitRepos = section
	return result, nil
}

// findRepos walks the given path looking for .git directories and returns Repository structs.
func (g *GitReposScanner) findRepos(ctx context.Context, root string) ([]domain.Repository, error) {
	var repos []domain.Repository

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip inaccessible directories
		}
		if d.IsDir() && d.Name() == ".git" {
			repoPath := filepath.Dir(path)
			repo := g.buildRepo(ctx, repoPath)
			repos = append(repos, repo)
			return filepath.SkipDir
		}
		return nil
	})

	return repos, err
}

// buildRepo creates a Repository struct by querying git for remote URL and current branch.
func (g *GitReposScanner) buildRepo(ctx context.Context, repoPath string) domain.Repository {
	repo := domain.Repository{
		Path: repoPath,
	}

	remoteURL, err := g.cmd.Run(ctx, "git", "-C", repoPath, "remote", "get-url", "origin")
	if err == nil {
		repo.Remote = remoteURL
	}

	branch, err := g.cmd.Run(ctx, "git", "-C", repoPath, "branch", "--show-current")
	if err == nil {
		repo.Branch = branch
	}

	return repo
}
