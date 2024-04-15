package fixes

import (
	"fmt"
	"path/filepath"

	"github.com/open-policy-agent/opa/format"
)

func Fmt(in []byte, opts *FmtOptions) (bool, []byte, error) {
	filename := "unknown.rego"

	opaFmtOpts := format.Opts{}

	if opts != nil {
		if opts.Filename != "" {
			filename = opts.Filename
		}

		opaFmtOpts = opts.OPAFmtOpts
	}

	formatted, err := format.SourceWithOpts(filepath.Base(filename), in, opaFmtOpts)
	if err != nil {
		return false, nil, fmt.Errorf("failed to format: %w", err)
	}

	return string(in) != string(formatted), formatted, nil
}
