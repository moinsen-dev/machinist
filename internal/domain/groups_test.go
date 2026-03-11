package domain

import (
	"reflect"
	"testing"
)

func TestRestoreGroupsReturns6Groups(t *testing.T) {
	groups := RestoreGroups()
	if len(groups) != 6 {
		t.Fatalf("expected 6 groups, got %d", len(groups))
	}

	if groups[0].ID != "01-homebrew" {
		t.Errorf("first group ID = %q, want %q", groups[0].ID, "01-homebrew")
	}
	if groups[len(groups)-1].ID != "06-macos" {
		t.Errorf("last group ID = %q, want %q", groups[len(groups)-1].ID, "06-macos")
	}
}

func TestAllGroupsHaveSnapshotFields(t *testing.T) {
	for _, g := range RestoreGroups() {
		if len(g.SnapshotFields) == 0 {
			t.Errorf("group %q has no SnapshotFields", g.ID)
		}
	}
}

func TestGroupByName(t *testing.T) {
	g, ok := GroupByName("homebrew")
	if !ok {
		t.Fatal("expected to find group 'homebrew'")
	}
	if g.ID != "01-homebrew" {
		t.Errorf("expected ID '01-homebrew', got %q", g.ID)
	}

	g, ok = GroupByName("configs")
	if !ok {
		t.Fatal("expected to find group 'configs'")
	}
	if g.ID != "03-configs" {
		t.Errorf("expected ID '03-configs', got %q", g.ID)
	}

	_, ok = GroupByName("foundation")
	if ok {
		t.Fatal("expected false for old group name 'foundation'")
	}
}

func TestGroupNames(t *testing.T) {
	names := GroupNames()
	if len(names) != 6 {
		t.Fatalf("expected 6 names, got %d", len(names))
	}
	if names[0] != "homebrew" {
		t.Errorf("first name = %q, want 'homebrew'", names[0])
	}
}

func TestRestoreGroups_SnapshotFieldsAreValid(t *testing.T) {
	snapType := reflect.TypeOf(Snapshot{})
	for _, g := range RestoreGroups() {
		for _, fieldName := range g.SnapshotFields {
			_, ok := snapType.FieldByName(fieldName)
			if !ok {
				t.Errorf("group %s references non-existent Snapshot field: %s", g.ID, fieldName)
			}
		}
	}
}

func TestHasDataReturnsFalseOnEmptySnapshot(t *testing.T) {
	snap := &Snapshot{}
	for _, g := range RestoreGroups() {
		if g.HasData(snap) {
			t.Errorf("group %q HasData should be false on empty snapshot", g.ID)
		}
	}
}

func TestHasDataReturnsTrueWhenFieldSet(t *testing.T) {
	snap := &Snapshot{
		Homebrew: &HomebrewSection{},
	}
	g, _ := GroupByName("homebrew")
	if !g.HasData(snap) {
		t.Error("expected HasData=true when Homebrew is set")
	}
}

func TestGroupHasData_Secrets(t *testing.T) {
	snap := &Snapshot{
		SSH: &SSHSection{},
	}
	g, _ := GroupByName("secrets")
	if !g.HasData(snap) {
		t.Error("expected HasData=true for secrets when SSH is set")
	}
}

func TestGroupHasData_Configs(t *testing.T) {
	snap := &Snapshot{
		Shell: &ShellSection{},
	}
	g, _ := GroupByName("configs")
	if !g.HasData(snap) {
		t.Error("expected HasData=true for configs when Shell is set")
	}
}

func TestStageCount(t *testing.T) {
	snap := &Snapshot{
		SSH: &SSHSection{},
		GPG: &GPGSection{},
	}
	g, _ := GroupByName("secrets")
	got := g.StageCount(snap)
	if got != 2 {
		t.Errorf("StageCount = %d, want 2", got)
	}
}

func TestStageCountZeroOnEmpty(t *testing.T) {
	snap := &Snapshot{}
	g, _ := GroupByName("homebrew")
	got := g.StageCount(snap)
	if got != 0 {
		t.Errorf("StageCount = %d, want 0", got)
	}
}
