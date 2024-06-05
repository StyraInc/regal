package refs

import (
	"context"
	"fmt"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/bundle"
	"github.com/open-policy-agent/opa/rego"

	rbundle "github.com/styrainc/regal/bundle"
	rio "github.com/styrainc/regal/internal/io"
	"github.com/styrainc/regal/pkg/builtins"
	"github.com/styrainc/regal/pkg/config"
)

// pq is a prepared query that finds ref names used in a module.
// pq is prepared at init time to make this functionality more
// efficient. In local testing, this reduced time by ~95%.
//
//nolint:gochecknoglobals
var pq *rego.PreparedEvalQuery

func init() {
	regalRules := rio.MustLoadRegalBundleFS(rbundle.Bundle)

	dataBundle := bundle.Bundle{
		Manifest: bundle.Manifest{
			Roots:    &[]string{"internal"},
			Metadata: map[string]any{"name": "internal"},
		},
		Data: map[string]any{
			"internal": map[string]any{
				"combined_config": map[string]any{
					"capabilities": rio.ToMap(config.CapabilitiesForThisVersion()),
				},
			},
		},
	}

	regoArgs := []func(*rego.Rego){
		rego.ParsedBundle("regal", &regalRules),
		rego.Module("module.rego", `
package lsp.completions

import rego.v1

import data.regal.ast

# ref_names returns a list of ref names that are used in the module.
# built-in functions are not included as they are provided by another completions provider.
# imports are not included as we need to use the imported_identifier instead
# (i.e. maybe an alias).
ref_names contains name if {
	some ref in ast.all_refs

	name := ast.ref_to_string(ref.value)

	not name in ast.builtin_functions_called
	not name in imports
}

# if a user has imported data.foo, then foo should be suggested.
# if they have imported data.foo as bar, then bar should be suggested.
# this also has the benefit of skipping future.* and rego.v1 as
# imported_identifiers will only match data.* and input.*
ref_names contains name if {
	some name in ast.imported_identifiers
}

# imports are not shown as we need to show the imported alias instead
imports contains ref if {
	some imp in ast.imports

	ref := ast.ref_to_string(imp.path.value)
}
`),
		rego.ParsedBundle("internal", &dataBundle),
		rego.Query(`data.lsp.completions.ref_names`),
		rego.Function2(builtins.RegalParseModuleMeta, builtins.RegalParseModule),
		rego.Function1(builtins.RegalLastMeta, builtins.RegalLast),
	}

	preparedQuery, err := rego.New(regoArgs...).PrepareForEval(context.Background())
	if err != nil {
		panic(err)
	}

	pq = &preparedQuery
}

// UsedInModule returns a list of ref names suitable for completion that are
// used in the module's code.
// See the rego above for more details on what's included and excluded.
// This function is run when the parse completes for a module.
func UsedInModule(ctx context.Context, module *ast.Module) ([]string, error) {
	rs, err := pq.Eval(ctx, rego.EvalInput(module))
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate rego query: %w", err)
	}

	if len(rs) == 0 || len(rs[0].Expressions) == 0 {
		// no refs found
		return []string{}, nil
	}

	foundRefs, ok := rs[0].Expressions[0].Value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected type %T", rs[0].Expressions[0].Value)
	}

	refNames := make([]string, len(foundRefs))
	for i, ref := range foundRefs {
		refNames[i], ok = ref.(string)
		if !ok {
			return nil, fmt.Errorf("unexpected type %T", ref)
		}
	}

	return refNames, nil
}
