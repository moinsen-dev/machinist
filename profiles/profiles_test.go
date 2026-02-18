package profiles_test

import (
	"testing"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/profiles"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestList(t *testing.T) {
	names, err := profiles.List()
	require.NoError(t, err)

	assert.GreaterOrEqual(t, len(names), 7, "should have at least 7 profiles")

	expected := []string{"minimal", "fullstack-js", "flutter-ios", "python-data", "go-dev", "rust-dev", "devops"}
	for _, name := range expected {
		assert.Contains(t, names, name, "should contain profile %q", name)
	}

	// Names should be sorted.
	for i := 1; i < len(names); i++ {
		assert.Less(t, names[i-1], names[i], "profiles should be sorted")
	}
}

func TestGet_Minimal(t *testing.T) {
	snap, err := profiles.Get("minimal")
	require.NoError(t, err)
	require.NotNil(t, snap)

	assert.Equal(t, "profile:minimal", snap.Meta.SourceHostname)
	assert.Equal(t, "profile", snap.Meta.MachinistVersion)

	require.NotNil(t, snap.Homebrew, "Homebrew section must be present")
	formulaeNames := make([]string, len(snap.Homebrew.Formulae))
	for i, f := range snap.Homebrew.Formulae {
		formulaeNames[i] = f.Name
	}
	assert.Contains(t, formulaeNames, "git")
	assert.Contains(t, formulaeNames, "starship")

	require.NotNil(t, snap.Shell, "Shell section must be present")
	assert.Equal(t, "/bin/zsh", snap.Shell.DefaultShell)
}

func TestGet_FullstackJS(t *testing.T) {
	snap, err := profiles.Get("fullstack-js")
	require.NoError(t, err)
	require.NotNil(t, snap)

	require.NotNil(t, snap.Node, "Node section must be present")
	assert.Equal(t, "fnm", snap.Node.Manager)
	assert.Equal(t, "v22", snap.Node.DefaultVersion)
	assert.GreaterOrEqual(t, len(snap.Node.GlobalPackages), 3)
}

func TestGet_NotFound(t *testing.T) {
	_, err := profiles.Get("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nonexistent")
}

func TestMerge(t *testing.T) {
	base, err := profiles.Get("minimal")
	require.NoError(t, err)

	override := &domain.Snapshot{
		Homebrew: &domain.HomebrewSection{
			Casks: []domain.Package{{Name: "docker"}},
		},
	}

	merged := profiles.Merge(base, override)

	require.NotNil(t, merged.Homebrew)

	// Merged should have all base formulae.
	formulaeNames := make([]string, len(merged.Homebrew.Formulae))
	for i, f := range merged.Homebrew.Formulae {
		formulaeNames[i] = f.Name
	}
	assert.Contains(t, formulaeNames, "git")
	assert.Contains(t, formulaeNames, "starship")

	// Merged should have both base casks AND override cask.
	caskNames := make([]string, len(merged.Homebrew.Casks))
	for i, c := range merged.Homebrew.Casks {
		caskNames[i] = c.Name
	}
	assert.Contains(t, caskNames, "visual-studio-code") // from base
	assert.Contains(t, caskNames, "iterm2")              // from base
	assert.Contains(t, caskNames, "docker")              // from override
}

func TestMerge_OverrideSection(t *testing.T) {
	base, err := profiles.Get("minimal")
	require.NoError(t, err)

	// Base (minimal) should NOT have a Node section.
	assert.Nil(t, base.Node, "minimal profile should not have a Node section")

	override := &domain.Snapshot{
		Node: &domain.NodeSection{
			Manager:        "fnm",
			DefaultVersion: "v22",
			GlobalPackages: []domain.Package{{Name: "typescript"}},
		},
	}

	merged := profiles.Merge(base, override)

	require.NotNil(t, merged.Node, "merged should have Node section from override")
	assert.Equal(t, "fnm", merged.Node.Manager)
	assert.Equal(t, "v22", merged.Node.DefaultVersion)
	assert.Len(t, merged.Node.GlobalPackages, 1)
}
