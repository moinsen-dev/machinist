package domain

import (
	"testing"
)

func TestRestoreGroupsReturns7Groups(t *testing.T) {
	groups := RestoreGroups()
	if len(groups) != 7 {
		t.Fatalf("expected 7 groups, got %d", len(groups))
	}

	expectedIDs := []string{
		"01-foundation", "02-shell", "03-runtimes",
		"04-editors", "05-infrastructure", "06-repos", "07-system",
	}
	for i, g := range groups {
		if g.ID != expectedIDs[i] {
			t.Errorf("group[%d].ID = %q, want %q", i, g.ID, expectedIDs[i])
		}
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
	g, ok := GroupByName("shell")
	if !ok {
		t.Fatal("expected to find group 'shell'")
	}
	if g.ID != "02-shell" {
		t.Errorf("expected ID '02-shell', got %q", g.ID)
	}

	_, ok = GroupByName("nonexistent")
	if ok {
		t.Fatal("expected false for nonexistent group")
	}
}

func TestGroupNames(t *testing.T) {
	names := GroupNames()
	if len(names) != 7 {
		t.Fatalf("expected 7 names, got %d", len(names))
	}
	if names[0] != "foundation" {
		t.Errorf("first name = %q, want 'foundation'", names[0])
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
	g, _ := GroupByName("foundation")
	if !g.HasData(snap) {
		t.Error("expected HasData=true when Homebrew is set")
	}
}

func TestStageCount(t *testing.T) {
	snap := &Snapshot{
		Homebrew: &HomebrewSection{},
		SSH:      &SSHSection{},
		GPG:      &GPGSection{},
	}
	g, _ := GroupByName("foundation")
	got := g.StageCount(snap)
	if got != 3 {
		t.Errorf("StageCount = %d, want 3", got)
	}
}

func TestStageCountZeroOnEmpty(t *testing.T) {
	snap := &Snapshot{}
	g, _ := GroupByName("foundation")
	got := g.StageCount(snap)
	if got != 0 {
		t.Errorf("StageCount = %d, want 0", got)
	}
}
