package fixes

import "github.com/open-policy-agent/opa/format"

type Options struct {
	Fmt *FmtOptions
}

type FmtOptions struct {
	Filename   string
	OPAFmtOpts format.Opts
}
