//nolint:gochecknoglobals
package embeds

import (
	"embed"
)

//go:embed templates
var EmbedTemplatesFS embed.FS

//go:embed schemas
var SchemasFS embed.FS
