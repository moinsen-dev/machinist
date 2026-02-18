package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/moinsen-dev/machinist/internal/scanner"
)

var (
	dimStyle     = lipgloss.NewStyle().Faint(true)
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))  // green
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))  // red
	scannerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("4"))  // blue
	counterStyle = lipgloss.NewStyle().Faint(true)
)

// newProgressWriter returns a ProgressFunc that writes scan progress to w.
func newProgressWriter(w io.Writer) scanner.ProgressFunc {
	return func(e scanner.ProgressEvent) {
		if !e.Done {
			// Starting a scanner — print on same line
			counter := counterStyle.Render(fmt.Sprintf("[%d/%d]", e.Index+1, e.Total))
			name := scannerStyle.Render(e.Name)
			fmt.Fprintf(w, "  %s Scanning %s...", counter, name)
			return
		}

		// Scanner finished
		if e.Err != nil {
			mark := errorStyle.Render("✗")
			dur := dimStyle.Render(fmt.Sprintf("(%.1fs)", e.Duration.Seconds()))
			reason := dimStyle.Render(shortError(e.Err))
			fmt.Fprintf(w, " %s %s %s\n", mark, dur, reason)
		} else {
			mark := successStyle.Render("✓")
			dur := dimStyle.Render(fmt.Sprintf("(%.1fs)", e.Duration.Seconds()))
			fmt.Fprintf(w, " %s %s\n", mark, dur)
		}
	}
}

// shortError returns a brief error message, stripping the "scanner xyz: " prefix.
func shortError(err error) string {
	msg := err.Error()
	if idx := strings.LastIndex(msg, ": "); idx != -1 {
		return msg[idx+2:]
	}
	return msg
}
