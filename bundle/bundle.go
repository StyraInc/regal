package bundle

import (
	"embed"

	rio "github.com/styrainc/regal/internal/io"
)

// Bundle FS will include the tests as well, but since that has negligible impact on the size of the binary,
// it's preferable to filter them out from the bundle than to e.g. create a separate directory for tests
//
//go:embed *
var Bundle embed.FS

// LoadedBundle contains the loaded contents of the Bundle.
var LoadedBundle = rio.MustLoadRegalBundleFS(Bundle)
