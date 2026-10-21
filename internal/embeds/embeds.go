package embeds

import (
	"embed"
)

var (
	//go:embed templates
	EmbedTemplatesFS embed.FS
	//go:embed schemas
	SchemasFS embed.FS
)
