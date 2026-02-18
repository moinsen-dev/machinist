package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ScannerItem represents a single scanner entry in the selection list.
type ScannerItem struct {
	Name        string
	Description string
	Category    string
	Selected    bool
}

// ScannerSelectModel is a bubbletea model that lets the user pick which
// scanners to include in a snapshot run.
type ScannerSelectModel struct {
	items   []ScannerItem
	cursor  int
	done    bool
	quitted bool
}

// NewScannerSelectModel creates a new model with the given items.
// All items are selected by default.
func NewScannerSelectModel(items []ScannerItem) ScannerSelectModel {
	// Ensure every item starts selected.
	for i := range items {
		items[i].Selected = true
	}
	return ScannerSelectModel{
		items: items,
	}
}

// Selected returns the names of the scanners the user chose.
func (m ScannerSelectModel) Selected() []string {
	var names []string
	for _, it := range m.items {
		if it.Selected {
			names = append(names, it.Name)
		}
	}
	return names
}

// Done reports whether the user confirmed the selection.
func (m ScannerSelectModel) Done() bool { return m.done }

// Quitted reports whether the user cancelled.
func (m ScannerSelectModel) Quitted() bool { return m.quitted }

// Init satisfies tea.Model.
func (m ScannerSelectModel) Init() tea.Cmd { return nil }

// Update satisfies tea.Model.
func (m ScannerSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyUp:
			if m.cursor > 0 {
				m.cursor--
			}
		case tea.KeyDown:
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case tea.KeySpace:
			if len(m.items) > 0 {
				m.items[m.cursor].Selected = !m.items[m.cursor].Selected
			}
		case tea.KeyEnter:
			m.done = true
			return m, tea.Quit
		case tea.KeyEsc:
			m.quitted = true
			return m, tea.Quit
		case tea.KeyRunes:
			switch string(msg.Runes) {
			case "q":
				m.quitted = true
				return m, tea.Quit
			case "a":
				allSelected := true
				for _, it := range m.items {
					if !it.Selected {
						allSelected = false
						break
					}
				}
				for i := range m.items {
					m.items[i].Selected = !allSelected
				}
			}
		}
	}
	return m, nil
}

// View satisfies tea.Model.
func (m ScannerSelectModel) View() string {
	if m.done {
		return ""
	}
	if m.quitted {
		return ""
	}

	titleStyle := lipgloss.NewStyle().Bold(true).MarginBottom(1)
	cursorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("86"))
	categoryStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).MarginTop(1)

	var b strings.Builder
	b.WriteString(titleStyle.Render("Select scanners to run:"))
	b.WriteString("\n")

	for i, item := range m.items {
		checkbox := "[ ]"
		if item.Selected {
			checkbox = "[x]"
		}

		line := fmt.Sprintf("%s %s %s", checkbox, item.Name, categoryStyle.Render("("+item.Category+")"))

		if i == m.cursor {
			line = cursorStyle.Render("> " + line)
		} else {
			line = "  " + line
		}
		b.WriteString(line)
		b.WriteString("\n")
	}

	b.WriteString(helpStyle.Render("space: toggle | enter: confirm | a: toggle all | q/esc: quit"))
	return b.String()
}
