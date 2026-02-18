package bundler

import (
	"bytes"
	"fmt"
	"text/template"

	machinist "github.com/moinsen-dev/machinist"
	"github.com/moinsen-dev/machinist/internal/domain"
)

// GenerateRestoreScript renders the restore shell script from a Snapshot
// using the embedded templates.
func GenerateRestoreScript(snapshot *domain.Snapshot) (string, error) {
	tmpl, err := template.ParseFS(machinist.TemplateFS, "templates/*.tmpl", "templates/stages/*.tmpl")
	if err != nil {
		return "", fmt.Errorf("parse templates: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "install.command.tmpl", snapshot); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}
	return buf.String(), nil
}

// GenerateChecklist renders the post-restore checklist markdown from a Snapshot.
func GenerateChecklist(snapshot *domain.Snapshot) (string, error) {
	tmpl, err := template.ParseFS(machinist.TemplateFS, "templates/checklist.md.tmpl")
	if err != nil {
		return "", fmt.Errorf("parse checklist template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "checklist.md.tmpl", snapshot); err != nil {
		return "", fmt.Errorf("execute checklist template: %w", err)
	}
	return buf.String(), nil
}

// GenerateReadme renders the README markdown for the DMG bundle from a Snapshot.
func GenerateReadme(snapshot *domain.Snapshot) (string, error) {
	tmpl, err := template.ParseFS(machinist.TemplateFS, "templates/README.md.tmpl")
	if err != nil {
		return "", fmt.Errorf("parse README template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "README.md.tmpl", snapshot); err != nil {
		return "", fmt.Errorf("execute README template: %w", err)
	}
	return buf.String(), nil
}
