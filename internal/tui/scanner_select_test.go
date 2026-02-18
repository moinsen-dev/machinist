package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sampleItems() []ScannerItem {
	return []ScannerItem{
		{Name: "homebrew", Description: "Scan Homebrew packages", Category: "packages"},
		{Name: "shell", Description: "Scan shell config", Category: "shell"},
		{Name: "vscode", Description: "Scan VS Code extensions", Category: "editors"},
	}
}

func TestScannerSelectModel_Init(t *testing.T) {
	m := NewScannerSelectModel(sampleItems())

	// All items should be selected by default.
	for i, item := range m.items {
		assert.True(t, item.Selected, "item %d (%s) should be selected by default", i, item.Name)
	}

	// Init should return nil cmd.
	cmd := m.Init()
	assert.Nil(t, cmd)
}

func TestScannerSelectModel_Toggle(t *testing.T) {
	m := NewScannerSelectModel(sampleItems())

	// Cursor starts at 0 (homebrew). Toggle it off.
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = model.(ScannerSelectModel)

	assert.False(t, m.items[0].Selected, "first item should be deselected after space")
	assert.True(t, m.items[1].Selected, "second item should remain selected")
	assert.True(t, m.items[2].Selected, "third item should remain selected")

	// Toggle it back on.
	model, _ = m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = model.(ScannerSelectModel)
	assert.True(t, m.items[0].Selected, "first item should be selected again")
}

func TestScannerSelectModel_Navigation(t *testing.T) {
	m := NewScannerSelectModel(sampleItems())
	assert.Equal(t, 0, m.cursor)

	// Move down.
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = model.(ScannerSelectModel)
	assert.Equal(t, 1, m.cursor)

	// Move down again.
	model, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = model.(ScannerSelectModel)
	assert.Equal(t, 2, m.cursor)

	// Move down at bottom stays at bottom.
	model, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = model.(ScannerSelectModel)
	assert.Equal(t, 2, m.cursor)

	// Move up.
	model, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = model.(ScannerSelectModel)
	assert.Equal(t, 1, m.cursor)
}

func TestScannerSelectModel_Selected(t *testing.T) {
	m := NewScannerSelectModel(sampleItems())

	// Deselect the first item (homebrew).
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = model.(ScannerSelectModel)

	selected := m.Selected()
	require.Len(t, selected, 2)
	assert.Equal(t, "shell", selected[0])
	assert.Equal(t, "vscode", selected[1])
}

func TestScannerSelectModel_Quit(t *testing.T) {
	m := NewScannerSelectModel(sampleItems())

	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	m = model.(ScannerSelectModel)

	assert.True(t, m.Quitted())
	assert.False(t, m.Done())
	assert.NotNil(t, cmd, "quit should return a tea.Quit cmd")
}

func TestScannerSelectModel_QuitEsc(t *testing.T) {
	m := NewScannerSelectModel(sampleItems())

	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = model.(ScannerSelectModel)

	assert.True(t, m.Quitted())
	assert.False(t, m.Done())
	assert.NotNil(t, cmd)
}

func TestScannerSelectModel_Enter(t *testing.T) {
	m := NewScannerSelectModel(sampleItems())

	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = model.(ScannerSelectModel)

	assert.True(t, m.Done())
	assert.False(t, m.Quitted())
	assert.NotNil(t, cmd, "enter should return a tea.Quit cmd")
}

func TestScannerSelectModel_View(t *testing.T) {
	m := NewScannerSelectModel(sampleItems())

	view := m.View()
	assert.Contains(t, view, "Select scanners to run:")
	assert.Contains(t, view, "homebrew")
	assert.Contains(t, view, "shell")
	assert.Contains(t, view, "vscode")
	assert.Contains(t, view, "[x]")
	assert.Contains(t, view, "packages")
}

func TestScannerSelectModel_ViewDone(t *testing.T) {
	m := NewScannerSelectModel(sampleItems())
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = model.(ScannerSelectModel)

	assert.Empty(t, m.View())
}

func TestScannerSelectModel_ToggleAll(t *testing.T) {
	m := NewScannerSelectModel(sampleItems())

	// All selected -> toggle all off.
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = model.(ScannerSelectModel)
	for _, it := range m.items {
		assert.False(t, it.Selected, "%s should be deselected", it.Name)
	}

	// All deselected -> toggle all on.
	model, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = model.(ScannerSelectModel)
	for _, it := range m.items {
		assert.True(t, it.Selected, "%s should be selected", it.Name)
	}
}
