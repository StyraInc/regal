package fixes

import (
	"bytes"
	"errors"
	"fmt"
	"os"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/format"
	"github.com/styrainc/regal/internal/parse"
)

type Fmt struct {
	// OPAFmtOpts are the options to pass to OPA's format.SourceWithOpts
	// function.
	OPAFmtOpts format.Opts
	// NameOverride allows this fix config to also be registered under another name, see note
	// in Name().
	NameOverride string
}

func (f *Fmt) Name() string {
	// this allows this fix config to also be registered under another name so that different
	// configurations can be registered under other linter rule names.
	if f.NameOverride != "" {
		return f.NameOverride
	}

	return "opa-fmt"
}

func (f *Fmt) Fix(fc *FixCandidate, opts *RuntimeOptions) ([]FixResult, error) {
	if opts == nil {
		return nil, errors.New("runtime options are required")
	}

	if fc.Filename == "" {
		return nil, errors.New("filename is required when formatting")
	}

	// Try to parse the contents first as v1, then as v0 ... and use the appropriate version to format

	module, version, err := parse.ModuleWithOpts(fc.Filename, string(fc.Contents), parse.ParserOptions())
	if err != nil {
		return nil, fmt.Errorf("failed to parse module: %w", err)
	}

	f.OPAFmtOpts.RegoVersion = version

	if version == ast.RegoV0 {
		f.OPAFmtOpts.RegoVersion = ast.RegoV0CompatV1
	}

	formatted, err := format.AstWithOpts(module, f.OPAFmtOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to format: %w", err)
	}

	if bytes.Equal(formatted, fc.Contents) {
		fmt.Fprintln(os.Stderr, "No changes needed for", fc.Filename)

		return nil, nil
	}

	fmt.Fprintln(os.Stderr, "Changes made to", fc.Filename)
	fmt.Fprintln(os.Stderr, string(formatted))

	return []FixResult{{
		Title:    f.Name(),
		Root:     opts.BaseDir,
		Contents: formatted,
	}}, nil
}
