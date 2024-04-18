package fixes

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/open-policy-agent/opa/format"
)

type Fmt struct {
	// NameOverride allows this fix config to also be registered under another name, see note
	// in Name().
	NameOverride string
	// OPAFmtOpts are the options to pass to OPA's format.SourceWithOpts
	// function.
	OPAFmtOpts format.Opts
}

func (f *Fmt) Name() string {
	// this allows this fix config to also be registered under another name so that different
	// configurations can be registered under other linter rule names.
	if f.NameOverride != "" {
		return f.NameOverride
	}

	return "opa-fmt"
}

func (f *Fmt) Fix(fc *FixCandidate, _ *RuntimeOptions) ([]FixResult, error) {
	if fc.Filename == "" {
		return nil, errors.New("filename is required when formatting")
	}

	formatted, err := format.SourceWithOpts(filepath.Base(fc.Filename), fc.Contents, f.OPAFmtOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to format: %w", err)
	}

	if string(formatted) == string(fc.Contents) {
		return nil, nil
	}

	return []FixResult{
		{
			Contents: formatted,
		},
	}, nil
}
