package bundle

import (
	"embed"
)

// Bundle FS will include the tests as well, but since that has negligible impact on the size of the binary,
// it's preferable to filter them out from the bundle than to e.g. create a separate directory for tests
//
//go:embed *
var Bundle embed.FS
