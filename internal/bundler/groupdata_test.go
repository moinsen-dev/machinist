package bundler

import (
	"testing"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestNewGroupTemplateData_SetsFields(t *testing.T) {
	snap := &domain.Snapshot{
		Meta: newMeta(),
		Homebrew: &domain.HomebrewSection{
			Formulae: []domain.Package{{Name: "git"}},
		},
		SSH: &domain.SSHSection{
			Keys: []string{"id_ed25519"},
		},
	}

	group := domain.RestoreGroup{
		ID:             "01-foundation",
		Name:           "foundation",
		Label:          "Foundation",
		ScriptName:     "01-foundation.sh",
		SnapshotFields: []string{"Homebrew", "SSH", "GPG", "Git"},
	}

	data := NewGroupTemplateData(snap, group)

	assert.Equal(t, "Foundation", data.GroupLabel)
	assert.Equal(t, "01-foundation", data.GroupID)
	assert.Equal(t, 2, data.StageCount) // Homebrew + SSH are non-nil
}

func TestNewGroupTemplateData_SnapshotFieldsAccessible(t *testing.T) {
	snap := &domain.Snapshot{
		Meta: newMeta(),
		Homebrew: &domain.HomebrewSection{
			Formulae: []domain.Package{{Name: "wget"}},
		},
	}

	group := domain.RestoreGroup{
		ID:             "01-foundation",
		Name:           "foundation",
		Label:          "Foundation",
		ScriptName:     "01-foundation.sh",
		SnapshotFields: []string{"Homebrew", "SSH", "GPG", "Git"},
	}

	data := NewGroupTemplateData(snap, group)

	// Snapshot fields should be accessible through embedding
	assert.Equal(t, "test-mac", data.Meta.SourceHostname)
	assert.NotNil(t, data.Homebrew)
	assert.Equal(t, "wget", data.Homebrew.Formulae[0].Name)
}

func TestNewGroupTemplateData_ZeroStageCount(t *testing.T) {
	snap := &domain.Snapshot{
		Meta: newMeta(),
	}

	group := domain.RestoreGroup{
		ID:             "03-runtimes",
		Name:           "runtimes",
		Label:          "Runtimes",
		ScriptName:     "03-runtimes.sh",
		SnapshotFields: []string{"Node", "Python", "Rust"},
	}

	data := NewGroupTemplateData(snap, group)

	assert.Equal(t, 0, data.StageCount)
	assert.Equal(t, "Runtimes", data.GroupLabel)
	assert.Equal(t, "03-runtimes", data.GroupID)
}
