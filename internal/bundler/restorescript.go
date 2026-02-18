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
