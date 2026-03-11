package bundler

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	machinist "github.com/moinsen-dev/machinist"
	"github.com/moinsen-dev/machinist/internal/domain"
)

// GenerateRestoreScript renders the restore shell script from a Snapshot
// using the embedded templates.
func GenerateRestoreScript(snapshot *domain.Snapshot) (string, error) {
	funcMap := template.FuncMap{
		"base": filepath.Base,
	}
	tmpl, err := template.New("").Funcs(funcMap).ParseFS(machinist.TemplateFS, "templates/*.tmpl", "templates/stages/*.tmpl")
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

// GenerateRestoreScripts renders per-group restore scripts and a thin
// orchestrator that runs them sequentially. It returns a map of
// filename -> content (e.g. "01-foundation.sh" -> "#!/bin/bash ...").
// Groups with no data in the snapshot are skipped.
func GenerateRestoreScripts(snapshot *domain.Snapshot) (map[string]string, error) {
	funcMap := template.FuncMap{
		"base": filepath.Base,
	}
	tmpl, err := template.New("").Funcs(funcMap).ParseFS(
		machinist.TemplateFS,
		"templates/*.tmpl",
		"templates/stages/*.tmpl",
		"templates/groups/*.tmpl",
	)
	if err != nil {
		return nil, fmt.Errorf("parse templates: %w", err)
	}

	scripts := make(map[string]string)
	var scriptNames []string

	for _, group := range domain.RestoreGroups() {
		if !group.HasData(snapshot) {
			continue
		}

		data := NewGroupTemplateData(snapshot, group)
		tmplName := group.ScriptName + ".tmpl"

		var buf bytes.Buffer
		if err := tmpl.ExecuteTemplate(&buf, tmplName, data); err != nil {
			return nil, fmt.Errorf("execute template %s: %w", tmplName, err)
		}

		scripts[group.ScriptName] = buf.String()
		scriptNames = append(scriptNames, group.ScriptName)
	}

	scripts["install.command"] = generateOrchestrator(snapshot, scriptNames)

	return scripts, nil
}

// generateOrchestrator builds a thin shell script that runs each group
// script sequentially with pass/fail tracking.
func generateOrchestrator(snapshot *domain.Snapshot, scriptNames []string) string {
	var b strings.Builder

	b.WriteString("#!/bin/bash\n")
	b.WriteString("set -uo pipefail\n\n")
	b.WriteString("cd \"$(dirname \"$0\")\"\n\n")

	fmt.Fprintf(&b, "# machinist restore orchestrator\n")
	fmt.Fprintf(&b, "# Generated: %s\n", snapshot.Meta.CreatedAt)
	fmt.Fprintf(&b, "# Source: %s (%s, %s)\n\n",
		snapshot.Meta.SourceHostname,
		snapshot.Meta.SourceOSVersion,
		snapshot.Meta.SourceArch)

	fmt.Fprintf(&b, "TOTAL=%d\n", len(scriptNames))
	b.WriteString("PASSED=0\n")
	b.WriteString("FAILED=0\n\n")

	b.WriteString("START_TIME=$(date +%s)\n\n")

	for _, name := range scriptNames {
		fmt.Fprintf(&b, "if [ -f \"%s\" ]; then\n", name)
		fmt.Fprintf(&b, "  echo \"==> Running %s ...\"\n", name)
		fmt.Fprintf(&b, "  if bash \"%s\"; then\n", name)
		b.WriteString("    PASSED=$((PASSED + 1))\n")
		b.WriteString("  else\n")
		b.WriteString("    FAILED=$((FAILED + 1))\n")
		b.WriteString("  fi\n")
		b.WriteString("fi\n\n")
	}

	b.WriteString("END_TIME=$(date +%s)\n")
	b.WriteString("ELAPSED=$((END_TIME - START_TIME))\n\n")

	b.WriteString("echo \"\"\n")
	b.WriteString("echo \"machinist restore completed in ${ELAPSED}s\"\n")
	b.WriteString("echo \"  $PASSED of $TOTAL groups succeeded\"\n")
	b.WriteString("if [ $FAILED -gt 0 ]; then\n")
	b.WriteString("  echo \"  $FAILED groups failed\"\n")
	b.WriteString("  exit 1\n")
	b.WriteString("fi\n")

	return b.String()
}
