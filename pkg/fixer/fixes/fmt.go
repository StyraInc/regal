package fixes

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/open-policy-agent/opa/format"
)

type Fmt struct {
	KeyOverride string
	// OPAFmtOpts are the options to pass to OPA's format.SourceWithOpts
	// function.
	OPAFmtOpts format.Opts
}

func (f *Fmt) Key() string {
	// this allows this fix config to also be registered under another key so that different
	// configurations can be registered under other linter rule keys.
	if f.KeyOverride != "" {
		return f.KeyOverride
	}

	return "opa-fmt"
}

func (f *Fmt) Fix(in []byte, opts *RuntimeOptions) (bool, []byte, error) {
	filename := ""

	if opts != nil && opts.Metadata.Filename != "" {
		filename = opts.Metadata.Filename
	}

	if filename == "" {
		return false, nil, errors.New("filename is required when formatting")
	}

	formatted, err := format.SourceWithOpts(filepath.Base(filename), in, f.OPAFmtOpts)
	if err != nil {
		return false, nil, fmt.Errorf("failed to format: %w", err)
	}

	// we always return true because the fix still completed successfully, and
	// then we can say that the violation with this instance's key was fixed too.
	return true, formatted, nil
}
