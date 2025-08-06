package bundle

import (
	"embed"
	"fmt"
	"os"
	"sync"
	"sync/atomic"

	"github.com/open-policy-agent/opa/v1/bundle"

	rio "github.com/styrainc/regal/internal/io"
)

// Bundle FS will include the tests as well, but since that has negligible impact on the size of the binary,
// it's preferable to filter them out from the bundle than to e.g. create a separate directory for tests.
var (
	//go:embed *
	regalBundle    embed.FS
	devPath        = os.Getenv("REGAL_BUNDLE_PATH")
	lastErrMsg     = atomic.Pointer[string]{}
	EmbeddedBundle = sync.OnceValue(func() *bundle.Bundle {
		return rio.MustLoadRegalBundleFS(regalBundle)
	})
	successLogOnce = sync.OnceFunc(func() {
		fmt.Fprintln(os.Stderr, "Successfully loaded development bundle")
	})
)

func init() {
	if devPath != "" {
		fmt.Fprintln(os.Stderr, "REGAL_BUNDLE_PATH set. Will attempt using development bundle from:", devPath)
	}
}

// LoadedBundle contains the Regal bundle.
func LoadedBundle() *bundle.Bundle {
	// For development, allow bundle to be loaded dynamically from path instead
	// of the normal one embedded in the compiled binary. This allows editing e.g.
	// LSP policies while the language server is running. This should be considered
	// *very* experimental at this point.
	if devPath != "" {
		b, err := rio.LoadRegalBundlePath(devPath)
		if err == nil {
			if last := lastErrMsg.Load(); last != nil {
				lastErrMsg.Store(nil)

				fmt.Fprintln(os.Stderr, "development bundle back to a good state, no longer using embedded bundle")
			}

			successLogOnce()

			return b
		}

		// Avoid flooding the console/logs with the same error message
		if curr, last := err.Error(), lastErrMsg.Load(); last == nil || *last != curr {
			fmt.Fprintf(os.Stderr, "error loading development bundle from %s:\n%v\n", devPath, err)

			lastErrMsg.Store(&curr)
		}

		// Now fallback to the embedded bundle if the development path fails, as the bundle may
		// be requested at any time (and very frequently!) from the various LSP commands, and we
		// don't want a broken language server while developing!
	}

	return EmbeddedBundle()
}
