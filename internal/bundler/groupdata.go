package bundler

import "github.com/moinsen-dev/machinist/internal/domain"

// GroupTemplateData embeds the Snapshot and adds group-specific fields
// needed by group templates (GroupLabel, GroupID, StageCount).
type GroupTemplateData struct {
	*domain.Snapshot
	GroupLabel string
	GroupID    string
	StageCount int
}

// NewGroupTemplateData creates the template data for a group.
func NewGroupTemplateData(snap *domain.Snapshot, group domain.RestoreGroup) GroupTemplateData {
	return GroupTemplateData{
		Snapshot:   snap,
		GroupLabel: group.Label,
		GroupID:    group.ID,
		StageCount: group.StageCount(snap),
	}
}
