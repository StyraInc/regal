//nolint:gochecknoglobals
package embeds

import "embed"

var EmbedBundleFS embed.FS

//go:embed templates
var EmbedTemplatesFS embed.FS

//go:embed schemas/regal-ast.json
var ASTSchema []byte
