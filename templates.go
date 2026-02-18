package machinist

import "embed"

//go:embed templates/*.tmpl templates/stages/*.tmpl
var TemplateFS embed.FS
