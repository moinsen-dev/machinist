package machinist

import "embed"

//go:embed templates/*.tmpl templates/stages/*.tmpl templates/groups/*.tmpl
var TemplateFS embed.FS
