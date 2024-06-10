package refs

import (
	"context"
	"fmt"
	"sync"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/bundle"
	"github.com/open-policy-agent/opa/rego"

	rbundle "github.com/styrainc/regal/bundle"
	rio "github.com/styrainc/regal/internal/io"
	"github.com/styrainc/regal/pkg/builtins"
	"github.com/styrainc/regal/pkg/config"

	_ "embed"
)

//go:embed rego/ref_names.rego
var refNamesRego string

// pq is a prepared query that finds ref names used in a module.
// pq is prepared at init time to make this functionality more
// efficient. In local testing, this reduced time by ~95%.
//
//nolint:gochecknoglobals
var pq *rego.PreparedEvalQuery

//nolint:gochecknoglobals
var pqInitOnce sync.Once

// initialize prepares the rego query for finding ref names used in a module.
// This is run and the resulting prepared query stored for performance reasons.
// This function is only used by language server code paths and so init() is not
// used.
func initialize() {
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
		rego.Module("mod.rego", refNamesRego),
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
	if pq == nil {
		pqInitOnce.Do(initialize)
	}

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
