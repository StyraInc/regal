package bundle

import (
	"embed"
	"sync"

	"github.com/open-policy-agent/opa/bundle"

	rio "github.com/styrainc/regal/internal/io"
)

// Bundle FS will include the tests as well, but since that has negligible impact on the size of the binary,
// it's preferable to filter them out from the bundle than to e.g. create a separate directory for tests
//
//go:embed *
var Bundle embed.FS

//nolint:gochecknoglobals
var (
	loadedBundle     bundle.Bundle
	loadedBundleOnce sync.Once
)

// LoadedBundle returns the loaded Regal bundle for this version of Regal.
func LoadedBundle() *bundle.Bundle {
	loadedBundleOnce.Do(func() {
		loadedBundle = rio.MustLoadRegalBundleFS(Bundle)
	})

	return &loadedBundle
}
