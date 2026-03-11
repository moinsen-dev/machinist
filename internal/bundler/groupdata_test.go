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
	}

	group := domain.RestoreGroup{
		ID:             "01-homebrew",
		Name:           "homebrew",
		Label:          "Homebrew Packages",
		ScriptName:     "01-homebrew.sh",
		SnapshotFields: []string{"Homebrew"},
	}

	data := NewGroupTemplateData(snap, group)

	assert.Equal(t, "Homebrew Packages", data.GroupLabel)
	assert.Equal(t, "01-homebrew", data.GroupID)
	assert.Equal(t, 1, data.StageCount) // Homebrew is non-nil
}

func TestNewGroupTemplateData_SnapshotFieldsAccessible(t *testing.T) {
	snap := &domain.Snapshot{
		Meta: newMeta(),
		Homebrew: &domain.HomebrewSection{
			Formulae: []domain.Package{{Name: "wget"}},
		},
	}

	group := domain.RestoreGroup{
		ID:             "01-homebrew",
		Name:           "homebrew",
		Label:          "Homebrew Packages",
		ScriptName:     "01-homebrew.sh",
		SnapshotFields: []string{"Homebrew"},
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
		ID:             "04-runtimes",
		Name:           "runtimes",
		Label:          "Runtime Installers",
		ScriptName:     "04-runtimes.sh",
		SnapshotFields: []string{"Node", "Python", "Rust"},
	}

	data := NewGroupTemplateData(snap, group)

	assert.Equal(t, 0, data.StageCount)
	assert.Equal(t, "Runtime Installers", data.GroupLabel)
	assert.Equal(t, "04-runtimes", data.GroupID)
}
